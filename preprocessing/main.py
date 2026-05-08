"""
preprocessing/main.py
─────────────────────
Full 5-stage word preprocessing pipeline.
 
Stage 1  SPARQL Seeding     → fetch seed entities from Wikidata
Stage 2  Vector Lookup      → avg-pool multi-word names → top-N fastText neighbours
Stage 3  Validity Filter    → spaCy POS, stopwords, Korp frequency gate
Stage 4  Dimension Reduction→ PCA 300 → 100 dims (fit per output file)
Stage 5  CSV Export         → server/wordfiles/<name>_vectors.csv
"""

from __future__ import annotations
 
import csv
import logging
import re
import sys
from collections import defaultdict
from pathlib import Path
from typing import Dict, List, Optional, Tuple
from sklearn.decomposition import PCA 
 
import numpy as np
from dotenv import load_dotenv

# ── Paths ─────────────────────────────────────────────────────────────────────
BASE_DIR = Path(__file__).resolve().parent
load_dotenv(dotenv_path=BASE_DIR / ".env.local")
 
from seeding import query_runner

# ── Configuration ─────────────────────────────────────────────────────────────
FASTTEXT_MODEL_PATH = BASE_DIR / "cc.sv.300.bin"
KORP_DIR            = BASE_DIR / "korp"
OUTPUT_DIR          = BASE_DIR.parent / "server" / "wordfiles"
SPACY_MODEL         = "sv_core_news_sm"
 
PCA_DIMS            = 100    # target dimensionality
NEIGHBOURS_PER_SEED = 70     # fastText neighbours to fetch per seed token
TOP_N_PER_SEED      = 50     # max valid neighbours to keep per seed
MIN_WORD_LEN        = 3      # drop very short tokens
MIN_KORP_FREQ       = 5      # min Korp occurrences (if list is available)
 
# spaCy POS tags we allow as game words
ALLOWED_POS = {"NOUN", "PROPN"}

# ── Query → output file mapping ───────────────────────────────────────────────
# Each SPARQL query is assigned a category label and an output CSV name.
# Multiple queries can share the same output CSV (they'll be merged).
CATEGORY_MAPPING: Dict[str, Tuple[str, str]] = {
    "swedish_celebrities": ("celebrity",  "celebrities_vectors.csv"),
    "swedish_music":       ("celebrity",  "celebrities_vectors.csv"),
    "swedish_companies":   ("company",    "companies_vectors.csv"),
    "global_brands":       ("company",    "companies_vectors.csv"),
    "swedish_characters":  ("character",  "characters_vectors.csv"),
    "swedish_food":        ("food",       "food_vectors.csv"),
    "swedish_geography":   ("geography",  "geography_vectors.csv"),
    "swedish_tv_and_film": ("media",      "media_vectors.csv"),
    "video_games":         ("game",       "games_vectors.csv"),
}
 

# ── Logging ───────────────────────────────────────────────────────────────────
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[
        logging.FileHandler(BASE_DIR / "pipeline.log"),
        logging.StreamHandler(sys.stdout),
    ],
)
log = logging.getLogger(__name__)

# ══════════════════════════════════════════════════════════════════════════════
# Helpers
# ══════════════════════════════════════════════════════════════════════════════
 
def _cosine_sim(a: np.ndarray, b: np.ndarray) -> float:
    na, nb = np.linalg.norm(a), np.linalg.norm(b)
    return float(np.dot(a, b) / (na * nb)) if na and nb else 0.0
 
 
def _extract_label(row: Dict[str, str]) -> Optional[str]:
    """Pull the Swedish human-readable label from any SPARQL result row."""
    # Try well-known label variable names first
    for key in (
        "personLabel", "artistLabel", "companyLabel", "brandLabel",
        "characterLabel", "foodLabel", "placeLabel", "workLabel", "gameLabel",
    ):
        val = row.get(key, "").strip()
        # Skip raw QIDs (e.g. "Q12345") and empty strings
        if val and not re.match(r"^Q\d+$", val):
            return val
    # Fallback: first value that looks like a word (not a URI or QID)
    for val in row.values():
        val = val.strip()
        if val and not val.startswith("http") and not re.match(r"^Q\d+$", val):
            return val
    return None

# ══════════════════════════════════════════════════════════════════════════════
# Model / resource loaders
# ══════════════════════════════════════════════════════════════════════════════
 
def load_fasttext(path: Path):
    """Load the binary fastText model (cc.sv.300.bin)."""
    try:
        import fasttext
    except ImportError:
        log.error("fasttext not installed. Run: pip install fasttext")
        sys.exit(1)
 
    if not path.exists():
        log.error(f"fastText model not found: {path}")
        sys.exit(1)
 
    log.info(f"Loading fastText model from {path} …")
    model = fasttext.load_model(str(path))
    log.info("fastText model loaded.")
    return model

def load_spacy(model_name: str):
    """Load a spaCy model with only the tagger component (fast)."""
    try:
        import spacy
    except ImportError:
        log.error("spaCy not installed. Run: pip install spacy")
        sys.exit(1)
 
    log.info(f"Loading spaCy model '{model_name}' …")
    try:
        nlp = spacy.load(model_name, disable=["parser", "ner", "senter"])
    except OSError:
        log.error(
            f"spaCy model '{model_name}' not found.\n"
            f"Run: python -m spacy download {model_name}"
        )
        sys.exit(1)
 
    log.info("spaCy loaded.")
    return nlp

def load_korp_freq(korp_dir: Path) -> Optional[Dict[str, int]]:
    """
    Load and merge all korp-*.csv files from `korp_dir` into a single
    word → total_frequency dict.
 
    All files share the columns `word` and `Totalt` (the cross-corpus
    total). Other columns differ per file (lemma, pos, ne_type, …) and
    are ignored. Frequencies are *summed* across files so a word that
    appears in several dumps gets a combined count.
 
    Returns None if the directory is absent or contains no CSV files,
    which gracefully disables the frequency filter downstream.
    """
    if not korp_dir.exists():
        log.warning(
            f"Korp directory not found at {korp_dir} — frequency filter disabled."
        )
        return None
 
    csv_files = sorted(korp_dir.glob("*.csv"))
    if not csv_files:
        log.warning(f"No CSV files found in {korp_dir} — frequency filter disabled.")
        return None
 
    freq: Dict[str, int] = defaultdict(int)
    total_rows = 0
 
    for csv_path in csv_files:
        file_rows = 0
        try:
            with csv_path.open(encoding="utf-8", newline="") as f:
                reader = csv.DictReader(f)
 
                # Normalise header names: strip whitespace and BOM
                if reader.fieldnames is None:
                    log.warning(f"  {csv_path.name}: could not read headers, skipping.")
                    continue
 
                headers = [h.lstrip("\ufeff").strip() for h in reader.fieldnames]
 
                # Validate that the two columns we need are present
                if "word" not in headers or "Totalt" not in headers:
                    log.warning(
                        f"  {csv_path.name}: missing 'word' or 'Totalt' column "
                        f"(found: {headers[:6]}…), skipping."
                    )
                    continue
 
                # Re-open with normalised fieldnames so DictReader uses them
                f.seek(0)
                reader = csv.DictReader(f, fieldnames=headers)
                next(reader)  # skip original header row
 
                for row in reader:
                    word = row.get("word", "").strip().lower()
                    if not word:
                        continue
                    try:
                        count = int(float(row.get("Totalt", 0) or 0))
                    except (ValueError, TypeError):
                        count = 0
                    freq[word] += count
                    file_rows += 1
 
        except Exception as exc:
            log.warning(f"  {csv_path.name}: error reading file — {exc}")
            continue
 
        log.info(f"  {csv_path.name}: {file_rows:,} rows merged.")
        total_rows += file_rows
 
    if not freq:
        log.warning("Korp: no frequency data loaded — filter disabled.")
        return None
 
    log.info(
        f"Korp complete — {len(freq):,} unique words, "
        f"{total_rows:,} total rows from {len(csv_files)} files."
    )
    return dict(freq)

# ══════════════════════════════════════════════════════════════════════════════
# Stage 2 — Vector Lookup
# ══════════════════════════════════════════════════════════════════════════════
 
def _neighbours_for_seed(model,seed_label: str,k: int = NEIGHBOURS_PER_SEED) -> List[Tuple[float, str]]:
    """
    Return (cosine_similarity, word) pairs for the k nearest Swedish
    fastText neighbours of `seed_label`.
 
    Strategy:
    • Single-word seed  → model.get_nearest_neighbors() directly (fast).
    • Multi-word seed   → avg-pool token vectors, then re-rank the union
                          of per-token neighbours by cosine sim to avg vector.
                          This avoids a full-vocabulary scan.
    """
    tokens = [t for t in seed_label.lower().split() if t.isalpha()]
    if not tokens:
        return []
 
    if len(tokens) == 1:
        try:
            raw = model.get_nearest_neighbors(tokens[0], k=k)
            return [(sim, w) for sim, w in raw]
        except Exception as exc:
            log.debug(f"  fastText error for '{tokens[0]}': {exc}")
            return []
 
    # Multi-word: avg-pool then re-rank
    avg_vec = np.mean([model.get_word_vector(t) for t in tokens], axis=0)
    candidates: set[str] = set()
    for token in tokens:
        try:
            for _, w in model.get_nearest_neighbors(token, k=k):
                candidates.add(w)
        except Exception:
            pass
 
    scored = [
        (_cosine_sim(avg_vec, model.get_word_vector(w)), w)
        for w in candidates
    ]
    scored.sort(reverse=True)
    return scored[:k]
 
 
def stage2_vector_lookup(seed_data: Dict[str, List[Dict[str, str]]],model) -> Dict[str, List[Tuple[str, str, float, np.ndarray]]]:
    """
    Stage 2: for every seed entity, fetch its nearest fastText neighbours.
 
    Returns:
        { output_csv: [(word, category, similarity, raw_300d_vector), …] }
    """
    log.info("=" * 60)
    log.info("Stage 2: Vector Lookup")
    log.info("=" * 60)
 
    # Group seed labels by output CSV
    per_output: Dict[str, List[Tuple[str, str]]] = defaultdict(list)
    for query_name, rows in seed_data.items():
        mapping = CATEGORY_MAPPING.get(query_name)
        if not mapping:
            log.warning(f"  No CATEGORY_MAPPING entry for query '{query_name}', skipping.")
            continue
        category, output_csv = mapping
        for row in rows:
            label = _extract_label(row)
            if label:
                per_output[output_csv].append((label, category))
 
    results: Dict[str, List[Tuple[str, str, float, np.ndarray]]] = defaultdict(list)
 
    for output_csv, seeds in per_output.items():
        log.info(f"  {output_csv}: processing {len(seeds)} seeds …")
        # Track seen words per file to deduplicate across seeds
        seen: Dict[str, float] = {}  # word → best similarity
 
        for label, category in seeds:
            neighbours = _neighbours_for_seed(model, label, k=NEIGHBOURS_PER_SEED)
            kept = 0
            for sim, word in neighbours:
                if kept >= TOP_N_PER_SEED:
                    break
                wl = word.lower()
                # Keep only if this is the first time we see it, or better sim
                if wl not in seen or sim > seen[wl]:
                    seen[wl] = sim
                    kept += 1
 
        # Now build the actual result list with vectors
        for word, sim in seen.items():
            vec = model.get_word_vector(word)
            # Use category of the first seed that produced this word
            # (we just need a representative category)
            cat = seeds[0][1] if seeds else "unknown"
            results[output_csv].append((word, cat, sim, vec))
 
        log.info(f"    → {len(results[output_csv]):,} unique candidate words")
 
    total = sum(len(v) for v in results.values())
    log.info(f"Stage 2 complete — {total:,} total candidate words across all files.")
    return dict(results)

# ══════════════════════════════════════════════════════════════════════════════
# Stage 3 — Validity Filter
# ══════════════════════════════════════════════════════════════════════════════
 
def _is_valid(word: str, nlp, korp_freq: Optional[Dict[str, int]]) -> bool:
    """Return True if the word passes all validity filters."""
    # Length & character sanity
    if len(word) < MIN_WORD_LEN:
        return False
    if not word.replace("-", "").isalpha():   # allow hyphenated words
        return False
 
    # Korp frequency gate (skip if list unavailable)
    if korp_freq is not None and korp_freq.get(word, 0) < MIN_KORP_FREQ:
        return False
 
    # spaCy POS + stopword check (process single token, fast)
    doc = nlp(word)
    if not doc:
        return False
    token = doc[0]
    if token.is_stop:
        return False
    if token.pos_ not in ALLOWED_POS:
        return False
 
    return True
 
 
def stage3_validity_filter(
    candidates: Dict[str, List[Tuple[str, str, float, np.ndarray]]],
    nlp,
    korp_freq: Optional[Dict[str, int]],
) -> Dict[str, List[Tuple[str, str, float, np.ndarray]]]:
    """Stage 3: drop words that fail POS, stopword, or Korp-frequency filters."""
    log.info("=" * 60)
    log.info("Stage 3: Validity Filter")
    log.info("=" * 60)
 
    filtered: Dict[str, List[Tuple[str, str, float, np.ndarray]]] = {}
 
    for output_csv, entries in candidates.items():
        valid = [
            (word, cat, sim, vec)
            for word, cat, sim, vec in entries
            if _is_valid(word, nlp, korp_freq)
        ]
        filtered[output_csv] = valid
        log.info(
            f"  {output_csv}: {len(entries):,} in → {len(valid):,} valid "
            f"({len(entries) - len(valid):,} dropped)"
        )
 
    log.info("Stage 3 complete.")
    return filtered

# ══════════════════════════════════════════════════════════════════════════════
# Stage 4 — Dimension Reduction
# ══════════════════════════════════════════════════════════════════════════════
 
def stage4_reduce_dimensions(validated: Dict[str, List[Tuple[str, str, float, np.ndarray]]],) -> Dict[str, List[Tuple[str, str, np.ndarray]]]:
    """
    Stage 4: fit PCA per output file and transform 300-d → PCA_DIMS-d.
 
    PCA is fit independently per file so that the reduced space best captures
    the semantic variance of that specific vocabulary slice (celebrities,
    geography, etc.) rather than all words combined.
    """
 
    log.info("=" * 60)
    log.info(f"Stage 4: Dimension Reduction (300 → {PCA_DIMS})")
    log.info("=" * 60)
 
    reduced: Dict[str, List[Tuple[str, str, np.ndarray]]] = {}
 
    for output_csv, entries in validated.items():
        if not entries:
            log.warning(f"  {output_csv}: 0 entries — skipping PCA.")
            reduced[output_csv] = []
            continue
 
        n_components = min(PCA_DIMS, len(entries), 300)
        if n_components < PCA_DIMS:
            log.warning(
                f"  {output_csv}: only {len(entries)} words — "
                f"reducing to {n_components} dims instead of {PCA_DIMS}."
            )
 
        matrix = np.array([vec for _, _, _, vec in entries], dtype=np.float32)
        pca = PCA(n_components=n_components, random_state=42)
        reduced_matrix = pca.fit_transform(matrix)
 
        variance_retained = pca.explained_variance_ratio_.sum()
        log.info(
            f"  {output_csv}: {len(entries):,} words, "
            f"variance retained: {variance_retained:.1%}"
        )
 
        # Pad to PCA_DIMS with zeros if we had fewer components
        if n_components < PCA_DIMS:
            pad = np.zeros((len(entries), PCA_DIMS - n_components), dtype=np.float32)
            reduced_matrix = np.concatenate([reduced_matrix, pad], axis=1)
 
        reduced[output_csv] = [
            (word, cat, reduced_matrix[i])
            for i, (word, cat, _sim, _vec) in enumerate(entries)
        ]
 
    log.info("Stage 4 complete.")
    return reduced

# ══════════════════════════════════════════════════════════════════════════════
# Stage 5 — CSV Export
# ══════════════════════════════════════════════════════════════════════════════
 
def stage5_export_csv(reduced: Dict[str, List[Tuple[str, str, np.ndarray]]], output_dir: Path) -> None:
    """Stage 5: write [word, category, v0 … v{PCA_DIMS-1}] CSVs."""
    log.info("=" * 60)
    log.info(f"Stage 5: CSV Export → {output_dir}")
    log.info("=" * 60)
 
    output_dir.mkdir(parents=True, exist_ok=True)
    dim_headers = [f"v{i}" for i in range(PCA_DIMS)]
    fieldnames = ["word", "category"] + dim_headers
 
    for output_csv, entries in reduced.items():
        if not entries:
            log.warning(f"  Skipping {output_csv} — no entries.")
            continue
 
        path = output_dir / output_csv
        with path.open("w", encoding="utf-8", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=fieldnames)
            writer.writeheader()
            for word, category, vec in entries:
                row: Dict = {"word": word, "category": category}
                # Pad / truncate vec to PCA_DIMS just in case
                for i in range(PCA_DIMS):
                    row[f"v{i}"] = f"{float(vec[i]):.6f}" if i < len(vec) else "0.000000"
                writer.writerow(row)
 
        log.info(f"  Wrote {len(entries):,} rows → {path}")
 
    log.info("Stage 5 complete.")

# ══════════════════════════════════════════════════════════════════════════════
# Pipeline entry point
# ══════════════════════════════════════════════════════════════════════════════
 
def pipeline() -> None:
    seeding_output_dir = BASE_DIR / "seeding" / "output"
    seeding_output_dir.mkdir(parents=True, exist_ok=True)
 
    # ── Stage 1: SPARQL Seeding ──────────────────────────────────────────────
    log.info("=" * 60)
    log.info("Stage 1: SPARQL Seeding")
    log.info("=" * 60)
    seed_data = query_runner.run_all_and_save(
        queries=query_runner.QUERIES,
        output_dir=seeding_output_dir,
    )
    log.info("Stage 1 complete.\n")
 
    # ── Load shared resources ────────────────────────────────────────────────
    model    = load_fasttext(FASTTEXT_MODEL_PATH)
    nlp      = load_spacy(SPACY_MODEL)
    korp     = load_korp_freq(KORP_DIR)
 
    # ── Stages 2–5 ───────────────────────────────────────────────────────────
    candidates  = stage2_vector_lookup(seed_data, model)
    validated   = stage3_validity_filter(candidates, nlp, korp)
    reduced     = stage4_reduce_dimensions(validated)
    stage5_export_csv(reduced, OUTPUT_DIR)
 
    log.info("")
    log.info("Pipeline complete! ✓")
    log.info(f"Output files in: {OUTPUT_DIR}")
 
 
if __name__ == "__main__":
    pipeline()
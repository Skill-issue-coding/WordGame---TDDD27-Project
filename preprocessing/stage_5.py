"""
Stage 5: Encode all words and entities using intfloat/multilingual-e5-large.

Inputs:
  - intermediate/stage4_general/general_words.csv   (general vocabulary from Korp)
  - intermediate/stage3_attrs/*.csv                 (entities enriched with wiki summary + attributes)

Output:
  - intermediate/stage5_encoded/embeddings.npy      (float32, shape N*1024, L2-normalised)
  - intermediate/stage5_encoded/vocab.json          (list of N canonical word/entity strings)

Entity passage format: "passage: <Name>. <wiki_summary>. <wiki_attributes>"
General word format:   "passage: <word>"

Entities take priority — if a general word overlaps with a known entity, the richer
entity passage is kept and the naked word form is dropped.
"""

import json
import sys
import logging
from pathlib import Path

import numpy as np
import pandas as pd
import torch

try:
    from shared import CATEGORY_MAPPING, _is_valid_label
except ImportError:
    CATEGORY_MAPPING: dict = {}
    def _is_valid_label(value: str) -> bool:  # type: ignore[misc]
        import re
        value = (value or "").strip()
        return bool(value) and not value.startswith("http") and not re.match(r"^Q\d+$", value) and any(c.isalpha() for c in value)

BASE_DIR       = Path(__file__).resolve().parent
STAGE3_DIR     = BASE_DIR / "intermediate" / "stage3_attrs"
STAGE4_CSV     = BASE_DIR / "intermediate" / "stage4_general" / "general_words.csv"
OUTPUT_DIR     = BASE_DIR / "intermediate" / "stage5_encoded"
SERVER_DIR     = BASE_DIR.parent / "server" / "wordfiles"

MODEL_NAME     = "intfloat/multilingual-e5-large"
BATCH_SIZE     = 128
SUMMARY_MAX_CHARS = 1500

def _setup_logger() -> logging.Logger:
    log_path = Path(__file__).resolve().parent / "pipeline.log"
    root = logging.getLogger()
    if not any(
        isinstance(h, logging.FileHandler) and h.baseFilename == str(log_path)
        for h in root.handlers
    ):
        handler = logging.FileHandler(log_path, encoding="utf-8")
        handler.setLevel(logging.INFO)
        handler.setFormatter(logging.Formatter("%(asctime)s [%(levelname)s] %(message)s"))
        root.addHandler(handler)
        root.setLevel(logging.INFO)
    return logging.getLogger(__name__)

log = _setup_logger()


# ── Passage builders ──────────────────────────────────────────────────────────

def entity_passage(name: str, summary: str, attrs: str) -> str:
    name    = (name    or "").strip()
    summary = (summary or "").strip()[:SUMMARY_MAX_CHARS]
    attrs   = (attrs   or "").strip()
    parts   = [p for p in (name, summary, attrs) if p]
    return "passage: " + ". ".join(parts)


def word_passage(word: str) -> str:
    return f"passage: {word.strip()}"


# ── Loaders ───────────────────────────────────────────────────────────────────

def load_entities() -> tuple[dict[str, tuple[str, str]], dict[str, str]]:
    """Returns ({key: (name, passage)}, {key: category})."""
    records:    dict[str, tuple[str, str]] = {}
    categories: dict[str, str]             = {}

    if not STAGE3_DIR.exists():
        print(f"Varning: {STAGE3_DIR} saknas — inga entiteter laddas.")
        return records, categories

    for csv_path in sorted(STAGE3_DIR.glob("*.csv")):
        cat = CATEGORY_MAPPING.get(csv_path.stem, ("general", ""))[0]
        df  = pd.read_csv(csv_path)

        label_col = next((c for c in df.columns if c.endswith("Label")), None)
        if label_col is None:
            label_col = next((c for c in ("name", "word") if c in df.columns), None)
        if label_col is None:
            print(f"Varning: ingen namn-kolumn i {csv_path.name}, hoppar över.")
            continue

        for _, row in df.iterrows():
            name = str(row.get(label_col, "")).strip()
            if not name or not _is_valid_label(name):
                continue
            key = name.lower()
            if key in records:
                continue  # first CSV wins (sorted → deterministic)
            summary = str(row.get("wiki_summary", ""))
            attrs   = str(row.get("wiki_attributes", ""))
            records[key]    = (name, entity_passage(name, summary, attrs))
            categories[key] = cat

    return records, categories


def load_general_words(entity_keys: set[str]) -> dict[str, tuple[str, str]]:
    """Returns {lowercase_key: (word, passage_text)}, skipping entity overlaps."""
    records: dict[str, tuple[str, str]] = {}

    if not STAGE4_CSV.exists():
        print(f"Varning: {STAGE4_CSV} saknas — inga generella ord laddas.")
        return records

    df = pd.read_csv(STAGE4_CSV)
    for word in df["word"].dropna().astype(str):
        word = word.strip()
        if not word:
            continue
        key = word.lower()
        if key in entity_keys or key in records:
            continue
        records[key] = (word, word_passage(word))

    return records


# ── Main ──────────────────────────────────────────────────────────────────────

def main():
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    device = "cuda" if torch.cuda.is_available() else "cpu"
    print(f"Device: {device}")
    log.info("Stage 5: start")
    log.info(f"Device: {device}")

    # 1. Collect items
    entity_records, entity_cats = load_entities()
    print(f"Entiteter (stage 3): {len(entity_records):,}")
    log.info(f"Stage 5: entities {len(entity_records)}")

    word_records = load_general_words(set(entity_records.keys()))
    print(f"Generella ord (stage 4): {len(word_records):,}")
    log.info(f"Stage 5: general words {len(word_records)}")

    # Entities first so their indices stay grouped, then general words
    entity_items = list(entity_records.items())   # [(key, (name, passage)), ...]
    word_items   = list(word_records.items())
    if not entity_items and not word_items:
        print("Inga poster att koda — avbryter.")
        sys.exit(1)

    vocab    = [name    for _, (name, _)       in entity_items] + \
               [name    for _, (name, _)       in word_items]
    passages = [passage for _, (_,    passage) in entity_items] + \
               [passage for _, (_,    passage) in word_items]
    sources  = [entity_cats.get(key, "general") for key, _ in entity_items] + \
               ["general" for _ in word_items]
    print(f"Totalt att koda: {len(passages):,}")
    log.info(f"Stage 5: total passages {len(passages)}")

    # 2. Load E5 model
    try:
        from sentence_transformers import SentenceTransformer
    except ImportError:
        print("sentence-transformers saknas. Kör: pip install sentence-transformers")
        sys.exit(1)

    print(f"Laddar modell '{MODEL_NAME}'...")
    log.info(f"Stage 5: loading model {MODEL_NAME}")
    model = SentenceTransformer(MODEL_NAME, device=device)

    # 3. Encode
    print(f"Kodar i batchar om {BATCH_SIZE}...")
    log.info(f"Stage 5: encoding batch_size={BATCH_SIZE}")
    embeddings = model.encode(
        passages,
        batch_size=BATCH_SIZE,
        show_progress_bar=True,
        normalize_embeddings=True,  # cosine sim == dot product after L2 norm
        convert_to_numpy=True,
    )

    embeddings = embeddings.astype(np.float32)

    # 4. Save
    emb_path     = OUTPUT_DIR / "embeddings.npy"
    vocab_path   = OUTPUT_DIR / "vocab.json"
    sources_path = OUTPUT_DIR / "sources.json"

    np.save(str(emb_path), embeddings)
    with vocab_path.open("w", encoding="utf-8") as f:
        json.dump(vocab, f, ensure_ascii=False)
    with sources_path.open("w", encoding="utf-8") as f:
        json.dump(sources, f, ensure_ascii=False)

    log.info(f"Stage 5: wrote {emb_path}")
    log.info(f"Stage 5: wrote {vocab_path}")
    log.info(f"Stage 5: wrote {sources_path}")

    SERVER_DIR.mkdir(parents=True, exist_ok=True)
    server_sources_path = SERVER_DIR / "sources.json"
    with server_sources_path.open("w", encoding="utf-8") as f:
        json.dump(sources, f, ensure_ascii=False)
    log.info(f"Stage 5: wrote {server_sources_path}")

    print(f"\nKlar! {len(vocab):,} poster kodade.")
    print(f"  embeddings : {emb_path}  (shape {embeddings.shape})")
    print(f"  vocab      : {vocab_path}")
    print(f"  sources    : {sources_path}")
    print(f"  sources    : {server_sources_path}  (server)")


if __name__ == "__main__":
    main()

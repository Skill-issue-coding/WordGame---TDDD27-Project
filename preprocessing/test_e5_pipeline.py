"""
test_e5_pipeline.py — Test intfloat/multilingual-e5-large on your actual game vocabulary.
 
This script:
  1. Loads your existing word sources:
       - preprocessing/seeding/output/*.csv    (entity seeds from Wikidata)
       - preprocessing/korp/*.csv              (general Swedish words + frequency)
       - preprocessing/kelly.xml               (Kelly word list)
  2. Optionally fetches Swedish Wikipedia lead paragraphs for entities (cached to disk).
  3. Encodes everything with mE5-large (cached to disk as .npz).
  4. Interactive REPL: type a word/phrase, see closest neighbors with similarity and source.
 
Usage:
  python test_e5_pipeline.py --build                       # label-only, fast
  python test_e5_pipeline.py --build --with-wiki           # fetch descriptions, slow but RECOMMENDED
  python test_e5_pipeline.py --build --limit-korp 5000     # smaller Korp sample for iteration
  python test_e5_pipeline.py --build --limit-entities 500
  python test_e5_pipeline.py                               # REPL (after --build)
 
Tip: First run `--build --limit-korp 3000 --limit-entities 300` to validate the pipeline,
then run the full build overnight.
 
Requirements:
  pip install sentence-transformers torch requests tqdm numpy lxml
"""

import argparse
import csv
import hashlib
import json
import os
import pathlib
import sys
import time
from typing import Optional

import numpy as np
import requests
from tqdm import tqdm

# ============================================================================
# Paths
# ============================================================================
HERE = pathlib.Path(__file__).resolve().parent
SEEDING_DIR = HERE / "seeding" / "output"
KORP_DIR = HERE / "korp"
KELLY_XML = HERE / "kelly.xml"
 
CACHE_DIR = HERE / ".cache_e5"
CACHE_DIR.mkdir(exist_ok=True)
WIKI_CACHE = CACHE_DIR / "wiki_summaries.json"
VOCAB_JSON = CACHE_DIR / "vocab.json"
VEC_NPZ = CACHE_DIR / "vectors.npz"
 
MODEL_NAME = "intfloat/multilingual-e5-large"
WIKI_REST = "https://sv.wikipedia.org/api/rest_v1/page/summary/{}"
UA = {"User-Agent": "wordgame-tddd27/0.1 (preprocessing-test)"}

# ============================================================================
# Try to use your existing loaders from shared.py; otherwise fall back.
# ============================================================================
try:
    sys.path.insert(0, str(HERE))
    from shared import load_kelly as _shared_load_kelly  # type: ignore
    HAS_SHARED_KELLY = True
except Exception:
    HAS_SHARED_KELLY = False
 
try:
    from shared import load_korp as _shared_load_korp  # type: ignore
    HAS_SHARED_KORP = True
except Exception:
    HAS_SHARED_KORP = False

def load_kelly_fallback(path: pathlib.Path) -> list[str]:
    """Minimal Kelly XML loader: extracts every <gf>…</gf> grundform."""
    if not path.exists():
        print(f"  ⚠ {path} saknas — hoppar över Kelly", file=sys.stderr)
        return []
    try:
        from lxml import etree # pip install lxml
    except ImportError:
        print("  ⚠ lxml saknas (pip install lxml) — hoppar över Kelly", file=sys.stderr)
        return []
    tree = etree.parse(str(path))
    words = []
    for gf in tree.iter("gf"):
        if gf.text and gf.text.strip():
            w = gf.text.strip().lower()
            if w.isalpha() and len(w) >= 2:
                words.append(w)
    return sorted(set(words))

def load_korp_fallback(korp_dir: pathlib.Path, limit: Optional[int] = None) -> list[tuple[str, int]]:
    """
    Read all CSVs in preprocessing/korp/ and aggregate (word, frequency).
    Expected Språkbanken Korp export columns include 'ord' (or 'grundform') and a freq column.
    We try a few common column names and fall back to first/second column.
    """
    if not korp_dir.exists():
        print(f"  ⚠ {korp_dir} saknas — hoppar över Korp", file=sys.stderr)
        return []
    agg: dict[str, int] = {}
    for csv_path in sorted(korp_dir.glob("*.csv")):
        with open(csv_path, "r", encoding="utf-8", newline="") as f:
            sample = f.read(4096)
            f.seek(0)
            try:
                dialect = csv.Sniffer().sniff(sample, delimiters=",;\t|")
            except csv.Error:
                dialect = csv.excel
            reader = csv.DictReader(f, dialect=dialect)
            fields = [c.lower() for c in (reader.fieldnames or [])]
            word_col = next((c for c in ("grundform", "ord", "lemma", "word") if c in fields), None)
            freq_col = next((c for c in ("frekvens", "freq", "count", "antal") if c in fields), None)
            pos_col = next((c for c in ("ordklass", "pos", "tag") if c in fields), None)
            for row in reader:
                row_lc = {k.lower(): v for k, v in row.items() if k}
                w = (row_lc.get(word_col) if word_col else None) or ""
                w = w.strip().lower()
                if not w or not w.isalpha() or len(w) < 2:
                    continue
                # Skip pronouns/determiners/conjunctions if POS column exists
                if pos_col:
                    pos = (row_lc.get(pos_col) or "").lower()
                    if pos in {"pn", "dt", "kn", "in", "pp", "ie", "ab", "ha"}:
                        # Pronoun, determiner, conjunction, interjection, preposition,
                        # infinitive marker, adverb, interrogative adverb — skip
                        continue
                try:
                    f_val = int(row_lc.get(freq_col, "1")) if freq_col else 1
                except ValueError:
                    f_val = 1
                agg[w] = agg.get(w, 0) + f_val
    out = sorted(agg.items(), key=lambda x: -x[1])
    if limit:
        out = out[:limit]
    return out

def load_kelly() -> list[str]:
    if HAS_SHARED_KELLY:
        try:
            return list(_shared_load_kelly())
        except Exception as e:
            print(f"  ⚠ shared.load_kelly() kraschade ({e}) — använder fallback", file=sys.stderr)
    return load_kelly_fallback(KELLY_XML)

def load_korp(limit: Optional[int] = None) -> list[tuple[str, int]]:
    if HAS_SHARED_KORP:
        try:
            data = _shared_load_korp()
            if data and isinstance(data[0], (tuple, list)):
                return list(data)[:limit] if limit else list(data)
            return [(w, 1) for w in data][:limit] if limit else [(w, 1) for w in data]
        except Exception as e:
            print(f"  ⚠ shared.load_korp() kraschade ({e}) — använder fallback", file=sys.stderr)
    return load_korp_fallback(KORP_DIR, limit=limit)

# ============================================================================
# Seeding outputs (entities)
# ============================================================================
def load_seeding_outputs(limit_per_file: Optional[int] = None) -> list[dict]:
    """
    Reads every CSV in preprocessing/seeding/output/.
    Filename (without extension) becomes the category, e.g.
        celebrities.csv → category = "celebrities".
    Tries to find label + QID columns under common names.
    """
    if not SEEDING_DIR.exists():
        print(f"  ⚠ {SEEDING_DIR} saknas — hoppar över entiteter", file=sys.stderr)
        return []
    out = []
    for csv_path in sorted(SEEDING_DIR.glob("*.csv")):
        category = csv_path.stem  # e.g. "celebrities"
        with open(csv_path, "r", encoding="utf-8", newline="") as f:
            sample = f.read(4096); f.seek(0)
            try:
                dialect = csv.Sniffer().sniff(sample, delimiters=",;\t|")
            except csv.Error:
                dialect = csv.excel
            reader = csv.DictReader(f, dialect=dialect)
            fields = [c.lower() for c in (reader.fieldnames or [])]
            label_col = next((c for c in ("label", "name", "title", "namn", "labelsv") if c in fields), None)
            qid_col = next((c for c in ("qid", "wikidata", "wikidataid", "item") if c in fields), None)
            sitelinks_col = next((c for c in ("sitelinks", "popularity", "score") if c in fields), None)
            n = 0
            for row in reader:
                row_lc = {k.lower(): v for k, v in row.items() if k}
                label = (row_lc.get(label_col) if label_col else None) or ""
                label = label.strip()
                if not label:
                    continue
                qid = (row_lc.get(qid_col) or "").strip() if qid_col else ""
                # Strip URL prefix if SPARQL returned http://www.wikidata.org/entity/Q123
                if "/" in qid:
                    qid = qid.rsplit("/", 1)[-1]
                try:
                    sitelinks = int(row_lc.get(sitelinks_col, "0")) if sitelinks_col else 0
                except (ValueError, TypeError):
                    sitelinks = 0
                out.append({
                    "label": label,
                    "category": category,
                    "qid": qid or None,
                    "sitelinks": sitelinks,
                })
                n += 1
                if limit_per_file and n >= limit_per_file:
                    break
    return out

# ============================================================================
# Swedish Wikipedia summary fetcher (with disk cache)
# ============================================================================
def load_wiki_cache() -> dict[str, Optional[str]]:
    if WIKI_CACHE.exists():
        return json.loads(WIKI_CACHE.read_text("utf-8"))
    return {}
 
 
def save_wiki_cache(cache: dict[str, Optional[str]]) -> None:
    WIKI_CACHE.write_text(json.dumps(cache, ensure_ascii=False), "utf-8")
 
 
def fetch_wiki_summaries(entities: list[dict], throttle_s: float = 0.05) -> dict[str, Optional[str]]:
    """For each entity label, fetch sv.wikipedia.org page summary. Cached to disk."""
    cache = load_wiki_cache()
    to_fetch = [e for e in entities if e["label"] not in cache]
    if not to_fetch:
        return cache
    print(f"  → Hämtar Wikipedia-sammanfattningar för {len(to_fetch)} entiteter (cachas)…")
    for i, e in enumerate(tqdm(to_fetch)):
        label = e["label"]
        try:
            url = WIKI_REST.format(requests.utils.quote(label.replace(" ", "_")))
            r = requests.get(url, headers=UA, timeout=15)
            if r.status_code == 200:
                cache[label] = r.json().get("extract")
            else:
                cache[label] = None
        except Exception:
            cache[label] = None
        time.sleep(throttle_s)
        if (i + 1) % 200 == 0:
            save_wiki_cache(cache)  # checkpoint
    save_wiki_cache(cache)
    return cache

# ============================================================================
# Build vocabulary: unified list of {label, kind, category, text, freq}
# ============================================================================
def build_vocabulary(
    with_wiki: bool,
    limit_entities: Optional[int] = None,
    limit_korp: Optional[int] = None,
) -> list[dict]:
    vocab: list[dict] = []
    seen: set[str] = set()
 
    print("Läser entiteter från seeding/output/…")
    entities = load_seeding_outputs()
    if limit_entities:
        entities = entities[:limit_entities]
    print(f"  {len(entities)} entiteter")
 
    wiki_summaries: dict[str, Optional[str]] = {}
    if with_wiki and entities:
        wiki_summaries = fetch_wiki_summaries(entities)
 
    for e in entities:
        key = e["label"].lower()
        if key in seen:
            continue
        seen.add(key)
        summary = wiki_summaries.get(e["label"]) if with_wiki else None
        # If we have a Wikipedia summary, embed THAT (the magic).
        # Otherwise fall back to the label itself.
        text = summary if summary else e["label"]
        vocab.append({
            "label": e["label"],
            "kind": "entity",
            "category": e["category"],
            "qid": e["qid"],
            "sitelinks": e["sitelinks"],
            "text": text,
            "has_wiki": bool(summary),
        })
 
    print("Läser Kelly…")
    kelly_words = load_kelly()
    print(f"  {len(kelly_words)} Kelly-ord")
    for w in kelly_words:
        if w.lower() in seen:
            continue
        seen.add(w.lower())
        vocab.append({
            "label": w,
            "kind": "word",
            "category": "kelly",
            "qid": None,
            "sitelinks": 0,
            "text": w,
            "has_wiki": False,
        })
 
    print("Läser Korp…")
    korp_words = load_korp(limit=limit_korp)
    print(f"  {len(korp_words)} Korp-ord")
    for w, freq in korp_words:
        if w.lower() in seen:
            continue
        seen.add(w.lower())
        vocab.append({
            "label": w,
            "kind": "word",
            "category": "korp",
            "qid": None,
            "sitelinks": int(freq),
            "text": w,
            "has_wiki": False,
        })
 
    print(f"\nTotal vokabulär: {len(vocab)} poster "
          f"({sum(1 for v in vocab if v['kind'] == 'entity')} entiteter, "
          f"{sum(1 for v in vocab if v['kind'] == 'word')} ord)")
    return vocab
 
 
# ============================================================================
# Encode with mE5-large (with disk cache)
# ============================================================================
def encode_vocabulary(vocab: list[dict], batch_size: int = 16) -> np.ndarray:
    from sentence_transformers import SentenceTransformer
    import torch
 
    print(f"\nLaddar {MODEL_NAME}…")
    t0 = time.time()
    device = "cuda" if torch.cuda.is_available() else "cpu"
    if device == "cpu":
        # On Apple Silicon, mps works too, but sentence-transformers handles it
        try:
            if torch.backends.mps.is_available():
                device = "mps"
        except AttributeError:
            pass
    print(f"  enhet: {device}")
    model = SentenceTransformer(MODEL_NAME, device=device)
    print(f"  modell laddad på {time.time() - t0:.1f}s")
 
    # mE5 expects "passage: " for documents and "query: " for queries
    passages = [f"passage: {v['text']}" for v in vocab]
 
    print(f"Kodar {len(passages)} poster med batch_size={batch_size}…")
    t0 = time.time()
    vectors = model.encode(
        passages,
        batch_size=batch_size,
        show_progress_bar=True,
        normalize_embeddings=True,  # L2-normalize → cosine = dot product
        convert_to_numpy=True,
    ).astype(np.float32)
    print(f"  kodning klar på {time.time() - t0:.1f}s "
          f"({len(passages) / max(time.time() - t0, 0.001):.1f} poster/s)")
    return vectors
 
 
def save_artifacts(vocab: list[dict], vectors: np.ndarray) -> None:
    VOCAB_JSON.write_text(json.dumps(vocab, ensure_ascii=False), "utf-8")
    np.savez_compressed(VEC_NPZ, vectors=vectors)
    print(f"\nSparade {VOCAB_JSON} ({VOCAB_JSON.stat().st_size / 1e6:.1f} MB)")
    print(f"Sparade {VEC_NPZ} ({VEC_NPZ.stat().st_size / 1e6:.1f} MB)")
 
 
def load_artifacts() -> tuple[list[dict], np.ndarray]:
    if not VOCAB_JSON.exists() or not VEC_NPZ.exists():
        print("Inga sparade artefakter. Kör först:  python test_e5_pipeline.py --build",
              file=sys.stderr)
        sys.exit(1)
    vocab = json.loads(VOCAB_JSON.read_text("utf-8"))
    vectors = np.load(VEC_NPZ)["vectors"]
    return vocab, vectors
 
 
# ============================================================================
# Interactive REPL
# ============================================================================
def repl() -> None:
    vocab, vectors = load_artifacts()
    n_entities = sum(1 for v in vocab if v["kind"] == "entity")
    n_with_wiki = sum(1 for v in vocab if v.get("has_wiki"))
    print(f"Laddat {len(vocab)} poster ({n_entities} entiteter, "
          f"{n_with_wiki} med Wikipedia-text)")
 
    from sentence_transformers import SentenceTransformer
    import torch
    device = "cuda" if torch.cuda.is_available() else "cpu"
    try:
        if device == "cpu" and torch.backends.mps.is_available():
            device = "mps"
    except AttributeError:
        pass
    print(f"Laddar {MODEL_NAME} på {device}…")
    model = SentenceTransformer(MODEL_NAME, device=device)
 
    TOP_N = 15
    print("\nSkriv ett ord, namn eller fras. 'q' för avsluta.\n")
 
    while True:
        try:
            q = input("> ").strip()
        except (EOFError, KeyboardInterrupt):
            print()
            break
        if not q or q.lower() in ("q", "quit", "exit"):
            break
 
        t0 = time.time()
        # mE5 query prefix is mandatory
        qv = model.encode(
            [f"query: {q}"],
            normalize_embeddings=True,
            convert_to_numpy=True,
        ).astype(np.float32)[0]
        sims = vectors @ qv  # cosine since both are L2-normalized
        top_idx = np.argsort(-sims)[: TOP_N + 5]  # extra for filtering self
 
        dt = (time.time() - t0) * 1000
        print(f"\n  kodade fråga + sökte {len(vocab)} poster på {dt:.0f} ms\n")
        print(f"  {'sim':>6}  {'kategori':<14}  {'kind':<7}  ord")
        print(f"  {'-' * 6}  {'-' * 14}  {'-' * 7}  {'-' * 30}")
 
        shown = 0
        for i in top_idx:
            v = vocab[i]
            # Hide exact match to the input
            if v["label"].lower() == q.lower():
                continue
            tag = "📖" if v.get("has_wiki") else "  "
            print(f"  {sims[i]:>6.3f}  {v['category']:<14}  {v['kind']:<7}  {tag} {v['label']}")
            shown += 1
            if shown >= TOP_N:
                break
        print()
 
 
# ============================================================================
# Main
# ============================================================================
def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--build", action="store_true",
                    help="Bygg vokabulär + koda allt med mE5-large.")
    ap.add_argument("--with-wiki", action="store_true",
                    help="Hämta svenska Wikipedia-sammanfattningar för entiteter "
                         "(rekommenderas — annars blir 'Avicii' bara nära andra namn).")
    ap.add_argument("--limit-entities", type=int, default=None,
                    help="Max antal entiteter (för snabb iteration).")
    ap.add_argument("--limit-korp", type=int, default=None,
                    help="Max antal Korp-ord (frekvens-rankat, vanliga först).")
    ap.add_argument("--batch-size", type=int, default=16,
                    help="Encoder-batchstorlek. Höj på GPU, sänk om OOM på CPU.")
    args = ap.parse_args()
 
    if args.build:
        vocab = build_vocabulary(
            with_wiki=args.with_wiki,
            limit_entities=args.limit_entities,
            limit_korp=args.limit_korp,
        )
        vectors = encode_vocabulary(vocab, batch_size=args.batch_size)
        save_artifacts(vocab, vectors)
        print("\n✅ Klart. Kör nu:  python test_e5_pipeline.py")
    else:
        repl()
 
 
if __name__ == "__main__":
    main()
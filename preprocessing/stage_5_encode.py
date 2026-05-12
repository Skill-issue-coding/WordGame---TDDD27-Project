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
from pathlib import Path

import numpy as np
import pandas as pd
import torch

BASE_DIR   = Path(__file__).resolve().parent
STAGE3_DIR = BASE_DIR / "intermediate" / "stage3_attrs"
STAGE4_CSV = BASE_DIR / "intermediate" / "stage4_general" / "general_words.csv"
OUTPUT_DIR = BASE_DIR / "intermediate" / "stage5_encoded"

MODEL_NAME     = "intfloat/multilingual-e5-large"
BATCH_SIZE     = 512
SUMMARY_MAX_CHARS = 1500


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

def load_entities() -> dict[str, tuple[str, str]]:
    """Returns {lowercase_key: (canonical_name, passage_text)}."""
    records: dict[str, tuple[str, str]] = {}

    if not STAGE3_DIR.exists():
        print(f"Varning: {STAGE3_DIR} saknas — inga entiteter laddas.")
        return records

    for csv_path in sorted(STAGE3_DIR.glob("*.csv")):
        df = pd.read_csv(csv_path)

        label_col = next((c for c in df.columns if c.endswith("Label")), None)
        if label_col is None:
            label_col = next((c for c in ("name", "word") if c in df.columns), None)
        if label_col is None:
            print(f"Varning: ingen namn-kolumn i {csv_path.name}, hoppar över.")
            continue

        for _, row in df.iterrows():
            name = str(row.get(label_col, "")).strip()
            if not name:
                continue
            key = name.lower()
            if key in records:
                continue  # first CSV wins (sorted → deterministic)
            summary = str(row.get("wiki_summary", ""))
            attrs   = str(row.get("wiki_attributes", ""))
            records[key] = (name, entity_passage(name, summary, attrs))

    return records


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

    # 1. Collect items
    entity_records = load_entities()
    print(f"Entiteter (stage 3): {len(entity_records):,}")

    word_records = load_general_words(set(entity_records.keys()))
    print(f"Generella ord (stage 4): {len(word_records):,}")

    # Entities first so their indices stay grouped, then general words
    all_items = list(entity_records.values()) + list(word_records.values())
    if not all_items:
        print("Inga poster att koda — avbryter.")
        sys.exit(1)

    vocab    = [name    for name, _       in all_items]
    passages = [passage for _,    passage in all_items]
    print(f"Totalt att koda: {len(passages):,}")

    # 2. Load E5 model
    try:
        from sentence_transformers import SentenceTransformer
    except ImportError:
        print("sentence-transformers saknas. Kör: pip install sentence-transformers")
        sys.exit(1)

    print(f"Laddar modell '{MODEL_NAME}'...")
    model = SentenceTransformer(MODEL_NAME, device=device)

    # 3. Encode
    print(f"Kodar i batchar om {BATCH_SIZE}...")
    embeddings = model.encode(
        passages,
        batch_size=BATCH_SIZE,
        show_progress_bar=True,
        normalize_embeddings=True,  # cosine sim == dot product after L2 norm
        convert_to_numpy=True,
    )

    embeddings = embeddings.astype(np.float32)

    # 4. Save
    emb_path   = OUTPUT_DIR / "embeddings.npy"
    vocab_path = OUTPUT_DIR / "vocab.json"

    np.save(str(emb_path), embeddings)
    with vocab_path.open("w", encoding="utf-8") as f:
        json.dump(vocab, f, ensure_ascii=False)

    print(f"\nKlar! {len(vocab):,} poster kodade.")
    print(f"  embeddings : {emb_path}  (shape {embeddings.shape})")
    print(f"  vocab      : {vocab_path}")


if __name__ == "__main__":
    main()

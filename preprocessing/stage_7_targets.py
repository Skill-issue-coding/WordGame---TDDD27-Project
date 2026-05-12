"""
Stage 7: Build the curated Contexto target word list.

Not all words make good Contexto targets. "de", "ett", "ha" are too
abstract; highly technical or rare terms frustrate players. This script
filters stage 4's general vocabulary and stage 3's entities down to a
list of concrete, recognisable Swedish words that work as game targets.

Input:
  - intermediate/stage4_general/general_words.csv  (word, lemma, pos, Totalt, in_kelly)
  - intermediate/stage3_attrs/*.csv                (entities with wiki_summary / wiki_attributes)
  - intermediate/stage5_encoded/vocab.json         (only words that were actually encoded make it)

Output:
  - server/wordfiles/targets.json   JSON list of target word strings

At runtime the Go backend picks a random entry from this list and calls
CalculateDistance() for every guess — no precomputed rankings needed.
"""

import json
import sys
from pathlib import Path

import pandas as pd

BASE_DIR   = Path(__file__).resolve().parent
STAGE3_DIR = BASE_DIR / "intermediate" / "stage3_attrs"
STAGE4_CSV = BASE_DIR / "intermediate" / "stage4_general" / "general_words.csv"
VOCAB_FILE = BASE_DIR / "intermediate" / "stage5_encoded" / "vocab.json"
OUTPUT_DIR = BASE_DIR.parent / "server" / "wordfiles"
OUT_FILE   = OUTPUT_DIR / "targets.json"

# Only these POS types are interesting as Contexto targets.
# PROPN (proper nouns) come from entities and are included separately.
TARGET_POS = {"NOUN"}

# Must appear at least this often in Korp AND be in Kelly to pass.
MIN_KORP_FREQ = 1_000

# Word length limits — too short = ambiguous, too long = obscure
MIN_WORD_LEN = 4
MAX_WORD_LEN = 20


def load_encoded_vocab() -> set[str]:
    """Words that were actually encoded in stage 5 (our source of truth)."""
    if not VOCAB_FILE.exists():
        print(f"Varning: {VOCAB_FILE} saknas — körs utan vocab-filter.")
        return set()
    with VOCAB_FILE.open("r", encoding="utf-8") as f:
        return {w.lower() for w in json.load(f)}


def collect_general_targets(encoded: set[str]) -> list[str]:
    if not STAGE4_CSV.exists():
        print(f"Varning: {STAGE4_CSV} saknas — inga generella målord.")
        return []

    df = pd.read_csv(STAGE4_CSV)
    df["Totalt"] = pd.to_numeric(df["Totalt"], errors="coerce").fillna(0)

    mask = (
        df["pos"].isin(TARGET_POS)
        & (df["Totalt"] >= MIN_KORP_FREQ)
        & df["in_kelly"].astype(str).str.lower().isin({"true", "1", "yes"})
        & (df["word"].str.len() >= MIN_WORD_LEN)
        & (df["word"].str.len() <= MAX_WORD_LEN)
    )
    words = df.loc[mask, "word"].dropna().astype(str).tolist()

    if encoded:
        words = [w for w in words if w.lower() in encoded]

    return words


def collect_entity_targets(encoded: set[str]) -> list[str]:
    if not STAGE3_DIR.exists():
        return []

    entities = []
    for csv_path in sorted(STAGE3_DIR.glob("*.csv")):
        df = pd.read_csv(csv_path)

        label_col = next((c for c in df.columns if c.endswith("Label")), None)
        if label_col is None:
            label_col = next((c for c in ("name", "word") if c in df.columns), None)
        if label_col is None:
            continue

        for name in df[label_col].dropna().astype(str):
            name = name.strip()
            if not name or len(name) < MIN_WORD_LEN or len(name) > MAX_WORD_LEN:
                continue
            # Only include if the entity actually has wiki context (better targets)
            row = df.loc[df[label_col] == name].iloc[0]
            has_context = bool(str(row.get("wiki_summary", "")).strip()) or bool(
                str(row.get("wiki_attributes", "")).strip()
            )
            if not has_context:
                continue
            if encoded and name.lower() not in encoded:
                continue
            entities.append(name)

    return entities


def main():
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    encoded = load_encoded_vocab()

    general  = collect_general_targets(encoded)
    entities = collect_entity_targets(encoded)

    # Deduplicate, entities take priority (they have richer context)
    seen: set[str] = set()
    targets: list[str] = []
    for word in entities + general:
        key = word.lower()
        if key not in seen:
            seen.add(key)
            targets.append(word)

    if not targets:
        print("Inga målord hittades — kontrollera att stage 3 och 4 körts.")
        sys.exit(1)

    targets.sort(key=str.lower)

    with OUT_FILE.open("w", encoding="utf-8") as f:
        json.dump(targets, f, ensure_ascii=False, indent=2)

    print(f"Klar! {len(targets):,} målord sparade till {OUT_FILE}")
    print(f"  varav {len(entities):,} entiteter och {len(general):,} generella ord")


if __name__ == "__main__":
    main()

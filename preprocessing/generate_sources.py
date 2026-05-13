"""
Backfill sources.json for an existing encoded vocabulary.

Reads the current vocab.json and reverse-looks up each word against
intermediate/stage3_attrs/*.csv to determine its category.  Words not found
in any entity CSV are labelled "general".

Run this once to give the Go server category metadata without having to
re-encode everything through stage 5 (which takes hours).

Writes:
  - intermediate/stage5_encoded/sources.json
  - server/wordfiles/sources.json
"""

import json
import re
from pathlib import Path

import pandas as pd

try:
    from shared import CATEGORY_MAPPING
except ImportError:
    CATEGORY_MAPPING: dict = {}

BASE_DIR    = Path(__file__).resolve().parent
STAGE3_DIR  = BASE_DIR / "intermediate" / "stage3_attrs"
ENCODED_DIR = BASE_DIR / "intermediate" / "stage5_encoded"
OUTPUT_DIR  = BASE_DIR.parent / "server" / "wordfiles"

VOCAB_FILE  = ENCODED_DIR / "vocab.json"


_QID_RE = re.compile(r"^Q\d+$")

def _is_valid_label(value: str) -> bool:
    value = (value or "").strip()
    return (bool(value)
            and not value.startswith("http")
            and not _QID_RE.match(value)
            and any(c.isalpha() for c in value))


def build_entity_category_map() -> dict[str, str]:
    """Returns {lowercase_name: category} for all valid entity labels."""
    cat_map: dict[str, str] = {}

    if not STAGE3_DIR.exists():
        print(f"Varning: {STAGE3_DIR} saknas.")
        return cat_map

    for csv_path in sorted(STAGE3_DIR.glob("*.csv")):
        cat = CATEGORY_MAPPING.get(csv_path.stem, ("general", ""))[0]
        df  = pd.read_csv(csv_path)

        label_col = next((c for c in df.columns if c.endswith("Label")), None)
        if label_col is None:
            label_col = next((c for c in ("name", "word") if c in df.columns), None)
        if label_col is None:
            continue

        for name in df[label_col].dropna().astype(str):
            name = name.strip()
            if _is_valid_label(name):
                cat_map.setdefault(name.lower(), cat)

    return cat_map


def main() -> None:
    if not VOCAB_FILE.exists():
        print(f"Fel: {VOCAB_FILE} saknas. Kör stage 5 först.")
        return

    with VOCAB_FILE.open("r", encoding="utf-8") as f:
        vocab: list[str] = json.load(f)

    print(f"Laddar entitetskategorier från {STAGE3_DIR}…")
    entity_cats = build_entity_category_map()
    print(f"  {len(entity_cats):,} entiteter med känd kategori")

    sources = [entity_cats.get(w.lower(), "general") for w in vocab]

    counts: dict[str, int] = {}
    for s in sources:
        counts[s] = counts.get(s, 0) + 1
    print("Kategorifördelning:")
    for cat, n in sorted(counts.items()):
        print(f"  {cat}: {n:,}")

    for out_dir in (ENCODED_DIR, OUTPUT_DIR):
        out_dir.mkdir(parents=True, exist_ok=True)
        out_path = out_dir / "sources.json"
        with out_path.open("w", encoding="utf-8") as f:
            json.dump(sources, f, ensure_ascii=False)
        print(f"Skrev {out_path}")


if __name__ == "__main__":
    main()

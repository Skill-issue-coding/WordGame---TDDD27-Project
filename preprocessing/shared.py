# shared.py
import csv
import logging
import re
import sys
from collections import defaultdict
from pathlib import Path
from typing import Dict, Optional
from dotenv import load_dotenv

BASE_DIR = Path(__file__).resolve().parent
load_dotenv(dotenv_path=BASE_DIR / ".env.local")

# ── Configuration ─────────────────────────────────────────────────────────────
FASTTEXT_MODEL_PATH = BASE_DIR / "cc.sv.300.bin"
KORP_DIR            = BASE_DIR / "korp"
OUTPUT_DIR          = BASE_DIR.parent / "server" / "wordfiles"
INTERMEDIATE_DIR    = BASE_DIR / "intermediate"
SPACY_MODEL         = "sv_core_news_sm"

PCA_DIMS            = 100
NEIGHBOURS_PER_SEED = 70
TOP_N_PER_SEED      = 50
MIN_WORD_LEN        = 3

# NEW: Dynamic Korp Frequencies
DEFAULT_KORP_FREQ = 300 # High threshold for general words (verbs, adjectives, standard nouns)
CATEGORY_KORP_FREQ = {
    "character": 5,     # Keep low for specific names
    "game": 5,          
    "media": 10,
    "celebrity": 20,
    "company": 50,
}

# NEW: Expanded POS tags for better party-game vocabulary
ALLOWED_POS = {"NOUN", "PROPN", "VERB", "ADJ"}

CATEGORY_MAPPING = {
    "swedish_celebrities": ("celebrity",  "celebrities_vectors.csv"),
    "swedish_music":       ("celebrity",  "celebrities_vectors.csv"),
    "swedish_companies":   ("company",    "companies_vectors.csv"),
    "global_brands":       ("company",    "companies_vectors.csv"),
    "swedish_characters":  ("character",  "characters_vectors.csv"),
    "swedish_food":        ("food",       "food_vectors.csv"),
    "swedish_geography":   ("geography",  "geography_vectors.csv"),
    "swedish_tv_and_film": ("media",      "media_vectors.csv"),
    "video_games":         ("game",       "games_vectors.csv"),
    "swedish_culture":     ("general",    "culture_vectors.csv"), 
}

# ── Logging Setup (Terminal clean, File detailed) ─────────────────────────────
file_handler = logging.FileHandler(BASE_DIR / "pipeline.log", encoding="utf-8")
file_handler.setLevel(logging.INFO)

console_handler = logging.StreamHandler(sys.stdout)
console_handler.setLevel(logging.WARNING)

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    handlers=[file_handler, console_handler],
)

# ── Shared Loaders & Helpers ──────────────────────────────────────────────────
def setup_dirs():
    INTERMEDIATE_DIR.mkdir(parents=True, exist_ok=True)
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

def _extract_label(row: Dict[str, str]) -> Optional[str]:
    for key in ("personLabel", "artistLabel", "companyLabel", "brandLabel", "characterLabel", "foodLabel", "placeLabel", "workLabel", "gameLabel"):
        val = row.get(key, "").strip()
        if val and not re.match(r"^Q\d+$", val): return val
    for val in row.values():
        val = val.strip()
        if val and not val.startswith("http") and not re.match(r"^Q\d+$", val): return val
    return None

def load_fasttext():
    import fasttext
    if not FASTTEXT_MODEL_PATH.exists():
        logging.error(f"fastText model not found: {FASTTEXT_MODEL_PATH}")
        sys.exit(1)
    return fasttext.load_model(str(FASTTEXT_MODEL_PATH))

def load_spacy():
    import spacy
    try:
        return spacy.load(SPACY_MODEL, disable=["parser", "ner", "senter"])
    except OSError:
        logging.error(f"spaCy model '{SPACY_MODEL}' not found.")
        sys.exit(1)

# In shared.py
def load_korp_freq() -> Optional[Dict[str, int]]:
    if not KORP_DIR.exists(): return None
    
    cache_path = INTERMEDIATE_DIR / "korp_cache.pkl"
    if cache_path.exists():
        import pickle
        with open(cache_path, "rb") as f:
            return pickle.load(f)

    freq = defaultdict(int)
    for csv_path in sorted(KORP_DIR.glob("*.csv")):
        try:
            with csv_path.open(encoding="utf-8", newline="") as f:
                reader = csv.DictReader(f)
                headers = [h.lstrip("\ufeff").strip() for h in reader.fieldnames or []]
                if "word" not in headers or "Totalt" not in headers: continue
                f.seek(0)
                reader = csv.DictReader(f, fieldnames=headers)
                next(reader)
                for row in reader:
                    word = row.get("word", "").strip().lower()
                    if word: freq[word] += int(float(row.get("Totalt", 0) or 0))
        except Exception: pass
    result = dict(freq) if freq else None
    
    # Save to cache
    if result:
        import pickle
        INTERMEDIATE_DIR.mkdir(parents=True, exist_ok=True)
        with open(cache_path, "wb") as f:
            pickle.dump(result, f)
            
    return result
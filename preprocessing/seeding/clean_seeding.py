import os
import glob
import pandas as pd
import requests
import time
import unicodedata
import re
from pathlib import Path

CURRENT_DIR = Path(__file__).resolve().parent

# === Configuration Paths ===
MAKTBAROMETERN_DIR = CURRENT_DIR / "maktbarometern" / "csv"
SEEDING_OUTPUT_DIR = CURRENT_DIR / "output"
CLEANED_DIR = CURRENT_DIR.parent / "intermediate" / "seeding_cleaned"

# === Score Thresholds ===
# Only entries with score >= the threshold for their platform are kept.
# Set to 0 to disable filtering for a platform.
# Platform scores vary widely (x: 3-8, tiktok: 7-16, facebook: 11-30,
# youtube: 15-51, instagram: 16-55, arets-makthavare: 33-142).
SCORE_LIMITS: dict[str, int] = {
    "arets-makthavare": 35,
    "facebook":         15,
    "instagram":        20,
    "tiktok":           13,
    "x":                6,
    "youtube":          18,
}
# Fallback for any file not listed above.
DEFAULT_SCORE_LIMIT = 0

def clean_text(text):
    """Normalizes Unicode and removes emojis/special characters."""
    if not isinstance(text, str):
        return text
    # Normalize full-width characters and accents
    text = unicodedata.normalize('NFKC', text)
    # Keep only letters, numbers, spaces, hyphens, periods, and apostrophes
    text = re.sub(r'[^\w\s\-\.\']', '', text)
    # Strip extra whitespace
    return re.sub(r'\s+', ' ', text).strip()

def extract_qid(val):
    """Detects if a string is a Wikidata Q-ID or URI."""
    val_str = str(val).strip()
    # Matches "Q17452" or "http://www.wikidata.org/entity/Q17452"
    match = re.search(r"(Q\d+)$", val_str)
    if match and (val_str == match.group(1) or "entity/Q" in val_str):
        return match.group(1)
    return None

def fetch_labels_from_wikidata(qids):
    """Fetches labels for a list of Q-IDs via the Wikidata API (batches of 50)."""
    labels = {}
    qids = list(set(qids))
    chunk_size = 50 
    
    print(f"Fetching {len(qids)} missing names from the Wikidata API...")
    
    for i in range(0, len(qids), chunk_size):
        chunk = qids[i:i+chunk_size]
        ids_str = "|".join(chunk)
        url = f"https://www.wikidata.org/w/api.php?action=wbgetentities&ids={ids_str}&props=labels&languages=sv|en&format=json"
        
        try:
            resp = requests.get(url, headers={"User-Agent": "WordGameBot/1.0"}).json()
            if 'entities' in resp:
                for qid, data in resp['entities'].items():
                    if 'labels' in data:
                        # Prefer Swedish, fallback to English
                        if 'sv' in data['labels']:
                            labels[qid] = data['labels']['sv']['value']
                        elif 'en' in data['labels']:
                            labels[qid] = data['labels']['en']['value']
            time.sleep(0.1) # Be polite to the API
        except Exception as e:
            print(f"Error fetching chunk: {e}")
            
    return labels

def process_maktbarometern():
    print("--- Processing Maktbarometern ---")
    os.makedirs(CLEANED_DIR, exist_ok=True)
    files = glob.glob(os.path.join(MAKTBAROMETERN_DIR, "*.csv"))
    if not files:
        print(f"No files found in {MAKTBAROMETERN_DIR}")
        return
        
    all_dfs = []
    for f in files:
        df = pd.read_csv(f)

        # Standardize the name column
        if 'name' not in df.columns and 'account_name' in df.columns:
            df['name'] = df['account_name']
        elif 'name' not in df.columns:
            continue

        cols_to_keep = ['name']
        if 'score' in df.columns:
            cols_to_keep.append('score')

        df = df[cols_to_keep].dropna(subset=['name']).copy()

        # Apply score threshold: derive platform key from filename (e.g. "2025-facebook.csv" → "facebook")
        if 'score' in df.columns:
            stem = Path(f).stem  # e.g. "2025-facebook"
            platform = stem.split('-', 1)[-1] if '-' in stem else stem  # e.g. "facebook"
            limit = SCORE_LIMITS.get(platform, DEFAULT_SCORE_LIMIT)
            if limit > 0:
                before = len(df)
                df['score'] = pd.to_numeric(df['score'], errors='coerce')
                df = df[df['score'] >= limit]
                print(f"  {Path(f).name}: score >= {limit} → kept {len(df)}/{before} rows")

        # Apply the Unicode/Emoji cleaner
        df['name'] = df['name'].apply(clean_text)

        # Drop rows that became empty after cleaning (e.g. accounts that were just emojis)
        df = df[df['name'] != ""]
        all_dfs.append(df)
        
    if all_dfs:
        combined = pd.concat(all_dfs, ignore_index=True)
        
        # Aggregate duplicates across platforms by keeping their highest score
        if 'score' in combined.columns:
            combined['score'] = pd.to_numeric(combined['score'], errors='coerce').fillna(0)
            combined = combined.groupby('name', as_index=False)['score'].max()
        else:
            combined = combined.drop_duplicates(subset=['name'])
            
        # Sort by score descending
        if 'score' in combined.columns:
            combined = combined.sort_values(by='score', ascending=False)
            
        out_path = os.path.join(CLEANED_DIR, "maktbarometern_cleaned.csv")
        combined.to_csv(out_path, index=False)
        print(f"Saved combined Maktbarometern to '{out_path}' ({len(combined)} influencers)")

def process_seeding():
    print("\n--- Processing Wikidata Seeding ---")
    os.makedirs(CLEANED_DIR, exist_ok=True)
    
    files = glob.glob(os.path.join(SEEDING_OUTPUT_DIR, "*.csv"))
    if not files:
        print(f"No files found in {SEEDING_OUTPUT_DIR}")
        return
        
    for f in files:
        df = pd.read_csv(f)
        filename = os.path.basename(f)
        
        # Auto-detect the label column (e.g., personLabel, gameLabel)
        label_col = next((col for col in df.columns if col.endswith("Label")), None)
        if not label_col:
            label_col = 'word' if 'word' in df.columns else None
            
        if not label_col:
            continue
                
        # 1. Identify missing names (Q-IDs)
        qids_to_fetch = set()
        for val in df[label_col].dropna():
            qid = extract_qid(val)
            if qid:
                qids_to_fetch.add(qid)
                
        # 2. Fetch missing names from Wikidata
        fetched_labels = {}
        if qids_to_fetch:
            fetched_labels = fetch_labels_from_wikidata(list(qids_to_fetch))
            
        # 3. Replace Q-IDs and Clean all text
        def update_and_clean(val):
            val_str = str(val).strip()
            qid = extract_qid(val_str)
            
            # If it's a Q-ID, replace it with the API result
            if qid:
                val_str = fetched_labels.get(qid, val_str)
            
            # If it's STILL a Q-ID (meaning Wikidata has no label at all), return empty to drop it
            if extract_qid(val_str):
                return ""
                
            return clean_text(val_str)

        df[label_col] = df[label_col].apply(update_and_clean)
        
        # Drop any entities that resolved to empty names
        df = df[df[label_col] != ""]
        
        # Drop exact duplicates to save pipeline processing time
        df = df.drop_duplicates(subset=[label_col])
        
        out_path = os.path.join(CLEANED_DIR, filename)
        df.to_csv(out_path, index=False)
        print(f"{filename}: Cleaned and saved {len(df)} entities.")

if __name__ == "__main__":
    process_maktbarometern()
    process_seeding()
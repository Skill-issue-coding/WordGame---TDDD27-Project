import os
import glob
import pandas as pd
import re
import sys

# Get the directory one level up (the preprocessing folder)
parent_dir = os.path.abspath(os.path.join(os.path.dirname(__file__), '..'))
# Add it to Python's search path
sys.path.insert(0, parent_dir)

try:
    from shared import load_custom_stopwords, CLEANED_KORP_DIR
    HAS_SHARED = True
except ImportError as e:
    HAS_SHARED = False
    print(f"Varning: Kunde inte importera från 'shared.py': {e}")
    CLEANED_KORP_DIR = os.path.join(parent_dir, "korp_cleaned")

# === Configuration ===
KORP_DIR = "korp"
MIN_FREQ = 5           # Minimum occurrences across ALL files
MIN_LEN = 2             # Minimum character length
MAX_LEN = 25            # Maximum character length

# Pre-compiled Regex: Letters only, allows internal hyphens.
VALID_WORD_PATTERN = re.compile(r"^[a-zåäöéü]+(?:-[a-zåäöéü]+)*$", re.IGNORECASE)

def is_valid_word(word):
    """Returns True if the word is a valid string of letters."""
    word = str(word).strip()
    
    if len(word) < MIN_LEN or len(word) > MAX_LEN:
        return False
        
    if not VALID_WORD_PATTERN.fullmatch(word):
        return False
        
    return True

def main():
    os.makedirs(CLEANED_KORP_DIR, exist_ok=True)
    korp_files = glob.glob(os.path.join(KORP_DIR, "*.csv"))
    
    if not korp_files:
        print(f"No CSVs found in {KORP_DIR}/")
        return

    # 1. Load Kelly and Stopwords
    stopwords = set()
    if HAS_SHARED:
        print("Laddar Stopwords...")
        stopwords = load_custom_stopwords()
        print(f" -> {len(stopwords)} stoppord laddade.")
    
    total_original = 0
    all_dfs = []

    print("\nLoading and pre-filtering Korp datasets...\n" + "="*40)

    for file in korp_files:
        filename = os.path.basename(file)
        
        try:
            df = pd.read_csv(file, low_memory=False, usecols=lambda c: c in ['word', 'Totalt'])
        except Exception as e:
            print(f"Error loading {filename}: {e}")
            continue

        if 'word' not in df.columns or 'Totalt' not in df.columns:
            continue

        original_len = len(df)
        total_original += original_len

        # 1. Clean data types and drop NA
        df = df.dropna(subset=['word'])
        df['Totalt'] = pd.to_numeric(df['Totalt'], errors='coerce').fillna(0)

        # 2. Lowercase immediately for accurate regex, stopword, and Kelly matching
        df['word'] = df['word'].astype(str).str.strip().str.lower()

        if stopwords:
            # Force to list for safer Pandas matching
            stop_list = list(stopwords) 
            
            # Print a debug line on the first file to PROVE they loaded
            if len(all_dfs) == 0: 
                print(f" -> Stopwords look like this: {stop_list[:10]}")
                if "och" not in stop_list:
                    print("VARNING: 'och' finns inte i dina inladdade stoppord!")
            
            df = df[~df['word'].isin(stop_list)]

        # 5. Apply the regex text filter
        mask = df['word'].apply(is_valid_word)
        df = df[mask]
        
        all_dfs.append(df)
        print(f"Loaded {filename}: Kept {len(df):,} strictly valid dictionary words out of {original_len:,}")

    if not all_dfs:
        print("No valid data found to process.")
        return

    print("\nCombining and aggregating frequencies...")
    combined_df = pd.concat(all_dfs, ignore_index=True)
    
    # 6. Group by word and sum the frequencies
    aggregated_df = combined_df.groupby('word', as_index=False)['Totalt'].sum()

    # 7. Apply the Minimum Frequency threshold
    final_df = aggregated_df[aggregated_df['Totalt'] >= MIN_FREQ]
    
    # Sort by frequency descending
    final_df = final_df.sort_values('Totalt', ascending=False)
    
    total_cleaned = len(final_df)

    output_path = os.path.join(OUTPUT_DIR, "korp_combined_cleaned.csv")
    final_df.to_csv(output_path, index=False)

    print("="*40)
    print(f"TOTAL ORIGINAL ROWS: {total_original:,}")
    print(f"FINAL UNIQUE, CLEANED WORDS: {total_cleaned:,}")
    print(f"Cleaned combined file saved to '{output_path}'.")

if __name__ == "__main__":
    main()
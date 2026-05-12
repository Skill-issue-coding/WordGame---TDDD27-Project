import os
import pandas as pd
import spacy
import time

try:
    from shared import (
        read_korp, 
        load_kelly, 
        load_custom_stopwords, 
        ALLOWED_POS,  # e.g., {"NOUN", "VERB", "ADJ", "PROPN"}
        DEFAULT_KORP_FREQ
    )
    HAS_SHARED = True
except ImportError as e:
    HAS_SHARED = False
    print(f"Varning: Kunde inte importera från shared.py: {e}")
    exit()

# === Configuration ===
OUTPUT_DIR = "intermediate/stage4_general"
SPACY_MODEL = "sv_core_news_sm"

def main():
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    
    print("=" * 50)
    print("STAGE 4: LINGUISTIC FILTERING (spaCy + Kelly)")
    print("=" * 50)

    # 1. Load Korp Data
    print("Laddar Korp-frekvenser...")
    if HAS_SHARED:
        korp_data = read_korp() 
        
        if isinstance(korp_data, dict):
            df = pd.DataFrame(list(korp_data.items()), columns=['word', 'Totalt'])
        elif isinstance(korp_data, list):
            df = pd.DataFrame(korp_data)
        else:
            df = korp_data # Assume it's already a DataFrame
    else:
        # Fallback to reading the combined CSV directly
        df = pd.read_csv("korp_cleaned/korp_combined_cleaned.csv")
    
    original_len = len(df)
    print(f" -> Hittade {original_len:,} ord i Korp.")

    # 2. Filter by Baseline Frequency
    # We enforce a strict minimum frequency for general words to ensure they are well-known
    df['Totalt'] = pd.to_numeric(df['Totalt'], errors='coerce').fillna(0)
    df = df[df['Totalt'] >= DEFAULT_KORP_FREQ]
    print(f" -> Efter frekvenskrav (>= {DEFAULT_KORP_FREQ}): {len(df):,} ord kvar.")

    # 3. Load Kelly & Stopwords
    kelly_words = set()
    stopwords = set()
    if HAS_SHARED:
        kelly_words = load_kelly()
        stopwords = load_custom_stopwords()
    
    # Drop custom stopwords immediately to save spaCy processing time
    if stopwords:
        df = df[~df['word'].isin(stopwords)]

    # 4. Process with spaCy (The heavy lifting)
    print(f"\nLaddar spaCy-modellen '{SPACY_MODEL}'...")
    try:
        # Disable parser and ner for massive speed improvements
        nlp = spacy.load(SPACY_MODEL, disable=['parser', 'ner', 'senter'])
    except OSError:
        print(f"Fel: spaCy-modellen saknas. Kör: python -m spacy download {SPACY_MODEL}")
        return

    words_to_process = df['word'].astype(str).tolist()
    valid_words = []
    
    print(f"Börjar ordklasstaggning (POS) av {len(words_to_process):,} ord...")
    start_time = time.time()

    # nlp.pipe is highly optimized for processing large lists of text
    for doc in nlp.pipe(words_to_process, batch_size=2048):
        token = doc[0] # Since we feed single words, there is only one token
        word_text = token.text
        pos = token.pos_
        
        # Check 1: Is it an allowed Part-Of-Speech? (Noun, Verb, Adjective)
        if pos not in ALLOWED_POS:
            continue
            
        # Check 2: spaCy's built-in stopword detection
        if token.is_stop:
            continue
            
        # Check 3: The Kelly List Synergy
        # If it's a valid POS, we keep it. BUT if spaCy is unsure, or it's a weird word, 
        # we check if its base form (lemma) exists in the official Kelly dictionary.
        lemma = token.lemma_.lower()
        is_in_kelly = lemma in kelly_words if kelly_words else False
        
        valid_words.append({
            'word': word_text,
            'lemma': lemma,
            'pos': pos,
            'in_kelly': is_in_kelly
        })

    print(f"Taggning klar på {time.time() - start_time:.1f} sekunder.")

    # 5. Merge and Finalize
    valid_df = pd.DataFrame(valid_words)
    final_df = pd.merge(df, valid_df, on='word', how='inner')
    
    # Sort by Korp frequency so the most common words are at the top
    final_df = final_df.sort_values(by='Totalt', ascending=False)

    out_path = os.path.join(OUTPUT_DIR, "general_words.csv")
    final_df.to_csv(out_path, index=False)

    print("\n" + "=" * 50)
    print(f"RESULTAT: Reducerade {original_len:,} -> {len(final_df):,} högkvalitativa ord.")
    print(f"Sparad till: {out_path}")
    print("=" * 50)

if __name__ == "__main__":
    main()
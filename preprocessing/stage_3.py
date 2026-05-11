import pickle
import logging
import re
from shared import (
    INTERMEDIATE_DIR, MIN_WORD_LEN, DEFAULT_KORP_FREQ, 
    CATEGORY_KORP_FREQ, ALLOWED_POS, load_spacy, load_korp_freq, load_kelly
)

# Regex to find any standard Swedish vowel (including é for words like pokémon/créme)
VOWEL_REGEX = re.compile(r'[aeiouyåäöé]', re.IGNORECASE)

def main():
    print("starting stage 3")
    logging.info("=" * 60)
    logging.info("Stage 3: Validation, Formatting & Kelly Filter")
    
    with open(INTERMEDIATE_DIR / "stage2_candidates.pkl", "rb") as f:
        candidates = pickle.load(f)

    nlp = load_spacy()
    # korp_freq = load_korp_freq()
    kelly_words = load_kelly()  # Load the Kelly dictionary
    filtered = {}

    for output_csv, entries in candidates.items():
        pre_filtered = []
        for e in entries:
            word, cat, sim, vec = e
            
            # 1. Length and Alpha check (allowing internal hyphens)
            if len(word) < MIN_WORD_LEN or not word.replace("-", "").isalpha():
                continue
                
            # 2. Formatting check: No dangling hyphens, max 1 hyphen total
            if word.startswith("-") or word.endswith("-") or word.count("-") > 1:
                continue
                
            # 3. Gibberish check: Must contain at least one vowel
            if not VOWEL_REGEX.search(word):
                continue
                
            # # 4. Dynamic Korp Frequency Check
            # req_freq = CATEGORY_KORP_FREQ.get(cat, DEFAULT_KORP_FREQ)
            # if korp_freq is None or korp_freq.get(word, 0) < req_freq:
            #     continue

            pre_filtered.append(e)
            
        if not pre_filtered:
            filtered[output_csv] = []
            continue

        # # Sort by Korp frequency BEFORE deduplication to favor common lemmas
        # if korp_freq:
        #     pre_filtered.sort(key=lambda x: korp_freq.get(x[0], 0), reverse=True)

        words = [e[0] for e in pre_filtered]
        docs = nlp.pipe(words, batch_size=1000)
        
        seen_lemmas = set()
        valid_entries = []
        
        for doc, entry in zip(docs, pre_filtered):
            if doc and not doc[0].is_stop and doc[0].pos_ in ALLOWED_POS:
                word_text = doc[0].text.lower()
                pos = doc[0].pos_
                
                # 5. Strict Spellcheck / English Filter
                # If it's a standard word (NOUN, VERB, ADJ), it MUST be in the Kelly dictionary.
                # (We skip this check for PROPN so that names and brands are still allowed)
                if pos in {"NOUN", "VERB", "ADJ"}:
                    if word_text not in kelly_words:
                        continue 

                # 6. Deduplicate by Lemma
                lemma = doc[0].lemma_.lower()
                if lemma not in seen_lemmas:
                    seen_lemmas.add(lemma)
                    valid_entries.append(entry)
                
        filtered[output_csv] = valid_entries

    with open(INTERMEDIATE_DIR / "stage3_validated.pkl", "wb") as f:
        pickle.dump(filtered, f)
        
    print("stage 3 complete")
    logging.info("Stage 3 complete.")

if __name__ == "__main__":
    main()
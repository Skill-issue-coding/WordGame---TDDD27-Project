import pickle
import logging
from shared import INTERMEDIATE_DIR, MIN_WORD_LEN, DEFAULT_KORP_FREQ, CATEGORY_KORP_FREQ, ALLOWED_POS, load_spacy, load_korp_freq

def main():
    print("starting stage 3")
    logging.info("=" * 60)
    logging.info("Stage 3: Validation & Lemma Deduplication")
    logging.info("=" * 60)
    
    with open(INTERMEDIATE_DIR / "stage2_candidates.pkl", "rb") as f:
        candidates = pickle.load(f)

    nlp = load_spacy()
    korp_freq = load_korp_freq()
    filtered = {}

    for output_csv, entries in candidates.items():
        pre_filtered = []
        for e in entries:
            word, cat, sim, vec = e
            
            # 1. Fetch dynamic frequency threshold based on category
            req_freq = CATEGORY_KORP_FREQ.get(cat, DEFAULT_KORP_FREQ)
            
            # 2. Pre-filter by length, alpha, and dynamic Korp threshold
            if len(word) >= MIN_WORD_LEN and word.replace("-", "").isalpha():
                if korp_freq is None or korp_freq.get(word, 0) >= req_freq:
                    pre_filtered.append(e)
        
        if not pre_filtered:
            filtered[output_csv] = []
            continue

        # 3. Sort by Korp frequency BEFORE deduplication.
        # This guarantees that the base lemma (usually highest freq) wins the race.
        if korp_freq:
            pre_filtered.sort(key=lambda x: korp_freq.get(x[0], 0), reverse=True)

        words = [e[0] for e in pre_filtered]
        docs = nlp.pipe(words, batch_size=1000)
        
        seen_lemmas = set()
        valid_entries = []
        
        # 4. Process via C-optimized batches and Deduplicate by Lemma
        for doc, entry in zip(docs, pre_filtered):
            if doc and not doc[0].is_stop and doc[0].pos_ in ALLOWED_POS:
                lemma = doc[0].lemma_.lower()
                
                # If we haven't seen this root word yet, add it!
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
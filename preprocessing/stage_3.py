import pickle
import logging
from shared import INTERMEDIATE_DIR, MIN_WORD_LEN, MIN_KORP_FREQ, ALLOWED_POS, load_spacy, load_korp_freq

def _is_valid(word: str, nlp, korp_freq) -> bool:
    if len(word) < MIN_WORD_LEN or not word.replace("-", "").isalpha(): return False
    if korp_freq is not None and korp_freq.get(word, 0) < MIN_KORP_FREQ: return False
    doc = nlp(word)
    if not doc or doc[0].is_stop or doc[0].pos_ not in ALLOWED_POS: return False
    return True

def main():
    print("starting stage 3")
    logging.info("=" * 60)
    logging.info("Stage 3: Validity Filter")

    # Load from Stage 2
    with open(INTERMEDIATE_DIR / "stage2_candidates.pkl", "rb") as f:
        candidates = pickle.load(f)

    nlp = load_spacy()
    korp_freq = load_korp_freq()
    filtered = {}

    for output_csv, entries in candidates.items():
        filtered[output_csv] = [e for e in entries if _is_valid(e[0], nlp, korp_freq)]

    # Save to disk for Stage 4
    with open(INTERMEDIATE_DIR / "stage3_validated.pkl", "wb") as f:
        pickle.dump(filtered, f)

    print("stage 3 complete")
    logging.info("Stage 3 complete.")

if __name__ == "__main__":
    main()
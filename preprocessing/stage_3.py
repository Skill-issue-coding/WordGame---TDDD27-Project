import logging
import pickle
import re

from shared import (
    ALLOWED_POS,
    INTERMEDIATE_DIR,
    MIN_WORD_LEN,
    load_custom_stopwords,
    load_kelly,
    load_spacy,
)

VOWEL_REGEX = re.compile(r"[aeiouyåäöé]", re.IGNORECASE)
ENTITY_CATEGORIES = {"celebrity", "company", "character", "game", "media", "geography"}
ENTITY_ALLOWED_POS = {"PROPN"}
ENTITY_SPECIAL_CHARS = {" ", "-", "'", ".", ":", "&"}


def _unpack_entry(entry):
    if len(entry) >= 8:
        word, cat, sim, vec, popularity, sitelinks, score, is_seed = entry[:8]
        return word, cat, sim, vec, float(popularity), float(sitelinks), float(score), bool(is_seed)

    word, cat, sim, vec = entry
    return word, cat, sim, vec, 0.0, 0.0, 0.0, False


def _is_valid_surface(word: str, category: str, is_seed: bool) -> bool:
    if len(word) < MIN_WORD_LEN:
        return False

    if word.startswith("-") or word.endswith("-") or "--" in word:
        return False

    if category in ENTITY_CATEGORIES or is_seed:
        compact = "".join(ch for ch in word if ch not in ENTITY_SPECIAL_CHARS)
        if not compact:
            return False
        if not any(ch.isalpha() for ch in compact):
            return False
        return all(ch.isalnum() or ch in ENTITY_SPECIAL_CHARS for ch in word)

    plain = word.replace("-", "")
    if not plain.isalpha():
        return False
    if not VOWEL_REGEX.search(word):
        return False
    return True


def main():
    print("starting stage 3")
    logging.info("=" * 60)
    logging.info("Stage 3: Validation, Formatting & Kelly Filter")

    with open(INTERMEDIATE_DIR / "stage2_candidates.pkl", "rb") as f:
        candidates = pickle.load(f)

    nlp = load_spacy()
    kelly_words = load_kelly()
    custom_stopwords = load_custom_stopwords()
    filtered = {}

    for output_csv, entries in candidates.items():
        pre_filtered = []
        for raw_entry in entries:
            word, cat, sim, vec, popularity, sitelinks, score, is_seed = _unpack_entry(raw_entry)
            if _is_valid_surface(word, cat, is_seed):
                pre_filtered.append((word, cat, sim, vec, popularity, sitelinks, score, is_seed))

        if not pre_filtered:
            filtered[output_csv] = []
            continue

        words = [entry[0] for entry in pre_filtered]
        docs = nlp.pipe(words, batch_size=1000)

        seen_keys = set()
        valid_entries = []

        for doc, entry in zip(docs, pre_filtered):
            word, cat, sim, vec, popularity, sitelinks, score, is_seed = entry
            token = next((t for t in doc if not t.is_space), None)
            if token is None:
                continue

            word_text = word.lower()
            lemma = token.lemma_.lower()
            pos = token.pos_
            is_entity = cat in ENTITY_CATEGORIES

            if not is_seed:
                if token.text.lower() in custom_stopwords or lemma in custom_stopwords:
                    continue

                if is_entity:
                    if pos not in ENTITY_ALLOWED_POS:
                        continue
                else:
                    if token.is_stop or pos not in ALLOWED_POS:
                        continue
                    if pos in {"NOUN", "VERB", "ADJ"} and token.text.lower() not in kelly_words:
                        continue

            dedupe_key = word_text if (is_entity or is_seed) else lemma
            if dedupe_key in seen_keys:
                continue

            seen_keys.add(dedupe_key)
            valid_entries.append((word_text, cat, sim, vec, popularity, sitelinks, score, is_seed))

        filtered[output_csv] = valid_entries

    with open(INTERMEDIATE_DIR / "stage3_validated.pkl", "wb") as f:
        pickle.dump(filtered, f)

    print("Stage 3 complete!")
    logging.info("Stage 3 complete.")


if __name__ == "__main__":
    main()

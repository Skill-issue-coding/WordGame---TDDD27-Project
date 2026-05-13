import csv
import logging
import pickle
import re
import time
from collections import defaultdict
from typing import Dict, List, Tuple

import numpy as np

from shared import (
    BASE_DIR,
    CATEGORY_KORP_FREQ,
    CATEGORY_MAPPING,
    DEFAULT_KORP_FREQ,
    INTERMEDIATE_DIR,
    MIN_WORD_LEN,
    _extract_label,
    load_custom_stopwords,
    load_fasttext,
    load_kelly,
    load_korp_freq,
    load_spacy,
    setup_dirs,
)

TOKEN_REGEX = re.compile(r"[0-9a-zåäöé]+", re.IGNORECASE)
VOWEL_REGEX = re.compile(r"[aeiouyåäöé]", re.IGNORECASE)

TOPIC_TYPE_SUFFIX = "_topic"
TOPIC_OUTPUT_CSV = "topics_vectors.csv"
NEIGHBORS_PER_SEED = 250

ENTITY_CATEGORIES = {"celebrity", "company", "character", "game", "media", "geography"}
ENTITY_SPECIAL_CHARS = {" ", "-", "'", ".", ":", "&"}
TOPIC_POS = {"NOUN", "ADJ", "VERB"}


def _normalize_word(text: str) -> str:
    text = " ".join((text or "").strip().split()).lower()
    return text.strip(" _-.,;:!?\"'()[]{}")


def _tokenize_label(label: str) -> List[str]:
    return [token.lower() for token in TOKEN_REGEX.findall(label or "")]


def _safe_float(value) -> float:
    if value is None:
        return 0.0
    try:
        return float(value)
    except (TypeError, ValueError):
        return 0.0


def _extract_seed_popularity(row: Dict[str, str]) -> Tuple[float, float, float]:
    sitelinks = max(0.0, _safe_float(row.get("sitelinks")))
    score = max(0.0, _safe_float(row.get("score")))
    rank = _safe_float(row.get("rank"))
    rank_signal = max(0.0, 101.0 - rank) if rank > 0 else 0.0
    popularity = sitelinks + (score * 4.0) + rank_signal
    if popularity <= 0:
        popularity = max(1.0, sitelinks, score)
    return popularity, sitelinks, score


def _entry_key(word: str, entry_type: str) -> str:
    return f"{word}::{entry_type}"


def _upsert_entry(
    seen: Dict[str, Tuple],
    word: str,
    entry_type: str,
    similarity: float,
    vector: np.ndarray,
    popularity: float,
    sitelinks: float,
    score: float,
    is_seed: bool,
) -> None:
    key = _entry_key(word, entry_type)
    existing = seen.get(key)
    if existing is None:
        seen[key] = (
            word,
            entry_type,
            float(similarity),
            vector,
            float(popularity),
            float(sitelinks),
            float(score),
            bool(is_seed),
        )
        return

    (
        existing_word,
        existing_type,
        existing_similarity,
        existing_vector,
        existing_popularity,
        existing_sitelinks,
        existing_score,
        existing_is_seed,
    ) = existing

    merged_is_seed = existing_is_seed or is_seed
    merged_similarity = max(existing_similarity, float(similarity))
    merged_vector = existing_vector

    if is_seed and not existing_is_seed:
        merged_vector = vector
    elif not existing_is_seed and not is_seed and float(similarity) > existing_similarity:
        merged_vector = vector

    seen[key] = (
        existing_word,
        existing_type,
        merged_similarity,
        merged_vector,
        max(existing_popularity, float(popularity)),
        max(existing_sitelinks, float(sitelinks)),
        max(existing_score, float(score)),
        merged_is_seed,
    )


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


def _basic_surface_ok(word: str) -> bool:
    if len(word) < MIN_WORD_LEN:
        return False
    if word.startswith("-") or word.endswith("-") or "--" in word:
        return False
    return any(ch.isalpha() for ch in word)


def main():
    setup_dirs()
    start_time = time.time()

    print("========================================")
    print(" Starting Simple Stage 2: Neighbor Expansion")
    print("========================================")
    logging.info("=" * 60)
    logging.info("Simple Stage 2: Neighbor Expansion")

    seed_entries_per_output = defaultdict(list)
    seeding_output_dir = BASE_DIR / "seeding" / "output"

    for csv_path in sorted(seeding_output_dir.glob("*.csv")):
        query_name = csv_path.stem
        mapping = CATEGORY_MAPPING.get(query_name)
        if not mapping:
            continue

        category, output_csv = mapping
        with csv_path.open("r", encoding="utf-8") as f:
            rows = list(csv.DictReader(f))

        for row in rows:
            label = _extract_label(row)
            if not label:
                continue

            normalized_label = _normalize_word(label)
            if not normalized_label:
                continue

            popularity, sitelinks, score = _extract_seed_popularity(row)
            seed_entries_per_output[output_csv].append(
                (normalized_label, category, popularity, sitelinks, score)
            )

    print(f"Loaded {sum(len(rows) for rows in seed_entries_per_output.values())} seed rows.")

    print("Loading fastText model... (This usually takes 10-20 seconds)")
    model_start = time.time()
    model = load_fasttext()
    print(f"-> Model loaded in {time.time() - model_start:.2f} seconds!\n")

    print("Loading spaCy and filters...")
    nlp = load_spacy()
    kelly_words = load_kelly()
    custom_stopwords = load_custom_stopwords()
    korp_freq = load_korp_freq() or {}

    results = defaultdict(dict)
    total_categories = len(seed_entries_per_output)

    for idx, (output_csv, seeds) in enumerate(seed_entries_per_output.items(), 1):
        total_seeds = len(seeds)
        print(f"[{idx}/{total_categories}] Processing {output_csv} ({total_seeds} seeds)...")
        cat_start = time.time()

        for seed_idx, (label, category, popularity, sitelinks, score) in enumerate(seeds, 1):
            if seed_idx % 100 == 0:
                print(f"  ... processed {seed_idx}/{total_seeds} seeds")

            valid_tokens = [
                token
                for token in _tokenize_label(label)
                if any(ch.isalpha() for ch in token) and token in model.wv
            ]
            if not valid_tokens:
                continue

            avg_vec = np.mean([model.wv[token] for token in valid_tokens], axis=0)
            _upsert_entry(
                seen=results[output_csv],
                word=label,
                entry_type=category,
                similarity=1.0,
                vector=avg_vec,
                popularity=popularity,
                sitelinks=sitelinks,
                score=score,
                is_seed=True,
            )

            neighbors = model.wv.similar_by_vector(avg_vec, topn=NEIGHBORS_PER_SEED)
            payload = []
            for neighbor_word, similarity in neighbors:
                normalized_neighbor = _normalize_word(neighbor_word)
                if not normalized_neighbor or normalized_neighbor == label:
                    continue
                if not _basic_surface_ok(normalized_neighbor):
                    continue

                payload.append(
                    (
                        normalized_neighbor,
                        float(similarity),
                        model.wv[neighbor_word],
                        popularity,
                        sitelinks,
                        score,
                    )
                )

            if not payload:
                continue

            docs = nlp.pipe((item[0] for item in payload), batch_size=512)
            for doc, item in zip(docs, payload):
                word_text, similarity, vector, seed_popularity, seed_sitelinks, seed_score = item
                token = next((t for t in doc if not t.is_space), None)
                if token is None:
                    continue

                pos = token.pos_
                lemma = token.lemma_.lower() if token.lemma_ else ""

                if pos == "PROPN":
                    if not _is_valid_surface(word_text, category, False):
                        continue
                    _upsert_entry(
                        seen=results[output_csv],
                        word=word_text,
                        entry_type=category,
                        similarity=similarity,
                        vector=vector,
                        popularity=max(1.0, seed_popularity * similarity * 0.2),
                        sitelinks=seed_sitelinks * similarity * 0.2,
                        score=seed_score * similarity * 0.2,
                        is_seed=False,
                    )
                    continue

                if pos not in TOPIC_POS:
                    continue
                if token.is_stop:
                    continue
                if word_text in custom_stopwords or (lemma and lemma in custom_stopwords):
                    continue
                if lemma and kelly_words and lemma not in kelly_words:
                    continue

                if korp_freq:
                    min_freq = CATEGORY_KORP_FREQ.get(category, DEFAULT_KORP_FREQ)
                    if lemma and korp_freq.get(lemma, 0) < min_freq:
                        continue

                if not _is_valid_surface(word_text, "general", False):
                    continue

                topic_type = f"{category}{TOPIC_TYPE_SUFFIX}"
                _upsert_entry(
                    seen=results[TOPIC_OUTPUT_CSV],
                    word=word_text,
                    entry_type=topic_type,
                    similarity=similarity,
                    vector=vector,
                    popularity=max(1.0, seed_popularity * similarity * 0.2),
                    sitelinks=seed_sitelinks * similarity * 0.2,
                    score=seed_score * similarity * 0.2,
                    is_seed=False,
                )

        print(
            f"-> Finished {output_csv} in {time.time() - cat_start:.2f}s. "
            f"Kept {len(results[output_csv])} unique words.\n"
        )

    output = {name: list(entries.values()) for name, entries in results.items()}

    print("Saving intermediate simple candidates to disk...")
    save_start = time.time()
    with open(INTERMEDIATE_DIR / "simple_stage2_candidates.pkl", "wb") as f:
        pickle.dump(output, f)
    print(f"-> Saved in {time.time() - save_start:.2f} seconds.")

    total_time = time.time() - start_time
    print("\n========================================")
    print(f" Simple Stage 2 Complete! (Total time: {total_time:.2f}s)")
    print("========================================")
    logging.info(f"Simple Stage 2 complete in {total_time:.2f} seconds.")


if __name__ == "__main__":
    main()

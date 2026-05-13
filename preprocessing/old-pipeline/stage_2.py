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
    CATEGORY_MAPPING,
    INTERMEDIATE_DIR,
    NEIGHBOURS_PER_SEED,
    TOP_N_PER_SEED,
    _extract_label,
    load_fasttext,
    setup_dirs,
)

TOKEN_REGEX = re.compile(r"[0-9a-zåäöé]+", re.IGNORECASE)


def _normalize_word(text: str) -> str:
    text = " ".join((text or "").strip().split()).lower()
    text = text.strip(" _-.,;:!?\"'()[]{}")
    return text


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


def _upsert_entry(
    seen: Dict[str, Tuple],
    word: str,
    category: str,
    similarity: float,
    vector: np.ndarray,
    popularity: float,
    sitelinks: float,
    score: float,
    is_seed: bool,
) -> None:
    existing = seen.get(word)
    if existing is None:
        seen[word] = (
            word,
            category,
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
        existing_category,
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

    seen[word] = (
        existing_word,
        existing_category,
        merged_similarity,
        merged_vector,
        max(existing_popularity, float(popularity)),
        max(existing_sitelinks, float(sitelinks)),
        max(existing_score, float(score)),
        merged_is_seed,
    )


def main():
    setup_dirs()
    start_time = time.time()

    print("========================================")
    print(" Starting Stage 2: Vector Lookup")
    print("========================================")
    logging.info("=" * 60)
    logging.info("Stage 2: Vector Lookup")

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

    results = defaultdict(list)
    total_categories = len(seed_entries_per_output)

    for idx, (output_csv, seeds) in enumerate(seed_entries_per_output.items(), 1):
        total_seeds = len(seeds)
        print(f"[{idx}/{total_categories}] Processing {output_csv} ({total_seeds} seeds)...")
        cat_start = time.time()
        seen = {}

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
                seen=seen,
                word=label,
                category=category,
                similarity=1.0,
                vector=avg_vec,
                popularity=popularity,
                sitelinks=sitelinks,
                score=score,
                is_seed=True,
            )

            neighbors = model.wv.similar_by_vector(avg_vec, topn=NEIGHBOURS_PER_SEED)
            kept = 0

            for neighbor_word, similarity in neighbors:
                if kept >= TOP_N_PER_SEED:
                    break

                similarity = float(similarity)
                if similarity <= 0:
                    continue

                normalized_neighbor = _normalize_word(neighbor_word)
                if not normalized_neighbor or normalized_neighbor == label:
                    continue

                _upsert_entry(
                    seen=seen,
                    word=normalized_neighbor,
                    category=category,
                    similarity=similarity,
                    vector=model.wv[neighbor_word],
                    popularity=max(1.0, popularity * similarity * 0.2),
                    sitelinks=sitelinks * similarity * 0.2,
                    score=score * similarity * 0.2,
                    is_seed=False,
                )
                kept += 1

        output_entries = list(seen.values())
        output_entries.sort(
            key=lambda entry: (
                not entry[7],  # seeds first
                -entry[4],     # then popularity
                -entry[2],     # then similarity
                entry[0],
            )
        )
        results[output_csv] = output_entries

        print(
            f"-> Finished {output_csv} in {time.time() - cat_start:.2f}s. "
            f"Kept {len(output_entries)} unique words.\n"
        )

    print("Saving intermediate candidate vectors to disk...")
    save_start = time.time()
    with open(INTERMEDIATE_DIR / "stage2_candidates.pkl", "wb") as f:
        pickle.dump(dict(results), f)
    print(f"-> Saved in {time.time() - save_start:.2f} seconds.")

    total_time = time.time() - start_time
    print("\n========================================")
    print(f" Stage 2 Complete! (Total time: {total_time:.2f}s)")
    print("========================================")
    logging.info(f"Stage 2 complete in {total_time:.2f} seconds.")


if __name__ == "__main__":
    main()

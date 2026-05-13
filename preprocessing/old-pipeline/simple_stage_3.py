import logging
import pickle
import time

import numpy as np
from sklearn.decomposition import PCA

from shared import INTERMEDIATE_DIR, PCA_DIMS


def main():
    print("starting simple stage 3")
    logging.info("=" * 60)
    logging.info("Simple Stage 3: Dimension Reduction")
    logging.info("=" * 60)

    with open(INTERMEDIATE_DIR / "simple_stage2_candidates.pkl", "rb") as f:
        candidates = pickle.load(f)

    def unpack(entry):
        if len(entry) >= 8:
            word, entry_type, _sim, vec, popularity, sitelinks, score, is_seed = entry[:8]
            return word, entry_type, vec, float(popularity), float(sitelinks), float(score), bool(is_seed)
        word, entry_type, _sim, vec = entry
        return word, entry_type, vec, 0.0, 0.0, 0.0, False

    all_vectors = []
    for entries in candidates.values():
        all_vectors.extend([unpack(entry)[2] for entry in entries])

    if not all_vectors:
        logging.warning("No vectors found in simple stage 2 output.")
        return

    matrix = np.array(all_vectors, dtype=np.float32)

    start = time.time()
    n_components = min(PCA_DIMS, len(matrix), 300)
    pca = PCA(n_components=n_components, random_state=42)
    pca.fit(matrix)
    print(f"-> PCA fitted in {time.time() - start:.2f}s")

    reduced = {}
    for output_csv, entries in candidates.items():
        if not entries:
            reduced[output_csv] = []
            continue

        unpacked_entries = [unpack(entry) for entry in entries]
        cat_matrix = np.array([entry[2] for entry in unpacked_entries], dtype=np.float32)
        reduced_matrix = pca.transform(cat_matrix)

        if n_components < PCA_DIMS:
            pad = np.zeros((len(entries), PCA_DIMS - n_components), dtype=np.float32)
            reduced_matrix = np.concatenate([reduced_matrix, pad], axis=1)

        reduced[output_csv] = [
            (
                word,
                entry_type,
                reduced_matrix[i],
                popularity,
                sitelinks,
                score,
                is_seed,
            )
            for i, (word, entry_type, _vec, popularity, sitelinks, score, is_seed) in enumerate(
                unpacked_entries
            )
        ]

    with open(INTERMEDIATE_DIR / "simple_stage3_reduced.pkl", "wb") as f:
        pickle.dump(reduced, f)

    print("simple stage 3 complete")
    logging.info("Simple Stage 3 complete.")


if __name__ == "__main__":
    main()

# stage_4.py
import pickle
import logging
import numpy as np
from sklearn.decomposition import PCA
from shared import INTERMEDIATE_DIR, PCA_DIMS

def main():
    print("starting stage 4")
    logging.info("=" * 60)
    logging.info(f"Stage 4: Dimension Reduction (300 → {PCA_DIMS})")

    # Load from Stage 3
    with open(INTERMEDIATE_DIR / "stage3_validated.pkl", "rb") as f:
        validated = pickle.load(f)

    reduced = {}
    for output_csv, entries in validated.items():
        if not entries:
            reduced[output_csv] = []
            continue

        n_components = min(PCA_DIMS, len(entries), 300)
        matrix = np.array([vec for _, _, _, vec in entries], dtype=np.float32)
        pca = PCA(n_components=n_components, random_state=42)
        reduced_matrix = pca.fit_transform(matrix)

        if n_components < PCA_DIMS:
            pad = np.zeros((len(entries), PCA_DIMS - n_components), dtype=np.float32)
            reduced_matrix = np.concatenate([reduced_matrix, pad], axis=1)

        reduced[output_csv] = [
            (word, cat, reduced_matrix[i])
            for i, (word, cat, _sim, _vec) in enumerate(entries)
        ]

    # Save to disk for Stage 5
    with open(INTERMEDIATE_DIR / "stage4_reduced.pkl", "wb") as f:
        pickle.dump(reduced, f)

    print("stage 4 complete")
    logging.info("Stage 4 complete.")

if __name__ == "__main__":
    main()
import pickle
import logging
import numpy as np
from sklearn.decomposition import PCA
from shared import INTERMEDIATE_DIR, PCA_DIMS

def main():
    print("starting stage 4")
    logging.info("=" * 60)
    logging.info("Stage 4: Dimension Reduction")
    logging.info("=" * 60)
    
    with open(INTERMEDIATE_DIR / "stage3_validated.pkl", "rb") as f:
        validated = pickle.load(f)

    all_vectors = []
    for entries in validated.values():
        all_vectors.extend([vec for _, _, _, vec in entries])
    
    matrix = np.array(all_vectors, dtype=np.float32)
    
    # 2. Fit PCA once on the entire vocabulary
    n_components = min(PCA_DIMS, len(matrix), 300)
    pca = PCA(n_components=n_components, random_state=42)
    pca.fit(matrix) # Global fit

    reduced = {}
    for output_csv, entries in validated.items():
        if not entries:
            reduced[output_csv] = []
            continue

        cat_matrix = np.array([vec for _, _, _, vec in entries], dtype=np.float32)
        # Transform using the global PCA
        reduced_matrix = pca.transform(cat_matrix)

        if n_components < PCA_DIMS:
            pad = np.zeros((len(entries), PCA_DIMS - n_components), dtype=np.float32)
            reduced_matrix = np.concatenate([reduced_matrix, pad], axis=1)

        reduced[output_csv] = [
            (word, cat, reduced_matrix[i])
            for i, (word, cat, _sim, _vec) in enumerate(entries)
        ]

    with open(INTERMEDIATE_DIR / "stage4_reduced.pkl", "wb") as f:
        pickle.dump(reduced, f)

    print("stage 4 complete")
    logging.info("Stage 4 complete.")

if __name__ == "__main__":
    main()
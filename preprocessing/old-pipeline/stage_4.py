import pickle
import logging
import numpy as np
from sklearn.decomposition import PCA
from shared import INTERMEDIATE_DIR, PCA_DIMS

"""
    MENTAL NOTE:

    See if its possible to move parts of this to the GPU.

    CuPy -> pip install cupy-cuda11x

    ```
    from cuml import PCA  # CuML is a drop-in replacement!
    import cupy as cp

    # ... inside main() ...

    # Move the entire matrix to GPU
    matrix_gpu = cp.array(all_vectors, dtype=cp.float32)

    # Fit PCA on the GPU
    n_components = min(PCA_DIMS, len(matrix_gpu), 300)
    pca = PCA(n_components=n_components, random_state=42)
    pca.fit(matrix_gpu)

    # ... inside your loop ...
    cat_matrix_gpu = cp.array([vec for _, _, _, vec in entries], dtype=cp.float32)

    # Transform on the GPU and pull back to CPU
    reduced_matrix_gpu = pca.transform(cat_matrix_gpu)
    reduced_matrix = reduced_matrix_gpu.get() # .get() moves it back to numpy array
    ```
"""

def main():
    print("starting stage 4")
    logging.info("=" * 60)
    logging.info("Stage 4: Dimension Reduction")
    logging.info("=" * 60)
    
    with open(INTERMEDIATE_DIR / "stage3_validated.pkl", "rb") as f:
        validated = pickle.load(f)

    def unpack(entry):
        if len(entry) >= 8:
            word, cat, _sim, vec, popularity, sitelinks, score, is_seed = entry[:8]
            return word, cat, vec, float(popularity), float(sitelinks), float(score), bool(is_seed)
        word, cat, _sim, vec = entry
        return word, cat, vec, 0.0, 0.0, 0.0, False

    all_vectors = []
    for entries in validated.values():
        all_vectors.extend([unpack(entry)[2] for entry in entries])
    
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

        unpacked_entries = [unpack(entry) for entry in entries]
        cat_matrix = np.array([entry[2] for entry in unpacked_entries], dtype=np.float32)
        # Transform using the global PCA
        reduced_matrix = pca.transform(cat_matrix)

        if n_components < PCA_DIMS:
            pad = np.zeros((len(entries), PCA_DIMS - n_components), dtype=np.float32)
            reduced_matrix = np.concatenate([reduced_matrix, pad], axis=1)

        reduced[output_csv] = [
            (
                word,
                cat,
                reduced_matrix[i],
                popularity,
                sitelinks,
                score,
                is_seed,
            )
            for i, (word, cat, _vec, popularity, sitelinks, score, is_seed) in enumerate(
                unpacked_entries
            )
        ]

    with open(INTERMEDIATE_DIR / "stage4_reduced.pkl", "wb") as f:
        pickle.dump(reduced, f)

    print("stage 4 complete")
    logging.info("Stage 4 complete.")

if __name__ == "__main__":
    main()

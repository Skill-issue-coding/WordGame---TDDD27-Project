import csv
import pickle
import numpy as np
from collections import defaultdict
import logging
import time
from shared import BASE_DIR, INTERMEDIATE_DIR, CATEGORY_MAPPING, NEIGHBOURS_PER_SEED, TOP_N_PER_SEED, load_fasttext, _extract_label, setup_dirs

"""
    MENTAL NOTE:

    See if its possible to move parts of this to the GPU.

    CuPy -> https://docs.cupy.dev/en/stable/install.html

    ```
    import cupy as cp # Import CuPy

    # --- START OF GPU VECTORIZED MATH ---
    # 1. Transfer vectors from CPU RAM to GPU VRAM
    cand_matrix = cp.array([model.get_word_vector(w) for w in candidates_list])
    avg_vec_gpu = cp.array(avg_vec)

    # 2. Calculate norms on GPU
    cand_norms = cp.linalg.norm(cand_matrix, axis=1, keepdims=True)
    cand_norms[cand_norms == 0] = 1 
    normalized_cands = cand_matrix / cand_norms

    avg_vec_norm = cp.linalg.norm(avg_vec_gpu)
    if avg_vec_norm != 0: 
        normalized_avg = avg_vec_gpu / avg_vec_norm

        # 3. Blazing fast GPU dot product
        similarities_gpu = cp.dot(normalized_cands, normalized_avg)

        # 4. Transfer the result back to CPU RAM for sorting
        similarities = similarities_gpu.get()
    else:
        similarities = np.zeros(len(candidates_list))

    # --- END OF GPU VECTORIZED MATH ---
    ```
"""

def main():
    setup_dirs()
    start_time = time.time()
    
    print("========================================")
    print(" Starting Stage 2: Vector Lookup")
    print("========================================")
    logging.info("=" * 60)
    logging.info("Stage 2: Vector Lookup (Vectorized)")
    
    # 1. Reconstruct seed_data from Stage 1 CSVs
    seed_data = defaultdict(list)
    seeding_output_dir = BASE_DIR / "seeding" / "output"
    for csv_path in seeding_output_dir.glob("*.csv"):
        query_name = csv_path.stem
        with csv_path.open("r", encoding="utf-8") as f:
            seed_data[query_name] = list(csv.DictReader(f))
            
    print(f"Loaded {len(seed_data)} seed files.")

    # 2. Setup Vector mapping
    print("Loading fastText model... (This usually takes 10-20 seconds)")
    model_start = time.time()
    model = load_fasttext()
    print(f"-> Model loaded in {time.time() - model_start:.2f} seconds!\n")
    
    per_output = defaultdict(list)
    for query_name, rows in seed_data.items():
        mapping = CATEGORY_MAPPING.get(query_name)
        if not mapping: continue
        category, output_csv = mapping
        for row in rows:
            if label := _extract_label(row):
                per_output[output_csv].append((label, category))

    results = defaultdict(list)
    total_categories = len(per_output)
    
    for idx, (output_csv, seeds) in enumerate(per_output.items(), 1):
        total_seeds = len(seeds)
        print(f"[{idx}/{total_categories}] Processing {output_csv} ({total_seeds} seeds)...")
        cat_start = time.time()
        seen = {}
        
        for seed_idx, (label, category) in enumerate(seeds, 1):
            if seed_idx % 100 == 0:
                print(f"  ... processed {seed_idx}/{total_seeds} seeds")
                
            tokens = [t for t in label.lower().split() if t.isalpha()]
            if not tokens: continue
            
            candidates = set()
            for t in tokens:
                try: 
                    # Gensim's most_similar returns a list of (word, similarity) tuples
                    neighbors = model.wv.most_similar(positive=[t], topn=NEIGHBOURS_PER_SEED)
                    candidates.update(w for w, _ in neighbors)
                except KeyError: 
                    pass # Token not in vocabulary
                
            candidates_list = list(candidates)
            if not candidates_list:
                continue
                
            # Gensim accesses vectors using model.wv[word]
            valid_tokens = [t for t in tokens if t in model.wv]
            if not valid_tokens:
                continue
                
            avg_vec = np.mean([model.wv[t] for t in valid_tokens], axis=0)

            # --- START OF VECTORIZED MATH ---
            cand_matrix = np.array([model.wv[w] for w in candidates_list])

            cand_norms = np.linalg.norm(cand_matrix, axis=1, keepdims=True)
            cand_norms[cand_norms == 0] = 1 # Prevent division by zero
            normalized_cands = cand_matrix / cand_norms

            avg_vec_norm = np.linalg.norm(avg_vec)
            if avg_vec_norm == 0: 
                continue
            normalized_avg = avg_vec / avg_vec_norm

            similarities = np.dot(normalized_cands, normalized_avg)

            scored = sorted(zip(similarities, candidates_list), reverse=True)[:NEIGHBOURS_PER_SEED]
            # --- END OF VECTORIZED MATH ---
            
            kept = 0
            for sim, word in scored:
                if kept >= TOP_N_PER_SEED: break
                wl = word.lower()
                if wl not in seen or sim > seen[wl]:
                    seen[wl] = sim
                    kept += 1

        print(f"-> Finished {output_csv} in {time.time() - cat_start:.2f}s. Kept {len(seen)} unique candidate words.\n")

        for word, sim in seen.items():
            cat = seeds[0][1] if seeds else "unknown"
            # Append the 300-dimension vector using Gensim's API
            results[output_csv].append((word, cat, sim, model.wv[word]))

    # 3. Save to disk for Stage 3
    print("Saving intermediate candidate vectors to disk...")
    save_start = time.time()
    with open(INTERMEDIATE_DIR / "stage2_candidates.pkl", "wb") as f:
        pickle.dump(dict(results), f)
    print(f"-> Saved in {time.time() - save_start:.2f} seconds.")
        
    total_time = time.time() - start_time
    print(f"\n========================================")
    print(f" Stage 2 Complete! (Total time: {total_time:.2f}s)")
    print(f"========================================")
    logging.info(f"Stage 2 complete in {total_time:.2f} seconds.")

if __name__ == "__main__":
    main()
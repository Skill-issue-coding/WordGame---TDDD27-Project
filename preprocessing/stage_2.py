# stage_2.py
import csv
import pickle
import numpy as np
from collections import defaultdict
import logging
from shared import BASE_DIR, INTERMEDIATE_DIR, CATEGORY_MAPPING, NEIGHBOURS_PER_SEED, TOP_N_PER_SEED, load_fasttext, _extract_label, setup_dirs


def main():
    setup_dirs()
    print("starting stage 2")
    logging.info("=" * 60)
    logging.info("Stage 2: Vector Lookup (Vectorized)")
    logging.info("=" * 60)
    
    # 1. Reconstruct seed_data from Stage 1 CSVs
    seed_data = defaultdict(list)
    seeding_output_dir = BASE_DIR / "seeding" / "output"
    for csv_path in seeding_output_dir.glob("*.csv"):
        query_name = csv_path.stem
        with csv_path.open("r", encoding="utf-8") as f:
            seed_data[query_name] = list(csv.DictReader(f))

    # 2. Setup Vector mapping
    model = load_fasttext()
    per_output = defaultdict(list)
    for query_name, rows in seed_data.items():
        mapping = CATEGORY_MAPPING.get(query_name)
        if not mapping: continue
        category, output_csv = mapping
        for row in rows:
            if label := _extract_label(row):
                per_output[output_csv].append((label, category))

    results = defaultdict(list)
    for output_csv, seeds in per_output.items():
        seen = {}
        for label, category in seeds:
            tokens = [t for t in label.lower().split() if t.isalpha()]
            if not tokens: continue
            
            candidates = set()
            for t in tokens:
                try: candidates.update(w for _, w in model.get_nearest_neighbors(t, k=NEIGHBOURS_PER_SEED))
                except Exception: pass
                
            candidates_list = list(candidates)
            if not candidates_list:
                continue
                
            avg_vec = np.mean([model.get_word_vector(t) for t in tokens], axis=0)
            
            cand_matrix = np.array([model.get_word_vector(w) for w in candidates_list])

            cand_norms = np.linalg.norm(cand_matrix, axis=1, keepdims=True)
            cand_norms[cand_norms == 0] = 1 # Prevent division by zero
            normalized_cands = cand_matrix / cand_norms

            avg_vec_norm = np.linalg.norm(avg_vec)
            if avg_vec_norm == 0: 
                continue
            normalized_avg = avg_vec / avg_vec_norm

            similarities = np.dot(normalized_cands, normalized_avg)

            scored = sorted(zip(similarities, candidates_list), reverse=True)[:NEIGHBOURS_PER_SEED]
            
            kept = 0
            for sim, word in scored:
                if kept >= TOP_N_PER_SEED: break
                wl = word.lower()
                if wl not in seen or sim > seen[wl]:
                    seen[wl] = sim
                    kept += 1

        for word, sim in seen.items():
            cat = seeds[0][1] if seeds else "unknown"
            results[output_csv].append((word, cat, sim, model.get_word_vector(word)))

    # 3. Save to disk for Stage 3
    with open(INTERMEDIATE_DIR / "stage2_candidates.pkl", "wb") as f:
        pickle.dump(dict(results), f)
        
    print("stage 2 complete")
    logging.info("Stage 2 complete.")

if __name__ == "__main__":
    main()
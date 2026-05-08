# stage_2.py
import csv
import pickle
import numpy as np
from collections import defaultdict
import logging
from shared import BASE_DIR, INTERMEDIATE_DIR, CATEGORY_MAPPING, NEIGHBOURS_PER_SEED, TOP_N_PER_SEED, load_fasttext, _extract_label, setup_dirs

def _cosine_sim(a: np.ndarray, b: np.ndarray) -> float:
    na, nb = np.linalg.norm(a), np.linalg.norm(b)
    return float(np.dot(a, b) / (na * nb)) if na and nb else 0.0

def main():
    setup_dirs()
    print("starting stage 2")
    logging.info("=" * 60)
    logging.info("Stage 2: Vector Lookup")
    
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
            # Average pooling logic inline for brevity
            tokens = [t for t in label.lower().split() if t.isalpha()]
            if not tokens: continue
            
            candidates = set()
            for t in tokens:
                try: candidates.update(w for _, w in model.get_nearest_neighbors(t, k=NEIGHBOURS_PER_SEED))
                except Exception: pass
                
            avg_vec = np.mean([model.get_word_vector(t) for t in tokens], axis=0)
            scored = sorted([(_cosine_sim(avg_vec, model.get_word_vector(w)), w) for w in candidates], reverse=True)[:NEIGHBOURS_PER_SEED]
            
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
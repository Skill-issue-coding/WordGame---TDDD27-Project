import csv
import pickle
import logging
from shared import INTERMEDIATE_DIR, OUTPUT_DIR, PCA_DIMS, setup_dirs

def main():
    setup_dirs()
    print("starting stage 5")
    logging.info("=" * 60)
    logging.info(f"Stage 5: CSV Export → {OUTPUT_DIR}")

    # Load from Stage 4
    with open(INTERMEDIATE_DIR / "stage4_reduced.pkl", "rb") as f:
        reduced = pickle.load(f)

    dim_headers = [f"v{i}" for i in range(PCA_DIMS)]
    fieldnames = ["word", "category"] + dim_headers

    for output_csv, entries in reduced.items():
        if not entries: continue
        path = OUTPUT_DIR / output_csv
        with path.open("w", encoding="utf-8", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=fieldnames)
            writer.writeheader()
            for word, category, vec in entries:
                row = {"word": word, "category": category}
                for i in range(PCA_DIMS):
                    row[f"v{i}"] = f"{float(vec[i]):.6f}" if i < len(vec) else "0.000000"
                writer.writerow(row)

    print(f"Pipeline complete! ✓ Output files in: {OUTPUT_DIR}")
    logging.info("Stage 5 complete.")

if __name__ == "__main__":
    main()
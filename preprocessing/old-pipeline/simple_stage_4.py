import csv
import logging
import pickle

from shared import BASE_DIR, INTERMEDIATE_DIR, PCA_DIMS, setup_dirs

OUTPUT_DIR_SIMPLE = BASE_DIR.parent / "server" / "wordfiles_simple"


def main():
    setup_dirs()
    OUTPUT_DIR_SIMPLE.mkdir(parents=True, exist_ok=True)

    print("starting simple stage 4")
    logging.info("=" * 60)
    logging.info(f"Simple Stage 4: CSV Export -> {OUTPUT_DIR_SIMPLE}")
    logging.info("=" * 60)

    with open(INTERMEDIATE_DIR / "simple_stage3_reduced.pkl", "rb") as f:
        reduced = pickle.load(f)

    dim_headers = [f"v{i}" for i in range(PCA_DIMS)]
    fieldnames = ["word", "category", "popularity", "sitelinks", "score", "is_seed"] + dim_headers

    for output_csv, entries in reduced.items():
        if not entries:
            continue
        path = OUTPUT_DIR_SIMPLE / output_csv
        with path.open("w", encoding="utf-8", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=fieldnames)
            writer.writeheader()
            for entry in entries:
                if len(entry) >= 7:
                    word, category, vec, popularity, sitelinks, score, is_seed = entry[:7]
                else:
                    word, category, vec = entry[:3]
                    popularity, sitelinks, score, is_seed = 0.0, 0.0, 0.0, False

                row = {
                    "word": word,
                    "category": category,
                    "popularity": f"{float(popularity):.6f}",
                    "sitelinks": f"{float(sitelinks):.6f}",
                    "score": f"{float(score):.6f}",
                    "is_seed": str(bool(is_seed)).lower(),
                }
                for i in range(PCA_DIMS):
                    row[f"v{i}"] = f"{float(vec[i]):.6f}" if i < len(vec) else "0.000000"
                writer.writerow(row)

    print(f"Simple pipeline complete! Output files in: {OUTPUT_DIR_SIMPLE}")
    logging.info("Simple Stage 4 complete.")


if __name__ == "__main__":
    main()

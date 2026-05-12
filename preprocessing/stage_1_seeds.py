import logging
from shared import BASE_DIR, setup_dirs
from seeding import query_runner
import csv

def process_maktbarometern():
    """Reads Maktbarometern CSVs, keeps top-ranked entities, and preserves scoring metadata."""
    
    makt_dir = BASE_DIR / "seeding" / "maktbarometern" / "csv"
    output_dir = BASE_DIR / "seeding" / "output"
    out_file = output_dir / "maktbarometern.csv"
    
    if not makt_dir.exists():
        logging.warning(f"Maktbarometern dir not found: {makt_dir}")
        return
        
    aggregated = {}
    
    for csv_path in makt_dir.glob("*.csv"):
        with open(csv_path, "r", encoding="utf-8") as f:
            reader = csv.DictReader(f)
            
            # Catch cases where the CSV delimiter might be a semicolon instead of comma
            if not reader.fieldnames or 'rank' not in reader.fieldnames:
                f.seek(0)
                reader = csv.DictReader(f, delimiter=';')

            for row in reader:
                try:
                    # FILTER 1: Only take the top 100 from each platform
                    rank = int(float(row["rank"]))
                    if rank > 100:
                        continue
                except (ValueError, KeyError):
                    continue  # Skip malformed rows

                name = " ".join(row.get("name", "").strip().split())
                if not name:
                    continue

                try:
                    score = float(row.get("score", 0) or 0)
                except ValueError:
                    score = 0.0

                norm = name.lower()
                row_rank_signal = max(0.0, 101.0 - float(rank))

                if norm not in aggregated:
                    aggregated[norm] = {
                        "wordLabel": name,
                        "score": 0.0,
                        "rank": rank,
                        # Synthetic popularity axis so Maktbarometern can share weighting logic
                        # with sitelinks-based Wikidata entities in Stage 2.
                        "sitelinks": 0.0,
                        "sources": 0,
                    }

                agg = aggregated[norm]
                agg["score"] += score
                agg["rank"] = min(int(agg["rank"]), rank)
                agg["sitelinks"] += row_rank_signal
                agg["sources"] = int(agg["sources"]) + 1

    if aggregated:
        formatted_rows = sorted(
            aggregated.values(),
            key=lambda row: (float(row["score"]) + float(row["sitelinks"])),
            reverse=True,
        )

        output_dir.mkdir(parents=True, exist_ok=True)
        with open(out_file, "w", encoding="utf-8", newline="") as f:
            writer = csv.DictWriter(
                f,
                fieldnames=["wordLabel", "score", "rank", "sitelinks", "sources"],
            )
            writer.writeheader()
            writer.writerows(formatted_rows)
            
        logging.info(
            f"Processed Maktbarometern: Saved {len(formatted_rows)} unique top influencers."
        )
        print(f"  Processed Maktbarometern: {len(formatted_rows)} unique entities.")

def main():
    setup_dirs()
    seeding_output_dir = BASE_DIR / "seeding" / "output"
    
    print("starting stage 1")
    logging.info("=" * 60)
    logging.info("\nStage 1: SPARQL Seeding")
    logging.info("=" * 60)
    
    query_runner.run_all_and_save(
        queries=query_runner.QUERIES,
        output_dir=seeding_output_dir,
    )

    print("\nProcessing Maktbarometern data...")
    process_maktbarometern()

    print("stage 1 complete")
    logging.info("Stage 1 complete.")

if __name__ == "__main__":
    main()

import logging
from shared import BASE_DIR, setup_dirs
from seeding import query_runner
import csv
from pathlib import Path

def process_maktbarometern():
    """Reads Maktbarometern CSVs, filters by rank, and formats for Stage 2."""
    from shared import BASE_DIR # Ensure BASE_DIR is accessible
    
    makt_dir = BASE_DIR / "seeding" / "maktbarometern" / "csv"
    output_dir = BASE_DIR / "seeding" / "output"
    out_file = output_dir / "maktbarometern.csv"
    
    if not makt_dir.exists():
        logging.warning(f"Maktbarometern dir not found: {makt_dir}")
        return
        
    seen_names = set()
    formatted_rows = []
    
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
                    if int(row["rank"]) > 100:
                        continue
                except (ValueError, KeyError):
                    continue # Skip malformed rows
                
                name = row.get("name", "").strip()
                
                # FILTER 2: Deduplicate and ignore empty
                if not name or name.lower() in seen_names:
                    continue
                    
                seen_names.add(name.lower())
                
                # Format to match Wikidata outputs
                formatted_rows.append({"wordLabel": name})
                
    if formatted_rows:
        output_dir.mkdir(parents=True, exist_ok=True)
        with open(out_file, "w", encoding="utf-8", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=["wordLabel"])
            writer.writeheader()
            writer.writerows(formatted_rows)
            
        logging.info(f"Processed Maktbarometern: Saved {len(formatted_rows)} unique top influencers.")
        print(f"  Processed Maktbarometern: {len(formatted_rows)} unique entities.")

def main():
    setup_dirs()
    seeding_output_dir = BASE_DIR / "seeding" / "output"
    
    print("starting stage 1")
    logging.info("=" * 60)
    logging.info("\nStage 1: SPARQL Seeding")
    logging.info("=" * 60)
    
    # query_runner.run_all_and_save(
    #     queries=query_runner.QUERIES,
    #     output_dir=seeding_output_dir,
    # )

    print("\nProcessing Maktbarometern data...")
    process_maktbarometern()

    print("stage 1 complete")
    logging.info("Stage 1 complete.")

if __name__ == "__main__":
    main()
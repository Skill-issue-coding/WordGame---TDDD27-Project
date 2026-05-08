import logging
from shared import BASE_DIR, setup_dirs
from seeding import query_runner

def main():
    setup_dirs()
    seeding_output_dir = BASE_DIR / "seeding" / "output"
    
    print("starting stage 1")
    logging.info("=" * 60)
    logging.info("Stage 1: SPARQL Seeding")
    logging.info("=" * 60)
    
    query_runner.run_all_and_save(
        queries=query_runner.QUERIES,
        output_dir=seeding_output_dir,
    )
    
    print("stage 1 complete")
    logging.info("Stage 1 complete.")

if __name__ == "__main__":
    main()
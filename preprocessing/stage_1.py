import logging
from pathlib import Path
from shared import BASE_DIR, setup_dirs
from seeding import query_runner
from seeding import clean_seeding

def _setup_logger() -> logging.Logger:
    log_path = Path(__file__).resolve().parent / "pipeline.log"
    root = logging.getLogger()
    if not any(
        isinstance(h, logging.FileHandler) and h.baseFilename == str(log_path)
        for h in root.handlers
    ):
        handler = logging.FileHandler(log_path, encoding="utf-8")
        handler.setLevel(logging.INFO)
        handler.setFormatter(logging.Formatter("%(asctime)s [%(levelname)s] %(message)s"))
        root.addHandler(handler)
        root.setLevel(logging.INFO)
    return logging.getLogger(__name__)

log = _setup_logger()

def main():
    setup_dirs()
    seeding_output_dir = BASE_DIR / "seeding" / "output"
    
    print("starting stage 1")
    log.info("=" * 60)
    log.info("Stage 1: SPARQL Seeding")
    log.info(f"Output dir: {seeding_output_dir}")
    log.info("=" * 60)
    
    query_runner.run_all_and_save(queries=query_runner.QUERIES, output_dir=seeding_output_dir)

    print("Rensar seeding-data (löser QID-etiketter, tar bort ogiltiga poster)…")
    log.info("Stage 1: running clean_seeding")
    clean_seeding.process_seeding()
    clean_seeding.process_maktbarometern()

    print("stage 1 complete")
    log.info("Stage 1 complete.")

if __name__ == "__main__":
    main()

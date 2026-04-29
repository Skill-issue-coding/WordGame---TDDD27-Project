import logging
from pathlib import Path
from typing import Dict, Iterable, List
from termcolor import colored
from .query import Query

LOG_FILE = Path(__file__).resolve().parent / "seeding_pipeline.log"
logging.basicConfig(
    filename=str(LOG_FILE),
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    datefmt="%Y-%m-%d %H:%M:%S"
)

BASE_DIR = Path(__file__).resolve().parent / "queries"

QUERY_SPECS = [
    ("swedish_celebrities", "swedish_celebrities.sparql"),
    ("swedish_companies", "swedish_companies.sparql"),
    ("global_brands", "global_brands.sparql"),
    ("video_games", "video_games.sparql"),
    ("swedish_tv_and_film", "swedish_tv_and_film.sparql"),
    ("swedish_music", "swedish_music.sparql"),
    ("swedish_food", "swedish_food.sparql"),             
    ("swedish_characters", "swedish_characters.sparql"),
    ("swedish_geography", "swedish_geography.sparql"),  
]

QUERIES_BY_NAME: Dict[str, Query] = {
    name: Query(name=name, query_filepath=str(BASE_DIR / filename))
    for name, filename in QUERY_SPECS
}

DEFAULT_QUERY_NAMES = [name for name, _ in QUERY_SPECS]
QUERIES: List[Query] = [QUERIES_BY_NAME[name] for name in DEFAULT_QUERY_NAMES]

def get_query(name: str) -> Query:
    try:
        return QUERIES_BY_NAME[name]
    except KeyError as exc:
        available = ", ".join(sorted(QUERIES_BY_NAME))
        raise KeyError(colored("Unknown query ", "red"),f"'{name}'. Available: {available}") from exc

def run_all_and_save(queries: Iterable[Query], output_dir: Path) -> Dict[str, List[Dict[str, str]]]:
    """Runs a batch of queries and saves them, logging successes and failures."""
    per_query_rows: Dict[str, List[Dict[str, str]]] = {}
    failed_queries: List[str] = []
    
    logging.info("=== Starting New Seeding Batch ===")
    
    for query in queries:
        print(f"Executing: {query.name}...")
        logging.info(f"Executing query: {query.name}")
        
        try:
            rows = query.run_and_save(output_dir)
            per_query_rows[query.name] = rows
            
            success_msg = f"Success: {query.name} ({len(rows)} rows saved)"
            print(success_msg)
            logging.info(success_msg)
            
        except Exception as exc:
            error_msg = f"FAILED: {query.name} - {str(exc)}"
            print(error_msg)
            logging.error(error_msg)
            failed_queries.append(query.name)
            
    print("\n--- SEEDING SUMMARY ---")
    if failed_queries:
        print(f"Completed with {len(failed_queries)} failures.")
        print(f"Failed queries: {', '.join(failed_queries)}")
        print(f"Check {LOG_FILE.name} for details.")
        logging.warning(f"Batch completed with failures: {failed_queries}")
    else:
        print("All queries executed successfully!")
        logging.info("Batch completed successfully with no errors.")
        
    return per_query_rows
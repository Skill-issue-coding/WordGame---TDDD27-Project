from pathlib import Path
from dotenv import load_dotenv

base_dir = Path(__file__).resolve().parent

env_path = base_dir / ".env.local"
load_dotenv(dotenv_path=env_path)

from seeding import query_runner

def pipeline():
    seeding_output_dir = base_dir / "seeding" / "output"
    seeding_output_dir.mkdir(parents=True, exist_ok=True)

    # --- SEEDS ---
    print("Starting Stage 1: Seeding SPARQL Queries...")
    
    seed_data = query_runner.run_all_and_save(
        queries=query_runner.QUERIES, 
        output_dir=seeding_output_dir
    )
    
    print("Stage 1 Complete. Seeds generated.")
    return seed_data

if __name__ == "__main__":
    pipeline()
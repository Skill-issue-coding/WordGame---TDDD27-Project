import os
import glob
import pandas as pd
import requests
import time

# === Configuration ===
# Look for the cleaned seeding files first, fallback to the raw output
INPUT_DIRS = ["seeding_cleaned", "seeding/output_cleaned", "seeding/output"]
OUTPUT_DIR = "intermediate/stage2_wiki"

def get_input_dir():
    for d in INPUT_DIRS:
        if os.path.exists(d) and glob.glob(os.path.join(d, "*.csv")):
            return d
    return None

def fetch_wikipedia_summaries(entities):
    """
    Fetches the introductory paragraph from Swedish Wikipedia for a list of entities.
    Uses a Session for connection pooling (much faster).
    """
    session = requests.Session()
    session.headers.update({"User-Agent": "WordGameBot/1.0 (WordGame TDDD27)"})
    
    url = "https://sv.wikipedia.org/w/api.php"
    summaries = {}
    
    print(f"Börjar hämta {len(entities)} sammanfattningar från sv.wikipedia.org...")
    
    for i, entity in enumerate(entities):
        if not isinstance(entity, str) or not entity.strip():
            continue
            
        params = {
            "action": "query",
            "format": "json",
            "titles": entity,
            "prop": "extracts",
            "exintro": True,      # Only get the intro paragraph
            "explaintext": True,  # Return plain text, not HTML
            "redirects": 1        # Automatically resolve redirects (e.g., "Zlatan" -> "Zlatan Ibrahimović")
        }
        
        try:
            resp = session.get(url, params=params, timeout=5).json()
            pages = resp.get("query", {}).get("pages", {})
            
            # The API returns a dict with the Page ID as the key. 
            # If the page doesn't exist, the key is "-1".
            for page_id, page_data in pages.items():
                if page_id != "-1" and "extract" in page_data:
                    summaries[entity] = page_data["extract"].strip()
                else:
                    summaries[entity] = "" # Not found on Wiki
                    
        except Exception as e:
            print(f"Fel vid hämtning av '{entity}': {e}")
            summaries[entity] = ""
            
        # Print progress every 100 entities
        if (i + 1) % 100 == 0:
            print(f" -> Hämtat {i + 1}/{len(entities)}...")
            
        # Polite delay to avoid getting IP-banned by Wikimedia
        time.sleep(0.05) 
        
    return summaries

def main():
    input_dir = get_input_dir()
    if not input_dir:
        print("Kunde inte hitta någon mapp med seeding-filer.")
        return
        
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    csv_files = glob.glob(os.path.join(input_dir, "*.csv"))
    
    print("=" * 50)
    print("STAGE 2: WIKIPEDIA CONTEXT ENRICHMENT")
    print("=" * 50)
    
    for file in csv_files:
        filename = os.path.basename(file)
        
        # Skip if we've already processed this file (useful if the script crashes halfway)
        output_path = os.path.join(OUTPUT_DIR, filename)
        if os.path.exists(output_path):
            print(f"Hoppar över {filename} (redan bearbetad).")
            continue
            
        df = pd.read_csv(file)
        
        # Find the column that contains the actual entity names
        label_col = next((col for col in df.columns if col.endswith("Label")), None)
        if not label_col:
            label_col = 'name' if 'name' in df.columns else ('word' if 'word' in df.columns else None)
            
        if not label_col:
            print(f"Varning: Hittade ingen namn-kolumn i {filename}. Hoppar över.")
            continue
            
        # Extract unique entities to fetch
        unique_entities = df[label_col].dropna().unique()
        print(f"\nFil: {filename} | Hittade {len(unique_entities)} unika entiteter.")
        
        # Fetch the summaries
        summaries_dict = fetch_wikipedia_summaries(unique_entities)
        
        # Map the summaries back to the dataframe
        df['wiki_summary'] = df[label_col].map(summaries_dict)
        
        # Calculate success rate
        found_count = (df['wiki_summary'] != "").sum()
        print(f"Klar med {filename}. Hittade wiki-text för {found_count}/{len(df)} entiteter.")
        
        # Save to intermediate/stage2_wiki/
        df.to_csv(output_path, index=False)

if __name__ == "__main__":
    main()
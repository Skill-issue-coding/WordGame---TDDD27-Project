import os
import glob
import pandas as pd
import requests
import time
import re

# ===Configuration ===
INPUT_DIR = "intermediate/stage2_wiki"
OUTPUT_DIR = "intermediate/stage3_attrs"

# The specific Wikidata properties we care about for the game
TARGET_PROPERTIES = {
    "P31": "Typ",         # Instance of (e.g., video game, tech company, human)
    "P106": "Yrke",       # Occupation (e.g., actor, singer)
    "P136": "Genre",      # Genre (music, games, film)
    "P452": "Bransch",    # Industry (for companies)
    "P178": "Utvecklare", # Developer (for games)
    "P641": "Sport",      # Sport (for athletes)
}

def extract_qid(val):
    """Safely extracts a Q-ID from a URL or string."""
    val_str = str(val).strip()
    match = re.search(r"(Q\d+)$", val_str)
    return match.group(1) if match else None

def fetch_wikidata_claims(qids, session):
    """Step 1: Fetch the raw claims (P-values) for our main entities."""
    claims_data = {}
    qids = list(set(qids))
    
    for i in range(0, len(qids), 50):
        chunk = qids[i:i+50]
        url = f"https://www.wikidata.org/w/api.php?action=wbgetentities&ids={'|'.join(chunk)}&props=claims&format=json"
        
        try:
            resp = session.get(url, timeout=10).json()
            if 'entities' in resp:
                for qid, data in resp['entities'].items():
                    if 'claims' in data:
                        claims_data[qid] = data['claims']
        except Exception as e:
            print(f"  [!] Nätverksfel vid hämtning av claims: {e}")
            
        time.sleep(0.05) # Rate limit protection
    return claims_data

def fetch_wikidata_labels(qids, session):
    """Step 2: Translate the raw property Q-IDs into readable Swedish words."""
    labels = {}
    qids = list(set(qids))
    
    for i in range(0, len(qids), 50):
        chunk = qids[i:i+50]
        url = f"https://www.wikidata.org/w/api.php?action=wbgetentities&ids={'|'.join(chunk)}&props=labels&languages=sv|en&format=json"
        
        try:
            resp = session.get(url, timeout=10).json()
            if 'entities' in resp:
                for qid, data in resp['entities'].items():
                    if 'labels' in data:
                        # Prefer Swedish, fallback to English
                        if 'sv' in data['labels']:
                            labels[qid] = data['labels']['sv']['value']
                        elif 'en' in data['labels']:
                            labels[qid] = data['labels']['en']['value']
        except Exception as e:
            pass
        time.sleep(0.05)
    return labels

def main():
    if not os.path.exists(INPUT_DIR):
        print(f"Fel: Hittar inte {INPUT_DIR}. Kör stage 2 först!")
        return
        
    os.makedirs(OUTPUT_DIR, exist_ok=True)
    csv_files = glob.glob(os.path.join(INPUT_DIR, "*.csv"))
    
    session = requests.Session()
    session.headers.update({"User-Agent": "WordGameBot/1.0 (WordGame TDDD27)"})

    print("=" * 50)
    print("STAGE 3: WIKIDATA ATTRIBUTE ENRICHMENT")
    print("=" * 50)

    for file in csv_files:
        filename = os.path.basename(file)
        output_path = os.path.join(OUTPUT_DIR, filename)
        
        if os.path.exists(output_path):
            print(f"Hoppar över {filename} (redan bearbetad).")
            continue
            
        df = pd.read_csv(file)
        
        # 1. Find the column containing the Wikidata URI/Q-ID (usually named 'person', 'game', 'company')
        qid_col = next(
            (col for col in df.columns
             if df[col].dropna().astype(str).str.contains("wikidata.org/entity/", na=False).any()),
            None,
        )
        
        # If no Q-ID column exists (like in Maktbarometern), just pass the file through unchanged
        if not qid_col:
            print(f"Fil: {filename} | Inga Wikidata-länkar hittades (troligen sociala medier). Passerar igenom.")
            df['wiki_attributes'] = ""
            df.to_csv(output_path, index=False)
            continue
            
        print(f"\nFil: {filename} | Hämtar attribut...")
        
        # Extract clean Q-IDs
        df['clean_qid'] = df[qid_col].apply(extract_qid)
        valid_qids = df['clean_qid'].dropna().unique().tolist()
        
        # 2. Fetch all claims
        claims_data = fetch_wikidata_claims(valid_qids, session)
        
        # 3. Harvest all the "Value Q-IDs" we need to translate
        needed_label_qids = set()
        for qid, claims in claims_data.items():
            for prop_id in TARGET_PROPERTIES.keys():
                if prop_id in claims:
                    for statement in claims[prop_id]:
                        try:
                            val_qid = statement['mainsnak']['datavalue']['value']['id']
                            needed_label_qids.add(val_qid)
                        except KeyError:
                            pass
                            
        # 4. Fetch the Swedish translations for those Q-IDs
        print(f" -> Översätter {len(needed_label_qids)} unika egenskaper till svenska...")
        property_labels = fetch_wikidata_labels(list(needed_label_qids), session)
        
        # 5. Assemble the final text strings
        attributes_list = []
        for index, row in df.iterrows():
            qid = row['clean_qid']
            if not qid or qid not in claims_data:
                attributes_list.append("")
                continue
                
            entity_claims = claims_data[qid]
            entity_text_parts = []
            
            for prop_id, prop_name in TARGET_PROPERTIES.items():
                if prop_id in entity_claims:
                    # Get all values for this property (e.g., a person might be both Actor and Singer)
                    vals = []
                    for statement in entity_claims[prop_id]:
                        try:
                            val_qid = statement['mainsnak']['datavalue']['value']['id']
                            if val_qid in property_labels:
                                vals.append(property_labels[val_qid])
                        except KeyError:
                            pass
                            
                    if vals:
                        # Construct "Yrke: skådespelare, sångare"
                        entity_text_parts.append(f"{prop_name}: {', '.join(vals)}.")
            
            attributes_list.append(" ".join(entity_text_parts))
            
        df['wiki_attributes'] = attributes_list
        df = df.drop(columns=['clean_qid'])
        
        found_count = sum(1 for a in attributes_list if a)
        print(f"✅ Klar! Hittade strukturerad data för {found_count}/{len(df)} entiteter.")
        df.to_csv(output_path, index=False)

if __name__ == "__main__":
    main()
import os
import glob
import time
import pandas as pd
import numpy as np
from sentence_transformers import SentenceTransformer
from sklearn.metrics.pairwise import cosine_similarity

try:
    from shared import load_kelly, load_korp_csvs, load_custom_stopwords
    HAS_SHARED = True
except ImportError as e:
    HAS_SHARED = False
    print(f"Kunde inte importera moduler från 'shared.py': {e}. Sökrymden kommer bli begränsad.")

def _normalize_word(value: str) -> str:
    return value.strip().lower()

def get_vocabulary():
    vocab = set()
    
    if HAS_SHARED:
        # 1. Load Kelly (Official Dictionary)
        print("Laddar Kelly-listan via shared.py...")
        kelly_words = {_normalize_word(w) for w in load_kelly() if isinstance(w, str) and w.strip()}
        vocab.update(kelly_words)
        print(f" -> Lade till {len(kelly_words)} ord från Kelly.")

        # 2. Load Korp (Frequency data)
        print("Laddar Korp-data via shared.py...")
        korp_rows = load_korp_csvs()
        korp_words = set()
        for row in korp_rows:
            # Try common column names for the word, fallback to the first column
            word = row.get('word') or row.get('Word') or row.get('token')
            if not word and row:
                word = list(row.values())[0]
            
            if word:
                korp_words.add(_normalize_word(str(word)))
        vocab.update(korp_words)
        print(f" -> Lade till {len(korp_words)} unika ord från {len(korp_rows)} Korp-rader.")

    # 3. Load Seeding Outputs (Celebrities, Companies, etc.)
    # We still use pandas here since your seeding outputs might have different column structures
    seed_files = glob.glob(os.path.join("seeding", "output", "*.csv"))
    if seed_files:
        print(f"Laddar entiteter från {len(seed_files)} seeding-filer...")
        seed_words = set()
        for file in seed_files:
            try:
                df = pd.read_csv(file)
                # Look for 'itemLabel' (SPARQL default) or 'word'
                col_name = 'itemLabel' if 'itemLabel' in df.columns else 'word'
                if col_name in df.columns:
                    seed_words.update(
                        _normalize_word(w)
                        for w in df[col_name].dropna().astype(str).tolist()
                        if w.strip()
                    )
            except Exception as e:
                print(f"Kunde inte ladda {file}: {e}")
        vocab.update(seed_words)
        print(f" -> Lade till {len(seed_words)} entiteter från seeding.")
    else:
        print("Inga seeding-filer hittades i 'seeding/output/'.")

    # 4. Filter out custom stopwords
    if HAS_SHARED:
        print("Tillämpar custom stopwords...")
        stopwords = load_custom_stopwords()
        if stopwords:
            original_len = len(vocab)
            # Subtract the stopwords set from the vocab set (ignoring case)
            vocab = {w for w in vocab if w not in stopwords}
            print(f" -> Tog bort {original_len - len(vocab)} ord som fanns i stopwords.")

    # Final cleanup: filter out empty strings and single characters
    clean_vocab = [w for w in vocab if isinstance(w, str) and len(w) > 1]
    
    return list(set(clean_vocab))

def main():
    print("=" * 60)
    print("Laddar intfloat/multilingual-e5-large i RAM...")
    print("Observera: Detta är en stor modell (~2.2 GB).")
    print("=" * 60)
    
    model = SentenceTransformer('intfloat/multilingual-e5-large')
    
    print("\nBygger den officiella sökrymden...")
    vocabulary = get_vocabulary()
    
    if not vocabulary:
        print("Fel: Sökrymden är tom. Avbryter.")
        return

    # Prefixing for E5 (symmetric search)
    prefixed_vocabulary = [f"query: {word}" for word in vocabulary]
    
    print(f"\nKodar {len(vocabulary)} unika ord/entiteter till vektorer...")
    start_encode = time.time()
    
    vocab_vectors = model.encode(
        prefixed_vocabulary, 
        show_progress_bar=True, 
        normalize_embeddings=True
    )
    
    print(f"Sökrymd vektoriserad på {time.time() - start_encode:.1f} sekunder.")
    print("=" * 60)

    TOP_N = 50

    while True:
        try:
            user_input = input("\nSkriv ett ord, namn eller fras (eller 'q' för avsluta): ").strip()
            
            if user_input.lower() in ['q', 'quit', 'exit']:
                break
            if not user_input:
                continue

            start_time = time.time()
            
            # Encode user input with the required E5 prefix
            target_vector = model.encode(
                [f"query: {user_input}"], 
                normalize_embeddings=True
            )
            
            similarities = cosine_similarity(target_vector, vocab_vectors)[0]
            closest_indices = np.argsort(similarities)[::-1]

            print(f"\nBeräkning klar på {time.time() - start_time:.3f} sekunder.")
            print(f"{'Likhet':<10} | Ord/Entitet")
            print("-" * 40)

            count = 0
            for idx in closest_indices:
                if count >= TOP_N:
                    break
                    
                word = vocabulary[idx]
                sim = similarities[idx]
                
                # Skip the exact match
                if word.lower() == user_input.lower():
                    continue
                    
                print(f"{sim:<10.4f} | {word}")
                count += 1
                
        except KeyboardInterrupt:
            print("\nAvslutar...")
            break
        except Exception as e:
            print(f"Ett fel uppstod: {e}")

if __name__ == "__main__":
    main()
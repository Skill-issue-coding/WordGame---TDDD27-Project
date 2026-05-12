import time
import numpy as np
from sentence_transformers import SentenceTransformer
from sklearn.metrics.pairwise import cosine_similarity

# Försök importera din Kelly-loader, annars använd en tom lista som grund
try:
    from shared import load_kelly
    HAS_KELLY = True
except ImportError:
    HAS_KELLY = False
    print("Varning: Kunde inte importera 'load_kelly' från 'shared.py'. Använder bara test-ord.")

def get_vocabulary():
    vocab = set()
    
    # 1. Ladda din officiella ordbok om den finns
    if HAS_KELLY:
        print("Laddar den officiella svenska ordboken (Kelly)...")
        kelly_words = load_kelly()
        vocab.update(kelly_words)
    
    # 2. Lägg till moderna/popkulturella ord som ofta saknas i gamla ordböcker
    # Detta är avgörande för att popkultur/entiteter ska kunna kopplas till något vettigt!
    moderna_ord = [
        "youtube", "spel", "minecraft", "influencer", "internet", "dator", 
        "streamer", "twitch", "esport", "fotboll", "sport", "mål", "boll", 
        "politik", "riksdag", "sverige", "regering", "klimat", "miljö",
        "musik", "konsert", "artist", "sång", "scen", "film", "bio", "skådespelare"
    ]
    vocab.update(moderna_ord)
    
    return list(vocab)

def main():
    print("=" * 60)
    print("Laddar KBLab/sentence-bert-swedish-cased i RAM...")
    print("=" * 60)
    model = SentenceTransformer('KBLab/sentence-bert-swedish-cased')
    
    print("Bygger sökrymd...")
    vocabulary = get_vocabulary()
    
    print(f"Kodar {len(vocabulary)} ord till vektorer... (Detta kan ta lite tid beroende på listans storlek)")
    start_encode = time.time()
    
    # SBERT kodar hela listan effektivt. 
    # Om din Kelly-lista är enorm (100k+ ord) kan du behöva sätta batch_size=256 och show_progress_bar=True
    vocab_vectors = model.encode(vocabulary, show_progress_bar=True)
    
    print(f"Sökrymd vektoriserad på {time.time() - start_encode:.1f} sekunder.")
    print(f"Redo! Söker bland {len(vocabulary)} ord.")
    print("=" * 60)

    TOP_N = 15

    while True:
        try:
            user_input = input("\nSkriv ett ord, namn eller fras (eller 'q' för avsluta): ").strip()
            
            if user_input.lower() in ['q', 'quit', 'exit']:
                break
            if not user_input:
                continue

            start_time = time.time()
            
            # 1. Skapa vektor för inputen
            target_vector = model.encode([user_input])
            
            # 2. Jämför mot hela sökrymden
            similarities = cosine_similarity(target_vector, vocab_vectors)[0]
            closest_indices = np.argsort(similarities)[::-1]

            print(f"\nBeräkning klar på {time.time() - start_time:.3f} sekunder.")
            print(f"{'Likhet':<10} | Ord i sökrymden")
            print("-" * 40)

            count = 0
            for idx in closest_indices:
                if count >= TOP_N:
                    break
                    
                word = vocabulary[idx]
                sim = similarities[idx]
                
                # Hoppa över ordet om det är exakt det användaren skrev
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
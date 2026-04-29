import fasttext
import fasttext.util
import pandas as pd
import spacy
import xml.etree.ElementTree as ET
import re
import os 

MODEL_NAME = 'cc.sv.300.bin'
OUTPUT_BASE = "../server/wordfiles/"

print(f"Ensuring output directory exists: {OUTPUT_BASE}")
os.makedirs(OUTPUT_BASE, exist_ok=True)

################## 1. LOAD MODELS ##################

print("Loading FastText model...")
ft = fasttext.load_model(MODEL_NAME)
fasttext.util.reduce_model(ft, 100)
print("Fasttext Model dimensions: ", ft.get_dimension())

print("Loading Swedish NLP model...")
try:
    nlp = spacy.load("sv_core_news_sm")
except OSError:
    print("Spacy model inte hittad. Kör: python -m spacy download sv_core_news_sm")
    exit()

# Helper function to convert the vector into a string for the CSV
def get_vector_string(word):
    vec = ft.get_word_vector(word)
    return " ".join(map(str, vec.tolist()))

################## 2. KELLY WORDS (Svensk Vokabulär) ##################
print("\n--- Bearbetar Kelly.xml ---")
tree = ET.parse('kelly.xml')
root = tree.getroot()

kelly_words = []
for entry in root.iter('LexicalEntry'):
    written_form = entry.find('.//feat[@att="writtenForm"]')
    pos = entry.find('.//feat[@att="kellyPartOfSpeech"]')
    
    if written_form is not None:
        word = written_form.get('val').lower()
        word_pos = pos.get('val') if pos is not None else ""
        
        if "noun" in word_pos or "adjective" in word_pos or "verb" in word_pos:
            kelly_words.append({
                'Word': word,
                'POS': word_pos, # <-- ADDED THIS TAG
                'Vector': get_vector_string(word)
            })

kelly_df = pd.DataFrame(kelly_words).drop_duplicates(subset=['Word'])
# Reorder columns to ensure consistency
kelly_df = kelly_df[['Word', 'POS', 'Vector']]
kelly_df.to_csv(OUTPUT_BASE + 'kelly_vectors.csv', index=False)
print(f"Sparade {len(kelly_df)} Kelly-ord till 'kelly_vectors.csv'")

################## 3. COMPANIES ##################
print("\n--- Bearbetar Företag ---")
pop_comp_df = pd.read_csv('popular-companies.csv', usecols=['Name'])
swe_comp_df = pd.read_csv('swedish-companies.csv', usecols=['Name'])
comp_df = pd.concat([pop_comp_df, swe_comp_df]).drop_duplicates()

def clean_company_name(name):
    name = str(name)
    if '(' in name and ')' in name:
        match = re.search(r'\((.*?)\)', name)
        if match:
            name = match.group(1)
    suffixes = [r'\s+AB$', r'\s+Inc\.?$', r'\s+Corp\.?$', r'\s+Ltd\.?$', r'\s+Group$']
    for suffix in suffixes:
        name = re.sub(suffix, '', name, flags=re.IGNORECASE)
    return name.strip()

comp_df['CleanName'] = comp_df['Name'].apply(clean_company_name)
comp_df = comp_df.drop_duplicates(subset=['CleanName'])

comp_df['Vector'] = comp_df['CleanName'].apply(lambda w: get_vector_string(w.replace(" ", "_")))
comp_df['POS'] = 'proper_noun'

comp_df = comp_df[['CleanName', 'POS', 'Vector']].rename(columns={'CleanName': 'Word'})

comp_df.to_csv(OUTPUT_BASE + 'companies_vectors.csv', index=False)
print(f"Sparade {len(comp_df)} Företag till 'companies_vectors.csv'")

################## 4. CELEBRITIES ##################
print("\n--- Bearbetar Kändisar ---")
celeb_df = pd.read_csv('celebrity.csv', usecols=['name', 'popularity'])
celeb_df = celeb_df.sort_values(by='popularity', ascending=False).head(2000)

celeb_df['Vector'] = celeb_df['name'].apply(lambda w: get_vector_string(w.replace(" ", "_")))
celeb_df['POS'] = 'proper_noun'

celeb_df = celeb_df[['name', 'POS', 'Vector']].rename(columns={'name': 'Word'})

celeb_df.to_csv(OUTPUT_BASE + 'celebrities_vectors.csv', index=False)
print(f"Sparade {len(celeb_df)} Kändisar till 'celebrities_vectors.csv'")

################## 5. MAKTBAROMETERN (Svenska Influencers/Media) ##################
print("\n--- Bearbetar Maktbarometern ---")
makt_files = [
    'maktbarometern/2025-arets-makthavare.csv',
    'maktbarometern/2025-facebook.csv',
    'maktbarometern/2025-instagram.csv',
    'maktbarometern/2025-tiktok.csv',
    'maktbarometern/2025-x.csv',
    'maktbarometern/2025-youtube.csv'
]

makt_dfs = []
for file in makt_files:
    try:
        df = pd.read_csv(file, usecols=['name'])
        makt_dfs.append(df)
    except Exception as e:
        print(f"Kunde inte läsa {file}. Kolla så mappen ligger rätt. ({e})")

if makt_dfs:
    makt_df = pd.concat(makt_dfs).drop_duplicates()

    def clean_makt_name(name):
        name = str(name)
        name = re.sub(r'[^\w\såäöÅÄÖ-]', '', name)
        return name.strip()

    makt_df['CleanName'] = makt_df['name'].apply(clean_makt_name)
    makt_df = makt_df[makt_df['CleanName'] != '']
    makt_df = makt_df.drop_duplicates(subset=['CleanName'])

    makt_df['Vector'] = makt_df['CleanName'].apply(lambda w: get_vector_string(w.replace(" ", "_")))
    makt_df['POS'] = 'proper_noun'

    makt_df = makt_df[['CleanName', 'POS', 'Vector']].rename(columns={'CleanName': 'Word'})

    makt_df.to_csv(OUTPUT_BASE + 'maktbarometern_vectors.csv', index=False)
    print(f"Sparade {len(makt_df)} profiler från Maktbarometern till 'maktbarometern_vectors.csv'")

################## 6. KORP (Vanliga ord) ##################
print("\n--- Bearbetar Korp-statistik ---")
stop_df = pd.read_csv('stoppord-mycket.csv', header=None, names=['word'])
stopwords = set(stop_df['word'].dropna().str.lower().tolist())

korp_df = pd.read_csv('korp-statistics.csv', skiprows=[1,2,3,4], usecols=['word'])
top_words_df = korp_df.dropna(subset=['word']).head(100000).copy()

korp_words = []
valid_word_pattern = re.compile(r'^[a-zåäö]+$')

print("Filtrerar Korp-ord... (Detta kan ta någon minut)")
for index, row in top_words_df.iterrows():
    word = str(row['word']).lower()
    
    if not valid_word_pattern.match(word):
        continue
    if word in stopwords:
        continue
        
    doc = nlp(word)
    if len(doc) > 0:
        token = doc[0]
        if token.pos_ in ['ADJ', 'VERB', 'NOUN']:
            korp_words.append({
                'Word': word,
                'POS': token.pos_,
                'Vector': get_vector_string(word)
            })
            
    if len(korp_words) >= 50000:
        break

korp_final_df = pd.DataFrame(korp_words).drop_duplicates(subset=['Word'])
korp_final_df = korp_final_df[['Word', 'POS', 'Vector']]
korp_final_df.to_csv(OUTPUT_BASE + 'korp_vectors.csv', index=False)
print(f"Sparade {len(korp_final_df)} Korp-ord till 'korp_vectors.csv'")

print("\nAll preprocessing klar!")
# WordGame - TDDD27 Project

## Preprocessing

The preprocessing pipeline lives in [preprocessing/README.md](preprocessing/README.md). It builds Swedish word and entity embeddings using `intfloat/multilingual-e5-large`, and outputs files into `server/wordfiles/`. Follow the setup and stage order in that document.

## Data Sources

- **Swedish stopwords:** From [peterdalle/svensktext](https://github.com/peterdalle/svensktext) (see `preprocessing/stopwords/`).

- **Common Swedish words and frequency:** From Sprakbanken Korp data in `preprocessing/korp/`.

- **Kelly word list:** `preprocessing/kelly.xml` from Sprakbanken.

- **Popular companies:** Source data in `preprocessing/` from companiesmarketcap.com and Kaggle.

- **Celebrities:** Source data in `preprocessing/` from Kaggle.

## Stack

- Preprocessing: Python pipeline (see [preprocessing/README.md](preprocessing/README.md))
- Backend: Go (loads `server/wordfiles/` at startup)
- Frontend: Next.js

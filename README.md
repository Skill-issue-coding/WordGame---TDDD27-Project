# WordGame - TDDD27 Project

## Word Vectors

- **FastText:** The swedish word vector has to be downloaded from [fasttext.cc/docs/en/crawl-vectors](https://fasttext.cc/docs/en/crawl-vectors).

## Acknowledgments

- **Swedish Stopwords:** The list of swedish stopwords (`python/stoppord.csv`) is sourced from [peterdalle/svensktext](https://github.com/peterdalle/svensktext) repository.

- **Common Swedish Words and Frequency:** The list of swedish words and frequency (`python/korp-statistics.csv`) is sourced from [spraakbanken.gu.se/](https://spraakbanken.gu.se/).

- **Kelly:** The list of Kelly swedish words (`python/kelly.xml`) is sourced from [spraakbanken.gu.se/](https://spraakbanken.gu.se/en/resources/kelly).

- **Popular Companies:** The list of popular companies (`python/popular-companies.csv`) is sourced from [companiesmarketcap.com/](https://companiesmarketcap.com/).

- **Swedish Companies:** The list of swedish companies (`python/swedish-companies.csv`) is sourced from [companiesmarketcap.com/sweden/largest-companies-in-sweden-by-market-cap/](https://companiesmarketcap.com/sweden/largest-companies-in-sweden-by-market-cap/).

- **Celebrities:** The list of celebrities (`python/Celebrity.csv`) is sourced from [kaggle.com/datasets/madhuripanchakshri/top-10000-celebrities-dataset](https://www.kaggle.com/datasets/madhuripanchakshri/top-10000-celebrities-dataset).

## Idea and stack

Vi har lite olika idéer för vårt projekt men man kan egentligen generalisera det till att:

Vi ska använda den svenska färdig tränade fasttext modellen och ta ut ett antal svenska ord, företag, kändisar osv. Sedan göra ett spel med det. Vad exakt för slags spel diskuterar vi fortfarande om, men vi har i alla fall kommit fram till en tech stack:

**Stack:**

Preprocessing: Python + FastText modell

Backend: Go

Frontend: Next.js

Vi kommer använda Python för "preprocessing" och då får vi en fil vi kan ladda in i en Go backend (med svenska ord sammankopplade med vektorer). Vår tanke är då att användaren ansluter till "tjänsten"/hemsidan och vår next.js frontend ansluter sig till vår backend genom websockets. Användaren kan då likt många online spel idag, skapa eller gå med i en lobby/ett rum och bjuda in sina vänner. Eventuellt ändra spelregler, "gamemode" eller liknande. Sedan starta spelet osv.

De spel idéer vi har just nu är:

**Alternativ 1 - Impostor Spel:**

Alla spelare förutom en får samma ord, den som inte får samma ord som dem andra är "impostor" och den får ett närliggande ord till det faktiska ordet. Exempelvis om ordet är "äpple" får impostorn "frukt" eller liknande. Man har rundor där alla ska skicka in ett ord för att bevisa att de inte är "impostorn", efter varje runda ska man rösta ut en person (eller hoppa över röstningen för just den rundan).

**Alternativ 2 - Contexto:**

Contexto fast att man spelar mot andra. Alltså att backenden väljer ut en ord helt random sen ska varje spelare försöka komma fram till det ordet så snabbt som möjligt genom att skriva in olika ord. När användaren skriver in ett ord så kommer den få ett "score" på hur nära det ordet är det korrekta ordet.

**Alternativ 3 - :**

Backenden tar åter igen fram ett ord. Alla ska skriva in ett ord de tycker är relevant/liknande på x antal sekunder (får kolla med max distans med GloVe så man inte skriver in ett random ord). Skriver 2 spelare in samma ord så får båda spelarna 0 poäng. Den med närmast avstånd till det ursprungliga ordet vinner.

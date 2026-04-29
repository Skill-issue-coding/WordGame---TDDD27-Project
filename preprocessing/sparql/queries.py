from dataclasses import dataclass
from typing import Dict, List


@dataclass(frozen=True)
class Query:
  name: str
  sparql: str

celebrity_query = """
PREFIX wd: <http://www.wikidata.org/entity/>
PREFIX wdt: <http://www.wikidata.org/prop/direct/>
PREFIX wikibase: <http://wikiba.se/ontology#>
PREFIX bd: <http://www.bigdata.com/rdf#>
PREFIX schema: <http://schema.org/>
PREFIX xsd: <http://www.w3.org/2001/XMLSchema#>

SELECT ?person ?personLabel ?sitelinks ?birthDate WHERE {
  ?person wdt:P31 wd:Q5.
  ?person wdt:P27 wd:Q34.

  ?person wdt:P569 ?birthDate.

  ?svArticle schema:about ?person ;
             schema:isPartOf <https://sv.wikipedia.org/> .

  ?person wikibase:sitelinks ?sitelinks.

  FILTER(
    ?sitelinks > 3 && ?sitelinks < 50 &&
    ?birthDate >= "1990-01-01T00:00:00Z"^^xsd:dateTime
  )

  SERVICE wikibase:label {
    bd:serviceParam wikibase:language "sv".
  }
}
ORDER BY DESC(?sitelinks)
LIMIT 500
"""

QUERIES: List[Query] = [
  Query("celebrities", celebrity_query),
]

QUERIES_BY_NAME: Dict[str, Query] = {query.name: query for query in QUERIES}


def get_query(name: str) -> Query:
  try:
    return QUERIES_BY_NAME[name]
  except KeyError as exc:
    available = ", ".join(sorted(QUERIES_BY_NAME))
    raise KeyError(f"Unknown query '{name}'. Available: {available}") from exc
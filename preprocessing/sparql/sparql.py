import random
import time
from typing import Dict, Iterable, List
from SPARQLWrapper import JSON, SPARQLWrapper
from queries import Query, QUERIES, get_query

ENDPOINT = "https://query.wikidata.org/sparql"

def run_sparql(query: str):
    sparql = SPARQLWrapper(ENDPOINT)
    sparql.setQuery(query)
    sparql.setReturnFormat(JSON)

    sparql.addCustomHttpHeader("User-Agent", "WordGameResearchBot/0.1 (mailto:emidj236@student.liu.se)")

    time.sleep(0.5 + random.random() * 0.5)

    return sparql.query().convert()


def run_queries(queries: Iterable[Query]) -> Dict[str, object]:
    results: Dict[str, object] = {}
    for query in queries:
        results[query.name] = run_sparql(query.sparql)
    return results


def run_all_queries() -> Dict[str, object]:
    return run_queries(QUERIES)


def run_queries_by_name(names: Iterable[str]) -> Dict[str, object]:
    resolved: List[Query] = [get_query(name) for name in names]
    return run_queries(resolved)


if __name__ == "__main__":
    run_all_queries()
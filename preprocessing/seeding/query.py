import csv
import os
import random
import socket
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Tuple
from urllib.error import HTTPError, URLError
import json

from SPARQLWrapper import JSON, SPARQLWrapper


@dataclass(frozen=True)
class Query:
    name: str
    query_filepath: str
    output_filepath: str = ""

    ENDPOINT = "https://query.wikidata.org/sparql"
    MAIL = os.getenv("MAIL", "")
    REQUEST_TIMEOUT_SECONDS = int(os.getenv("WDQS_TIMEOUT_SECONDS", "45"))
    BASE_DELAY_SECONDS = float(os.getenv("WDQS_BASE_DELAY_SECONDS", "3.0"))
    MAX_RETRIES = int(os.getenv("WDQS_MAX_RETRIES", "4"))
    BACKOFF_BASE = float(os.getenv("WDQS_BACKOFF_BASE", "2.0"))
    RETRY_STATUS_CODES = {429, 500, 502, 503, 504}

    def __post_init__(self) -> None:
        if not self.output_filepath:
            object.__setattr__(self, "output_filepath", f"{self.name}.csv")
        if not self.MAIL:
            raise ValueError("Environment variable 'MAIL' must be set for the Wikidata User-Agent policy.")

    def _load_query(self) -> str:
        return Path(self.query_filepath).read_text(encoding="utf-8")

    @staticmethod
    def _sleep_with_jitter(base_seconds: float) -> None:
        time.sleep(base_seconds + random.random() * 0.5)

    def run_query(self) -> Dict[str, object]:
        # ... (Keep your existing run_query logic exactly the same) ...
        query_text = self._load_query()
        for attempt in range(self.MAX_RETRIES + 1):
            sparql = SPARQLWrapper(self.ENDPOINT)
            sparql.setQuery(query_text)
            sparql.setReturnFormat(JSON)
            sparql.setTimeout(self.REQUEST_TIMEOUT_SECONDS)

            sparql.addCustomHttpHeader(
                "User-Agent",
                f"WordGameResearchBot/0.1 (mailto:{self.MAIL})",
            )

            self._sleep_with_jitter(self.BASE_DELAY_SECONDS)

            try:
                # REMOVE this line:
                # return sparql.query().convert()
                
                # ADD these lines instead:
                result = sparql.query()
                raw_json_str = result.response.read().decode("utf-8")
                return json.loads(raw_json_str, strict=False)

            except HTTPError as exc:
                if exc.code not in self.RETRY_STATUS_CODES or attempt >= self.MAX_RETRIES:
                    raise
            except (URLError, socket.timeout):
                if attempt >= self.MAX_RETRIES:
                    raise

            backoff_seconds = self.BASE_DELAY_SECONDS * (self.BACKOFF_BASE ** attempt)
            self._sleep_with_jitter(backoff_seconds)

        raise RuntimeError("Failed to execute query after retries.")

    def parse_results(self, result: Dict[str, object]) -> Tuple[List[str], List[Dict[str, str]]]:
        """Parses Wikidata JSON response into headers and dictionary rows."""
        head = result.get("head", {})
        headers = list(head.get("vars", []))
        bindings = result.get("results", {}).get("bindings", [])
        
        rows: List[Dict[str, str]] = []
        for binding in bindings:
            row: Dict[str, str] = {
                header: binding.get(header, {}).get("value", "") 
                for header in headers
            }
            rows.append(row)
            
        return headers, rows

    def run_and_save(self, output_dir: Path) -> List[Dict[str, str]]:
        """Executes the query, saves it to CSV, and returns the parsed rows."""
        raw_result = self.run_query()
        headers, rows = self.parse_results(raw_result)

        output_path = output_dir / self.output_filepath
        output_path.parent.mkdir(parents=True, exist_ok=True)
        
        with output_path.open("w", encoding="utf-8", newline="") as handle:
            writer = csv.DictWriter(handle, fieldnames=headers)
            writer.writeheader()
            writer.writerows(rows)

        return rows
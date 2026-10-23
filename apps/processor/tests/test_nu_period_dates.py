"""Two-level date review for the Nu parser.

Level 1 (document): the statement period is unreadable -> one parked
document review item (raw preserved); movements still parse on their own.
Level 2 (movement): with no year from the period, a movement whose own
date carries no year parks as a movement review item instead of silently
emptying its normalized date.

Artifacts are synthetic (tmpdir), no local PDF needed. Run with:
<python.exe> tests/test_nu_period_dates.py
"""

import json
import shutil
import sys
import tempfile
from pathlib import Path

PROCESSOR_ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(PROCESSOR_ROOT))

from parser.nu_account import NuAccountParser

# Real header layout: "Periodo" label, then the range on the next line.
REAL_PERIOD_EXTRACT = """Hola
Wilhelm David Buitrago Garcia
Periodo
01 - 30 ABR 2026
Movimientos
10 abr
Enviaste a alguien
-$100.000,00
14 abr
Recibiste de alguien
+$16.500,00
Rendimiento total de tu cuenta
"""


def run_parser(extract_text):
    base = Path(tempfile.mkdtemp(prefix="nu-period-"))
    try:
        pid = "t"
        artifact_dir = base / pid
        (artifact_dir / "extraction").mkdir(parents=True)
        (artifact_dir / "manifest.json").write_text(json.dumps({}), encoding="utf-8")
        (artifact_dir / "inspection.json").write_text(json.dumps({}), encoding="utf-8")
        (artifact_dir / "extraction" / "001.txt").write_text(extract_text, encoding="utf-8")
        return NuAccountParser(pid, base_path=str(base)).parse()
    finally:
        shutil.rmtree(base, ignore_errors=True)


def test_real_period_normalizes_movements():
    doc = run_parser(REAL_PERIOD_EXTRACT)
    assert doc.metadata.get("periodo") == "01 - 30 ABR 2026", doc.metadata.get("periodo")
    assert len(doc.movements) == 2, len(doc.movements)
    assert doc.movements[0].date_normalized == "2026-04-10", doc.movements[0].date_normalized
    assert doc.movements[1].date_normalized == "2026-04-14"
    assert len(doc.document_review) == 0, "a readable period parks nothing"


def test_unreadable_period_parks_document_not_movements():
    extract = """Hola
Usuario
Periodo
no se pudo determinar
Movimientos
14 abr 2026
Recibiste de alguien
+$16.500,00
Rendimiento total de tu cuenta
"""
    doc = run_parser(extract)
    # Level 1: the document period is parked once, raw preserved.
    assert len(doc.document_review) == 1, (
        f"unreadable period must park one document item, got {len(doc.document_review)}"
    )
    assert doc.document_review[0].raw_period == "no se pudo determinar"
    # The movement still parses: it carries its own year.
    assert len(doc.movements) == 1, len(doc.movements)
    assert doc.movements[0].date_normalized == "2026-04-14", doc.movements[0].date_normalized
    assert len(doc.review_required) == 0, "a movement with its own year is not parked"


def test_movement_without_year_parks_as_movement_review():
    extract = """Hola
Usuario
Periodo
no se pudo determinar
Movimientos
14 abr
Recibiste de alguien
+$16.500,00
Rendimiento total de tu cuenta
"""
    doc = run_parser(extract)
    # Level 2: no year anywhere: the movement parks with its raw date.
    assert len(doc.movements) == 0, "a movement without a derivable date must not emit an empty date"
    assert len(doc.review_required) == 1, (
        f"movement must park, got {len(doc.review_required)}"
    )
    assert doc.review_required[0].date_original == "14 abr"
    assert "date" in doc.review_required[0].reason.lower()
    # The document period is still parked, independently.
    assert len(doc.document_review) == 1


def test_period_on_same_line_variant():
    extract = """Hola
Usuario
Periodo: 01 - 30 ABR 2026
Movimientos
10 abr
Pago a Tigo
-$3.000,00
Rendimiento total de tu cuenta
"""
    doc = run_parser(extract)
    assert doc.metadata.get("periodo") == "01 - 30 ABR 2026", doc.metadata.get("periodo")
    assert len(doc.movements) == 1, len(doc.movements)
    assert doc.movements[0].date_normalized == "2026-04-10"


if __name__ == "__main__":
    tests = [
        test_real_period_normalizes_movements,
        test_unreadable_period_parks_document_not_movements,
        test_movement_without_year_parks_as_movement_review,
        test_period_on_same_line_variant,
    ]
    for t in tests:
        t()
        print(f"ok  {t.__name__}")
    print("Period/date tests passed")

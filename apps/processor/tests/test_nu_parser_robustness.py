"""Robustness tests for the Nu parser: deviations must park for human
review, never fabricate a corrupted movement or drop one silently.

The artifacts are synthetic, generated in a tmpdir: no local PDF artifact
is needed. Run with: <python.exe> tests/test_nu_parser_robustness.py
"""

import json
import shutil
import sys
import tempfile
from pathlib import Path

PROCESSOR_ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(PROCESSOR_ROOT))

from parser.nu_account import NuAccountParser

# Real Nu statement format, as provided by the maintainer: dd-mes dates,
# one-line descriptions, amounts with the sign glued to the dollar sign.
REAL_NU_EXTRACT = """Hola
Wilhelm David Buitrago Garcia
Movimientos
10 abr
Enviaste a Wilhelm David Buitrago Garcia
-$100.000,00
10 abr
Pago a Tigo
-$3.000,00
10 abr
Enviaste a Wilhelm David Buitrago Garcia
-$10.000,00
14 abr
Compra en SPOTIFY con tarjeta débito
-$30.500,00
14 abr
Recibiste de Cristian Andres Villarreal Orozco
+$16.500,00
14 abr
Recibiste de Ronald Steven Perez Baquero
+$670.600,00
14 abr
Recibiste de Angela Maria Maestre Sierra
+$406.385,00
15 abr
Enviaste a KELLYS DAYANA NUEZ TERNERA
-$780.000,00
Rendimiento total de tu cuenta
"""


def make_artifact(base, processing_id, extract_text):
    artifact_dir = base / processing_id
    (artifact_dir / "extraction").mkdir(parents=True)
    (artifact_dir / "manifest.json").write_text(json.dumps({}), encoding="utf-8")
    (artifact_dir / "inspection.json").write_text(json.dumps({}), encoding="utf-8")
    (artifact_dir / "extraction" / "001.txt").write_text(extract_text, encoding="utf-8")


def run_parser(extract_text):
    base = Path(tempfile.mkdtemp(prefix="nu-test-"))
    try:
        make_artifact(base, "t", extract_text)
        return NuAccountParser("t", base_path=str(base)).parse()
    finally:
        shutil.rmtree(base, ignore_errors=True)


def test_real_format_parses_unchanged():
    doc = run_parser(REAL_NU_EXTRACT)
    assert len(doc.movements) == 8, f"expected 8 movements, got {len(doc.movements)}"
    assert len(doc.review_required) == 0, (
        f"real format must not park anything, got {len(doc.review_required)}"
    )
    first = doc.movements[0]
    assert first.date_original == "10 abr", first.date_original
    assert first.description_original == "Enviaste a Wilhelm David Buitrago Garcia"
    assert first.amount_original == "-$100.000,00"
    assert first.amount_normalized == -100000, first.amount_normalized


def test_amount_variant_line_is_parked():
    # Variant amount (space after the sign) must not be swallowed as
    # description and must not take the next movement's amount.
    extract = """Movimientos
01 mar
Compra en almacen
+$700.000,00
02 mar
Pago de servicios
+ $350.000,00
03 mar 2026
Otro movimiento
-$100.000,00
Rendimiento total de tu cuenta
"""
    doc = run_parser(extract)
    assert len(doc.movements) == 2, f"expected A and C only, got {len(doc.movements)}"
    assert len(doc.review_required) == 1, (
        f"the variant-amount movement must be parked, got {len(doc.review_required)}"
    )
    parked = doc.review_required[0]
    assert parked.date_original == "02 mar", parked.date_original
    assert "amount" in parked.reason.lower(), parked.reason
    # C parses, with its date raw preserved.
    assert doc.movements[1].amount_original == "-$100.000,00"
    assert doc.movements[1].date_original == "03 mar 2026"


def test_absorbed_structured_lines_parked():
    # The description of the parsed movement must never contain an
    # amount-like or date-like line: that means it ate the next movement.
    extract = """Movimientos
01 mar
Compra en tienda
+ $50.000,00
Descuento aplicado
-$15.000,00
Rendimiento total de tu cuenta
"""
    doc = run_parser(extract)
    assert len(doc.movements) == 0, "a corrupted movement must not be emitted"
    assert len(doc.review_required) == 1, (
        f"absorption must park the movement, got {len(doc.review_required)}"
    )
    assert "absorb" in doc.review_required[0].reason.lower(), doc.review_required[0].reason


def test_date_with_year_is_tolerated():
    extract = """Movimientos
10 abr 2026
Recibiste de alguien
+$16.500,00
Rendimiento total de tu cuenta
"""
    doc = run_parser(extract)
    assert len(doc.movements) == 1, f"date with year must parse, got {len(doc.movements)}"
    assert doc.movements[0].date_original == "10 abr 2026", doc.movements[0].date_original
    assert doc.movements[0].amount_normalized == 16500


if __name__ == "__main__":
    tests = [
        test_real_format_parses_unchanged,
        test_amount_variant_line_is_parked,
        test_absorbed_structured_lines_parked,
        test_date_with_year_is_tolerated,
    ]
    for t in tests:
        t()
        print(f"ok  {t.__name__}")
    print("Robustness tests passed")

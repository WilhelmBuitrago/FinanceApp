"""Smoke: process_pdf runs end to end. The other suites cover its stages
one by one, but nothing exercised the orchestration itself — which is how
a call-site NameError (a leftover dstPath) shipped invisible to the stage
tests and to py_compile: only a real run of the pipeline finds it.

Only the PDF read is stubbed. The artifacts, the parse, the validation
and the manifest all run for real.

Run with: <python.exe> tests/test_nu_pipeline_smoke.py
"""

import json
import os
import sys
import tempfile
from pathlib import Path

PROCESSOR_ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(PROCESSOR_ROOT))

from src import main

WITH_PERIOD = """Hola
Juan Perez
CC-12345678
Número de Cuenta
4012345678
Periodo
01 - 30 ABR 2026
Movimientos
14 abr
Recibiste de alguien
+$16.500,00
Rendimiento total de tu cuenta
"""


def test_process_pdf_end_to_end():
    base = Path(tempfile.mkdtemp(prefix="pipeline-"))

    inspection = {"file": {"sha256": "smoke", "size_bytes": 1}, "warnings": []}
    extraction = {"pages": [{"page_number": 1, "text": WITH_PERIOD}]}

    # The only stubbed boundary: reading the PDF. Everything downstream —
    # artifacts on disk, the parser, the validator, the manifest — is real.
    main.inspect_and_extract = lambda pdf_path: (inspection, extraction)

    previous = os.environ.get("ARTIFACTS_PATH")
    os.environ["ARTIFACTS_PATH"] = str(base)
    try:
        manifest = main.process_pdf("smoke.pdf")
    finally:
        if previous is None:
            os.environ.pop("ARTIFACTS_PATH", None)
        else:
            os.environ["ARTIFACTS_PATH"] = previous

    processing_id = manifest["processing_id"]
    if manifest["status"] != "completed":
        raise AssertionError(f"pipeline failed: {manifest['errors']} {manifest['validation']}")
    if manifest["movements_count"] < 1:
        raise AssertionError("the fixture movement did not survive the pipeline")

    artifact_dir = base / processing_id
    for expected in ("inspection.json", "extraction/001.txt", "manifest.json", "parsed.json"):
        if not (artifact_dir / expected).exists():
            raise AssertionError(f"missing artifact: {expected}")

    written = json.loads((artifact_dir / "manifest.json").read_text(encoding="utf-8"))
    if written["processor_version"] != main.PROCESSOR_VERSION:
        raise AssertionError("manifest version drifted from the component version")
    if written["contract_version"] != main.CONTRACT_VERSION:
        raise AssertionError("manifest contract version drifted")


if __name__ == "__main__":
    test_process_pdf_end_to_end()
    print("Pipeline smoke tests passed")

"""Layer 0: when the statement period is unreadable on the first pass, the
processor silently re-extracts once. A partial write or a failed read can
leave the text incomplete; re-extracting usually fixes it. The user is
never notified of the retry: only its failure surfaces, as a document
review item.

Run with: <python.exe> tests/test_nu_extraction_retry.py
"""

import json
import shutil
import sys
import tempfile
from pathlib import Path

PROCESSOR_ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(PROCESSOR_ROOT))

from src.main import parse_with_retry

WITH_PERIOD = """Hola
Usuario
Periodo
01 - 30 ABR 2026
Movimientos
14 abr
Recibiste de alguien
+$16.500,00
Rendimiento total de tu cuenta
"""

WITHOUT_PERIOD = """Hola
Usuario
Movimientos
14 abr
Recibiste de alguien
+$16.500,00
Rendimiento total de tu cuenta
"""


def make_artifact(base, pid, extract_text):
    artifact_dir = base / pid
    (artifact_dir / "extraction").mkdir(parents=True, exist_ok=True)
    (artifact_dir / "manifest.json").write_text(json.dumps({}), encoding="utf-8")
    (artifact_dir / "inspection.json").write_text(json.dumps({}), encoding="utf-8")
    (artifact_dir / "extraction" / "001.txt").write_text(extract_text, encoding="utf-8")


def test_retry_fixes_unreadable_period():
    base = Path(tempfile.mkdtemp(prefix="retry-"))
    try:
        make_artifact(base, "r", WITHOUT_PERIOD)
        calls = {"n": 0}

        def extractor(pdf_path):
            # The second read of the same PDF finds the header that the
            # partial first pass missed.
            calls["n"] += 1
            make_artifact(base, "r", WITH_PERIOD)
            return {"pages": [{"page_number": 1, "text": WITH_PERIOD}]}

        doc = parse_with_retry("dummy.pdf", str(base), "r", {}, extractor)
        assert calls["n"] == 1, f"expected exactly one silent retry, got {calls['n']}"
        assert len(doc.document_review) == 0, "the retry read the period: nothing is parked"
        assert doc.metadata.get("periodo") == "01 - 30 ABR 2026"
    finally:
        shutil.rmtree(base, ignore_errors=True)


def test_retry_failure_keeps_document_review():
    base = Path(tempfile.mkdtemp(prefix="retry2-"))
    try:
        make_artifact(base, "r", WITHOUT_PERIOD)
        calls = {"n": 0}

        def extractor(pdf_path):
            calls["n"] += 1
            make_artifact(base, "r", WITHOUT_PERIOD)
            return {"pages": [{"page_number": 1, "text": WITHOUT_PERIOD}]}

        doc = parse_with_retry("dummy.pdf", str(base), "r", {}, extractor)
        assert calls["n"] == 1, "the retry happens exactly once"
        assert len(doc.document_review) == 1, (
            "a failed retry leaves the document review parked"
        )
        assert doc.document_review[0].raw_period == ""
    finally:
        shutil.rmtree(base, ignore_errors=True)


def test_no_retry_when_period_readable():
    base = Path(tempfile.mkdtemp(prefix="retry3-"))
    try:
        make_artifact(base, "r", WITH_PERIOD)

        def extractor(pdf_path):
            raise AssertionError("must not re-extract when the period read fine")

        doc = parse_with_retry("dummy.pdf", str(base), "r", {}, extractor)
        assert len(doc.document_review) == 0
        assert len(doc.movements) == 1
    finally:
        shutil.rmtree(base, ignore_errors=True)


if __name__ == "__main__":
    tests = [
        test_retry_fixes_unreadable_period,
        test_retry_failure_keeps_document_review,
        test_no_retry_when_period_readable,
    ]
    for t in tests:
        t()
        print(f"ok  {t.__name__}")
    print("Retry tests passed")

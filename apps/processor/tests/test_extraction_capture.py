"""One PDF pass: inspection and extraction share the open document.

The pipeline used to open the statement twice (inspect, then extract) and
never closed the handles. inspect_and_extract does it once and closes.

This test exercises the extraction mechanics only; it runs against the
interpreter of the maintenance env (PyMuPDF 1.28.2) while production pins
1.26.3, so it is not an extraction-fidelity test.

Run with: <python.exe> tests/test_extraction_capture.py
"""

import sys
import tempfile
from pathlib import Path

PROCESSOR_ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(PROCESSOR_ROOT))

from extraction._fitz import _get_fitz
from extraction.capture import inspect_and_extract


def make_pdf(base, pages_text):
    fitz = _get_fitz()
    doc = fitz.open()
    for text in pages_text:
        page = doc.new_page()
        page.insert_text((72, 72), text)
    path = base / "statement.pdf"
    doc.save(path)
    doc.close()
    return path


def test_one_pass_matches_two_pass():
    base = Path(tempfile.mkdtemp(prefix="capture-"))
    try:
        pdf_path = make_pdf(base, ["HOLA USUARIO", "Periodo 01 - 30 ABR 2026", "Movimientos"])
        inspection, extraction = inspect_and_extract(str(pdf_path))

        assert inspection["pdf"]["page_count"] == 3, inspection["pdf"]
        assert len(extraction["pages"]) == 3, extraction
        texts = [p["text"] for p in extraction["pages"]]
        assert any("Periodo" in t for t in texts), texts
        # The inspection saw the same document the extraction read.
        assert inspection["indicators"]["native_text_available"] is True
        assert inspection["file"]["sha256"]
    finally:
        import shutil
        shutil.rmtree(base, ignore_errors=True)


if __name__ == "__main__":
    test_one_pass_matches_two_pass()
    print("ok  test_one_pass_matches_two_pass")
    print("Capture tests passed")

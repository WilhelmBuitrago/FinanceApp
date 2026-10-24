"""One PDF pass for the whole extraction stage.

inspect_pdf and extract_native_text each opened the statement (the
pipeline called them back to back, so every statement was read twice) and
neither closed the handle. inspect_and_extract shares one open document
between both and closes it when done. The individual functions remain for
callers that need only one half (the silent retry re-extracts alone).
"""

from extraction._fitz import _get_fitz
from extraction.extractor import extract_pages
from extraction.inspection import inspect_document


def inspect_and_extract(path):
    """Inspect and extract in a single pass over the statement."""
    fitz = _get_fitz()
    doc = fitz.open(path)
    try:
        inspection = inspect_document(doc, path)
        extraction = extract_pages(doc)
        return inspection, extraction
    finally:
        doc.close()

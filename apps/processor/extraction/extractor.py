import os
import sys
from io import StringIO

def _get_fitz():
    old_stdout = sys.stdout
    old_stderr = sys.stderr
    sys.stdout = StringIO()
    sys.stderr = StringIO()
    try:
        import fitz
    finally:
        sys.stdout = old_stdout
        sys.stderr = old_stderr
    return fitz

def extract_native_text(path):
    fitz = _get_fitz()
    doc = fitz.open(path)
    pages = []
    for i in range(doc.page_count):
        page = doc[i]
        text = page.get_text()
        pages.append({"page_number": i+1, "text": text})
    return {"pages": pages}

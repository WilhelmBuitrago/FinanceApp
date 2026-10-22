"""Shared PyMuPDF importer.

pymupdf prints to stdout/stderr on import (deprecation warnings), which would
corrupt the JSON contract this processor writes to stdout. Every module that
needs fitz imports it from here so the safeguard exists exactly once.
"""

import sys
from io import StringIO


def _get_fitz():
    # Suppress deprecation warning from fitz -> pymupdf which prints to stdout
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

"""Shared PyMuPDF importer.

pymupdf prints to stdout/stderr on import (deprecation warnings), which would
corrupt the JSON contract this processor writes to stdout. Every module that
needs fitz imports it from here so the safeguard exists exactly once.
"""

import sys
from io import StringIO


class PasswordProtectedPDF(Exception):
    """The user, not the code, holds the key: the statement is locked and
    the extraction cannot read a single page. The name is the message:
    the processor prints it instead of a traceback."""


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


def open_pdf(path):
    """Open a PDF for reading, naming the one common condition the user
    can fix: a protected file. fitz.open itself succeeds on a locked
    document; the pages refuse afterwards, so the check belongs here,
    before any page work."""
    fitz = _get_fitz()
    doc = fitz.open(path)
    if doc.needs_pass:
        doc.close()
        raise PasswordProtectedPDF(
            f"el PDF {path} está protegido con contraseña: "
            "ábrelo y guárdalo sin protección para procesarlo"
        )
    return doc

"""A password-protected statement used to surface as a raw traceback:
the user saw a wall of frames instead of 'the file is locked'. The
reader now names the condition, and the processor prints that name."""

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from extraction._fitz import PasswordProtectedPDF, open_pdf


class FakeDoc:
    def __init__(self, needs_pass):
        self.needs_pass = needs_pass
        self.closed = False

    def close(self):
        self.closed = True


class FakeFitz:
    def __init__(self, needs_pass):
        self._doc = FakeDoc(needs_pass)

    def open(self, path):
        return self._doc


def test_open_pdf_rejects_a_locked_file_with_a_named_error(monkeypatch=None):
    import extraction._fitz as fitzmod

    locked = FakeFitz(True)
    original_get = fitzmod._get_fitz
    fitzmod._get_fitz = lambda: locked
    try:
        try:
            open_pdf("locked.pdf")
        except PasswordProtectedPDF as exc:
            assert "contraseña" in str(exc).lower() or "password" in str(exc).lower()
            assert locked._doc.closed, "the locked handle must not stay open"
        else:
            raise AssertionError("a locked PDF must raise PasswordProtectedPDF")
    finally:
        fitzmod._get_fitz = original_get


def test_open_pdf_returns_a_readable_document():
    import extraction._fitz as fitzmod

    readable = FakeFitz(False)
    original_get = fitzmod._get_fitz
    fitzmod._get_fitz = lambda: readable
    try:
        doc = open_pdf("plain.pdf")
        assert doc is readable._doc
    finally:
        fitzmod._get_fitz = original_get


if __name__ == "__main__":
    test_open_pdf_rejects_a_locked_file_with_a_named_error()
    test_open_pdf_returns_a_readable_document()
    print("password tests passed")

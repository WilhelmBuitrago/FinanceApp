"""Failure modes of the artifact writer.

The backend reads manifest.json as the job's terminal result. A partial
manifest (a crash mid-write) would be read as a valid artifact, and a run
that dies mid-pipeline would leave its half-written artifacts on the
user's disk with nothing referencing them.

Only the PDF read is stubbed; the artifacts, the parse, the validation
and the writes all run for real.

Run with: <python.exe> tests/test_artifact_failure_modes.py
"""

import json
import os
import sys
import tempfile
from pathlib import Path
from types import SimpleNamespace

PROCESSOR_ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(PROCESSOR_ROOT))

from extraction.artifacts import write_artifacts
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


class Boom(Exception):
    pass


class ExplodingPage(dict):
    """A page whose text explodes when written: simulates the disk going
    away in the middle of one extraction page."""

    def get(self, key, default=None):
        if key == "text":
            raise Boom("disk went away mid-page")
        return super().get(key, default)


def _explode(*_args, **_kwargs):
    raise Boom("injected failure")


def _run_pipeline(base):
    """The only stubbed boundary: reading the PDF."""
    inspection = {"file": {"sha256": "fm", "size_bytes": 1}, "warnings": []}
    extraction = {"pages": [{"page_number": 1, "text": WITH_PERIOD}]}
    main.inspect_and_extract = lambda pdf_path: (inspection, extraction)

    previous = os.environ.get("ARTIFACTS_PATH")
    os.environ["ARTIFACTS_PATH"] = str(base)
    try:
        return main.process_pdf("fm.pdf")
    finally:
        if previous is None:
            os.environ.pop("ARTIFACTS_PATH", None)
        else:
            os.environ["ARTIFACTS_PATH"] = previous


def _contents(path):
    return sorted(p.name for p in Path(path).iterdir())


def test_success_writes_complete_json_with_no_temporary_leftovers():
    base = Path(tempfile.mkdtemp(prefix="artifacts-ok-"))
    manifest = _run_pipeline(base)

    artifact_dir = base / manifest["processing_id"]
    leftovers = [p.name for p in artifact_dir.iterdir() if p.name.endswith(".tmp")]
    if leftovers:
        raise AssertionError(f"temporary files survived the write: {leftovers}")
    for name in ("manifest.json", "parsed.json"):
        # A truncated file would fail to parse here.
        json.loads((artifact_dir / name).read_text(encoding="utf-8"))


def test_a_failed_manifest_write_leaves_nothing_behind():
    base = Path(tempfile.mkdtemp(prefix="artifacts-boom-"))
    real_dump = json.dump

    calls = {"n": 0}

    def fail_at_manifest(*args, **kwargs):
        # write_artifacts dumps inspection.json first (call 1); the manifest
        # is the pipeline's second dump.
        calls["n"] += 1
        if calls["n"] >= 2:
            raise Boom("disk went away mid-write")
        return real_dump(*args, **kwargs)

    json.dump = fail_at_manifest
    raised = None
    try:
        _run_pipeline(base)
    except Boom as exc:
        raised = exc
    finally:
        json.dump = real_dump

    if raised is None:
        raise AssertionError("the injected failure did not propagate")
    if calls["n"] < 2:
        raise AssertionError("the failure never reached the manifest write")
    if _contents(base):
        raise AssertionError(
            f"a partial run survived the failure: {_contents(base)}"
        )


def test_a_failing_parse_leaves_no_partial_run():
    base = Path(tempfile.mkdtemp(prefix="artifacts-parse-"))
    real_retry = main.parse_with_retry
    main.parse_with_retry = _explode
    raised = None
    try:
        _run_pipeline(base)
    except Boom:
        raised = True
    finally:
        main.parse_with_retry = real_retry

    if not raised:
        raise AssertionError("the injected failure did not propagate")
    if _contents(base):
        raise AssertionError(
            f"a partial run survived the failure: {_contents(base)}"
        )


def test_a_failing_inspection_creates_nothing():
    base = Path(tempfile.mkdtemp(prefix="artifacts-inspect-"))
    main.inspect_and_extract = _explode
    raised = None
    try:
        main.process_pdf("bad.pdf")
    except Boom:
        raised = True
    finally:
        main.inspect_and_extract = lambda pdf_path: (
            {"file": {"sha256": "fm", "size_bytes": 1}, "warnings": []},
            {"pages": [{"page_number": 1, "text": WITH_PERIOD}]},
        )

    if not raised:
        raise AssertionError("the injected failure did not propagate")
    if _contents(base):
        raise AssertionError(
            f"an artifact directory was created before the failure: {_contents(base)}"
        )


# ---- P1-bis: the intermediate artifacts are atomic too ----


def test_write_artifacts_success_leaves_no_temporary_anywhere():
    base = Path(tempfile.mkdtemp(prefix="artifacts-writer-"))
    write_artifacts(
        str(base),
        {"file": {"sha256": "w", "size_bytes": 1}, "warnings": []},
        {"pages": [{"page_number": 1, "text": "page one"}, {"page_number": 2, "text": "page two"}]},
    )

    extraction_dir = base / "extraction"
    for name in ("inspection.json", "extraction/001.txt", "extraction/002.txt"):
        if not (base / name).exists():
            raise AssertionError(f"missing artifact: {name}")
    leftovers = [str(p.relative_to(base)) for p in base.rglob("*.tmp")]
    if leftovers:
        raise AssertionError(f"temporary files survived the write: {leftovers}")
    page_one = extraction_dir / "001.txt"
    if page_one.read_text(encoding="utf-8") != "page one":
        raise AssertionError("the page content did not round-trip")


def test_a_page_write_failure_never_leaves_a_partial_page():
    base = Path(tempfile.mkdtemp(prefix="artifacts-page-"))
    raised = None
    try:
        write_artifacts(
            str(base),
            {"file": {"sha256": "w", "size_bytes": 1}, "warnings": []},
            {"pages": [{"page_number": 1, "text": "page one"}, ExplodingPage(page_number=2)]},
        )
    except Boom:
        raised = True

    if not raised:
        raise AssertionError("the injected failure did not propagate")
    extraction_dir = base / "extraction"
    if not (extraction_dir / "001.txt").exists():
        raise AssertionError("the page written before the failure was lost")
    if (extraction_dir / "001.txt").read_text(encoding="utf-8") != "page one":
        raise AssertionError("the earlier page was corrupted by the failure")
    if (extraction_dir / "002.txt").exists():
        raise AssertionError(
            "a partial page reads as complete: 002.txt exists after the failed write"
        )
    # The inspection was written before the pages and stays complete.
    json.loads((base / "inspection.json").read_text(encoding="utf-8"))


def test_a_failing_page_inside_the_pipeline_leaves_no_run():
    base = Path(tempfile.mkdtemp(prefix="artifacts-pipeline-page-"))
    main.inspect_and_extract = lambda pdf_path: (
        {"file": {"sha256": "fm", "size_bytes": 1}, "warnings": []},
        {"pages": [{"page_number": 1, "text": WITH_PERIOD}, ExplodingPage(page_number=2)]},
    )

    previous = os.environ.get("ARTIFACTS_PATH")
    os.environ["ARTIFACTS_PATH"] = str(base)
    raised = None
    try:
        main.process_pdf("fm.pdf")
    except Boom:
        raised = True
    finally:
        if previous is None:
            os.environ.pop("ARTIFACTS_PATH", None)
        else:
            os.environ["ARTIFACTS_PATH"] = previous

    if not raised:
        raise AssertionError("the injected failure did not propagate")
    if _contents(base):
        raise AssertionError(f"a partial run survived the failure: {_contents(base)}")


def test_the_retry_cleanup_removes_stale_tmp_files():
    base = Path(tempfile.mkdtemp(prefix="artifacts-retry-tmp-"))
    processing_id = "retry-tmp"
    artifact_dir = base / processing_id
    extraction_dir = artifact_dir / "extraction"
    extraction_dir.mkdir(parents=True)
    (extraction_dir / "001.txt").write_text("stale page", encoding="utf-8")
    (extraction_dir / "001.txt.tmp").write_text("interrupted page", encoding="utf-8")

    # First parse parks a document review (period unreadable) and forces
    # the retry; the second is clean.
    class FakeParser:
        calls = 0

        def __init__(self, processing_id, base_path=None):
            self.processing_id = processing_id

        def parse(self):
            FakeParser.calls += 1
            if FakeParser.calls == 1:
                return SimpleNamespace(document_review=["period unreadable"])
            return SimpleNamespace(document_review=[])

    real_parser = main.NuAccountParser
    real_native = main.extract_native_text
    main.NuAccountParser = FakeParser
    main.extract_native_text = lambda pdf_path: {
        "pages": [{"page_number": 1, "text": "fresh page"}]
    }
    raised = None
    try:
        doc = main.parse_with_retry("fm.pdf", str(base), processing_id, {}, None)
    except Exception as exc:  # noqa: BLE001 — the assertion below reports the real cause
        raised = exc
    finally:
        main.NuAccountParser = real_parser
        main.extract_native_text = real_native

    if raised is not None:
        raise AssertionError(f"the retry raised instead of recovering: {raised!r}")
    if doc.document_review:
        raise AssertionError("the retry did not recover: the document review survived it")
    names = sorted(p.name for p in extraction_dir.iterdir())
    if "001.txt.tmp" in names:
        raise AssertionError(f"the retry cleanup left a stale temporary: {names}")
    if "001.txt" not in names:
        raise AssertionError(f"the retry did not rewrite the fresh page: {names}")
    if (extraction_dir / "001.txt").read_text(encoding="utf-8") != "fresh page":
        raise AssertionError("the rewritten page is not the fresh one")


if __name__ == "__main__":
    test_success_writes_complete_json_with_no_temporary_leftovers()
    test_a_failed_manifest_write_leaves_nothing_behind()
    test_a_failing_parse_leaves_no_partial_run()
    test_a_failing_inspection_creates_nothing()
    test_write_artifacts_success_leaves_no_temporary_anywhere()
    test_a_page_write_failure_never_leaves_a_partial_page()
    test_a_failing_page_inside_the_pipeline_leaves_no_run()
    test_the_retry_cleanup_removes_stale_tmp_files()
    print("Artifact failure-mode tests passed")

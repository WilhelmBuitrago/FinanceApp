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


class Boom(Exception):
    pass


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


if __name__ == "__main__":
    test_success_writes_complete_json_with_no_temporary_leftovers()
    test_a_failed_manifest_write_leaves_nothing_behind()
    test_a_failing_parse_leaves_no_partial_run()
    test_a_failing_inspection_creates_nothing()
    print("Artifact failure-mode tests passed")

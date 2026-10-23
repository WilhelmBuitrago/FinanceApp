import os
import sys
import uuid
import warnings
import logging
from pathlib import Path

os.environ['PYTHONWARNINGS'] = 'ignore'
warnings.filterwarnings('ignore')
logging.basicConfig(level=logging.INFO, format='%(asctime)s [%(levelname)s] %(message)s', stream=sys.stderr)
log = logging.getLogger('processor')
# Normalize path and add processor root to sys.path
processor_root = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(processor_root))
from extraction.inspection import inspect_pdf
from extraction.extractor import extract_native_text
from extraction.artifacts import write_artifacts
from parser.nu_account import NuAccountParser
from validation.validator import NuAccountValidator

PROCESSOR_VERSION = "0.1.0"
CONTRACT_VERSION = "1"


def parse_with_retry(pdf_path, artifacts_base, processing_id, inspection, extract_text=None):
    """Parse the extraction, silently retrying once when the statement
    period is unreadable.

    A partial write or a failed read can leave the text incomplete:
    re-extracting the PDF and parsing again usually fixes it. The retry is
    internal: the user is never told it happened. Only a retry that also
    fails surfaces, as the document review the parser already emits.
    """
    extractor = extract_text or extract_native_text
    parser = NuAccountParser(processing_id, base_path=artifacts_base)
    doc = parser.parse()
    if not doc.document_review:
        return doc

    log.warning(f"period unreadable on first pass; re-extracting once pdf={pdf_path}")
    extraction_dir = os.path.join(artifacts_base, processing_id, 'extraction')
    if os.path.isdir(extraction_dir):
        for stale in os.listdir(extraction_dir):
            if stale.endswith('.txt'):
                os.remove(os.path.join(extraction_dir, stale))
    extraction = extractor(pdf_path)
    write_artifacts(os.path.join(artifacts_base, processing_id), inspection, extraction)

    doc = parser.parse()
    if doc.document_review:
        log.warning("period still unreadable after retry; parking document review")
    return doc


def process_pdf(pdf_path):
    processing_id = str(uuid.uuid4())
    log.info(f"process_pdf start pdf={pdf_path} processing_id={processing_id}")
    log.info(f"inspection start")
    inspection = inspect_pdf(pdf_path)
    log.info(f"inspection done sha256={inspection['file']['sha256']}")
    log.info(f"extraction start")
    extraction = extract_native_text(pdf_path)
    log.info(f"extraction done")
    artifacts_base = os.environ.get('ARTIFACTS_PATH')
    if not artifacts_base:
        base = Path(__file__).resolve().parent.parent.parent
        artifacts_base = str(base / "data" / "artifacts")
    artifact_dir = os.path.join(artifacts_base, processing_id)
    os.makedirs(artifact_dir, exist_ok=True)
    log.info(f"writing artifacts to {artifact_dir}")
    write_artifacts(artifact_dir, inspection, extraction)

    # Create minimal manifest so parser can load it
    import json as _json
    minimal_manifest = {"processing_id": processing_id, "input": {"sha256": inspection["file"]["sha256"]}}
    with open(os.path.join(artifact_dir, "manifest.json"), "w", encoding="utf-8") as f:
        _json.dump(minimal_manifest, f)
    log.info(f"minimal manifest written")

    # Parsing & validation for Nu account statements. The silent retry
    # (Layer 0) lives inside: the user only sees the outcome.
    log.info(f"parser start")
    parsed_doc = parse_with_retry(dstPath, artifacts_base, processing_id, inspection)
    log.info(f"parser done movements={len(parsed_doc.movements)}")
    log.info(f"validation start")
    validator = NuAccountValidator()
    validation_result = validator.validate(parsed_doc)
    log.info(f"validation done valid={validation_result.get('valid')}")

    manifest = {
        "processing_id": processing_id,
        "processor_version": PROCESSOR_VERSION,
        "contract_version": CONTRACT_VERSION,
        "input": {
            "sha256": inspection["file"]["sha256"],
            "filename": os.path.basename(pdf_path),
            "size_bytes": inspection["file"]["size_bytes"]
        },
        "stages": {
            "inspection": "completed",
            "native_extraction": "completed",
            "parsing": "completed",
            "validation": "completed"
        },
        "status": "completed" if validation_result.get('valid') else "validation_failed",
        "warnings": inspection.get("warnings", []),
        "errors": [],
        "validation": validation_result,
        "parsed_metadata": {
            "titular": parsed_doc.metadata.get('titular'),
            "cc": parsed_doc.metadata.get('cc'),
            "numero_cuenta": parsed_doc.metadata.get('numero_cuenta'),
            "periodo": parsed_doc.metadata.get('periodo')
        },
        "movements_count": len(parsed_doc.movements)
    }
    import json
    with open(os.path.join(artifact_dir, "manifest.json"), "w", encoding="utf-8") as f:
        json.dump(manifest, f, ensure_ascii=False, indent=2)
    log.info(f"manifest written")
    # Save parsed document for later use
    with open(os.path.join(artifact_dir, "parsed.json"), "w", encoding="utf-8") as f:
        json.dump({
            "metadata": parsed_doc.metadata,
            "summary": parsed_doc.summary,
            "movements": [m.__dict__ for m in parsed_doc.movements],
            "review_required": [r.__dict__ for r in parsed_doc.review_required],
            "document_review": [d.__dict__ for d in parsed_doc.document_review]
        }, f, ensure_ascii=False, indent=2)
    log.info(f"parsed.json written movements={len(parsed_doc.movements)}")
    log.info(f"process_pdf complete")
    return manifest

if __name__ == "__main__":
    import sys
    import json
    if len(sys.argv) < 2:
        sys.stderr.write("usage: python apps/processor/src/main.py <pdf_path>\n")
        sys.exit(1)
    try:
        m = process_pdf(sys.argv[1])
        sys.stdout.write(json.dumps(m, ensure_ascii=False))
    except Exception as e:
        log.error(f"Processor error: {e}", exc_info=True)
        sys.stderr.write(f"Processor error: {e}\n")
        sys.exit(1)

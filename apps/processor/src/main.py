import os
import sys
import uuid
import warnings

os.environ['PYTHONWARNINGS'] = 'ignore'
warnings.filterwarnings('ignore')
# Make project root importable for apps.processor.*
project_root = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..', '..'))
sys.path.insert(0, project_root)
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))
from extraction.inspection import inspect_pdf
from extraction.extractor import extract_native_text
from extraction.artifacts import write_artifacts
from parser.nu_account import NuAccountParser
from validation.validator import NuAccountValidator

PROCESSOR_VERSION = "0.1.0"
CONTRACT_VERSION = "1"

def process_pdf(pdf_path):
    processing_id = str(uuid.uuid4())
    inspection = inspect_pdf(pdf_path)
    extraction = extract_native_text(pdf_path)
    base = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..', '..'))
    artifact_dir = os.path.join(base, "data", "artifacts", processing_id)
    write_artifacts(artifact_dir, inspection, extraction)

    # Create minimal manifest so parser can load it
    import json as _json
    minimal_manifest = {"processing_id": processing_id, "input": {"sha256": inspection["file"]["sha256"]}}
    with open(os.path.join(artifact_dir, "manifest.json"), "w", encoding="utf-8") as f:
        _json.dump(minimal_manifest, f)

    # Parsing & validation for Nu account statements
    parser = NuAccountParser(processing_id)
    parsed_doc = parser.parse()
    validator = NuAccountValidator()
    validation_result = validator.validate(parsed_doc)

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
    # Save parsed document for later use
    with open(os.path.join(artifact_dir, "parsed.json"), "w", encoding="utf-8") as f:
        json.dump({
            "metadata": parsed_doc.metadata,
            "summary": parsed_doc.summary,
            "movements": [m.__dict__ for m in parsed_doc.movements]
        }, f, ensure_ascii=False, indent=2)
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
        sys.stderr.write(f"Processor error: {e}\n")
        sys.exit(1)

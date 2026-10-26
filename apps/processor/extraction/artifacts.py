import json
import os


def write_json_atomic(path, payload):
    """Write JSON beside its target, then move it into place.

    A reader (the backend polling the manifest as the job's result) must
    never observe a partial file: an interrupted plain write leaves a
    truncated JSON that still parses as "an artifact". The move is a
    same-volume rename: atomic on Windows too.
    """
    tmp_path = path + ".tmp"
    with open(tmp_path, "w", encoding="utf-8") as f:
        json.dump(payload, f, ensure_ascii=False, indent=2)
    os.replace(tmp_path, path)


def write_artifacts(base_dir, inspection, extraction):
    os.makedirs(base_dir, exist_ok=True)
    os.makedirs(os.path.join(base_dir, "extraction"), exist_ok=True)
    with open(os.path.join(base_dir, "inspection.json"), "w", encoding="utf-8") as f:
        json.dump(inspection, f, ensure_ascii=False, indent=2)
    for page in extraction.get("pages", []):
        fname = f"{page['page_number']:03d}.txt"
        with open(os.path.join(base_dir, "extraction", fname), "w", encoding="utf-8") as f:
            f.write(page.get("text", ""))

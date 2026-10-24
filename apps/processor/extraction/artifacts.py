import json
import os


def write_artifacts(base_dir, inspection, extraction):
    os.makedirs(base_dir, exist_ok=True)
    os.makedirs(os.path.join(base_dir, "extraction"), exist_ok=True)
    with open(os.path.join(base_dir, "inspection.json"), "w", encoding="utf-8") as f:
        json.dump(inspection, f, ensure_ascii=False, indent=2)
    for page in extraction.get("pages", []):
        fname = f"{page['page_number']:03d}.txt"
        with open(os.path.join(base_dir, "extraction", fname), "w", encoding="utf-8") as f:
            f.write(page.get("text", ""))

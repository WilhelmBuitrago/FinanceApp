import hashlib
import os

from extraction._fitz import _get_fitz


def sha256_file(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(8192), b""):
            h.update(chunk)
    return h.hexdigest()


def inspect_document(doc, path):
    """Inspection over an already-open document: the caller owns the
    handle (one pass shared with the extraction)."""
    size_bytes = os.path.getsize(path)
    sha256 = sha256_file(path)
    version = doc.metadata.get("format", "")
    page_count = doc.page_count
    pages = []
    for i in range(page_count):
        page = doc[i]
        rect = page.rect
        text = page.get_text()
        image_count = len(page.get_images())
        pages.append({
            "page_number": i+1,
            "width": rect.width,
            "height": rect.height,
            "text_character_count": len(text),
            "image_count": image_count
        })
    native_text_available = any(p["text_character_count"] > 0 for p in pages)
    image_content_available = any(p["image_count"] > 0 for p in pages)
    mixed_content = native_text_available and image_content_available
    warnings = []
    if not native_text_available:
        warnings.append("PAGE_WITHOUT_NATIVE_TEXT")
    if mixed_content:
        warnings.append("MIXED_CONTENT")
    return {
        "file": {"name": os.path.basename(path), "size_bytes": size_bytes, "sha256": sha256},
        "pdf": {"page_count": page_count, "version": version},
        "pages": pages,
        "indicators": {
            "native_text_available": native_text_available,
            "image_content_available": image_content_available,
            "mixed_content": mixed_content
        },
        "warnings": warnings
    }


def inspect_pdf(path):
    fitz = _get_fitz()
    doc = fitz.open(path)
    try:
        return inspect_document(doc, path)
    finally:
        doc.close()

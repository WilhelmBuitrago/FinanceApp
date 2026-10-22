from extraction._fitz import _get_fitz

def extract_native_text(path):
    fitz = _get_fitz()
    doc = fitz.open(path)
    pages = []
    for i in range(doc.page_count):
        page = doc[i]
        text = page.get_text()
        pages.append({"page_number": i+1, "text": text})
    return {"pages": pages}

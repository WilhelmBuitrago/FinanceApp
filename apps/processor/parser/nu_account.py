"""Parser for Nu Account statements from processor artefacts."""

import json
import os
import re
from dataclasses import dataclass
from typing import List, Dict

from normalization.nu_normalizer import normalize_date, normalize_amount


@dataclass
class ParsedMovement:
    date_original: str
    date_normalized: str
    description_original: str
    amount_original: str
    amount_normalized: int
    page_number: int
    sequence: int


@dataclass
class ParsedNuAccountDocument:
    metadata: Dict
    summary: Dict
    movements: List[ParsedMovement]
    review_required: List['ReviewItem'] = None


@dataclass
class ReviewItem:
    """A movement the pipeline could not normalize: the raw is preserved so a
    human can verify it against the PDF instead of trusting the parser."""
    date_original: str
    description_original: str
    amount_original: str
    page_number: int
    reason: str


class NuAccountParser:
    START_MARKER = "Movimientos"
    END_MARKER = "Rendimiento total de tu cuenta"
    FOOTER_MARKER = "Los rendimientos que genera tu cuenta los pagamos a diario"

    def __init__(self, processing_id: str, base_path: str = None):
        if base_path is None:
            base_path = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..', '..', 'data', 'artifacts'))
        self.processing_id = processing_id
        self.base_path = base_path
        self.artifact_dir = os.path.join(base_path, processing_id)

    def _load_json(self, filename: str) -> dict:
        path = os.path.join(self.artifact_dir, filename)
        with open(path, 'r', encoding='utf-8') as f:
            return json.load(f)

    def _load_extraction_pages(self) -> List[str]:
        extraction_dir = os.path.join(self.artifact_dir, 'extraction')
        files = sorted([f for f in os.listdir(extraction_dir) if f.endswith('.txt')])
        pages = []
        for fname in files:
            page_num = int(re.search(r'(\d+)', fname).group(1))
            path = os.path.join(extraction_dir, fname)
            with open(path, 'r', encoding='utf-8') as f:
                text = f.read()
            pages.append((page_num, text))
        pages.sort(key=lambda x: x[0])
        return pages

    def _parse_metadata(self, first_page_text: str) -> Dict:
        lines = [l.strip() for l in first_page_text.splitlines() if l.strip()]
        metadata = {
            'titular': None,
            'cc': None,
            'placa': None,
            'numero_cuenta': None,
            'periodo': None,
        }

        for i, line in enumerate(lines):
            if re.search(r'\bCC[-\s]?', line, re.I):
                m = re.search(r'CC[-\s]?(\d+)', line, re.I)
                if m:
                    metadata['cc'] = m.group(1)
            if re.search(r'Número de Cuenta', line, re.I):
                if i + 1 < len(lines):
                    metadata['numero_cuenta'] = lines[i+1].strip()
            if re.search(r'Nu Placa', line, re.I) or re.search(r'Placa', line, re.I):
                # Placa value is next line
                if i + 1 < len(lines):
                    # Avoid capturing label itself
                    candidate = lines[i+1].strip()
                    if not re.search(r'(Placa|CC|Número|Período)', candidate, re.I):
                        metadata['placa'] = candidate
            if re.search(r'Período', line, re.I):
                if i + 1 < len(lines):
                    metadata['periodo'] = lines[i+1].strip()

        # Titular heuristic: find first line after greeting that looks like a person name
        # Name: 2-4 words, only letters/accent, no digits, not containing keywords
        name_keywords = ['hola', 'llegó', 'extracto', 'tarjeta', 'resumen', 'dinero', 'rendimientos', 'nu', 'financiera', 'calle', 'nit', 'si solo', 'disponible']
        for i, line in enumerate(lines):
            if line.lower().startswith('hola'):
                # search next few lines
                for j in range(i+1, min(i+6, len(lines))):
                    candidate = lines[j]
                    # skip keywords
                    low = candidate.lower()
                    if any(k in low for k in name_keywords):
                        continue
                    if re.search(r'\d', candidate):
                        continue
                    # must have at least two words
                    if len(candidate.split()) >= 2 and re.match(r'^[A-Za-zÁÉÍÓÚáéíóúÑñ\s]+$', candidate):
                        metadata['titular'] = candidate
                        break
                if metadata['titular']:
                    break
        # Fallback to line 3
        if not metadata['titular'] and len(lines) > 2:
            metadata['titular'] = lines[2]

        return metadata

    def parse(self) -> ParsedNuAccountDocument:
        manifest = self._load_json('manifest.json')
        inspection = self._load_json('inspection.json')
        pages = self._load_extraction_pages()
        if not pages:
            raise ValueError("No extraction pages found")

        first_page_text = pages[0][1]
        metadata = self._parse_metadata(first_page_text)

        movements: List[ParsedMovement] = []
        review_required: List[ReviewItem] = []
        inside_movements = False
        sequence = 0

        for page_number, text in pages:
            lines = text.splitlines()
            for idx, raw_line in enumerate(lines):
                line = raw_line.strip()
                if not line:
                    continue

                # Footer ignore
                if self.FOOTER_MARKER.lower() in line.lower():
                    continue

                # Start marker
                if self.START_MARKER.lower() in line.lower():
                    inside_movements = True
                    continue

                # End marker
                if self.END_MARKER.lower() in line.lower():
                    inside_movements = False
                    continue

                if not inside_movements:
                    continue

                # Movement parsing: date line pattern dd mmm
                date_match = re.match(r'^(\d{1,2}\s*[A-Za-z]{3})$', line, re.I)
                if date_match:
                    date_original = line
                    # Next lines: description and amount
                    description_parts = []
                    amount_original = None
                    # Look ahead within same page
                    look_ahead_idx = idx + 1
                    # Collect description until amount line found
                    while look_ahead_idx < len(lines):
                        next_line = lines[look_ahead_idx].strip()
                        # Stop if next date line or marker
                        if re.match(r'^(\d{1,2}\s*[A-Za-z]{3})$', next_line, re.I):
                            break
                        if self.START_MARKER.lower() in next_line.lower() or self.END_MARKER.lower() in next_line.lower():
                            break
                        # Amount pattern
                        if re.search(r'^[+-]\$\s*\d', next_line):
                            amount_original = next_line
                            look_ahead_idx += 1
                            break
                        # Accumulate description
                        if next_line:
                            description_parts.append(next_line)
                        look_ahead_idx += 1

                    if amount_original:
                        description_original = ' '.join(description_parts).strip()
                        # Normalize
                        try:
                            date_normalized = normalize_date(metadata.get('periodo') or '', date_original)
                        except Exception:
                            date_normalized = ''
                        try:
                            amount_normalized = normalize_amount(amount_original)
                        except Exception as exc:
                            # Not importable and not droppable: keep the raw
                            # for human review with the exact reason.
                            review_required.append(ReviewItem(
                                date_original=date_original,
                                description_original=' '.join(description_parts).strip(),
                                amount_original=amount_original,
                                page_number=page_number,
                                reason=f"unparseable amount: {exc}",
                            ))
                            continue
                        sequence += 1
                        movements.append(ParsedMovement(
                            date_original=date_original,
                            date_normalized=date_normalized,
                            description_original=description_original,
                            amount_original=amount_original,
                            amount_normalized=amount_normalized,
                            page_number=page_number,
                            sequence=sequence
                        ))
                    # Continue loop, don't process description lines again
                    continue

        summary = {
            'processing_id': self.processing_id,
            'page_count': len(pages),
            'movements_count': len(movements),
            'movements_need_review': len(review_required),
        }

        return ParsedNuAccountDocument(
            metadata=metadata,
            summary=summary,
            movements=movements,
            review_required=review_required,
        )

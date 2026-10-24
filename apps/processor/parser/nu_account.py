"""Parser for Nu Account statements from processor artefacts."""

import json
import os
import re
from dataclasses import dataclass

from normalization.nu_normalizer import normalize_amount, normalize_date


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
    metadata: dict
    summary: dict
    movements: list[ParsedMovement]
    review_required: list['ReviewItem'] = None
    document_review: list['DocumentReviewItem'] = None


@dataclass
class ReviewItem:
    """A movement the pipeline could not normalize: the raw is preserved so a
    human can verify it against the PDF instead of trusting the parser."""
    date_original: str
    description_original: str
    amount_original: str
    page_number: int
    reason: str
    # Present when the amount normalized fine: a movement parked for its
    # date travels with its amount so the document-period resolution can
    # import it as soon as the year is confirmed.
    amount_normalized: int = 0


@dataclass
class DocumentReviewItem:
    """The document-level review parking: the statement period could not be
    read from the extraction. The raw is preserved for the human."""
    raw_period: str
    reason: str


class NuAccountParser:
    START_MARKER = "Movimientos"
    END_MARKER = "Rendimiento total de tu cuenta"
    FOOTER_MARKER = "Los rendimientos que genera tu cuenta los pagamos a diario"

    # Months accepted on a date line: a bare "12 cuotas" is a description,
    # not a date, so a valid month is required.
    _MONTH_NAMES = (
        'ene|feb|mar|abr|may|jun|jul|ago|sep|set|oct|nov|dic'
        '|jan|apr|aug|dec'
    )
    # Tolerant date: the strict shape plus any trailing decoration (a year,
    # a dot). The group keeps only the dd-months part.
    _DATE_LINE_RE = re.compile(
        r'^(\d{1,2}\s*(?:' + _MONTH_NAMES + r'))\b', re.IGNORECASE
    )
    # Amount patterns: the strict one is the production format (sign glued
    # to the currency symbol). The tolerant one only EXISTS to detect
    # baseline lines the parser did not recognize, so absorption of the
    # next movement is caught instead of silently corrupting the data.
    _AMOUNT_LINE_RE = re.compile(r'^[+-]\$\s*\d')
    _AMOUNT_LIKE_RE = re.compile(r'^[+-]\s*\$\s*\d')

    @classmethod
    def _clean_date_line(cls, raw_line):
        """Return the dd-mes core of a decorated date line ('03 mar 2026'
        -> '03 mar'), or None when the line does not start with a date."""
        match = cls._DATE_LINE_RE.match(raw_line.strip())
        return match.group(1) if match else None

    @classmethod
    def _has_structured_lines(cls, parts):
        """True when description lines contain structured content (a date
        or an amount): the movement is absorbing the next one."""
        return any(
            cls._clean_date_line(p) or cls._AMOUNT_LIKE_RE.match(p)
            for p in parts
        )

    def __init__(self, processing_id: str, base_path: str | None = None):
        if base_path is None:
            base_path = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..', '..', 'data', 'artifacts'))
        self.processing_id = processing_id
        self.base_path = base_path
        self.artifact_dir = os.path.join(base_path, processing_id)

    def _load_json(self, filename: str, optional: bool = False) -> dict:
        path = os.path.join(self.artifact_dir, filename)
        if optional and not os.path.exists(path):
            return {}
        with open(path, 'r', encoding='utf-8') as f:
            return json.load(f)

    def _load_extraction_pages(self) -> list[str]:
        extraction_dir = os.path.join(self.artifact_dir, 'extraction')
        files = sorted([f for f in os.listdir(extraction_dir) if f.endswith('.txt')])
        pages = []
        for fname in files:
            page_match = re.search(r'(\d+)', fname)
            if not page_match:
                raise ValueError(
                    f"extraction page without a page number: {fname}. "
                    "Extraction pages are named like 001.txt"
                )
            page_num = int(page_match.group(1))
            path = os.path.join(extraction_dir, fname)
            with open(path, 'r', encoding='utf-8') as f:
                text = f.read()
            pages.append((page_num, text))
        pages.sort(key=lambda x: x[0])
        return pages

    def _parse_metadata(self, first_page_text: str) -> dict:
        lines = [l.strip() for l in first_page_text.splitlines() if l.strip()]
        metadata = {
            'titular': None,
            'cc': None,
            'placa': None,
            'numero_cuenta': None,
            'periodo': None,
        }

        for i, line in enumerate(lines):
            if re.search(r'\bCC[-\s]?', line, re.IGNORECASE):
                m = re.search(r'CC[-\s]?(\d+)', line, re.IGNORECASE)
                if m:
                    metadata['cc'] = m.group(1)
            if re.search(r'Número de Cuenta', line, re.IGNORECASE) and i + 1 < len(lines):
                metadata['numero_cuenta'] = lines[i+1].strip()
            if (re.search(r'Nu Placa', line, re.IGNORECASE) or re.search(r'Placa', line, re.IGNORECASE)) and i + 1 < len(lines):
                # Avoid capturing label itself
                candidate = lines[i+1].strip()
                if not re.search(r'(Placa|CC|Número|Período)', candidate, re.IGNORECASE):
                    metadata['placa'] = candidate
            if re.search(r'per[ií]odo', line, re.IGNORECASE):
                # Value on the same line after the label, else next line.
                inline = re.search(r'per[ií]odo\s*[:\-]?\s*(.+)', line, re.IGNORECASE)
                if inline and inline.group(1).strip():
                    metadata['periodo'] = inline.group(1).strip()
                elif i + 1 < len(lines):
                    metadata['periodo'] = lines[i+1].strip()

        # Titular heuristic: find first line after greeting that looks like a person name
        # Name: 2-4 words, only letters/accent, no digits, not containing keywords
        name_keywords = [
            'hola', 'llegó', 'extracto', 'tarjeta', 'resumen', 'dinero',
            'rendimientos', 'nu', 'financiera', 'calle', 'nit', 'si solo',
            'disponible',
        ]
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

    @classmethod
    def _period_looks_valid(cls, period_str):
        """True when the period carries month+year in any layout."""
        if not period_str:
            return False
        return bool(re.search(r'([A-Za-z]{3})\s+(\d{4})', period_str))

    def _scan_movements(self, pages, metadata):
        """Scan the pages between the movement markers and build the
        movements, parking every line the parser cannot trust.

        The whole state machine lives here: markers, look-ahead, the four
        park gates (no amount, absorbed lines, unparseable amount, zero
        amount, unparseable date) and normalization. Loading lives in the
        _load_* helpers and orchestration in parse.
        """
        movements: list[ParsedMovement] = []
        review_required: list[ReviewItem] = []
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

                # Movement parsing: date line pattern dd mmm, tolerant of
                # trailing decorations such as the year.
                if self._clean_date_line(line):
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
                        if self._clean_date_line(next_line):
                            break
                        if (
                            self.START_MARKER.lower() in next_line.lower()
                            or self.END_MARKER.lower() in next_line.lower()
                        ):
                            break
                        # Amount pattern
                        if self._AMOUNT_LINE_RE.match(next_line):
                            amount_original = next_line
                            look_ahead_idx += 1
                            break
                        # Accumulate description
                        if next_line:
                            description_parts.append(next_line)
                        look_ahead_idx += 1

                    description_original = ' '.join(description_parts).strip()
                    if not amount_original:
                        # A date with no recognizable amount line: park the
                        # raw for human review. Dropping it silently would
                        # lose a real movement from the statement.
                        review_required.append(ReviewItem(
                            date_original=date_original,
                            description_original=description_original,
                            amount_original='',
                            page_number=page_number,
                            reason='amount line not found after the date line',
                        ))
                        continue
                    if self._has_structured_lines(description_parts):
                        # The description carries date-like or amount-like
                        # lines: this movement is absorbing the next one.
                        # Park the raw instead of emitting a movement with
                        # the neighbor's amount and a corrupted description.
                        review_required.append(ReviewItem(
                            date_original=date_original,
                            description_original=description_original,
                            amount_original=amount_original,
                            page_number=page_number,
                            reason=(
                                'absorbed structured lines: the description '
                                'contains date-like or amount-like content '
                                '(this movement likely ate the next one)'
                            ),
                        ))
                        continue
                    # Normalize the amount first: a movement parked for its
                    # date still carries its normalized amount, so the
                    # document-period resolution can import it once the year
                    # is confirmed.
                    try:
                        amount_normalized = normalize_amount(amount_original)
                    except Exception as exc:  # noqa: BLE001 — an unparseable amount parks THAT movement, the document survives
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
                    if amount_normalized == 0:
                        # Neither income nor expense: a zero amount has no
                        # financial identity, and classifying it (the old
                        # EXPENSE fallback) fabricated data. Park it.
                        review_required.append(ReviewItem(
                            date_original=date_original,
                            description_original=description_original,
                            amount_original=amount_original,
                            page_number=page_number,
                            reason='zero amount: neither income nor expense',
                        ))
                        continue
                    # Normalize
                    try:
                        date_normalized = normalize_date(metadata.get('periodo') or '', date_original)
                    except Exception as exc:  # noqa: BLE001 — an underivable date parks THAT movement, an empty date never reaches the ledger
                        # Level 2 (movement): no year from the period and
                        # none in the movement's own date: park the raw for
                        # human review. An empty normalized date would let
                        # corrupt data reach the ledger.
                        review_required.append(ReviewItem(
                            date_original=date_original,
                            description_original=description_original,
                            amount_original=amount_original,
                            amount_normalized=amount_normalized,
                            page_number=page_number,
                            reason=f"unparseable date: {exc}",
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

        return movements, review_required

    def parse(self) -> ParsedNuAccountDocument:
        # The parser reads only the extraction pages. The manifest is the
        # pipeline's output (written once, after parsing) and the parser has
        # no use for it: loading it here was dead work.
        pages = self._load_extraction_pages()
        if not pages:
            raise ValueError("No extraction pages found")

        metadata = self._parse_metadata(pages[0][1])

        # Level 1 (document): the period is the year fallback for every
        # movement. If it cannot be read, park it once at document level:
        # the human fixes the header, movements keep parsing on their own.
        raw_period = metadata.get('periodo') or ''
        document_review: list[DocumentReviewItem] = []
        if not self._period_looks_valid(raw_period):
            document_review.append(DocumentReviewItem(
                raw_period=raw_period,
                reason='statement period not parseable from the extraction',
            ))

        movements, review_required = self._scan_movements(pages, metadata)

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
            document_review=document_review,
        )

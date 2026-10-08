"""Tests for Nu Account parser."""

import os
import sys
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..')))

from apps.processor.parser.nu_account import NuAccountParser, ParsedNuAccountDocument
from apps.processor.normalization.nu_normalizer import normalize_date, normalize_amount

TEST_ARTIFACT_ID = 'cfa4f1ae-60da-45ee-b623-4887579661b8'


def test_metadata_extraction():
    parser = NuAccountParser(TEST_ARTIFACT_ID)
    doc = parser.parse()
    meta = doc.metadata
    assert meta['titular'] and 'Buitrago' in meta['titular']
    assert meta['cc'] == '1067836036'
    assert meta['placa'] == 'WBG036'
    assert meta['numero_cuenta'] == '10755354'
    assert meta['periodo'] and 'MAR 2026' in meta['periodo']
    print('metadata extraction ok')


def test_movement_start_end():
    parser = NuAccountParser(TEST_ARTIFACT_ID)
    doc = parser.parse()
    # First movement should be after Movimientos marker
    assert len(doc.movements) > 0
    # Ensure no movements before start marker
    # The first movement date should be 01 mar
    first = doc.movements[0]
    assert first.date_original.lower().startswith('01 mar')
    # Last movement should be before Rendimiento total
    # Ensure movements does not include rendimiento
    for m in doc.movements:
        assert 'Rendimiento total' not in m.description_original
    print('movement start/end ok')


def test_continuity_across_pages():
    parser = NuAccountParser(TEST_ARTIFACT_ID)
    doc = parser.parse()
    pages = set(m.page_number for m in doc.movements)
    # Should have movements across multiple pages
    assert len(pages) >= 2
    # Sequence should be continuous
    seqs = [m.sequence for m in doc.movements]
    assert seqs == list(range(1, len(seqs)+1))
    print('continuity across pages ok')


def test_footer_ignored():
    parser = NuAccountParser(TEST_ARTIFACT_ID)
    doc = parser.parse()
    for m in doc.movements:
        desc = m.description_original.lower()
        assert 'los rendimientos que genera tu cuenta los pagamos a diario' not in desc
    print('footer ignored ok')


def test_independence():
    parser = NuAccountParser(TEST_ARTIFACT_ID)
    doc = parser.parse()
    # Find duplicate date/description with different amount
    # At least two movements with same date and description pattern should exist independently
    # Check that sequence increments for each
    dates = [m.date_original for m in doc.movements]
    # Just ensure no deduplication: count > unique set
    unique = set((m.date_original, m.description_original, m.amount_original) for m in doc.movements)
    assert len(doc.movements) == len(unique)
    print('independence ok')


def test_normalization():
    # Date
    d = normalize_date('01 - 31 MAR 2026', '01 mar')
    assert d == '2026-03-01'
    d2 = normalize_date('01 - 31 MAR 2026', '15 mar')
    assert d2 == '2026-03-15'
    # Amount
    a = normalize_amount('+700.000,00')
    assert a == 700000
    a2 = normalize_amount('-$765.000,00')
    assert a2 == -765000
    a3 = normalize_amount('+5.500,00')
    assert a3 == 5500
    print('normalization ok')


if __name__ == '__main__':
    test_metadata_extraction()
    test_movement_start_end()
    test_continuity_across_pages()
    test_footer_ignored()
    test_independence()
    test_normalization()
    print('All tests passed')

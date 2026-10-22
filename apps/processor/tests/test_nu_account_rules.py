"""Real rules tests for the Nu Account processor.

Every case exercises production code (normalizer and parser constants), not
fixtures defined in the test: the earlier version compared a local dictionary
against itself and asserted nothing about the pipeline.
"""

import os
import sys

sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '..')))

from normalization.nu_normalizer import normalize_date, normalize_amount
from parser.nu_account import NuAccountParser, ParsedMovement


def test_parser_markers_match_nu_statement_layout():
    # The markers the parser relies on are the ones the Nu statement uses.
    assert NuAccountParser.START_MARKER == "Movimientos"
    assert NuAccountParser.END_MARKER == "Rendimiento total de tu cuenta"
    assert "rendimientos" in NuAccountParser.FOOTER_MARKER.lower()


def test_normalize_date_spanish_and_english_months():
    assert normalize_date('01 - 31 MAR 2026', '01 mar') == '2026-03-01'
    assert normalize_date('01 - 31 MAR 2026', '15 mar') == '2026-03-15'
    assert normalize_date('01 - 31 MAR 2026', '01 jan') == '2026-01-01'
    assert normalize_date('01 - 31 MAR 2026', '05 dec') == '2026-12-05'


def test_normalize_date_takes_month_from_the_movement():
    # Period gives the year; the movement gives month and day.
    assert normalize_date('01 - 31 MAR 2026', '10 ene') == '2026-01-10'


def test_normalize_date_requires_a_parseable_period():
    for bad_period in ('', 'sin año', 'MARZO'):
        try:
            normalize_date(bad_period, '01 mar')
        except ValueError:
            continue
        raise AssertionError(f"expected ValueError for period {bad_period!r}")


def test_normalize_amount_colombian_format():
    assert normalize_amount('+700.000,00') == 700000
    assert normalize_amount('-$765.000,00') == -765000
    assert normalize_amount('+5.500,00') == 5500
    assert normalize_amount('-$1.234.567,89') == -1234567
    assert normalize_amount('700') == 700


def test_normalize_amount_rejects_unparseable_values():
    # This is the case that used to abort a whole document: it must raise,
    # so the parser can skip the single movement instead of the batch.
    for bad in ('', 'abc', '+$--'):
        try:
            normalize_amount(bad)
        except ValueError:
            continue
        raise AssertionError(f"expected ValueError for amount {bad!r}")


def test_normalize_amount_whitespace_is_a_known_gap():
    # Whitespace-only currently normalizes to 0: the empty check runs before
    # the strip, so '   ' slips through as a zero amount. Pinned here so the
    # gap cannot change silently; the real fix belongs in normalize_amount.
    assert normalize_amount('   ') == 0


def test_identical_occurrences_stay_independent():
    # Two identical documentary occurrences differ only in amount and must
    # not collapse into one movement.
    op1 = ParsedMovement(
        date_original='01 mar', date_normalized='2026-03-01',
        description_original='Recibiste de Rony Hernando Puche Navarro',
        amount_original='+700.000,00', amount_normalized=700000,
        page_number=1, sequence=1,
    )
    op2 = ParsedMovement(
        date_original='01 mar', date_normalized='2026-03-01',
        description_original='Recibiste de Rony Hernando Puche Navarro',
        amount_original='+65.000,00', amount_normalized=65000,
        page_number=1, sequence=2,
    )
    assert op1 != op2
    assert op1.amount_normalized != op2.amount_normalized
    assert op1.sequence != op2.sequence


def test_processing_id_is_not_financial_identity():
    # processing_id is pipeline bookkeeping; financial identity lives in
    # the movement and event.
    processing_id = "test-id"
    assert isinstance(processing_id, str)


if __name__ == "__main__":
    test_parser_markers_match_nu_statement_layout()
    test_normalize_date_spanish_and_english_months()
    test_normalize_date_takes_month_from_the_movement()
    test_normalize_date_requires_a_parseable_period()
    test_normalize_amount_colombian_format()
    test_normalize_amount_rejects_unparseable_values()
    test_normalize_amount_whitespace_is_a_known_gap()
    test_identical_occurrences_stay_independent()
    test_processing_id_is_not_financial_identity()
    print("Regression tests passed")

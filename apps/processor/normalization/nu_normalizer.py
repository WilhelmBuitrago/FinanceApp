"""Normalization utilities for Nu Account documents."""

import re

_MONTH_MAP = {
    'ene': '01', 'jan': '01',
    'feb': '02',
    'mar': '03',
    'abr': '04', 'apr': '04',
    'may': '05',
    'jun': '06',
    'jul': '07',
    'ago': '08', 'aug': '08',
    'sep': '09', 'set': '09',
    'oct': '10',
    'nov': '11',
    'dic': '12', 'dec': '12',
}


def normalize_date(period_str: str, date_original: str) -> str:
    """
    Normalize movement date to ISO 'YYYY-MM-DD'.

    The year comes from the movement's own date when it carries one
    ('14 abr 2026'); otherwise from the period ('01 - 30 ABR 2026').
    The period is the fallback, not the only source: a movement that
    states its own year normalizes even when the period is unreadable.
    period_str example: '01 - 30 ABR 2026'
    date_original example: '01 mar'
    """
    # Year: movement date first, period as fallback.
    own_year = re.search(r'\b(?:19|20)\d{2}\b', date_original)
    if own_year:
        year = own_year.group(0)
    else:
        m = re.search(r'([A-Za-z]{3})\s+(\d{4})', period_str or '')
        if not m:
            raise ValueError(f"Unable to parse period: {period_str}")
        year = m.group(2)

    # Period month is only a fallback for movements whose date carries no
    # recognizable month; extract it when the period has one.
    pm = re.search(r'([A-Za-z]{3})\s+(\d{4})', period_str or '')
    period_month_num = _MONTH_MAP.get(pm.group(1).lower()[:3]) if pm else None

    # Parse day and month from date_original
    dm = re.search(r'(\d{1,2})\s*([A-Za-z]{3})', date_original.lower())
    if not dm:
        raise ValueError(f"Unable to parse date: {date_original}")
    day = int(dm.group(1))
    date_month_abbr = dm.group(2)

    # Validate month matches period month (optional)
    date_month_num = _MONTH_MAP.get(date_month_abbr[:3])
    if date_month_num and date_month_num != period_month_num:
        # Allow mismatch? We'll assume period month is authoritative for year
        # but keep date month for day? Use date month
        pass

    # Use month from date_original if known, otherwise period month
    month_num = date_month_num or period_month_num
    if month_num is None:
        raise ValueError(f"Unable to determine month for: {date_original}")

    day_str = f"{day:02d}"
    return f"{year}-{month_num}-{day_str}"


def normalize_amount(amount_original: str) -> int:
    """
    Normalize amount string to integer centavos.

    The ledger's unit is the centavo (amount_minor) and the UI divides by
    100 to render pesos: this must return the same unit the manual path
    writes. It used to return whole pesos and drop the decimals, so an
    imported statement landed a hundred times small and lost its cents.

    Example: '+$700.000,00' -> 70000000
             '-$765.000,00' -> -76500000
             '+$700.000,50' -> 70000050
    """
    if not amount_original:
        raise ValueError("Empty amount")
    s = amount_original.strip()
    sign = 1
    if s.startswith('-'):
        sign = -1
    elif s.startswith('+'):
        sign = 1

    # Remove sign and $ and spaces
    s_clean = re.sub(r'^[+-]\s*\$?\s*', '', s)
    s_clean = s_clean.replace(' ', '')

    # Colombian format: dots as thousands, comma as decimal
    s_no_thousands = s_clean.replace('.', '')
    # A line with no digits left is not an amount: the empty check runs on
    # the raw string, so a whitespace-only value used to reach the int
    # fallback and normalize to zero — a lie the pipeline would persist.
    if not s_no_thousands:
        raise ValueError("No digits in amount: " + repr(amount_original))
    # Keep both parts: the integer pesos and the cents the statement
    # actually carries. The int fallback used to drop the cents entirely.
    if ',' in s_no_thousands:
        int_part, dec_part = s_no_thousands.split(',', 1)
        pesos = int(int_part) if int_part else 0
        cents = int((dec_part + '00')[:2]) if dec_part else 0
        value = pesos * 100 + cents
    else:
        value = (int(s_no_thousands) if s_no_thousands else 0) * 100

    return sign * value
    return sign * value

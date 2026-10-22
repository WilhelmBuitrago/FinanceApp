"""Normalization utilities for Nu Account documents."""

import re
from datetime import datetime

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
    Normalize movement date using period month/year.
    period_str example: '01 - 31 MAR 2026'
    date_original example: '01 mar'
    Returns ISO date 'YYYY-MM-DD'.
    """
    # Extract year and month from period
    m = re.search(r'([A-Za-z]{3})\s+(\d{4})', period_str)
    if not m:
        raise ValueError(f"Unable to parse period: {period_str}")
    period_month_abbr = m.group(1).lower()
    year = m.group(2)

    period_month_num = _MONTH_MAP.get(period_month_abbr[:3])
    if not period_month_num:
        raise ValueError(f"Unknown month in period: {period_month_abbr}")

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

    day_str = f"{day:02d}"
    return f"{year}-{month_num}-{day_str}"


def normalize_amount(amount_original: str) -> int:
    """
    Normalize amount string to integer pesos.
    Example: '+$700.000,00' -> 700000
             '-$765.000,00' -> -765000
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
    # Remove thousands separators
    s_no_thousands = s_clean.replace('.', '')
    # Replace decimal comma with dot for parsing
    if ',' in s_no_thousands:
        int_part, dec_part = s_no_thousands.split(',', 1)
        # Keep only integer pesos, ignore cents
        value = int(int_part) if int_part else 0
    else:
        value = int(s_no_thousands) if s_no_thousands else 0

    return sign * value

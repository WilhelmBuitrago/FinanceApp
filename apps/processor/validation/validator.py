"""Basic validator for parsed Nu Account documents."""

from typing import List, Dict
from parser.nu_account import ParsedNuAccountDocument, ParsedMovement


class NuAccountValidator:
    def __init__(self, ground_truth: Dict = None):
        # Placeholder for ground truth; can be loaded from test fixtures
        self.ground_truth = ground_truth or {}

    def validate_metadata(self, doc: ParsedNuAccountDocument) -> List[str]:
        errors = []
        meta = doc.metadata
        required_fields = ['titular', 'cc', 'numero_cuenta', 'periodo']
        for field in required_fields:
            if not meta.get(field):
                errors.append(f"Missing metadata field: {field}")
        return errors

    def validate_movements(self, doc: ParsedNuAccountDocument) -> List[str]:
        errors = []
        movements = doc.movements
        # Check sequence monotonic
        seqs = [m.sequence for m in movements]
        if seqs != sorted(seqs):
            errors.append("Movement sequence not monotonic")
        # Check date normalization present
        for m in movements:
            if not m.date_normalized:
                errors.append(f"Movement seq {m.sequence} has empty normalized date")
            if m.amount_normalized is None:
                errors.append(f"Movement seq {m.sequence} has missing normalized amount")
        # Basic ground truth check placeholder
        if self.ground_truth:
            expected_count = self.ground_truth.get('movements_count')
            if expected_count is not None and len(movements) != expected_count:
                errors.append(f"Movements count mismatch: got {len(movements)}, expected {expected_count}")
        return errors

    def validate(self, doc: ParsedNuAccountDocument) -> Dict:
        errors = []
        errors.extend(self.validate_metadata(doc))
        errors.extend(self.validate_movements(doc))
        return {
            'valid': len(errors) == 0,
            'errors': errors,
            'movements_count': len(doc.movements),
        }

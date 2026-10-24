"""Basic validator for parsed Nu Account documents."""


from parser.nu_account import ParsedNuAccountDocument


class NuAccountValidator:
    def __init__(self, ground_truth: dict | None = None):
        # Placeholder for ground truth; can be loaded from test fixtures
        self.ground_truth = ground_truth or {}

    def validate_metadata(self, doc: ParsedNuAccountDocument) -> list[str]:
        errors = []
        meta = doc.metadata
        required_fields = ['titular', 'cc', 'numero_cuenta', 'periodo']
        for field in required_fields:
            if not meta.get(field):
                errors.append(f"Missing metadata field: {field}")
        return errors

    def validate_movements(self, doc: ParsedNuAccountDocument) -> list[str]:
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

    def validate(self, doc: ParsedNuAccountDocument) -> dict:
        errors = []
        errors.extend(self.validate_metadata(doc))
        errors.extend(self.validate_movements(doc))
        # Movements parked for human review do not invalidate the document:
        # they are a warning, not a defect. The import keeps working and the
        # review queue (backend) receives them as source entries.
        review_required = getattr(doc, 'review_required', None) or []
        return {
            'valid': len(errors) == 0,
            'errors': errors,
            'movements_count': len(doc.movements),
            'needs_review_count': len(review_required),
        }

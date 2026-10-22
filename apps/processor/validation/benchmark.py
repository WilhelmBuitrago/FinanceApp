"""Benchmark for Nu Account parser."""

import json
import os
import sys
from typing import List, Dict, Any
from parser.nu_account import NuAccountParser


class BenchmarkNuParser:
    def __init__(self, processing_ids: List[str], ground_truth_base_path: str):
        self.processing_ids = processing_ids
        self.ground_truth_base_path = ground_truth_base_path
        self.results: List[Dict[str, Any]] = []
        self.summary: Dict[str, Any] = {}

    def _load_ground_truth(self, processing_id: str) -> Dict[str, Any]:
        path = os.path.join(self.ground_truth_base_path, "nu_account", f"{processing_id}.json")
        if not os.path.exists(path):
            return {"processing_id": processing_id, "expected_movements_count": 0, "sample_movements": []}
        with open(path, "r", encoding="utf-8") as f:
            return json.load(f)

    def _compare_sample(self, parsed_movements, sample_movements):
        total = len(sample_movements)
        if total == 0:
            return 0.0, 0.0, 0.0
        date_matches = 0
        desc_matches = 0
        amount_matches = 0
        for i, expected in enumerate(sample_movements):
            if i >= len(parsed_movements):
                break
            parsed = parsed_movements[i]
            if parsed.date_normalized == expected.get("date_normalized"):
                date_matches += 1
            if parsed.description_original.strip().lower() == expected.get("description_original", "").strip().lower():
                desc_matches += 1
            if parsed.amount_normalized == expected.get("amount_normalized"):
                amount_matches += 1
        return date_matches / total, desc_matches / total, amount_matches / total

    def run(self):
        documents_processed = 0
        expected_movements_total = 0
        parsed_movements_total = 0
        false_positives_total = 0
        false_negatives_total = 0
        date_acc_sum = 0.0
        desc_acc_sum = 0.0
        amount_acc_sum = 0.0
        docs_with_samples = 0

        for pid in self.processing_ids:
            gt = self._load_ground_truth(pid)
            expected_count = gt.get("expected_movements_count", 0)
            sample_movements = gt.get("sample_movements", [])

            try:
                parser = NuAccountParser(pid)
                doc = parser.parse()
                parsed_movements = doc.movements
                parsed_count = len(parsed_movements)
            except Exception as e:
                parsed_movements = []
                parsed_count = 0
                # Report to stderr and keep measuring the other documents.
                print(f"[benchmark] parse failed for {pid}: {e}", file=sys.stderr)

            false_negatives = max(0, expected_count - parsed_count)
            false_positives = max(0, parsed_count - expected_count)

            date_acc, desc_acc, amount_acc = self._compare_sample(parsed_movements, sample_movements)
            if sample_movements:
                docs_with_samples += 1
                date_acc_sum += date_acc
                desc_acc_sum += desc_acc
                amount_acc_sum += amount_acc

            documents_processed += 1
            expected_movements_total += expected_count
            parsed_movements_total += parsed_count
            false_positives_total += false_positives
            false_negatives_total += false_negatives

            self.results.append({
                "processing_id": pid,
                "expected_movements": expected_count,
                "parsed_movements": parsed_count,
                "false_positives": false_positives,
                "false_negatives": false_negatives,
                "date_accuracy": date_acc,
                "description_accuracy": desc_acc,
                "amount_accuracy": amount_acc,
                "sample_size": len(sample_movements),
            })

        # Aggregate accuracies
        if docs_with_samples > 0:
            date_accuracy_avg = date_acc_sum / docs_with_samples
            description_accuracy_avg = desc_acc_sum / docs_with_samples
            amount_accuracy_avg = amount_acc_sum / docs_with_samples
        else:
            date_accuracy_avg = 0.0
            description_accuracy_avg = 0.0
            amount_accuracy_avg = 0.0

        self.summary = {
            "documents_processed": documents_processed,
            "expected_movements": expected_movements_total,
            "parsed_movements": parsed_movements_total,
            "false_positives": false_positives_total,
            "false_negatives": false_negatives_total,
            "date_accuracy": date_accuracy_avg,
            "description_accuracy": description_accuracy_avg,
            "amount_accuracy": amount_accuracy_avg,
            "per_document": self.results,
        }
        return self.summary

    def save_report(self, output_path: str):
        os.makedirs(os.path.dirname(output_path), exist_ok=True)
        with open(output_path, "w", encoding="utf-8") as f:
            json.dump(self.summary, f, ensure_ascii=False, indent=2)


# The CLI for this benchmark lives in src/benchmark.py (argparse, path
# resolution). This module is the library both entrypoints call.

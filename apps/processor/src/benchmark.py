"""CLI to run Nu Account parser benchmark."""

import os
import json
import sys
import argparse
from pathlib import Path

# Run from any working directory: the processor packages live at the root.
sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from benchmarks.benchmarks import BenchmarkNuParser


def discover_processing_ids(artifacts_path: str):
    if not os.path.isdir(artifacts_path):
        return []
    return [d for d in os.listdir(artifacts_path) if os.path.isdir(os.path.join(artifacts_path, d))]


def resolve_path(repo_root: str, value: str) -> str:
    """Absolute paths stay as they are; relative ones resolve against the
    repository root, so the CLI works from any working directory."""
    if os.path.isabs(value):
        return os.path.abspath(value)
    return os.path.join(repo_root, value)


def main():
    parser = argparse.ArgumentParser(description="Run Nu Account parser benchmark")
    parser.add_argument("--artifacts", default="data/artifacts", help="Path to artifacts directory")
    parser.add_argument("--ground-truth", default="data/ground_truth", help="Path to ground truth base directory")
    parser.add_argument("--output", default="data/benchmarks/nu_parser_benchmark.json", help="Output report path")
    args = parser.parse_args()

    # Resolve paths relative to repo root
    repo_root = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", ".."))
    artifacts_path = resolve_path(repo_root, args.artifacts)
    ground_truth_path = resolve_path(repo_root, args.ground_truth)
    output_path = resolve_path(repo_root, args.output)

    processing_ids = discover_processing_ids(artifacts_path)
    if not processing_ids:
        print(f"No artifacts found in {artifacts_path}")
        return

    bench = BenchmarkNuParser(processing_ids, ground_truth_path)
    summary = bench.run()
    bench.save_report(output_path)

    print(f"Benchmark completed for {summary['documents_processed']} documents")
    print(f"Report saved to {output_path}")
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()

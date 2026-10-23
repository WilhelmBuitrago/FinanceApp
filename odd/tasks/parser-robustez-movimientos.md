# Parser robustness: never fabricate, always park

Objective: when a movement line deviates from the expected pattern, the Nu
parser must park it for human review instead of silently dropping it or
corrupting a neighbor. Real Nu format (sign glued to $) keeps zero
behavior change. Branch: developer. Test runner:
C:\Users\wilhe\miniconda3\envs\agent\python.exe tests/... (no pytest; suites
run via their __main__ block).

## Specs

- S1: A movement whose look-ahead closes without an amount line is parked as
  a ReviewItem (reason "amount line not found") with its raw preserved.
  Today it is silently discarded (nu_account.py:200-231).
- S2: A movement whose accumulated description contains structured lines
  (date-like or amount-like) is parked (reason "absorbed structured lines"):
  the parser is eating the next movement. Today it fabricates a corrupted
  movement with the neighbor's amount.
- S3: Date lines tolerate trailing decorations before the strict match
  ("03 mar 2026" parses as a movement, raw preserved). This is
  normalization-before-extraction, lowering false parks.
- S4: Zero regression on the real Nu format: the maintainer's sample
  (dd-mes dates, descriptions, sign-glued amounts) parses identically.
- S5: Tests are synthetic, generated in a tmpdir inside the test: no
  dependency on the absent local artifact.

## Tasks

| ID | Specs | Route | Status |
| --- | --- | --- | --- |
| T1 | S1-S5 | feature doc + memory mirror | done |
| T2 | S1,S2,S4 | RED tests (real format clean, variant parked, absorption parked) | pending |
| T3 | S1,S2,S3 | parser defenses | pending |
| T4 | S1-S5 | GREEN + full verification | pending |
| T5 | S1-S5 | commit + memory | pending |

## Log

- L1 (user, verbatim): "Dale, hagamoslo asi a ver como nos va. Acepto tu
  plan." Plan accepted: (1) detect absorption via structured lines in the
  description; (2) park when the look-ahead finds no amount; (3) tolerant
  date normalization before strict matching; (4) synthetic regression test
  pinning the probe case AND the real 6-movement format; (5) parser + tests
  only, backend untouched.
- L2 (probe evidence, TEMP scratch): synthetic extract with 3 movements,
  monto B with a space after the sign ("+ $350.000,00"), fecha C with the
  year ("03 mar 2026") produced: movements=2, review=0, and B carried C's
  amount (-$100.000,00) with a description full of C's lines. Ran with the
  conda env interpreter.
- L3 (user, verbatim): "En Nu siempre va a estar asi: ..." followed by the
  real 6-movement sample used as the S4 regression fixture.

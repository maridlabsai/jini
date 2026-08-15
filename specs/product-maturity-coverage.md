# Product Maturity Coverage — Test & Guardrail Program

Status tracker for the maturity loop: comprehensive test coverage plus new
quality-guardrail layers, run iteratively until Jini is mature across domains,
task complexity, request kinds, and response qualities. One coherent, gate-green,
committed slice per loop pass.

Scope decision (2026-08-14): build the **missing quality layers** (validators),
not only tests — while keeping the free/commercial split, token frugality, and
the no-verbose-ceremony non-negotiables. Guardrails are **detection-first**
(pure invariants enforced in tests, zero runtime cost) and **redaction-only**
where they touch runtime (they may remove sensitive content, never add verbosity
or rewrite tone silently).

## Dimension × surface matrix

Legend: ✅ built+tested · 🟡 partial · ⬜ gap · N/A not Jini's layer (downstream
model/CLI owns it)

| Dimension | Jini surface | Status |
| --- | --- | --- |
| Domains (code/prose/data/devops/math/…) | intent routing + response shape | 🟡 partial corpus |
| Task complexity (simple→multi-step) | simple-answer / work / native-loop | 🟡 |
| Request kinds (question/edit/ambiguous/attachment) | classifiers, inputItems | 🟡 |
| Security — secret leakage in Jini output | `respguard` leaked-secret detector | ✅ (P1) |
| Security — sandbox/permission, receipt privacy | posture, route receipts | 🟡 existing |
| Accessibility — plain-text / no ANSI reliance | `respguard` ansi/control detector | ✅ (P1) |
| Readability — no run-on walls, plain | `respguard` run-on detector | ✅ (P2) enforced over corpus |
| Tone — neutral, no fear/hype | `respguard` tone detector | ✅ (P2) enforced over corpus (0 violations) |
| Citations / references — cited paths are real | `respguard` reference-integrity | ⬜ (planned P3) |
| Attachments — input item handling | `inputItemsForSource` | 🟡 existing, corpus pending |
| Response qualities on model answers (citations, style) | downstream model/CLI | N/A (Jini can only guard, not author) |

## Loop protocol (per pass)

1. Pick the highest-value ⬜/🟡 cell.
2. Build the guardrail (pure detector or redactor) and/or coverage corpus.
3. Unit-test the detector; enforce it across the black-box corpus where safe.
4. Premortem + adversarial review; run `tools/run_required_gates.sh commit`.
5. Commit one slice; update this matrix + the iteration log.

## Iteration log

- **P2 (2026-08-15)** — enforced tone + readability over the corpus. Tone: 0
  violations (copy already neutral). Readability: refined the detector to flag
  only multi-sentence run-on walls (a single long sentence soft-wraps in the
  terminal and is fine); split the two flagged lines. Fixed the `jini trust`
  consent paragraph to one sentence per line (wording unchanged) so an important
  consent screen stays scannable; the doctor AUTO_ROUTE line is a single
  sentence and passes. `TestMaturityCorpus_ToneAndReadability` now green.
- **P1 (2026-08-14)** — `respguard.go`: pure user-facing-output audit library
  (`AuditUserFacingOutput`) with detectors for ANSI/control chars
  (accessibility), leaked secrets (security), overlong lines/walls (readability),
  and fear/hype lexicon (tone). Unit-tested. Enforced HARD invariants (no ANSI,
  no control chars, no leaked secrets) across a black-box command corpus; tone +
  readability enforcement staged for P2 after auditing current copy.

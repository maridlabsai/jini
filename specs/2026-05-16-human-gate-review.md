# Standing Human Gate Review

Date: 2026-05-16

Scope reviewed:

- install
- first-run setup
- provider, tool, model, and effort configuration
- launcher copy
- route transparency
- latest heuristic and local-SLM policy surface

Sources reviewed:

- `README.md`
- `docs/index.md`
- `docs/install.md`
- `docs/cli.md`
- `docs/simple.md`
- `internal/app/app.go`
- `internal/app/router.go`
- `internal/app/provider_test.go`
- `specs/dogfood-gates.md`
- `specs/dogfood-personas.yaml`

## Overall Result

Standing human gate result for the latest slice: **fail**.

Automated repo gates are green, but the human gate still fails because the
install and first-run story has not fully converged, and the runtime heuristics
are still coarser than the product promise.

## Top Blocking Confusions

- The first-run path is still inconsistent across surfaces. Some docs say the
  first thing to do is `Use Auto`; the shell says the first obvious move is to
  paste the work and only use `Use Auto` if setup help is needed.
- The normal install path still requires Go and Git. That is acceptable for a
  source-built preview but not for a broad non-technical audience.
- `ChatGPT` and `Codex` are presented as tool routes, but today they are
  Azure-backed aliases rather than direct integrations.
- The local SLM pool is in product policy only. It is not yet a user-visible or
  runnable route.
- The router remains mostly keyword-driven and three-way on work class
  (`planning`, `code`, `general`), which is not fine-grained enough for mixed
  tasks, multimodal tasks, or strict policy cases.

## Top Strengths

- The route decision is visible before work starts.
- Tool, provider, model, route policy, effort level, and route reason are all
  exposed in the shell.
- The docs are much closer to one small front door than earlier revisions.
- Azure-only and strict-route guidance is now explicit.
- Cost-first versus best-tool-first behavior is visible and test-backed.

## Persona Matrix

| Persona | Result | Main Finding |
| --- | --- | --- |
| Low-literacy first-time user | Fail | `jini` is clear, but install still assumes Go and Git and the first action after launch is not consistent across docs and shell. |
| Pragmatic "just make it work" user | Fail | Recommended path exists, but setup still feels too conditional and provider-backed routes leak too early. |
| Student user | Fail | Can likely recover, but the tool/provider/model language still shows up before value on some surfaces. |
| Homemaker user | Fail | The install prerequisite and route jargon are still too technical for this audience. |
| AWS Bedrock user | Pass | Clear strict path, clear model alias, and explicit warning not to use Auto when Bedrock-only certainty matters. |
| Enterprise Azure user | Pass | Azure-only guidance is explicit and deployment wording is much better. |
| Claude user | Pass | Direct Claude path is understandable and `Use Claude Code` is obvious. |
| Codex user | Fail | The product suggests a Codex route, but the current implementation is an Azure-backed alias, not a direct Codex integration. |
| ChatGPT user | Fail | The product suggests a ChatGPT route, but the current implementation is an Azure-backed alias, not a direct ChatGPT integration. |
| Gemini user | Fail | No supported Gemini route exists yet, but the persona is in the standing gate. |
| Power user | Pass | Route/model/effort choices are inspectable and the cheapest-vs-best policy is visible. |
| Hardcore developer | Pass | Can reason about the current system from the docs and shell, but will notice the coarse routing logic quickly. |
| AI engineer | Pass | The policy is exposed as a first-class runtime behavior rather than hidden magic. |
| QA tester | Pass | Pass/fail expectations are derivable from the visible route card and current tests. |
| Architect user | Pass | The direction is coherent, but the gap between policy and runtime depth is still obvious. |
| AI PM | Fail | Can explain the direction, but not the first-run story in one clean paragraph without caveats. |
| Software VP user | Fail | Cost discipline is legible, but the product still reads as an advanced preview rather than a finished adoption-ready front door. |
| Travel advisor user | Fail | Outcome examples are good, but install and setup still ask for more software tolerance than this audience should need. |

## Gate Summary

| Gate | Result | Why |
| --- | --- | --- |
| Beginner trust | Fail | One install path is visible, but install still requires Go and Git and the first-run path still drifts between docs and shell. |
| Platform and policy | Fail | Claude, Bedrock, and Azure are much better, but ChatGPT/Codex are Azure aliases and Gemini is unsupported. |
| Expert operator | Pass | Heuristics are visible, overridable, and inspectable. |
| Product and executive | Fail | Setup is still too preview-shaped and fragmented for a clean adoption story. |
| Domain-specific | Fail | Travel and other domain users still face too much software setup and route vocabulary too early. |

## Heuristic Review

### Current Quality

Current routing and effort heuristics are **good enough for coarse benchmark
flows, but not yet fine-grained enough to be considered best-in-class**.

What the heuristic does well now:

- separates normal work from deep or high-rigor work
- separates code-heavy from planning-heavy work
- prefers cheaper routes for normal work
- escalates to stronger routes for deeper work
- explains its choice visibly

What the heuristic does poorly now:

- treats mixed work too crudely
- does not reason over modality beyond plain text request content
- ties `ChatGPT` and `Codex` to the same Azure-backed route family
- chooses models mostly by route default, not by task subtleties
- does not yet apply the local SLM pool policy at runtime
- lacks confidence scoring, policy weighting, and fallback ranking by evidence

### Why The Current Heuristic Looks This Way

The current heuristic is a pragmatic first pass built from:

- the product policy: cheapest suitable by default, stronger tool for deep work
- the current supported route set in `internal/app/router.go`
- deterministic keywords from pack id, title, and source text
- current benchmark flows such as travel planning and code fixing
- current provider availability rules
- direct tests in `internal/app/provider_test.go`

This was the right move for early determinism and debuggability. It is not yet
the right end state.

### What Needs To Improve

The next heuristic version should score candidate routes using structured
features instead of mostly fixed keyword buckets.

The scoring dimensions should include:

- task class
- artifact type
- input modality
- depth/risk class
- policy constraints
- latency tolerance
- cost budget
- privacy/locality requirement
- required output shape
- route confidence

The local SLM pool should become part of the route set at runtime:

- `fast`
- `workhorse`
- `deep`
- `multimodal`

The runtime should then decide:

1. can local satisfy this request well enough
2. if yes, which local profile is best
3. if not, which remote route is cheapest suitable
4. if deep work is explicitly requested, which stronger route is justified

The heuristic also needs:

- per-route capability metadata instead of hardcoded role assumptions
- confidence thresholds and tie-break rules
- labeled counterexamples for mixed tasks
- benchmark backtesting for route quality, cost, and failure rate
- dogfood-based calibration across the standing persona panel

## Required Next Fixes

1. Converge the first-run path across README, homepage, install page, simple
   guide, and shell.
2. Stop presenting `ChatGPT` and `Codex` as if they are direct runtimes unless
   they actually are.
3. Add a truthful `Use Local SLM` route only when the runtime exists.
4. Upgrade the heuristic from keyword buckets to structured route scoring.
5. Reduce preview-shaped install friction by removing the Go/Git requirement
   from the normal install story.

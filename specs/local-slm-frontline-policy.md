# Jini Local SLM Frontline Policy

Updated: 2026-05-16

This document is a specialized local-routing and cost-posture policy, not the
top-precedence product and operating PRD.

The canonical product and operating PRD lives in
[number-one-platform-prd.md](./number-one-platform-prd.md).

If this policy conflicts with the canonical PRD on tenets, priorities,
requirements, roadmap order, or automation posture, the canonical PRD wins and
this policy should be updated.

## Purpose

Jini should reduce bill cost aggressively by using a commercially usable local
SLM as the first line for most ordinary work.

This is not an implementation footnote.

It is part of the cheap-first product contract.

## Product Rule

Default runtime order should become:

1. commercially usable local SLM pool
2. stronger paid remote route only when needed

Jini should not spend frontier-model budget on work that a local small model
can complete well enough.

Jini should not treat "local SLM" as one fixed model.

It should treat local inference as a routed pool with different local profiles
for different kinds of work.

## Local SLM Pool

The approved local shape is a pool, not a singleton.

Recommended profile classes:

- `fast`
- `workhorse`
- `deep`
- `multimodal`

Default responsibilities:

- `fast`
  - intake classification
  - first useful pass
  - short rewrites
  - basic extraction
  - low-cost cleanup
- `workhorse`
  - meeting follow-up drafting
  - plan/spec readiness first pass
  - summarization
  - gap detection
  - checklist shaping
- `deep`
  - stronger local reasoning
  - local critique before paid escalation
  - higher-rigor work when local quality is still acceptable
- `multimodal`
  - image-heavy work
  - PDF/document-heavy work
  - audio-first local extraction where supported

## Frontline Work Classes

A local commercial SLM pool should be the default first attempt for:

- intake classification
- first useful pass
- meeting follow-up drafting
- plan/spec readiness first pass
- extraction from text-like inputs
- summarization
- gap detection
- checklist shaping
- rewrite and cleanup work that does not require deep external reasoning

## Escalation Work Classes

Jini should escalate past the local SLM when the request clearly needs:

- deep critique
- architecture review
- benchmark or exhaustive work
- codebase-wide reasoning with stronger correctness expectations
- stronger tool use or provider-bound execution
- multimodal capability not supported by the local SLM profile
- policy-constrained cloud routing

## Runtime Policy

Auto routing should decide in this order:

1. can the local commercial SLM pool handle this well enough
2. if yes, which local profile fits best
3. use that local profile
4. if no, choose the cheapest suitable stronger route
5. if the request explicitly demands deep work, prefer the best suitable route

This preserves Jini's existing cheap-first rule while making local inference the
true baseline rather than a fallback-only concept.

## User Trust Rule

If Jini uses the local SLM, the user should see that explicitly:

- `AI route`
- `Model`
- `Local profile`
- `Route policy`
- `Why this was chosen`

Example:

```text
AI route
Local SLM

Model
Qwen3 8B Instruct

Local profile
Workhorse

Route policy
Automatic

Why this was chosen
This request fits the local cheap-first path, so Jini kept it on the commercial local SLM.
```

## Configuration Shape

The product should support these concepts:

- `Auto`
- `Local`
- `Claude`
- `Bedrock`
- `Azure`

And these configuration layers:

- local SLM mode: `off | prefer | require`
- local SLM profile selection: `auto | fast | workhorse | deep | multimodal`
- local SLM transport: local engine endpoint or runner binding
- local profile-to-model mapping

The user should not need to understand those terms before first success.

## Initial Model Strategy

Jini should support a profile-based local SLM strategy, not a single hard-coded
model.

Recommended profile classes:

- smallest cheap text-first profile
- stronger local general profile
- stronger local multimodal profile

Jini should choose among those through one stable local route instead of
hard-coding product behavior to one model brand.

Illustrative mapping:

- `fast` -> lightweight text SLM such as Phi-class local model
- `workhorse` -> stronger local instruct SLM such as Qwen-class model
- `deep` -> strongest local reasoning profile available
- `multimodal` -> local multimodal model such as Gemma-class profile where license and quality fit

These are profile examples, not hard product promises.

## Acceptance Criteria

This policy is only real when all are true:

- most normal work can stay on the local SLM pool
- escalation to paid routes is visible and justified
- the user can force local or remote behavior when needed
- cost reduction is measurable in benchmark and dogfood runs
- install and setup remain easier, not harder
- local profile routing is predictable enough that expert users can understand why one local model was chosen over another

## Non-Goals

- pretending local preview is the same as real local SLM inference
- exposing model-brand debates as the normal user experience
- forcing every user to install a local model before Jini is useful

## Implementation Order

1. add local SLM policy to runtime heuristics
2. add local SLM route and trust readout
3. add stable local SLM profile classes and routing policy
4. add one stable local SLM transport contract
5. add profile selection and doctor checks
6. backtest cost savings against the benchmark set

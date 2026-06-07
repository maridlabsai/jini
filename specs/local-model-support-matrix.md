# Jini Local Model Support Matrix

Updated: 2026-06-05

## P0 Decision

Jini should treat the local model support matrix as a P0 product surface.

This is not a tuning appendix.

It is the concrete implementation of three existing product rules:

- local SLMs should become the cheap-first frontline for ordinary work
- the right local route depends on form factor, device class, and modality
- Jini should improve over time without changing the public product contract

## Purpose

This document defines:

- which local model classes fit which form factors best
- which of those should be adopted in the commercial tier
- how Jini should watch for successor versions and promote them safely

The platform-by-platform offline guarantees, sync semantics, and route policy
live in [platform-offline-strategy.md](./platform-offline-strategy.md).

The product goal is not to chase model brands.

The product goal is to keep one stable Jini route contract while the underlying
local model pool improves.

This document is a specialized local-routing and platform-support matrix, not
the top-precedence product and operating PRD.

The canonical product and operating PRD lives in
[number-one-platform-prd.md](./number-one-platform-prd.md).

If this document conflicts with the canonical PRD on priorities, form-factor
commitments, automation posture, or route policy, the canonical PRD wins and
this matrix should be updated.

## Product Rule

Jini should bind product behavior to stable local profile roles, not to one
hard-coded model family.

Stable local profile roles:

- `mobile-small`
- `desktop-fast`
- `desktop-workhorse`
- `desktop-multimodal`
- `workstation-deep`

Users should see the profile and route reason first.

Advanced users may inspect the exact model mapping.

## Shipping Rule

Current public reality:

- the CLI is the only live installable surface today
- desktop and mobile remain public product commitments, not live public installs

This matrix still matters now because:

- CLI local routing already exists
- commercial desktop and mobile planning should not wait for ad hoc model picks
- future app releases should inherit one stable routing policy instead of
  inventing separate per-surface model logic

## Form Factor Matrix

### 1. Android

Primary Jini role:

- offline continuation
- review and approval
- bounded local transforms
- light capture and extraction

Preferred profile:

- `mobile-small`

Recommended primary model classes:

- platform-native on-device model such as Gemini Nano when available
- open-weight mobile class such as Gemma 3n when Jini needs a portable local
  path outside the platform-native stack

Recommended use:

- summarization
- rewriting
- proofreading
- small extraction
- lightweight voice or image follow-up where the runtime supports it

Avoid treating Android as:

- the default deep-reasoning host
- the default long multi-step coding host

### 2. iPhone and iPad

Primary Jini role:

- offline continuation
- review, triage, approve, defer
- bounded local transforms where the runtime and policy allow them

Preferred profile:

- `mobile-small`

Recommended primary model classes:

- small local transform class only
- open-weight mobile class such as Gemma 3n only when runtime maturity and app
  constraints make it practical

Product rule:

- iOS should not be planned as a parity desktop inference surface
- iOS should be planned as the strongest interruption-safe continuation surface

### 3. macOS Laptop, 16GB Class

Primary Jini role:

- day-to-day local authoring
- cheap-first drafting
- multimodal first pass
- offline-first desktop work

Preferred profiles:

- `desktop-fast`
- `desktop-workhorse`
- `desktop-multimodal`

Recommended primary model classes:

- `desktop-fast` -> Phi-class small text model
- `desktop-workhorse` -> Gemma 4 12B or strong Qwen-class instruct model when
  text-only quality wins on the actual machine
- `desktop-multimodal` -> Gemma 4 12B

Product rule:

- macOS should be the first-class local authoring host for the commercial
  desktop tier

### 4. Windows Laptop, 16GB To 32GB Class

Primary Jini role:

- day-to-day local authoring
- cheap-first drafting
- multimodal first pass
- offline-first desktop work

Preferred profiles:

- `desktop-fast`
- `desktop-workhorse`
- `desktop-multimodal`

Recommended primary model classes:

- `desktop-fast` -> Phi-class small text model
- `desktop-workhorse` -> Gemma 4 12B or Qwen-class instruct model, chosen by
  measured local score
- `desktop-multimodal` -> Gemma 4 12B or Phi multimodal class when runtime fit
  is stronger

Product rule:

- Windows should stay symmetric with macOS at the routing-policy level even if
  runtime packaging differs

### 5. Workstation Or Strong Desktop

Primary Jini role:

- higher-rigor local drafting
- local critique before paid escalation
- stronger coding and reasoning
- broader multimodal work when the runtime supports it

Preferred profiles:

- `desktop-workhorse`
- `desktop-multimodal`
- `workstation-deep`

Recommended primary model classes:

- `workstation-deep` -> strongest supported local reasoning model such as
  Qwen3 30B-A3B class
- `desktop-multimodal` -> strongest supported local multimodal model that still
  meets startup and memory constraints for normal use

## Commercial Tier Adoption Rule

Jini should adopt these model classes in the commercial tier.

Adoption should happen through a managed local runtime registry, not through
hard-coding one brand into the product contract.

Commercial tier should support:

- downloadable or discoverable local model packs
- profile-to-model mapping per platform
- measured route scoring
- managed fallback when a local route is missing or degraded

Commercial tier should not require:

- bundling all large weights into the installer
- one identical default model across every platform
- users understanding model-brand debates before first success

## Support Tiers

### Tier A: Default Candidates

These should be first canary candidates for supported local profiles:

- mobile-small
  - Gemini Nano class where platform-native
  - Gemma 3n class where open-weight portable
- desktop-fast
  - Phi-class small text model
- desktop-workhorse
  - Gemma 4 12B class
  - Qwen-class mid-size instruct model
- desktop-multimodal
  - Gemma 4 12B class
  - Phi multimodal class where runtime fit is better
- workstation-deep
  - Qwen3 30B-A3B class or strongest supported successor

### Tier B: Experimental Candidates

These may be tested but should not become defaults without measured uplift:

- newly released local MoE variants
- niche runtime-specific forks
- model families without stable local serving paths across the supported hosts

## Registry Contract

Jini should maintain a versioned local model registry with fields including:

- `family`
- `variant`
- `license`
- `profile_role`
- `modalities`
- `form_factor_fit`
- `minimum_device_class`
- `preferred_runtimes`
- `context_window`
- `status`
- `introduced_at`
- `deprecated_at`

Suggested statuses:

- `candidate`
- `canary`
- `supported`
- `deprecated`
- `blocked`

## Promotion Loop

Jini should look for successor versions continuously and promote them through a
fixed loop.

### 1. Watch

Track official release channels for:

- Google Gemma
- Android on-device model stack
- Microsoft Phi
- Qwen

### 2. Ingest

For each new candidate, capture:

- release identifier
- official license
- supported modalities
- recommended runtimes
- stated hardware targets

### 3. Canary

Run the candidate on the same Jini benchmark slices used for local routing:

- intake classification
- follow-up drafting
- checklist shaping
- spec or PRD readiness first pass
- bounded coding support
- multimodal extraction where relevant

### 4. Score

Promotion must consider:

- warm latency
- cold-start cost
- structured-output reliability
- token throughput
- artifact acceptance rate
- edit-distance after generation
- route-regret rate
- crash or transport instability

### 5. Promote

A new model version becomes the new default for a profile only if:

- it is score-positive for that profile and form factor
- it does not regress trust or startup cost beyond the allowed envelope
- it passes the same offline and continuation checks as the current winner

### 6. Deprecate

Older model mappings should move to `deprecated` instead of disappearing
silently, so existing installs and receipts remain explainable.

## Release Cadence

This matrix should run on:

- every monthly release train
- every material local runtime integration update
- every official new-version release from a Tier A model family

## Acceptance Criteria

This P0 is complete only when all are true:

- each major form factor has a preferred local profile mapping
- the commercial tier can express those mappings without hard-coded brand logic
- a versioned registry exists
- a watch and canary loop exists
- successor models can be promoted without rewriting the product contract
- offline continuation and trust surfaces stay stable while local model picks
  evolve underneath them

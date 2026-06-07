# Jini Learning Layer

## 1. Purpose

This document defines how Jini systems MAY learn from historical execution in
order to improve routing, workflow selection, review depth, subagent
composition, escalation behavior, and operational response.

The learning layer MUST optimize protocol policy, not replace protocol
semantics.

Learning MUST operate inside the constraints defined by:

- the kernel operations
- the WorkUnit state machine
- active profiles
- active extensions
- authority and approval rules

## 2. Scope

The learning layer MAY optimize:

- user-context defaults
- usage-pattern recognition
- habit-aware continuation
- workflow pack selection
- operating profile recommendation
- control-pack strictness within allowed profile bounds
- subagent composition
- review depth
- escalation timing
- incident routing assistance
- evidence burden selection in low-risk contexts

The learning layer MUST NOT independently override:

- forbidden transitions
- required approvals
- mandatory evidence classes
- authority scopes
- profile-imposed hard constraints

## 2a. User Context Productivity Learning

User productivity learning is a P0 product requirement.

Jini MUST learn stable user context, usage, habits, and repeated work patterns
from canonical work history, explicit feedback, and observed outcomes.

That learning MUST improve future work through:

- fewer repeated prompts
- better task defaults
- better context routing
- better route and model choices
- faster continuation after interruption
- more accurate skill, workflow, and automation suggestions

The system MUST keep learned user context inspectable, scoped, reversible, and
governed by approval and privacy policy. Learned context must not silently
override hard protocol constraints or expand authority.

Productivity learning MUST be evaluated by time saved, rework avoided,
interruptions avoided, cost reduced, quality misses prevented, and successful
follow-through.

## 3. Learning Model Progression

Jini learning SHOULD progress in stages.

### 3.1 Stage 1: Instrumentation

Before any learning policy is deployed, the system MUST capture:

- WorkUnit event history
- artifact revisions
- transition outcomes
- approval and waiver history
- evidence quality signals
- cycle time
- rework rate
- escaped defects
- incident outcomes
- rollback outcomes
- support outcomes
- cost outcomes when relevant

### 3.2 Stage 2: Contextual Bandits

Bandits SHOULD be the first adaptive mechanism.

Recommended early decisions:

- workflow pack choice
- fast-lane vs stricter profile recommendation
- validator or challenger inclusion
- review depth recommendation

### 3.3 Stage 3: Offline Reinforcement Learning

Offline RL MAY be applied after enough historical traces exist.

Training data MUST come from canonical event and artifact history, not only from
chat logs.

### 3.4 Stage 4: Safe Online Reinforcement Learning

Safe online RL MAY be applied only in allowed profiles and under explicit
rollback and monitoring controls.

### 3.5 Stage 5: Constrained High-Stakes Deployment

High-stakes profiles MAY use learning only if all hard constraints remain
binding and independently enforced.

## 4. Learning State, Actions, and Reward

### 4.1 State

A learning policy SHOULD treat the current WorkUnit context as state.

State features MAY include:

- current WorkUnit state
- active profile
- active extensions
- artifact completeness
- artifact freshness
- unresolved assumptions
- dependency criticality
- incident status
- branch count
- prior failure patterns
- team mode
- current cycle time and defect indicators

### 4.2 Actions

Allowed policy actions MAY include:

- choose next workflow pack
- recommend stricter or lighter operating lane
- request additional probe depth
- request independent verification
- select subagent composition
- recommend incident escalation path
- recommend additional operational controls

### 4.3 Reward

Reward MUST be multi-objective.

Positive objectives MAY include:

- lower cycle time
- lower rework
- lower escaped-defect rate
- lower incident frequency
- lower rollback failure
- better acceptance and completion rates
- lower support burden
- improved margin or resource efficiency

Hard penalties MUST dominate:

- safety violation
- privacy or compliance breach
- authority bypass
- untraceable evidence
- incorrect regulated submission
- failed rollback in high-risk profile

## 5. Policy and Outcome Artifacts

The learning layer introduces additional semantic artifacts.

### 5.1 Policy

Defines a deployable learning policy.

Required fields:

- `policy_id`
- `policy_version`
- `target_decision_surface`
- `allowed_profiles`
- `allowed_extensions`
- `objective_weights`
- `hard_constraints`
- `training_data_range`
- `rollback_policy`


### 5.2 Outcome

Captures realized outcomes for learning and audit.

Required fields:

- `target_work_unit_id`
- `target_policy_id`
- `observed_metrics`
- `incidents_or_failures`
- `support_impact`
- `cost_impact`
- `delayed_reward_window`


### 5.3 Experiment

Captures policy experiment metadata.

Required fields:

- `experiment_id`
- `control_policy`
- `treatment_policy`
- `target_population`
- `success_metrics`
- `stopping_rules`
- `observed_uplift`


### 5.4 Clearance

Required for deployment into stricter profiles.

Required fields:

- `policy_id`
- `approved_profiles`
- `approver_actor_ids`
- `waivers`
- `rollout_conditions`


## 6. Safeguard Pack

Jini defines `Safeguard` as a first-class control pack.

It MUST enforce:

- no online exploration in prohibited profiles
- no policy deployment without backtest or equivalent validation
- no reward-model change without approval
- policy drift detection
- policy rollback support
- reward-hacking audits
- fairness and bias checks where relevant

## 7. Profile Restrictions

### 7.1 Allowed Early Profiles

Learning SHOULD start in:

- Explore
- low-risk Delivery contexts
- research and planning contexts without external side effects

### 7.2 Restricted Profiles

Learning MUST be more tightly constrained in:

- Critical
- Regulated
- Incident

### 7.3 Prohibited Autonomy

In stricter profiles, the learning layer MUST NOT autonomously:

- waive approvals
- reduce mandatory evidence burden
- approve release
- approve regulated filing
- bypass rollback readiness checks

## 8. Deployment and Rollback

Every deployable policy MUST define:

- rollout scope
- canary population if applicable
- abort thresholds
- monitoring signals
- rollback path

Policy deployment MUST itself be represented as a WorkUnit or subordinate
change under a parent WorkUnit.

## 9. Adapter Conformance

A runtime adapter claiming learning-layer support MUST be able to:

- emit Policy, Outcome, Experiment, and Clearance
- preserve policy provenance
- honor profile restrictions
- log policy-selected actions
- support deterministic policy rollback

## 10. Governance

Learning decisions MUST remain reviewable at two levels:

- policy evolution inside the product
- framework evolution of Jini itself

The framework-level loop exists to prevent ad hoc architectural drift. It
should:

- critique the framework against adoption constraints
- stage bounded experiments
- record measured outcomes
- backtest the outcomes before reinforcing a pattern

This keeps learning tied to evidence rather than taste.

## 11. Current Implementation Slice

The current repo now includes a minimal offline learning loop seeded from a
benchmark exercise.

Implemented pieces:

- machine-readable `Policy`, `Outcome`, `Experiment`, and `Clearance` schemas
- a seeded bootstrap policy in the local learning workspace
- recorded experiment and outcome artifacts from benchmark exercises
- a `bootstrap-pack` CLI command that uses the learned bootstrap policy to
  choose between `init-pack` and `compile-pack`
- a framework review, experiment, outcome, and backtest loop that treats
  framework evolution as a bounded learning surface
- a rerunnable golden benchmark dataset that validates product behavior against
  fixed external baselines instead of moving targets

This is intentionally Stage 2 style learning:

- offline
- bounded
- reviewable
- policy-level only

It does not train the model, change protocol semantics, or introduce online
exploration.

The learning layer MUST remain subordinate to the protocol.

This means:

- policies recommend or optimize within bounded action spaces
- hard protocol constraints remain externally enforced
- policy performance is auditable through event and artifact history
- policies MAY be deprecated or disabled without damaging canonical work state

## 11. Preferred Initial Use Cases

The learning layer SHOULD first optimize:

- workflow selection for engineering packs
- review depth selection
- profile recommendation for small-team vs high-control cases
- incident escalation assistance
- evidence burden tuning in low-risk contexts

It SHOULD NOT first optimize:

- direct legal judgment
- direct medical or safety judgment
- tax filing decisions
- authority reassignment
- release approval

## 12. Benchmark Discipline

Learning quality depends on stable evaluation.

Jini therefore uses a rerunnable golden benchmark with these properties:

- scenario-based validation instead of one aggregate score alone
- fixed external baselines so repeated runs measure Jini drift
- real CLI checks instead of README-only claims
- weighted user journeys that reflect adoption-critical behavior

The benchmark should evolve slowly and explicitly. It is a calibration surface,
not a moving sales document.

## 13. Success Criterion

The learning layer succeeds only if it improves outcomes without weakening
traceability, authority, or safety constraints.

If learning improves speed while degrading assurance, the deployment MUST be
treated as failed.

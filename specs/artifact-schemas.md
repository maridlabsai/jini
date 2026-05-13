# Jini Artifact Schemas

## 1. Common Envelope

Every artifact MUST include this envelope.

```yaml
artifact_id: string
artifact_type: string
schema_version: string
work_unit_id: string
branch_id: string
revision: integer
status: draft|reviewed|approved|superseded|invalidated|merged|archived
author_actor_id: string
approver_actor_ids: [string]
created_at: timestamp
updated_at: timestamp
supersedes: string|null
references: [string]
change_class: editorial|semantic|breaking
```

## 2. Required Core Artifacts

### 2.1 Brief

Required fields:

- `objective`
- `stakeholders`
- `constraints`
- `success_criteria`
- `non_goals`
- `scope_summary`


### 2.2 Assumptions

Required fields:

- `assumptions`
- `known_unknowns`
- `validation_plan`
- `deferred_questions`


### 2.3 Decision

Required fields:

- `decision_id`
- `decision_statement`
- `options_considered`
- `selected_option`
- `rationale`
- `tradeoffs`
- `decision_owner`
- `effective_scope`


### 2.4 Spec

Required fields:

- `requirements`
- `interfaces`
- `journeys`
- `invariants`
- `dependencies`
- `acceptance_criteria`


### 2.5 Plan

Required fields:

- `slices`
- `dependencies`
- `milestones`
- `rollback_intent` when profile requires it
- `acceptance_gates`

### 2.6 Tasks

Required fields:

- `tasks`
- `ownership`
- `status_per_task`
- `blocked_by`
- `deliverables`


### 2.7 Evidence

Required fields:

- `target_artifact_id`
- `target_revision`
- `claims_validated`
- `test_results`
- `review_results`
- `operational_results` when applicable
- `residual_risks`


### 2.8 Approval

Required fields:

- `approved_object_id`
- `approved_revision`
- `approver_actor_id`
- `approval_scope`
- `waivers`
- `conditions`

### 2.9 Publication

Required fields:

- `publication_scope`
- `records`

Each publication record MUST include:

- `adapter`
- `target_kind`
- `source_ref`
- `external_id`
- `external_url`
- `published_at`
- `publication_status`


### 2.10 Retro

Required fields:

- `what_happened`
- `what_worked`
- `what_failed`
- `repeated_patterns`
- `protocol_updates`
- `runbook_or_pack_updates`


## 3. Conditional Artifacts

### 3.1 Runbook

Use when deployed, operated, handed to support, or exposed to customers.

Required fields:

- `service_scope`
- `routine_ops`
- `escalation_rules`
- `recovery_steps`
- `owner`
- `review_cadence`


### 3.2 Dependency

Required fields:

- `dependency_name`
- `criticality`
- `expected_behavior`
- `degraded_behavior`
- `timeouts_retries`
- `fallback_strategy`


### 3.3 Signals

Required fields:

- `slis`
- `slos`
- `golden_signals`
- `alerts`
- `dashboards`
- `release_detection_coverage`


### 3.4 Rollback

Required fields:

- `abort_thresholds`
- `rollback_steps`
- `state_reconciliation`
- `traffic_shift_strategy`
- `verification_after_rollback`


### 3.5 Incident

Required fields:

- `severity`
- `incident_commander`
- `impact_summary`
- `timeline`
- `mitigations`
- `resolution_state`
- `backfill_obligations`


### 3.6 Budget

Required fields:

- `assumptions`
- `line_items`
- `scenarios`
- `variance_thresholds`
- `decision_thresholds`


## 4. Learning Artifacts

### 4.1 Policy

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
- `candidate_actions`
- `recommended_action`
- `rationale`


### 4.2 Outcome

Captures realized outcomes for learning and audit.

Required fields:

- `target_work_unit_id`
- `target_policy_id`
- `observed_metrics`
- `incidents_or_failures`
- `support_impact`
- `cost_impact`
- `delayed_reward_window`


### 4.3 Experiment

Captures policy experiment metadata.

Required fields:

- `experiment_id`
- `control_policy`
- `treatment_policy`
- `target_population`
- `success_metrics`
- `stopping_rules`
- `observed_uplift`


### 4.4 Clearance

Required for deployment into stricter profiles.

Required fields:

- `policy_id`
- `approved_profiles`
- `waivers`
- `rollout_conditions`


### 3.7 Filing

Required fields:

- `jurisdiction`
- `filing_period`
- `required_inputs`
- `calculation_rules`
- `signoff_requirements`
- `submission_deadline`


### 3.8 Submission

Required fields:

- `target_filing_id`
- `submitted_at`
- `submitted_by`
- `submission_proof`
- `accepted_or_rejected`


### 3.9 Literature

Required fields:

- `research_question`
- `sources`
- `key_findings`
- `gaps`
- `relevance_map`


### 3.10 Method

Required fields:

- `design`
- `data_inputs`
- `steps`
- `validity_risks`
- `reproducibility_notes`


### 3.11 Sources

Required fields:

- `source_entries`
- `credibility_notes`
- `coverage_gaps`
- `refresh_expectations`


### 3.12 Itinerary

Required fields:

- `stops`
- `timing`
- `bookings`
- `fallbacks`
- `constraints`


### 3.13 Scenarios

Required fields:

- `scenario_list`
- `base_case`
- `upside_case`
- `downside_case`
- `decision_use`


### 3.14 Inventory

Required fields:

- `document_list`
- `required_for`
- `source_locations`
- `missing_items`
- `retention_requirements`


## 4. Traceability Requirement

Evidence MUST be traceable to the artifact revision it validates.

At minimum, systems SHOULD support chains such as:

- `acceptance_criterion -> evidence`
- `decision -> approval`
- `spec revision -> plan -> task -> evidence`

## 5. Derived Bundles

The protocol MAY define bundles such as "release packet" or "research packet",
but a bundle MUST be modeled as a relation over semantic artifacts rather than
as a replacement artifact type.

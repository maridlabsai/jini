# Jini Profiles

## 1. Purpose

Profiles define operating density and vertical strictness without changing the
kernel.

A profile is a named composition of:

- required extensions
- required artifacts
- approval depth
- evidence burden
- operational obligations

## 2. Operating Profiles

### 2.1 Explore

For:

- learning
- prototyping
- low-risk ideation
- personal planning

Required minimum:

- Brief
- lightweight Decision
- lightweight Evidence

Traits:

- minimal ceremony
- no central approver
- fast iteration


### 2.2 Delivery

For:

- normal software/product work
- services delivery
- standard planning workflows

Required minimum:

- Brief
- Spec
- Plan
- Tasks
- Evidence

Traits:

- named owner
- moderate evidence burden


### 2.3 Critical

For:

- production changes with meaningful customer or business impact
- platform changes
- major releases

Required minimum:

- Delivery artifacts
- Signals
- Rollback
- Runbook
- independent verification

Traits:

- named operator
- named rollback authority
- stronger operational controls


### 2.4 Regulated

For:

- healthcare
- defense
- law enforcement
- finance or legal filing

Required minimum:

- Critical artifacts
- stricter Approval semantics
- audit-grade provenance
- segregation of duties when required

Traits:

- traceability
- stronger authority model
- stricter waiver controls


### 2.5 Incident

For:

- active incidents
- mission-impacting failures

Required minimum:

- Incident
- severity declaration
- incident commander
- backfill obligations

Traits:

- overrides normal sequencing
- does not waive provenance
- requires post-incident updates


## 3. Vertical Profiles

Verticals are defined as compositions of extension axes and operating profiles.

### 3.1 Software / SaaS

- base profile: Delivery or Critical
- common extensions:
  - Business: SaaS
  - Modality: software
  - Environment: cloud

### 3.2 Agentic AI

- base profile: Delivery or Critical
- additional requirements:
  - tool boundaries
  - memory model
  - eval burden
  - human override expectations

### 3.3 Mobile

- base profile: Delivery or Critical
- additional requirements:
  - offline behavior
  - sync rules
  - device constraints

### 3.4 Web and Search

- base profile: Delivery or Critical
- additional requirements:
  - accessibility
  - information architecture
  - ranking or abuse controls when relevant

### 3.5 DevOps / Platform

- base profile: Critical
- required artifacts:
  - Dependency
  - Signals
  - Rollback
  - Runbook

### 3.6 Healthcare

- base profile: Regulated
- additional requirements:
  - safety-critical controls
  - health-data constraints
  - explicit human override paths

### 3.7 Defense

- base profile: Regulated
- additional requirements:
  - classification handling
  - disconnected or air-gapped operation
  - multi-party authorization

### 3.8 Law Enforcement

- base profile: Regulated
- additional requirements:
  - evidentiary integrity
  - custody/disclosure obligations
  - field operations support

### 3.9 Retail

- base profile: Delivery or Critical
- additional requirements:
  - inventory correctness
  - pricing correctness
  - fraud controls
  - seasonal operating scale

### 3.10 E-signature / DocuSign-like

- base profile: Critical or Regulated
- additional requirements:
  - legal enforceability
  - identity assurance
  - tamper evidence
  - retention policy

### 3.11 Rideshare / Dispatch

- base profile: Critical
- additional requirements:
  - dispatch state
  - geospatial dependency awareness
  - safety controls
  - real-time degradation handling

### 3.12 Research Projects

- base profile: Explore or Delivery
- additional requirements:
  - Literature
  - Method
  - provenance and reproducibility evidence

### 3.13 Fields of Study

- base profile: Explore
- additional requirements:
  - Sources or equivalent
  - concept synthesis
  - uncertainty tracking

### 3.14 Travel Planning

- base profile: Explore or Delivery
- additional requirements:
  - Itinerary
  - Budget when cost matters
  - fallback/cancellation contingencies

### 3.15 Budgeting

- base profile: Explore or Delivery
- additional requirements:
  - Budget
  - Scenarios
  - variance thresholds

### 3.16 Tax / Regulated Filing

- base profile: Regulated
- additional requirements:
  - Filing
  - Inventory
  - Calculation traceability
  - Submission
  - mandatory signoff where required

## 4. Profile Selection Rule

When multiple profiles might apply, the stricter profile SHOULD win unless the
user or governing system explicitly approves a lighter profile.

## 5. Process Budget Rule

Every profile MUST define a maximum process burden and a minimum control burden.
This prevents both bureaucracy and under-control.

# Jini Memory

## 1. Purpose

Jini uses a lightweight Memory layer so work compounds across sessions without
forcing every workflow to reload raw material.

The Memory layer is intentionally small:

- `knowledge/`
- `projects/`
- `people/`

These directories are not the protocol itself. They are the persistent context
surface that feeds WorkUnits, packs, and derived views.

## 2. Directory Roles

### 2.1 `knowledge/`

Evergreen facts, concepts, patterns, domain maps, and shared operating rules.

Use it for:

- product and market notes
- architecture principles
- domain glossaries
- durable research summaries
- reusable constraints and heuristics

Avoid storing one-off project chatter here.

### 2.2 `projects/`

The compounding unit of active work.

Each project folder should contain:

- a short project index
- linked WorkUnits
- source material
- synthesized summaries
- derived views like PRDs

Projects are where research becomes product direction and then build work.

### 2.3 `people/`

Stakeholder memory and working preferences.

Use it for:

- role and ownership context
- review preferences
- risk tolerance
- communication style
- recurring concerns and decision biases

Do not turn this into a CRM dump. Keep it operational.

## 3. Loading Rules

Jini SHOULD load context in three tiers:

### 3.1 Always Loaded

- top-level project index
- current WorkUnit summary
- critical stakeholder notes

### 3.2 Navigated

- project sub-indexes
- research summaries
- active decisions

### 3.3 On Demand

- raw transcripts
- full datasets
- long attachments
- historical deep archives

This keeps context useful without turning every run into a token burn.

## 4. Summary-First Rule

Jini SHOULD prefer:

- summaries before transcripts
- definitions before SQL
- PRD view before raw notes
- project index before scanning folders

The protocol still binds to canonical artifacts, but the working surface should
be summary-first.

## 5. Minimal Folder Contracts

### 5.1 `knowledge/index.md`

Should answer:

- what durable knowledge exists
- where the most important references live
- what is safe to reuse broadly

### 5.2 `projects/index.md`

Should answer:

- what projects are active
- which project is canonical for a given topic
- which WorkUnits are currently live

### 5.3 `people/index.md`

Should answer:

- who matters most to current work
- which dossiers are high-signal
- which people require special review or communication handling

## 6. Project Folder Contract

Each project folder SHOULD contain:

- `README.md` or `index.md`
- `context/`
- `research/`
- `workunits/`
- `views/`

This keeps research, canonical artifacts, and rendered outputs close together.

## 7. Protocol Relationship

The Memory layer feeds Jini, but does not replace:

- WorkUnit state
- semantic artifacts
- guarded transitions
- approvals
- evidence

The rule is simple:

- Memory stores source material and reusable working context
- Jini artifacts store canonical state

## 8. Design Goal

The Memory layer should make Jini:

- more adaptive through better context selection
- more accurate through source traceability
- more detailed through layered summaries
- more efficient through smaller default context loads

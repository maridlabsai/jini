# Travel Curated Experience Framework Review

Updated: 2026-05-19

## Review Personas

### Travel Product Lead

Focus:

- whether the framework leads to better first outputs than generic itinerary
  drafting

### UX Researcher

Focus:

- whether the user can understand the flow without travel-agent training

### Competitive Analyst

Focus:

- whether the framework learns the right lessons from Layla and Navan Edge
  without claiming unsupported parity

### Trust And Operations Lead

Focus:

- whether execution boundaries stay explicit enough for future booking or
  disruption handling

## Round 1 Findings

### Finding 1

The first draft was too close to "make travel better" language and did not pin
the experience to a clear sequence.

Risk:

- the team could still ship generic itinerary output and claim compliance

### Finding 2

The benchmark references were too vague.

Risk:

- the team would cite Layla or Navan Edge loosely instead of translating
  concrete product lessons

### Finding 3

The planning-versus-execution boundary was under-specified.

Risk:

- future commercial work could quietly smuggle booking-like behavior into the
  public surface

### Finding 4

Continuity was present but not explicit enough about what must survive across
devices.

Risk:

- mobile and desktop teams could interpret "resume trip" differently

## Revisions Applied

- added a six-layer public experience model
- made scoped brief, curated options, itinerary object, smart references,
  continuity, and confirmation-first trust explicit
- added concrete benchmark references and translated lessons
- split shipping now versus later
- added explicit relationship to the commercial repo
- strengthened the shared invariants and rejection rules

## Rationalized Position

The right public travel posture is:

- Layla-like curation first
- Navan-like confirmation and preference discipline second
- live execution only when the infrastructure exists

That gives Jini a competitive travel direction without making false claims
about booking or disruption capability today.

## Final Verdict

`PASS`

The framework is now concrete enough to guide product work and strict enough
to prevent fake parity claims.

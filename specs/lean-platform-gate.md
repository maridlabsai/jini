# Lean Platform Gate

This gate keeps Jini lean, efficient, cheap to run, and easy to buy.

## Gate Categories

### 1. Cost Discipline

The product must preserve `lowest-total-cost-to-useful-outcome` and `cheap-by-default`.
Token frugality is P0 and must be treated as a first-order gate, not a generic
cost optimization.

That includes:

- `token-frugality-p0`
- `throttle-driven-platform-switching`
- `task-shaped-model-selection`
- `power-and-battery-aware-routing`

### 2. Latency Discipline

The product must preserve `time-to-first-useful-result` and `resume-cost`.

### 3. Command-Surface Discipline

The product must preserve `one-stable-surface`, low `command-surface-count`, and `no-compatibility-aliases` in the taught surface.

This discipline also applies to skills and agent interactions. Specialist
helpers must stay reachable through natural intake and progressive disclosure;
they must not become a second command tree or visible agent control plane.
The free tier must not include a skills-based OS productivity suite.

### 4. Visible Efficiency

The product must preserve `visible-efficiency` so users can understand why Jini saved time or money.

### 5. Buyability

The product must preserve `fewer-product-ideas-better-execution` and keep the public story easy to explain, deploy, govern, benchmark, and justify.

## Reject Conditions

Reject any change that:

- teaches a non-canonical alias
- adds a new taught command without removing another
- adds a multiword taught command when a standard one-word command exists
- slows first-turn or continuation paths without a measurable user benefit
- increases premium routing without measurable quality or speed gain
- increases token load, transcript replay, or verbose output without measurable
  quality, trust, or safety gain
- adds product ceremony before useful output
- teaches skill or agent vocabulary as a prerequisite to normal use
- ships developer agents, tester agents, `skills`, `delegate`, or a skills-based OS productivity suite in the free tier
- shows agent trees, role theater, or orchestration logs by default
- prevents Jini from moving between underlying platforms when throttling or
  quota pressure makes the current route the wrong one
- removes or weakens task-shaped model selection so a single model/profile is
  used regardless of task depth, modality, or complexity
- removes or weakens powered-mode full power execution when local capability is
  useful and safe
- removes or weakens low-battery or thermal-aware execution so local routes
  keep burning device resources when a smaller, deferred, or remote route would
  preserve the outcome
- splits offline and online execution into separate transcripts, task ids, or
  route histories instead of stitching them into one session timeline

## Required Regression Inputs

- `token-frugality-p0`
- `offline-online-session-stitching`
- `lowest-total-cost-to-useful-outcome`
- `one-stable-surface`
- `cheap-by-default`
- `visible-efficiency`
- `fewer-product-ideas-better-execution`
- `cost-per-successful-task`
- `time-to-first-useful-result`
- `resume-cost`
- `command-surface-count`
- `skill-agent-interaction-simplicity`
- `route-feedback-health`
- `route-feedback-impact`
- `no-compatibility-aliases`
- `throttle-driven-platform-switching`
- `task-shaped-model-selection`
- `power-and-battery-aware-routing`

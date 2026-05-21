# Lean Platform Gate

This gate keeps Jini lean, efficient, cheap to run, and easy to buy.

## Gate Categories

### 1. Cost Discipline

The product must preserve `lowest-total-cost-to-useful-outcome` and `cheap-by-default`.

### 2. Latency Discipline

The product must preserve `time-to-first-useful-result` and `resume-cost`.

### 3. Command-Surface Discipline

The product must preserve `one-stable-surface`, low `command-surface-count`, and `no-compatibility-aliases` in the taught surface.

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
- adds product ceremony before useful output

## Required Regression Inputs

- `lowest-total-cost-to-useful-outcome`
- `one-stable-surface`
- `cheap-by-default`
- `visible-efficiency`
- `fewer-product-ideas-better-execution`
- `cost-per-successful-task`
- `time-to-first-useful-result`
- `resume-cost`
- `command-surface-count`
- `no-compatibility-aliases`

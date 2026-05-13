---
title: Simple Guide
description: The plain-language path for understanding what Jini does without learning framework terms first.
---

Jini helps you answer three questions:

- What is done?
- What happens next?
- What is still missing?

That is all.

<div class="section-card">
  <strong>If you only remember one thing:</strong>
  <p>Jini is a way to slow down just enough to tell the truth about work before you call it finished.</p>
</div>

You do not need to learn the whole framework first.

## When Jini Helps

Jini helps when work gets messy.

For example:

- a meeting ends and nobody is sure what to do next
- a plan looks finished, but nobody knows if it is really ready
- a decision was made, but the reason is getting lost
- a problem is fixed, but nobody knows if the work is truly closed

## First Try

Run this:

```bash
jini try-example research-prd
```

You will see lines like:

```text
STATE  awaiting_verification
NEXT   Verify
MISSING-LATER
  - Approval
```

Read them like this:

- `STATE`: where the work is now
- `NEXT`: the next honest step
- `MISSING-LATER`: what will cause trouble if you ignore it

## Why This Matters

Many teams say work is done too early.

Jini helps you slow down just enough to see:

- what is really done
- what still needs checking
- what is missing before you call the work finished

## Four Common Cases

### Meeting Follow-up

Question:

**Did that meeting produce real follow-through, or just notes?**

Run:

```bash
jini try-example meeting-followup
```

### Research To PRD

Question:

**Is this spec really ready, or does it only look ready?**

Run:

```bash
jini try-example research-prd
```

### Vendor Selection

Question:

**Can we still explain this choice later?**

Run:

```bash
jini try-example vendor-selection
```

### Incident Response

Question:

**Is the incident really closed?**

Run:

```bash
jini try-example incident-response
```

## Smallest Useful Habit

Use this after important work:

```bash
jini status-pack <path-to-work>
```

It helps you see:

- what state the work is in
- what happens next
- what is missing

## If You Want More

<div class="section-card">
  <h3>Go next</h3>
  <div class="on-this-page">
    <a href="./index.md">Home</a>
    <a href="./examples.md">Examples</a>
    <a href="./cli.md">CLI Guide</a>
    <a href="./install.md">Install</a>
  </div>
</div>

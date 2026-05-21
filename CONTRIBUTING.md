# Contributing to Jini

Thanks for the interest.

## Before Opening a Pull Request

1. Open an issue first for large changes, new packs, or adapter additions.
2. Keep the protocol core small. Prefer packs, routines, adapters, or docs over
   expanding the core surface.
3. Preserve the framework rules:
   - lean beats broad when the two conflict
   - defaults should stay simple
   - durable state belongs in artifacts, not chat history

## Local Validation

Run the current CLI suite before sending changes:

```bash
make test-cli
make test-docs
make readiness
```

If you want the full public regression lane, use:

```bash
make test
```

If you change docs, examples, or install metadata, make sure command names,
bundle names, and pack names still line up.

If you change install, setup, launcher copy, route selection, `provider doctor`,
or beginner docs, also review the standing persona gate in
[specs/dogfood-gates.md](./specs/dogfood-gates.md).

If you make a major product decision or push a meaningful rewrite slice, also
review the scorecard gate in:

- [specs/competitive-kpis.yaml](./specs/competitive-kpis.yaml)
- [specs/golden-competitive-benchmark.yaml](./specs/golden-competitive-benchmark.yaml)
- [specs/rewrite-score-baseline.yaml](./specs/rewrite-score-baseline.yaml)

Use those files before deciding and before pushing. The rewrite does not pass
just because the change feels simpler locally. It must still preserve or
improve the scorecard lead and clear the locked rewrite floor.

Those slices do not pass by being technically correct alone. They pass only if:

- a low-literacy first-time user can start
- an AWS Bedrock user can force the route they expect
- an Azure enterprise user can stay inside the Azure path confidently
- a pragmatic “just make it work” user can get value without learning internals
- Claude, Codex, ChatGPT, and Gemini-style users can understand their path
- power users, AI engineers, QA testers, architects, AI PMs, software leaders,
  students, homemakers, and domain users can still make sense of install,
  setup, and usage without hidden policy surprises

## Pull Request Expectations

- one focused change per pull request
- clear description of user impact
- notes on risks or follow-up work when relevant
- tests for behavior changes
- scorecard impact called out for major rewrite slices
- push-ready evidence when product shape changed:
  - current benchmark result
  - current rewrite floor status
  - affected dogfood persona result when relevant

## Good First Contributions

- doc clarifications
- pack examples
- install and onboarding polish
- issue template improvements
- adapter conformance improvements

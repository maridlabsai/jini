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
python3 tools/jini.py --help
python3 -m unittest tests/test_jini_cli.py -v
```

If you change docs, examples, or install metadata, make sure command names,
bundle names, and pack names still line up.

## Pull Request Expectations

- one focused change per pull request
- clear description of user impact
- notes on risks or follow-up work when relevant
- tests for behavior changes

## Good First Contributions

- doc clarifications
- pack examples
- install and onboarding polish
- issue template improvements
- adapter conformance improvements

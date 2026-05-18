# Go Runtime Preview

This directory contains the first Go-based Jini runtime slice.

Current scope:

- `jini`
- `jini check`
- `jini open`
- `jini run` (launcher alias)

This runtime is intentionally narrow.

It reads the existing remembered-work record and work directory without
changing them:

- `.jini/current-work.json`
- `work-unit.yaml`
- `views/`
- `artifacts/`
- `exports/`

Install locally with:

```bash
./install.sh
```

Run locally with:

```bash
jini
```

The rewrite contract that defines the intended product shape lives in:

- [specs/product-rewrite-contract.md](../../specs/product-rewrite-contract.md)

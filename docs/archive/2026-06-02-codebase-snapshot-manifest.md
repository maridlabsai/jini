# Codebase Snapshot Manifest

Updated: 2026-06-02

## Purpose

This manifest records the archived baseline snapshots created before the next
major Jini initiative proceeds.

Both the free/open repo and the commercial repo were archived locally as Git
bundles and verified immediately after creation.

## Snapshot Inventory

### Free/Open Repo

- repo: `/Users/sharad.sharma/Developer/jini`
- head: `fe5b94e1aa1d58b013f71d1843935ec795edf2e8`
- bundle:
  `/Users/sharad.sharma/Developer/jini-archives/2026-06-02/jini-fe5b94e.bundle`
- creation command:
  `git -C /Users/sharad.sharma/Developer/jini bundle create /Users/sharad.sharma/Developer/jini-archives/2026-06-02/jini-fe5b94e.bundle --all`
- verification command:
  `git -C /Users/sharad.sharma/Developer/jini bundle verify /Users/sharad.sharma/Developer/jini-archives/2026-06-02/jini-fe5b94e.bundle`

### Commercial Repo

- repo: `/Users/sharad.sharma/Developer/jini-commercial`
- head: `b40ebffd4a01b2b389b0413117332129b973e4cd`
- bundle:
  `/Users/sharad.sharma/Developer/jini-archives/2026-06-02/jini-commercial-b40ebff.bundle`
- creation command:
  `git -C /Users/sharad.sharma/Developer/jini-commercial bundle create /Users/sharad.sharma/Developer/jini-archives/2026-06-02/jini-commercial-b40ebff.bundle --all`
- verification command:
  `git -C /Users/sharad.sharma/Developer/jini-commercial bundle verify /Users/sharad.sharma/Developer/jini-archives/2026-06-02/jini-commercial-b40ebff.bundle`

## Restore Commands

### Restore Free/Open Repo

```bash
git clone /Users/sharad.sharma/Developer/jini-archives/2026-06-02/jini-fe5b94e.bundle jini-restored
```

### Restore Commercial Repo

```bash
git clone /Users/sharad.sharma/Developer/jini-archives/2026-06-02/jini-commercial-b40ebff.bundle jini-commercial-restored
```

## Verification Notes

- both bundles recorded complete history
- both bundles were verified successfully on 2026-06-02
- this manifest exists so future migration work can always roll back to the
  exact pre-initiative state

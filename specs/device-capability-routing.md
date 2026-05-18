# Device Capability Routing

Updated: 2026-05-16

## Purpose

Jini should choose the best local SLM path for the user's actual device, not
just for the task.

The product goal is:

- highest useful productivity
- lowest justified expense
- automatic use of newly unlocked local capabilities after OS, runtime, driver,
  or Jini upgrades

## Core Rule

Local routing must consider all of these together:

1. task shape
2. device hardware
3. OS and OS version
4. installed local runtime stack
5. measured local reliability

Task-only routing is not enough.

## Capability Probe Surface

Jini should probe and persist:

- OS
- OS version
- CPU architecture
- CPU count
- total memory
- accelerator class
- local runtime class
- derived device class
- derived local profile availability
- endpoint/runtime signature
- Jini version that captured the profile
- capability registry version
- capture timestamp

This should be stored as a versioned repo-local device profile.

## Device Classes

Jini should classify the local machine into one of:

- `tiny`
- `laptop-light`
- `laptop-strong`
- `workstation`
- `gpu-heavy`

These are product-facing heuristics, not hardware marketing terms.

## Local Profile Availability

Jini should convert the device class into profile support states:

- `available`
- `limited`
- `unavailable`

For these local profiles:

- `local-fast`
- `local-workhorse`
- `local-deep`
- `local-multimodal`

The final profile state should be the intersection of:

- hardware potential
- local runtime/backend presence
- model/profile configuration

## Upgrade Rule

Jini must re-probe when:

- the cached device profile is stale
- the Jini version changed
- the capability registry version changed
- the OS or OS version changed
- the CPU architecture changed
- the local runtime/backend changed
- the configured local endpoint changed
- the configured local profile mapping changed

This is how Jini rides on top of newly unlocked capabilities over time instead
of freezing local heuristics to the first install state.

## User Trust Rule

When Local SLM is active, the user should be able to see:

- device class
- local accelerator class
- local runtime class
- local profile
- model
- why this choice was made

## Cost Rule

The policy remains:

- cheapest suitable route first
- stronger route only when justified

But "cheapest suitable" must be device-aware.

An unusably slow or unstable local route is not cheap in productivity terms.

## Acceptance Criteria

This slice is only complete when all are true:

- device profile is versioned and cached
- upgrade-aware re-probing exists
- local profile availability changes by device class and backend readiness
- route scoring uses device class
- provider doctor exposes device capability state
- the work has an independent validator gate

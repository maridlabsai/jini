# Decision-Tree Backtrack — Design (forward-looking)

Status: Planning sketch (NOT build-ready). Depends on the native agentic loop,
which is not yet designed/built. Tier: **Pro only** (user decision 2026-08-01).
Repo: commercial (paid), per [[paid-autopilot-boundary]] and the tier boundary.

## Vision

During an autonomous (native-loop) run, Jini records a TREE of decision points.
The user can view the tree, see the alternatives that existed at each point,
pick a different branch at any node, and Jini **rewinds workspace + agent
context to that node and re-runs down the alternate path.** It is "inspect the
agent's reasoning and redirect it" — time-travel for an autonomous run — not a
passive audit log. This is a flagship Pro differentiator against one-shot
agents.

## Tiering (user-ratified)

- **Free / Plus** native loop: executes; may keep a COMPACT LINEAR trail
  ("what it did") for receipts/transparency — no enumerated alternatives, no
  view, no backtrack.
- **Pro**: full decision TREE with alternatives at each node, `jini tree`
  view, and backtrack/branch. Recording the alternatives-tree is the expensive
  part and pairs with Pro. View + backtrack are Pro-gated (fail closed).

## Decisive dependency (the highest-leverage planning insight)

Backtrack REQUIRES the native loop, because the system must own and restore
state at each decision point. Delegated `claude --print` / `codex exec` run
one-shot and cannot be rewound per-decision (at best re-prompted whole-task).
Therefore:

> **The native loop must be designed from day one with two seams: (1)
> decision-point enumeration — at each juncture the loop surfaces the top-N
> viable options + rationale, acts on one, records the rest; (2) checkpoint
> hooks — snapshot workspace + context at each node.**

Even the free/basic loop should carry these seams (as no-ops / linear) so the
Pro tree layer is an addition, not a rewrite. Get this wrong and Pro backtrack
becomes a native-loop rebuild later.

## What a "decision point" is

A juncture where the agent chose among ≥2 viable options, e.g.: approach/plan
selection; which file(s) to target; the specific edit; which command to run;
branch on a tool result (test failed → fix vs revert). Not per-token — that is
too granular and too costly. The loop asks the model for candidate actions with
rationale at these junctures, picks one, and records the alternatives.

## State captured per node

- **Workspace**: git checkpoint. Use a dedicated ref namespace
  (`refs/jini/checkpoints/<run>/<node>`) or commits so the user's own history
  isn't polluted; shadow-init git if the project isn't a repo.
- **Agent context**: serialized message/state history →
  `.jini/decision-tree/<run>/<node>.json`.
- **Alternatives**: the enumerated not-taken options + rationale, so the user
  can pick one without re-derivation.
- **Metadata**: parent node, chosen option, timestamp, cost (tokens/$),
  result summary.

## Tree structure + persistence

Nodes = decision points; the run is a root→leaf path. Backtracking creates a
new branch from a node, so the structure grows into a tree. Persist under
`.jini/decision-tree/<run-id>/` (per-node files + a tree index). Cap depth /
branch count and prune to bound combinatorial growth.

## Surfaces (Pro)

- `jini tree` — render the tree (ANSI): nodes, the taken path, alternatives per
  node, per-node cost, outcome. `--format json` for tooling.
- `jini tree branch <node> [--choose <alt>]` — restore the node's workspace
  checkpoint + context, then re-run the loop from there taking the alternate
  option (or re-decide with the chosen alt as a constraint). Records a new
  branch.

## Honest framing: exploration, not deterministic replay

Re-running a branch will not reproduce exactly (model sampling, external state).
Frame it to the user as "explore the alternate path," not "replay" — otherwise
the promise is false. This matters for messaging (keep it simple + honest).

## Safety

- Workspace restore mutates files → guard: refuse/stash if there are uncommitted
  changes outside jini's checkpoints; only ever touch paths within the trusted
  directory (ties into the posture trust model — backtrack is an autonomous
  action and needs the same consent).
- Backtrack is itself an autonomous edit/execute action → gated by the same
  directory-trust + posture consent as any autonomous run.

## Public seam (mirrors the autopilot seam)

The public native loop exposes a `DecisionRecorder` interface (record node,
enumerate alternatives hook, checkpoint hook). Free/basic uses a no-op or
linear recorder; the commercial module registers the full tree recorder +
`jini tree`/backtrack (Pro-gated via `feature_policy`). Keeps paid IP in the
commercial repo; public ships only the seam + basic linear trail.

## Risks

- **Cost**: enumerating alternatives = more/larger model calls. Pro users pay;
  acceptable, but make enumeration breadth tunable.
- **Restore fidelity**: context restore is message-history replay, not true
  process state — good enough, but not perfect.
- **Combinatorial growth**: cap + prune.
- **Determinism**: see framing above.

## Phasing (once the native loop exists)

1. Native loop with decision-point + checkpoint SEAMS (public; free/basic).
2. Commercial: tree recording with alternatives per node.
3. Commercial: `jini tree` view (Pro-gated).
4. Commercial: backtrack + branch (restore + re-run).

## Bottom line

This does not start now — it is downstream of the native-loop initiative. Its
one hard demand on that initiative is that the loop be built WITH
decision-point enumeration + checkpoint seams from the start. Capture that as a
requirement on the native-loop design so this Pro capability is a layer, not a
rewrite.

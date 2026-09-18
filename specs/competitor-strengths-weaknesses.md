# Competitor Strengths & Weaknesses (forum-grounded)

Evidence-grounded intelligence on the major AI-coding competitors, drawn from 2026
public developer sentiment (Reddit, HN, DEV, review sites, the OpenAI forum, GitHub
issues). Each **strength** is a bar Jini must match or honestly concede; each
**weakness** is a wedge Jini's architecture is meant to exploit. This file is the
source of truth for the competitive-posture test (`competitor_posture_test.go`) and
feeds `competitive-intel.yml` + the golden benchmark.

Legend: 🟢 Jini matches/beats · 🟡 Jini partially covers · 🔴 Jini concedes (by design)

## OpenAI Codex
- **Strengths:** strong agentic terminal workflows; deep ChatGPT ecosystem tie-in.
- **Weaknesses:** weekly caps / Max plans hit zero in <90 min; GPT-5-Codex
  regression complaints; "paying for a tool that blocks you feels misleading."
- **Jini wedge:** 🟢 no throttle wall (fallback ladder) · 🟢 cost transparency
  (savings ledger) · 🟢 multi-provider, no single-vendor lock-in.

## AWS Kiro
- **Strengths:** spec-driven development — specs become *enforced constraints*, catch
  problems before bugs, produce traceable code + a living blueprint.
- **Weaknesses:** writing a good spec is hard; small tasks feel slower (spec
  overkill); punishing credit economics; bad for solo devs / rapid prototyping;
  credits drain even while idle.
- **Jini wedge:** 🟢 free-first (no credit anxiety) · 🟢 fast path for small tasks
  (no forced ceremony) · 🟡 planning/spec posture (has plan/semi/autonomous handoff;
  not a full spec IDE) · 🟢 solo-dev-friendly.

## Cursor
- **Strengths:** best-in-class Tab autocomplete; whole-codebase understanding;
  multi-file Composer refactors; **multi-model switching per task**; familiar VS
  Code UI.
- **Weaknesses:** no BYOK; opaque credit system; codebase-indexing privacy concerns;
  degrades on large repos (>100k LOC); inconsistent output quality.
- **Jini wedge:** 🟢 BYO keys/providers · 🟢 multi-model routing (matches the
  strength) · 🟢 cost transparency vs opaque credits · 🔴 no GUI/Tab-complete
  (terminal-native by design — concede the IDE-autocomplete lane).

## Windsurf
- **Strengths:** Cascade Flow plans across files with a single coherent context;
  "drives itself" (fewer prompts); largest free tier of the GUI tools.
- **Weaknesses:** GUI/IDE lock-in; still credit-metered at scale.
- **Jini wedge:** 🟢 free-first (matches/beats the free-tier draw) · 🟡 coherent
  multi-file context (native loop; not a Cascade-class planner yet) · 🟢 un-metered.

## Claude Code
- **Strengths:** deepest agentic autonomy; terminal-native; 1M-token context;
  multi-agent parallel tasks.
- **Weaknesses:** expensive; rate-limit drains (21%→100% on one prompt); capacity
  can't keep up → throttling + silent quality degradation.
- **Jini wedge:** 🟢 un-metered / throttle-survival · 🟢 free-first · 🔴 raw
  single-model frontier depth (Jini is "cheapest-capable", not "always-frontier" —
  concede the max-capability lane, win on flow + cost).

## Cross-cutting industry findings (apply to all)
- **Context engineering is THE differentiator** — reliably maintaining/updating
  project context as work progresses. 🟡 (Jini has memory + repo context; a
  first-class evolution target — see [[staying-current-is-vital]]).
- **Every tool is worse on existing codebases than greenfield.** 🟡 shared industry
  gap; an explicit quality target for Jini.
- **Reasoning-block replay causes self-reinforcing hallucination.** 🟢 Jini's
  cheap-first, context-frugal routing avoids replaying expensive reasoning state.
- **Hallucination ∝ cost** (wasted runs = spend); ~35% higher without fresh data.
  🟢 Jini's model-freshness (`jini check model`, catalog) keeps models current.
- **Cost efficiency + hallucination control now surpass raw capability** as the top
  developer evaluation criteria. 🟢 This is exactly Jini's wedge.

## Must-match strengths (Jini cannot concede these)
1. Multi-model flexibility (Cursor) → routing ✅
2. Free/low-friction entry (Windsurf) → free-first ✅
3. Terminal-native agentic autonomy (Claude Code) → native loop ✅
4. Planning before code (Kiro) → plan/semi/autonomous handoff 🟡

## Exploitable weaknesses (Jini's message must press these)
1. Throttle/credit walls (Codex, Kiro, Claude Code) → un-metered ✅
2. Opaque cost (Cursor, Kiro) → savings ledger ✅
3. Vendor lock-in / no BYOK (Cursor, Codex, Claude Code) → BYO providers ✅
4. Spec overhead for small tasks (Kiro) → fast path ✅

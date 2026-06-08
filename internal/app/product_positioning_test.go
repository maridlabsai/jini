package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductSettlingDecisionsGateCLIWedgeAndTierBoundaries(t *testing.T) {
	root := repoRootForMigrationTest(t)

	settling := readProductPositioningFile(t, root, "specs/product-settling-decisions.md")
	for _, want := range []string{
		"Jini is a CLI-first AI work router and durable session layer.",
		"Those tools and flows can be routes, adapters, or proof scenarios. They are not",
		"the product identity.",
		"The first product people should notice is the CLI.",
		"Anything that does not improve that wedge is not P0 for GTM.",
		"Claude Code and Codex first-minute parity is the highest-precedence",
		"Questions answer compactly and do not create work.",
		"Bare entities ask for intent and do not create work.",
		"Configured CLI route names require real installed-CLI handoff",
		"provider API routing is a separate route type.",
		"Adapter breadth is P0 only for familiar tools users already trust",
		"## CLI Handoff Decision",
		"provider API routes may exist, but they must not be marketed as Codex,",
		"Adapter support ships in waves:",
		"Wave 0, foundation: one handoff contract, route receipts, `doctor` detection",
		"Wave 1, terminal agents: Codex, Claude Code, Gemini CLI, Aider, and OpenCode.",
		"Wave 2, local and gateway routes: Ollama, LM Studio, OpenRouter, and LiteLLM.",
		"Wave 3, editor and cloud agents: Continue, Cline/Roo, Cursor, Windsurf, and",
		"P0.6: require adapter support waves to be scorecard-gated",
		"No hard-coded entity-to-template routing.",
		"P0.3: require the scorecard to include an intent-first routing outcome gate",
		"The free tier should prove Jini's routing and session value without giving away",
		"the commercial OS.",
		"Free tier does not include:",
		"developer-agent fleets",
		"skills-based OS productivity suite",
		"Commercial tier is where Jini becomes an agent and skills based OS productivity",
		"No new Jini conversation style.",
		"saved work is resumed through `status`, `continue`, `open`, or natural title matching",
		"no `Start/Keep` interruption model",
		"no visible `Switch` startup control",
		"Offline is a route state, not a separate product.",
		"Until the CLI wedge is noticeably strong, defer broad expansion.",
		"No drift without explicit agreement.",
		"Older broad PRDs, research notes, and platform plans are background only.",
		"Protected product and PRD surfaces must not change casually.",
		"The canonical near-term PRD must stay smaller than the older platform plans",
		"make competitor research or user learning sound like automatic feature scope",
		"bash tools/product_prd_drift_gate.sh",
	} {
		if !strings.Contains(settling, want) {
			t.Fatalf("product settling decisions must preserve %q", want)
		}
	}

	canonicalPRD := readProductPositioningFile(t, root, "specs/number-one-platform-prd.md")
	for _, want := range []string{
		"[product-settling-decisions.md](./product-settling-decisions.md)",
		"Jini is a CLI-first AI work router and durable session layer",
		"The near-term product is not the broad OS.",
		"Core charter: intent-first Claude/Codex parity outranks feature expansion.",
		"Route between familiar CLIs, providers, gateways, and local/offline models.",
		"Treat configured CLI routes as real installed-CLI handoffs",
		"provider API routes separately from CLI handoff routes.",
		"Bare `jini` is a task prompt, not a dashboard.",
		"[launcher-intake-design.md](./launcher-intake-design.md)",
		"[number-one-development-plan.md](./number-one-development-plan.md)",
		"Token frugality is P0.",
		"## Tier Boundary",
		"Commercial value must be materially higher than the free CLI: managed",
	} {
		if !strings.Contains(canonicalPRD, want) {
			t.Fatalf("canonical PRD must point to settled product positioning %q", want)
		}
	}

	readme := readProductPositioningFile(t, root, "README.md")
	for _, want := range []string{
		"Jini is a CLI-first AI work router and durable session",
		"switching among configured provider, model, and local routes",
		"Real downstream CLI handoff for names like `codex` and `claude-code` is a P0",
		"Adapter support waves:",
		"Wave 1: Codex, Claude Code, Gemini CLI, Aider, and OpenCode",
		"Wave 2: Ollama, LM Studio, OpenRouter, and LiteLLM-compatible gateways",
		"Wave 0 and Wave 1 are runtime-supported when the downstream CLI is installed",
		"Wave 2 and Wave 3 are planned targets, not shipped claims.",
		"Meeting, plan-readiness, travel, and vendor-comparison flows are proof",
		"[specs/product-settling-decisions.md](specs/product-settling-decisions.md)",
	} {
		if !strings.Contains(readme, want) {
			t.Fatalf("README must teach settled product positioning %q", want)
		}
	}
}

func TestFocusedDeliveryChainGatesPRDDesignAndImplementation(t *testing.T) {
	root := repoRootForMigrationTest(t)

	launcherDesign := readProductPositioningFile(t, root, "specs/launcher-intake-design.md")
	for _, want := range []string{
		"This is the active dev design for the CLI front door.",
		"Bare `jini` renders the same task prompt with or without saved work.",
		"Startup is not a saved-work dashboard.",
		"Do not add a new front-door interaction pattern from implementation alone.",
		"update `product-settling-decisions.md` first in the",
	} {
		if !strings.Contains(launcherDesign, want) {
			t.Fatalf("launcher design must preserve focused delivery rule %q", want)
		}
	}
	for _, stale := range []string{
		"startup should show a compact resume card",
		"keep `Start` and `Continue` as plain-language actions",
		"produce the artifact before the long explanation",
	} {
		if strings.Contains(launcherDesign, stale) {
			t.Fatalf("launcher design must not preserve stale launcher rule %q", stale)
		}
	}

	developmentPlan := readProductPositioningFile(t, root, "specs/number-one-development-plan.md")
	for _, want := range []string{
		"This is the active implementation plan. It is intentionally narrow.",
		"Make the CLI good enough for users to prefer trying Jini again.",
		"The core charter is intent-first Claude/Codex parity.",
		"Intent-First CLI Parity",
		"bare entities create `Task Snapshot`, itinerary, or other artifacts",
		"Paused work can restart only through `product-settling-decisions.md`.",
		"If that trace cannot be written in one short paragraph, the cut is too broad.",
	} {
		if !strings.Contains(developmentPlan, want) {
			t.Fatalf("development plan must preserve focused implementation rule %q", want)
		}
	}
	for _, stale := range []string{
		"make the CLI and apps feel like quantum jumps",
		"monthly public release train",
		"Workstream E: CLI And App Surface Specialization",
		"desktop as the rich review and artifact-edit surface",
	} {
		if strings.Contains(developmentPlan, stale) {
			t.Fatalf("development plan must not preserve stale broad-plan rule %q", stale)
		}
	}
}

func TestCanonicalPRDStaysReducedToCurrentGTMWedge(t *testing.T) {
	root := repoRootForMigrationTest(t)

	canonicalPRD := readProductPositioningFile(t, root, "specs/number-one-platform-prd.md")
	if lines := strings.Count(canonicalPRD, "\n") + 1; lines > 170 {
		t.Fatalf("canonical PRD must stay reduced; got %d lines", lines)
	}
	for _, want := range []string{
		"## P0 Outcome Requirements",
		"Bare `jini` is a task prompt, not a dashboard.",
		"no visible `Switch` startup control",
		"no `Start/Keep` interruption model",
		"no Working Draft for obvious file edits",
		"no hard-coded entity-to-template routing",
		"## Market And Learning Guards",
		"No competitor finding becomes active scope unless the decision record changes.",
		"task-first startup even with saved work",
		"real downstream CLI handoff and adapter support waves",
		"CLI UX, PRD drift, and scorecard gates in commit gates",
	} {
		if !strings.Contains(canonicalPRD, want) {
			t.Fatalf("canonical PRD must preserve current GTM requirement %q", want)
		}
	}
	for _, stale := range []string{
		"agentic automation in all aspects of running this company is non negotiable",
		"Active work` first",
		"Other active work`",
		"Type `Switch`",
		"Start/Keep way of thinking",
		"#### P0.10 Competitive release pressure",
		"#### P0.11 Compounding user productivity learning",
		"Commercial includes:",
		"Free includes:",
	} {
		if strings.Contains(canonicalPRD, stale) {
			t.Fatalf("canonical PRD must not preserve stale requirement %q", stale)
		}
	}
}

func TestPRDImplementationTraceCoversP0Requirements(t *testing.T) {
	root := repoRootForMigrationTest(t)

	trace := readProductPositioningFile(t, root, "specs/prd-implementation-trace.md")
	for _, want := range []string{
		"maps the canonical P0 requirements",
		"Start from a natural task in the current directory",
		"Edit local files directly when clear and safe",
		"Fail closed with exact ambiguity",
		"Answer simple questions compactly",
		"Ask intent for bare entities without artifacts",
		"Route between familiar CLIs, providers, gateways, and local/offline models",
		"Treat configured CLI routes as installed-CLI handoffs",
		"`cli_handoff.go`, `generateWithConfiguredProviderDecision`",
		"fake downstream CLI handoff smoke test",
		"Keep route, token, and runtime diagnostics inspectable",
		"Block regressions before commit and push",
		"macOS CLI handoff runs a Gatekeeper trust check before execution",
	} {
		if !strings.Contains(trace, want) {
			t.Fatalf("PRD implementation trace must preserve %q", want)
		}
	}
}

func readProductPositioningFile(t *testing.T, root, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

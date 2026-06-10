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
		"[product-streamline-redline.md](./product-streamline-redline.md)",
		"continues only while the Go kernel can preserve",
		"[agentic-development-operating-model.md](./agentic-development-operating-model.md)",
		"mandatory for non-trivial Jini engineering cuts",
		"coordinator-owned process",
		"must not become public UX, free-tier command",
		"Questions answer compactly and do not create work.",
		"Bare entities ask for intent and do not create work.",
		"Configured CLI route names require real installed-CLI handoff",
		"provider API routing is a separate route type.",
		"Adapter breadth is P0 only for familiar tools users already trust",
		"## CLI Handoff Decision",
		"provider API routes may exist, but they must not be marketed as Codex,",
		"route setup guidance is public through `jini route help`",
		"provider/local route setup without printing secrets",
		"Adapter support ships in waves:",
		"Wave 0, foundation: one handoff contract, route receipts, `doctor` detection",
		"Wave 1, terminal agents: Codex, Claude Code, Gemini CLI, Aider, and OpenCode.",
		"Wave 2, local and gateway routes: Ollama, LM Studio, OpenRouter, and LiteLLM.",
		"Wave 3, editor and cloud agents: Continue, Cline/Roo, Cursor, Windsurf, and",
		"P0.6: require adapter support waves to be scorecard-gated",
		"No hard-coded entity-to-template routing.",
		"P0.3: require the scorecard to include an intent-first routing outcome gate",
		"The free tier should prove Jini's routing and session value without giving away",
		"future commercial automation.",
		"CLI is available now. App surfaces, when shipped, are available to both free and",
		"commercial-only feature names fail closed in the",
		"real entitlement checks",
		"Subscription gates capabilities inside those",
		"Free tier does not include:",
		"developer-agent fleets",
		"skills-based OS productivity suite",
		"Commercial tier is the future managed automation layer for skills, delegation,",
		"Commercial value must be materially higher than the free surfaces:",
		"developer and tester agents hidden behind normal Jini outcomes",
		"No new Jini conversation style.",
		"saved work is resumed through `status`, `continue`, `open`, or natural title matching",
		"no `Start/Keep` interruption model",
		"no visible `Switch` startup control",
		"Offline is a route state, not a separate product.",
		"Until the CLI wedge is noticeably strong, defer broad expansion.",
		"macOS app PRD, UX design, HLD, and LLD for a Codex desktop-caliber surface",
		"The macOS app may proceed only as a focused desktop surface",
		"Focused implementation is the development philosophy.",
		"macOS app PRD: `specs/macos-app-prd.md`",
		"macOS app UX design: `specs/macos-app-ux-design.md`",
		"macOS app HLD: `specs/macos-app-hld.md`",
		"macOS app LLD: `specs/macos-app-lld.md`",
		"smallest change that advances the active CLI wedge",
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
		"[number-one-platform-hld.md](./number-one-platform-hld.md)",
		"[number-one-platform-lld.md](./number-one-platform-lld.md)",
		"[launcher-intake-design.md](./launcher-intake-design.md)",
		"[number-one-development-plan.md](./number-one-development-plan.md)",
		"[macOS app planning](./macos-app-prd.md)",
		"macOS app HLD/LLD for a Codex desktop-caliber session and artifact surface",
		"No release ships unless competitor-parity golden transcript gates",
		"Token frugality is P0.",
		"## Tier Boundary",
		"CLI is available now. App surfaces, when shipped, are available to both free",
		"and commercial users. Subscription gates capabilities, not the ability to",
		"install or open Jini.",
		"Commercial value must be materially higher than the free surfaces:",
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

	hld := readProductPositioningFile(t, root, "specs/number-one-platform-hld.md")
	for _, want := range []string{
		"This high-level design translates",
		"Jini does not ship an iteration unless the competitor-parity transcript gates",
		"Jini is five runtime layers:",
		"Intent boundary",
		"Simple questions must stop at step 2.",
		"Saved work is passive context.",
	} {
		if !strings.Contains(hld, want) {
			t.Fatalf("HLD must preserve architecture contract %q", want)
		}
	}

	lld := readProductPositioningFile(t, root, "specs/number-one-platform-lld.md")
	for _, want := range []string{
		"This low-level design defines the executable contracts",
		"Simple factual questions must not print `Result ready.`, `Task Snapshot`,",
		"`whats teh",
		"Do not create `current-work.json` for simple factual questions.",
		"If the transcript would surprise a Claude Code or",
		"rewrite-trigger candidate",
	} {
		if !strings.Contains(lld, want) {
			t.Fatalf("LLD must preserve runtime contract %q", want)
		}
	}

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
		"The development philosophy is focused implementation: ship the smallest",
		"stop before the work",
		"Make the CLI good enough for users to prefer trying Jini again.",
		"The core charter is intent-first Claude/Codex parity.",
		"Intent-First CLI Parity",
		"bare entities create `Task Snapshot`, itinerary, or other artifacts",
		"Paused work can restart only through `product-settling-decisions.md`.",
		"Non-trivial engineering cuts must use the internal sub-agent divide-and-conquer",
		"The coordinator owns scope splits, disjoint write sets, integration, and",
		"operating-model scope split and evidence",
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

func TestMacOSAppPRDPinsCodexParityWithoutProductDrift(t *testing.T) {
	root := repoRootForMigrationTest(t)

	prd := readProductPositioningFile(t, root, "specs/macos-app-prd.md")
	for _, want := range []string{
		"This is a specialized PRD for the Jini macOS app.",
		"Build a macOS app at the Codex desktop-caliber UX bar",
		"The app is not a new product",
		"same session model as the CLI",
		"Task first, not dashboard first.",
		"Claude/Codex familiarity beats Jini-specific workflow invention.",
		"Offline is a route state, not a separate product.",
		"Token frugality is visible when useful and silent when not useful.",
		"route evidence",
		"file diffs",
		"signed, notarized",
		"hardened runtime enabled",
		"Developer ID signed, notarized, stapled",
		"sandbox-entitlement reviewed",
		"Redacted diagnostics",
		"macOS app is available to free and commercial users.",
		"Subscription gates capabilities inside the app, not the app download",
		"No Start/Keep model.",
		"No visible Switch startup control.",
		"No saved-work dashboard on bare launch.",
		"No free-tier skills-based OS productivity suite.",
		"Default implementation direction is a Tauri 2 shell over the Go core.",
		"[macos-app-hld.md](./macos-app-hld.md) compares Tauri 2 with native",
		"Go core owns intent, routing, session graph",
		"PRD/HLD/LLD trace covers every P0 requirement.",
		"macOS file access review proves whether Phase 1 ships hardened-runtime-only",
		"validate the Tauri Phase 1 stack and trigger the native SwiftUI/AppKit escape",
		"Competitor scorecard includes Codex desktop parity checks",
	} {
		if !strings.Contains(prd, want) {
			t.Fatalf("macOS app PRD must preserve focused desktop requirement %q", want)
		}
	}
	for _, stale := range []string{
		"Start/Keep way of thinking",
		"desktop and mobile apps",
		"agentic automation in all aspects of running this company is non negotiable",
		"Current work\n-",
	} {
		if strings.Contains(prd, stale) {
			t.Fatalf("macOS app PRD must not preserve stale or broad requirement %q", stale)
		}
	}
}

func TestMacOSAppUXDesignPinsFirstMinuteDesktopContract(t *testing.T) {
	root := repoRootForMigrationTest(t)

	ux := readProductPositioningFile(t, root, "specs/macos-app-ux-design.md")
	for _, want := range []string{
		"This is the active UX design for the Jini macOS app.",
		"same session model as the CLI",
		"Use a three-pane desktop supervision shell",
		"left sidebar for projects, sessions, and search",
		"center thread for compact task input",
		"right inspector for progress, diffs, artifacts, route evidence",
		"Dashboard-first: too slow to first useful output",
		"Default launch:",
		"Simple question:",
		"Direct edit:",
		"Forbidden first-minute states:",
		"Start/Keep modal",
		"Switch control",
		"saved-work dashboard",
		"Task Snapshot for factual questions",
		"Working Draft for clear file edits",
		"route ceremony before a route matters",
		"agent role list before useful output",
		"Pass condition: `what is the capital of france?` returns `Paris.`",
		"Pass condition: no Working Draft, no generic artifact, no unnecessary plan.",
		"Pass condition: route named `codex` never silently becomes provider API.",
		"Pass condition: offline work does not fork the session.",
		"Approval request anatomy:",
		"Keyboard And Menu Contract",
		"VoiceOver labels for route state",
		"Golden transcripts pass for simple answer, direct edit, ambiguous edit",
		"Visual QA rejects Start/Keep, Switch, Task Snapshot, Working Draft",
	} {
		if !strings.Contains(ux, want) {
			t.Fatalf("macOS app UX design must preserve first-minute desktop contract %q", want)
		}
	}
	for _, stale := range []string{
		"Start/Keep way of thinking",
		"Current work\n-",
		"Type Switch",
		"Other active work",
	} {
		if strings.Contains(ux, stale) {
			t.Fatalf("macOS app UX design must reject stale UX language %q", stale)
		}
	}
}

func TestMacOSAppHLDDefinesGoCoreTauriShellBoundary(t *testing.T) {
	root := repoRootForMigrationTest(t)

	hld := readProductPositioningFile(t, root, "specs/macos-app-hld.md")
	for _, want := range []string{
		"This high-level design translates",
		"This HLD covers the macOS app only.",
		"Phase 1 uses a Tauri 2 shell over the Go core.",
		"Go owns product runtime behavior.",
		"Tauri owns the macOS window, shell packaging, updater, webview, and native",
		"TypeScript/Rust code must stay presentation and IPC only.",
		"SwiftUI/AppKit may be added only for narrow native controls",
		"Renderer can request side effects but cannot perform them.",
		"Only the Go sidecar may perform side effects.",
		"Phase 1 uses stdio JSON-RPC for the Go sidecar.",
		"Local TCP is not allowed for Phase 1",
		"introduce a `SessionIdentityMap`",
		"never create a session identity for transient simple answers",
		"macOS File Access Posture",
		"App Sandbox with security-scoped project bookmarks",
		"saved work is passive",
		"simple factual questions stop before work creation",
		"`codex` route never silently becomes provider API.",
		"offline work appends to the same session graph.",
		"Free and commercial users can install and open the macOS app.",
		"Entitlements gate capabilities, not the app shell.",
		"free app must not expose agent-role theater",
		"[macos-app-lld.md](./macos-app-lld.md) defines the app protocol",
	} {
		if !strings.Contains(hld, want) {
			t.Fatalf("macOS app HLD must preserve architecture boundary %q", want)
		}
	}
	for _, stale := range []string{
		"Electron",
		"local TCP is allowed",
		"provider API alias",
		"agent-role theater as default UX",
		"Windows:",
	} {
		if strings.Contains(hld, stale) {
			t.Fatalf("macOS app HLD must reject stale or unsafe architecture %q", stale)
		}
	}
}

func TestMacOSAppLLDDefinesExecutableContracts(t *testing.T) {
	root := repoRootForMigrationTest(t)

	lld := readProductPositioningFile(t, root, "specs/macos-app-lld.md")
	for _, want := range []string{
		"This low-level design defines the executable contracts for the Jini macOS app.",
		"Go owns product logic.",
		"The renderer owns presentation only.",
		"jini app serve --stdio --surface macos",
		"JSON-RPC-like messages over stdout/stdin",
		"no local TCP in Phase 1",
		"every request includes an idempotency key",
		"every event includes a monotonically increasing `sequence`",
		"`last_event_id` and `last_sequence`",
		"Retrying the same key returns the same terminal response and does not repeat",
		"projection_resync_required",
		"`app.handshake`",
		"`turn.submit`",
		"`approval.resolve`",
		"`route.status`",
		"`diagnostics.export`",
		"TransientResponseVM",
		"Simple factual questions return `TransientResponseVM`, not `TurnVM`.",
		"`TransientResponseVM` has no `session_id`",
		"RouteEvidenceVM",
		"ApprovalRequestVM",
		"action_digest",
		"target_digest",
		"scope_options",
		"Every approval binds `action_digest`, `target_digest`, `session_id`,",
		"Scoped approvals require explicit TTL, match rules, revocation surface, and",
		"ApprovalReceiptVM",
		"OfflineDebtVM",
		"DiagnosticsPreviewVM",
		"`TurnVM` is only for durable session turns.",
		"`compact_answer` cannot create session state by itself.",
		"Approval ids are single-use.",
		"`route_kind` must distinguish `cli_handoff`, `provider_api`, `gateway`,",
		"CLI handoff routes do not fall back to provider APIs.",
		"SessionIdentityMap",
		"One legacy pack dir maps to at most one `app_session_id`.",
		"Transient simple answers never allocate an `app_session_id`.",
		"App Sandbox with security-scoped project bookmarks",
		"Renderer must not change when persistence migrates",
		"simple answer returns `TransientResponseVM` with no session creation",
		"protocol idempotency prevents duplicate side effects",
		"route `codex` unavailable state does not fallback to provider API",
		"HLD and LLD are protected by PRD drift gate.",
	} {
		if !strings.Contains(lld, want) {
			t.Fatalf("macOS app LLD must preserve executable contract %q", want)
		}
	}
	for _, stale := range []string{
		"Working Draft for clear file edits",
		"Start/Keep",
		"Switch control",
		"renderer owns route choice",
		"simple answer returns compact turn with no session creation",
	} {
		if strings.Contains(lld, stale) {
			t.Fatalf("macOS app LLD must reject stale contract %q", stale)
		}
	}
}

func TestProductStreamlineRedlineDefinesRewriteTriggers(t *testing.T) {
	root := repoRootForMigrationTest(t)

	redline := readProductPositioningFile(t, root, "specs/product-streamline-redline.md")
	for _, want := range []string{
		"Current competitor research supports one clear product shape:",
		"OpenAI Codex CLI is a local terminal coding agent",
		"Claude Code emphasizes real developer tools, explicit permission",
		"Gemini CLI is an interactive terminal REPL",
		"GitHub Copilot cloud agent shows the platform direction",
		"Aider reinforces the basics",
		"## Design Alternatives",
		"## Selected Approach",
		"Jini continues with Approach B until a rewrite trigger fires.",
		"## Rewrite Triggers",
		"Three P0 first-minute transcript incidents recur",
		"without entering starter-pack or artifact-rendering code",
		"cannot map to a PRD outcome, HLD boundary, LLD contract, and",
		"Architecture quality does not compensate for a bad transcript.",
		"## Research Cadence",
	} {
		if !strings.Contains(redline, want) {
			t.Fatalf("streamline redline must preserve %q", want)
		}
	}
}

func TestAgenticDevelopmentOperatingModelPinsInternalDivideAndConquer(t *testing.T) {
	root := repoRootForMigrationTest(t)

	operatingModel := readProductPositioningFile(t, root, "specs/agentic-development-operating-model.md")
	for _, want := range []string{
		"internal engineering operating model, not a public product surface",
		"Use divide-and-conquer sub-agents by default for non-trivial Jini engineering",
		"This rule is mandatory for material cuts that change product behavior,",
		"Every non-trivial cut must name its trace before completion:",
		"lead agent remains accountable for the final answer, final diff, gate",
		"Reasoning: map the problem, constraints, invariants, and likely failure modes.",
		"Planning: turn the goal into ordered slices with explicit scope and gates.",
		"Design: define contracts, boundaries, data flow, and user-facing behavior.",
		"Coding: implement bounded changes inside an assigned write set.",
		"Testing: create or run focused regression checks and required gates.",
		"Code review: independently inspect the diff for bugs, regressions, missing",
		"Delegation is capped at two levels.",
		"one owner per writable file or glob",
		"Sub-agents can make Jini development stronger. They must not make Jini harder to",
		"If users need to understand the agent tree to trust or operate the default CLI,",
	} {
		if !strings.Contains(operatingModel, want) {
			t.Fatalf("agentic development operating model must preserve %q", want)
		}
	}

	executionPolicy := readProductPositioningFile(t, root, "specs/execution-routing-policy.md")
	for _, want := range []string{
		"does not create public `delegate` commands",
		"coordinator-owned divide-and-conquer is required for non-trivial Jini",
		"sub-agent write scopes must be disjoint and evidence-bound",
		"max delegation depth is 2",
		"overlapping sub-agent write scopes -> stop parallel work and serialize",
	} {
		if !strings.Contains(executionPolicy, want) {
			t.Fatalf("execution routing policy must preserve sub-agent boundary %q", want)
		}
	}

	delegationSlice := readProductPositioningFile(t, root, "specs/skills-and-delegation-slice.md")
	for _, want := range []string{
		"coordinator-owned process is mandatory for non-trivial engineering cuts",
		"is not a public UX promise",
		"internal engineering sub-agent controls",
		"exposes internal engineering sub-agent orchestration as a free-tier command",
		"It must not show agent trees, role theater, or step-by-step orchestration logs",
	} {
		if !strings.Contains(delegationSlice, want) {
			t.Fatalf("skills/delegation slice must preserve sub-agent tier boundary %q", want)
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
		"no `Result ready`, `Task Snapshot`, `Saved:`, or `Next: jini ...` shell around",
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
		"[number-one-platform-hld.md](./number-one-platform-hld.md)",
		"[number-one-platform-lld.md](./number-one-platform-lld.md)",
		"Start from a natural task in the current directory",
		"Edit local files directly when clear and safe",
		"Fail closed with exact ambiguity",
		"Answer simple questions compactly",
		"simple factual question tests including typo transcript",
		"Ask intent for bare entities without artifacts",
		"Route between familiar CLIs, providers, gateways, and local/offline models",
		"Treat configured CLI routes as installed-CLI handoffs",
		"`cli_handoff.go`, `generateWithConfiguredProviderDecision`",
		"fake downstream CLI handoff smoke test",
		"Gatekeeper rejection fail-closed regression",
		"Keep route, token, and runtime diagnostics inspectable",
		"Block regressions before commit and push",
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

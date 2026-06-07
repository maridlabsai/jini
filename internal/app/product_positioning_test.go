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
		"The free tier should prove Jini's routing and session value without giving away",
		"the commercial OS.",
		"Free tier does not include:",
		"developer-agent fleets",
		"skills-based OS productivity suite",
		"Commercial tier is where Jini becomes an agent and skills based OS productivity",
		"No new Jini conversation style.",
		"no `Start/Keep` interruption model",
		"Offline is a route state, not a separate product.",
		"Until the CLI wedge is noticeably strong, defer broad expansion.",
		"Protected product and PRD surfaces must not change casually.",
		"bash tools/product_prd_drift_gate.sh",
	} {
		if !strings.Contains(settling, want) {
			t.Fatalf("product settling decisions must preserve %q", want)
		}
	}

	canonicalPRD := readProductPositioningFile(t, root, "specs/number-one-platform-prd.md")
	for _, want := range []string{
		"[product-settling-decisions.md](./product-settling-decisions.md)",
		"Jini should be built as a CLI-first AI work router and durable session layer",
		"The near-term GTM product is not the broad OS.",
		"direct task intake, file edits, route switching,",
	} {
		if !strings.Contains(canonicalPRD, want) {
			t.Fatalf("canonical PRD must point to settled product positioning %q", want)
		}
	}

	readme := readProductPositioningFile(t, root, "README.md")
	for _, want := range []string{
		"Jini is a CLI-first AI work router and durable session",
		"switching among configured CLIs and model routes",
		"Meeting, plan-readiness, travel, and vendor-comparison flows are proof",
		"[specs/product-settling-decisions.md](specs/product-settling-decisions.md)",
	} {
		if !strings.Contains(readme, want) {
			t.Fatalf("README must teach settled product positioning %q", want)
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

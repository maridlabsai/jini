package app

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type featureAccess struct {
	ID       string
	Label    string
	Tier     string
	Allowed  bool
	Reason   string
	Fallback string
}

func featureAccessForID(featureID string) featureAccess {
	id := exactCommandToken(featureID)
	switch id {
	case "core-cli", "route-control", "local-routes", "configured-cli-handoff":
		return featureAccess{
			ID:       id,
			Label:    "Core Jini CLI",
			Tier:     "free",
			Allowed:  true,
			Reason:   "Core CLI, local routes, and configured CLI handoff stay available without a Jini subscription.",
			Fallback: "Run `jini route help` to configure local or BYO routes.",
		}
	case "commercial-skills", "commercial-delegation", "commercial-agents", "commercial-automation":
		label := commercialFeatureLabel(id)
		if currentSubscriptionTier() != "commercial" {
			return featureAccess{
				ID:       id,
				Label:    label,
				Tier:     "commercial",
				Allowed:  false,
				Reason:   label + " requires commercial entitlement and is not part of the free CLI.",
				Fallback: "Use `jini route help` for free local, BYO provider, and installed-CLI routing.",
			}
		}
		return featureAccess{
			ID:       id,
			Label:    label,
			Tier:     "commercial",
			Allowed:  false,
			Reason:   label + " is commercial-only and is not implemented in this public CLI build.",
			Fallback: "Use the normal `jini` task prompt, `jini route`, and `jini status` flows.",
		}
	default:
		return featureAccess{
			ID:       id,
			Label:    "Unknown capability",
			Tier:     "unknown",
			Allowed:  false,
			Reason:   "Unknown Jini capability.",
			Fallback: "Run `jini commands` to see the supported command surface.",
		}
	}
}

func commercialOnlyCommandAccess(raw string) (featureAccess, bool) {
	switch interactiveCommandName(raw) {
	case "skills", "skill":
		return featureAccessForID("commercial-skills"), true
	case "delegate", "delegation":
		return featureAccessForID("commercial-delegation"), true
	case "agents", "agent", "developers", "testers":
		return featureAccessForID("commercial-agents"), true
	case "automation", "automations":
		return featureAccessForID("commercial-automation"), true
	default:
		return featureAccess{}, false
	}
}

func renderFeatureAccessDenied(w io.Writer, access featureAccess) {
	fmt.Fprintln(w, "Not available in this CLI.")
	fmt.Fprintf(w, "%s\n", strings.TrimSpace(access.Reason))
	if fallback := strings.TrimSpace(access.Fallback); fallback != "" {
		fmt.Fprintf(w, "%s\n", fallback)
	}
}

func currentSubscriptionTier() string {
	switch exactCommandToken(firstNonEmpty(os.Getenv("JINI_SUBSCRIPTION_TIER"), os.Getenv("JINI_TIER"), "free")) {
	case "commercial", "paid", "team", "enterprise":
		return "commercial"
	default:
		return "free"
	}
}

func commercialFeatureLabel(featureID string) string {
	switch featureID {
	case "commercial-skills":
		return "Skills OS"
	case "commercial-delegation":
		return "Managed delegation"
	case "commercial-agents":
		return "Developer and tester agents"
	case "commercial-automation":
		return "Managed automation"
	default:
		return "Commercial feature"
	}
}

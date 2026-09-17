package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Automated model quality gate — `jini check model <route>`. Onboarding a model
// must be zero-touch AND quality-floored, with no human judgment. This runs a
// small set of OBJECTIVE, deterministic capability probes (instruction-following,
// output-contract adherence, code structure, constraint counting) against the
// route's live model and returns VERIFIED / UNVERIFIED. It is a capability floor
// that catches broken or incapable models — not a quality ranking (the golden
// competitive benchmark and real usage do that). Deterministic + LLM-judge-free
// keeps it fast, bias-free, and safe to run automatically in an onboarding gate.

type modelProbe struct {
	name   string
	prompt string
	// pass evaluates the model's raw reply objectively.
	pass func(reply string) bool
}

const modelProbeSystemPrompt = "You are a precise assistant. Follow the instruction exactly. Output only what is asked, with no preamble, explanation, or code fences unless the instruction asks for them."

// modelProbes are intentionally objective — a capable model passes them
// reliably, a broken/incapable one does not, and no human or model judges.
var modelProbes = []modelProbe{
	{
		name:   "instruction-following",
		prompt: "Reply with exactly the single word READY and nothing else.",
		pass: func(reply string) bool {
			r := strings.ToUpper(strings.Trim(strings.TrimSpace(reply), ".!\"'` \n"))
			return r == "READY"
		},
	},
	{
		name:   "output-contract",
		prompt: `Return only this JSON object and nothing else: {"ok": true, "n": 7}`,
		pass: func(reply string) bool {
			start := strings.Index(reply, "{")
			end := strings.LastIndex(reply, "}")
			if start < 0 || end <= start {
				return false
			}
			var v struct {
				OK bool `json:"ok"`
				N  int  `json:"n"`
			}
			if err := json.Unmarshal([]byte(reply[start:end+1]), &v); err != nil {
				return false
			}
			return v.OK && v.N == 7
		},
	},
	{
		name:   "code-structure",
		prompt: "Write a Go function named add that takes two ints and returns their sum. Output only the code.",
		pass: func(reply string) bool {
			r := strings.ToLower(reply)
			return strings.Contains(r, "func add") && strings.Contains(reply, "+")
		},
	},
	{
		name:   "constraint-count",
		prompt: "List exactly three fruits, one per line, with no numbering, bullets, or extra text.",
		pass: func(reply string) bool {
			lines := 0
			for _, ln := range strings.Split(strings.TrimSpace(reply), "\n") {
				if strings.TrimSpace(ln) != "" {
					lines++
				}
			}
			return lines == 3
		},
	},
}

// modelCheckVerifyThreshold is the pass fraction required to be VERIFIED.
const modelCheckVerifyThreshold = 0.75 // 3 of 4 probes

type modelProbeResult struct {
	name   string
	passed bool
	err    error
}

// runModelCheck probes the model behind a provider route and reports a verdict.
// Exit 0 = VERIFIED, 1 = UNVERIFIED or not runnable — so it doubles as a gate.
func runModelCheck(mode string, stdout, stderr io.Writer) int {
	mode = normalizeToolMode(strings.TrimSpace(mode))
	if mode == "" {
		fmt.Fprintln(stderr, "Usage: jini check model <route> (e.g. groq, cerebras, openai)")
		return 1
	}
	if cliHandoffMode(mode) {
		fmt.Fprintf(stderr, "Model check targets provider/model routes; %q runs your installed CLI, which manages its own model.\n", mode)
		return 1
	}
	fallback := enrichRouteDecisionForRequest(providerGenerationRequest{}, detectRouteForToolMode(mode, true))
	provider := providerForDecision(providerGenerationRequest{}, fallback)
	if provider.ID == "local-preview" || provider.Status != "ok" {
		fmt.Fprintf(stderr, "Route %q is not ready to probe. Run `jini doctor`.\n", mode)
		return 1
	}

	fmt.Fprintf(stdout, "Probing %s …\n", firstNonEmpty(fallback.ToolLabel, provider.Label, mode))
	ctx := context.Background()
	passed := 0
	for _, probe := range modelProbes {
		result := runOneModelProbe(ctx, provider, probe)
		mark := "FAIL"
		if result.passed {
			mark = "ok  "
			passed++
		}
		if result.err != nil {
			fmt.Fprintf(stdout, "%s %-22s (%v)\n", mark, probe.name, result.err)
		} else {
			fmt.Fprintf(stdout, "%s %-22s\n", mark, probe.name)
		}
	}

	total := len(modelProbes)
	score := float64(passed) / float64(total)
	verified := score >= modelCheckVerifyThreshold
	verdict := "UNVERIFIED"
	if verified {
		verdict = "VERIFIED"
	}
	fmt.Fprintf(stdout, "\n%s — %d/%d probes passed. %s\n", verdict, passed, total,
		verifiedFollowup(verified))
	if verified {
		return 0
	}
	return 1
}

func verifiedFollowup(verified bool) string {
	if verified {
		return "Meets the capability floor for a trusted route."
	}
	return "Below the capability floor — usable, but not verified for automatic routing."
}

func runOneModelProbe(ctx context.Context, provider providerConfig, probe modelProbe) modelProbeResult {
	request := providerGenerationRequest{Source: probe.prompt, Title: "model probe", Standalone: true}
	reply, err := generateProviderText(ctx, provider, request, modelProbeSystemPrompt, probe.prompt)
	if err != nil {
		return modelProbeResult{name: probe.name, err: err}
	}
	return modelProbeResult{name: probe.name, passed: probe.pass(reply)}
}

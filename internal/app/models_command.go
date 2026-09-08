package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// Model auto-discovery — `jini models [provider]`. Onboarding a new model must be
// zero-touch: Jini lists whatever a configured provider's /models endpoint
// returns, so a model published today is usable today (with the shape's model
// override) without a Jini release. This is the "never wait for a model" path.

// listProviderModels fetches the model ids a configured provider exposes, using
// the same models-endpoint request the validator builds. Returns sorted ids.
func listProviderModels(ctx context.Context, shape byoShape, client *http.Client) ([]string, error) {
	key := strings.TrimSpace(configValue(shape.KeyEnv))
	if key == "" {
		return nil, fmt.Errorf("%s not configured (%s)", shape.Label, shape.KeyEnv)
	}
	base := firstNonEmpty(configValue(shape.baseEnv), shape.defaultBase)
	req, err := shape.buildRequest(ctx, base, key)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach %s: %v", shape.Label, err)
	}
	defer resp.Body.Close()
	if err := classifyProviderHTTPError(shape.Label, shape.KeyEnv, resp.StatusCode); err != nil {
		return nil, err
	}
	var decoded struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("%s returned a model list Jini could not read", shape.Label)
	}
	ids := make([]string, 0, len(decoded.Data))
	for _, m := range decoded.Data {
		if id := strings.TrimSpace(m.ID); id != "" {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids, nil
}

// runModels lists the models available on configured routable providers. With no
// argument it covers every configured shape; with a name, just that one.
func runModels(args []string, stdout, stderr io.Writer) int {
	specific := len(args) > 0 && strings.TrimSpace(args[0]) != "" && !strings.HasPrefix(strings.TrimSpace(args[0]), "-")

	var shapes []byoShape
	if specific {
		shape, ok := resolveBYOShape(args[0])
		if !ok {
			fmt.Fprintf(stderr, "Unknown provider %q. Known: %s\n", args[0], strings.Join(sortedBYOShapeIDs(), ", "))
			return 1
		}
		shapes = []byoShape{shape}
	} else {
		for _, id := range sortedBYOShapeIDs() {
			shapes = append(shapes, byoShapes[id])
		}
	}

	client := byoValidationClient()
	ctx := context.Background()
	anyConfigured := false
	for _, shape := range shapes {
		if !shape.routable() {
			continue
		}
		if strings.TrimSpace(configValue(shape.KeyEnv)) == "" {
			if specific {
				fmt.Fprintf(stderr, "%s not configured. Set %s, then rerun.\n", shape.Label, shape.KeyEnv)
				return 1
			}
			continue
		}
		anyConfigured = true
		models, err := listProviderModels(ctx, shape, client)
		if err != nil {
			fmt.Fprintf(stdout, "%s: %v\n", shape.Label, err)
			continue
		}
		current := firstNonEmpty(configValue(shape.modelEnv), shape.defaultModel)
		fmt.Fprintf(stdout, "%s (%d models · %s selects a model):\n", shape.Label, len(models), shape.modelEnv)
		for _, id := range models {
			marker := "  "
			if id == current {
				marker = "→ "
			}
			fmt.Fprintf(stdout, "%s%s\n", marker, id)
		}
		fmt.Fprintln(stdout)
	}

	if !specific && !anyConfigured {
		fmt.Fprintln(stdout, "No BYO providers configured. Set a key (e.g. GROQ_API_KEY), then `jini models`.")
		fmt.Fprintln(stdout, "Any model a provider lists is usable immediately via its model env (e.g. GROQ_MODEL=<id>) — no Jini update needed.")
	}
	return 0
}

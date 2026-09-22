package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// jini mcp — Model Context Protocol client foundation. Closes the single
// biggest ecosystem-parity gap with Claude Code (specs/pre-viral-readiness.md
// gap #3: "MCP client — let Jini consume MCP servers (tools/resources) like
// Claude Code"). This first slice loads the Claude-Code-shaped mcp.json config
// and lists configured servers with a PATH-readiness check. The live MCP
// JSON-RPC handshake and tool invocation are a deliberate follow-up.

type mcpServerConfig struct {
	Command   string   `json:"command"`
	Args      []string `json:"args,omitempty"`
	Transport string   `json:"transport,omitempty"`
}

type mcpConfig struct {
	MCPServers map[string]mcpServerConfig `json:"mcpServers"`
}

// mcpConfigLoader is overridable in tests so the suite never reads a real
// on-disk config (which would make listing non-hermetic).
var mcpConfigLoader = defaultMCPConfigLoader

// mcpLookPath resolves a command on PATH; overridable so tests do not depend on
// the host's PATH.
var mcpLookPath = exec.LookPath

// mcpConfigPath returns <JINI_HOME>/.jini/mcp.json — the same home that holds
// providers.json / mode.json, so JINI_HOME sandboxes it for tests and dogfood.
func mcpConfigPath() (string, error) {
	home, err := executionModeHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".jini", "mcp.json"), nil
}

// defaultMCPConfigLoader reads and parses mcp.json. A missing file is not an
// error — it is the normal "no servers configured yet" state.
func defaultMCPConfigLoader() (mcpConfig, error) {
	var cfg mcpConfig
	path, err := mcpConfigPath()
	if err != nil {
		return cfg, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("mcp.json is not valid JSON: %w", err)
	}
	return cfg, nil
}

type mcpServerStatus struct {
	Name      string `json:"name"`
	Command   string `json:"command"`
	Transport string `json:"transport"`
	Ready     bool   `json:"ready"`
}

// mcpServerStatuses resolves each configured server into a listing entry,
// marking whether its command exists on PATH. Sorted by name for deterministic
// output (map iteration order is otherwise random).
func mcpServerStatuses(cfg mcpConfig) []mcpServerStatus {
	statuses := make([]mcpServerStatus, 0, len(cfg.MCPServers))
	for name, srv := range cfg.MCPServers {
		transport := strings.TrimSpace(srv.Transport)
		if transport == "" {
			transport = "stdio" // Claude Code's default transport
		}
		ready := false
		if cmd := strings.TrimSpace(srv.Command); cmd != "" {
			if _, err := mcpLookPath(cmd); err == nil {
				ready = true
			}
		}
		statuses = append(statuses, mcpServerStatus{
			Name:      name,
			Command:   srv.Command,
			Transport: transport,
			Ready:     ready,
		})
	}
	sort.Slice(statuses, func(i, j int) bool { return statuses[i].Name < statuses[j].Name })
	return statuses
}

// runMCP implements `jini mcp` and `jini mcp list`: list each configured MCP
// server (name, command, transport) and mark PATH readiness. Optional
// `--format json`. Exit 0 on success, 1 on a bad subcommand or unreadable config.
func runMCP(args []string, stdout, stderr io.Writer) int {
	asJSON := false
	sub := ""
	formatPending := false // saw "--format"; the next bare token is its value
	for _, a := range args {
		switch {
		case a == "--json" || a == "--format=json":
			asJSON = true
		case a == "--format":
			formatPending = true
		case formatPending:
			formatPending = false
			if a == "json" {
				asJSON = true
			}
		case strings.HasPrefix(a, "--"):
			// ignore other flags
		default:
			sub = a
		}
	}
	if sub != "" && sub != "list" {
		fmt.Fprintf(stderr, "Unknown mcp subcommand %q. Usage: jini mcp [list] [--format json]\n", sub)
		return 1
	}

	cfg, err := mcpConfigLoader()
	if err != nil {
		fmt.Fprintf(stderr, "Could not load MCP config: %v\n", err)
		return 1
	}

	statuses := mcpServerStatuses(cfg)

	if asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(struct {
			Servers []mcpServerStatus `json:"servers"`
		}{Servers: statuses})
		return 0
	}

	renderMCPList(stdout, statuses)
	return 0
}

func renderMCPList(w io.Writer, statuses []mcpServerStatus) {
	if len(statuses) == 0 {
		fmt.Fprintln(w, "No MCP servers configured.")
		fmt.Fprintln(w, "Add one to <JINI_HOME>/.jini/mcp.json, e.g.:")
		fmt.Fprintln(w, `  {"mcpServers": {"files": {"command": "mcp-server-filesystem", "args": ["."], "transport": "stdio"}}}`)
		return
	}
	for _, s := range statuses {
		mark := "not found"
		if s.Ready {
			mark = "ready"
		}
		fmt.Fprintf(w, "  %s — %s (%s) [%s]\n", s.Name, s.Command, s.Transport, mark)
	}
}

package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stubMCPConfig injects a config loader so tests never read a real on-disk
// mcp.json, and a PATH resolver so readiness is deterministic (commands in
// `ready` resolve, everything else does not).
func stubMCPConfig(t *testing.T, cfg mcpConfig, err error, ready ...string) {
	t.Helper()
	oldLoader := mcpConfigLoader
	oldLook := mcpLookPath
	t.Cleanup(func() {
		mcpConfigLoader = oldLoader
		mcpLookPath = oldLook
	})
	mcpConfigLoader = func() (mcpConfig, error) { return cfg, err }
	readySet := map[string]bool{}
	for _, r := range ready {
		readySet[r] = true
	}
	mcpLookPath = func(file string) (string, error) {
		if readySet[file] {
			return "/usr/bin/" + file, nil
		}
		return "", errors.New("not found")
	}
}

func TestMCPConfigParsingViaLoader(t *testing.T) {
	raw := `{"mcpServers": {"files": {"command": "mcp-server-filesystem", "args": ["."], "transport": "stdio"}}}`
	var cfg mcpConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		t.Fatalf("unmarshal Claude-Code-shaped config: %v", err)
	}
	srv, ok := cfg.MCPServers["files"]
	if !ok {
		t.Fatalf("expected server %q to parse", "files")
	}
	if srv.Command != "mcp-server-filesystem" {
		t.Fatalf("command = %q, want mcp-server-filesystem", srv.Command)
	}
	if len(srv.Args) != 1 || srv.Args[0] != "." {
		t.Fatalf("args = %v, want [.]", srv.Args)
	}
	if srv.Transport != "stdio" {
		t.Fatalf("transport = %q, want stdio", srv.Transport)
	}
}

// stubMCPHome points the global-home resolver at a sandbox dir for the default
// loader tests (TestMain pins executionModeHomeDir, so JINI_HOME alone is not
// enough — override the var per-test, mirroring savings_ledger_test.go).
func stubMCPHome(t *testing.T, home string) {
	t.Helper()
	old := executionModeHomeDir
	t.Cleanup(func() { executionModeHomeDir = old })
	executionModeHomeDir = func() (string, error) { return home, nil }
}

func TestMCPDefaultLoaderMissingFileIsNotError(t *testing.T) {
	stubMCPHome(t, t.TempDir()) // fresh dir, no .jini/mcp.json
	cfg, err := defaultMCPConfigLoader()
	if err != nil {
		t.Fatalf("missing mcp.json must not be an error, got %v", err)
	}
	if len(cfg.MCPServers) != 0 {
		t.Fatalf("missing config must yield no servers, got %d", len(cfg.MCPServers))
	}
}

func TestMCPDefaultLoaderReadsSandboxedConfig(t *testing.T) {
	home := t.TempDir()
	stubMCPHome(t, home)
	dir := filepath.Join(home, ".jini")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	raw := `{"mcpServers": {"db": {"command": "mcp-db", "transport": "stdio"}}}`
	if err := os.WriteFile(filepath.Join(dir, "mcp.json"), []byte(raw), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}
	cfg, err := defaultMCPConfigLoader()
	if err != nil {
		t.Fatalf("load sandboxed config: %v", err)
	}
	if _, ok := cfg.MCPServers["db"]; !ok {
		t.Fatalf("expected server %q from sandboxed mcp.json", "db")
	}
}

func TestMCPListRendersReadinessAndSortsByName(t *testing.T) {
	cfg := mcpConfig{MCPServers: map[string]mcpServerConfig{
		"zeta":  {Command: "mcp-zeta", Transport: "stdio"},
		"alpha": {Command: "mcp-alpha"}, // transport defaults to stdio
	}}
	stubMCPConfig(t, cfg, nil, "mcp-alpha") // only alpha is on PATH

	var out, errBuf bytes.Buffer
	if code := runMCP([]string{"list"}, &out, &errBuf); code != 0 {
		t.Fatalf("exit = %d, want 0; stderr=%q", code, errBuf.String())
	}
	got := out.String()

	if ai, zi := strings.Index(got, "alpha"), strings.Index(got, "zeta"); ai < 0 || zi < 0 || ai > zi {
		t.Fatalf("servers must be listed sorted by name; got:\n%s", got)
	}
	if !strings.Contains(got, "alpha — mcp-alpha (stdio) [ready]") {
		t.Fatalf("alpha (on PATH) must be marked ready; got:\n%s", got)
	}
	if !strings.Contains(got, "zeta — mcp-zeta (stdio) [not found]") {
		t.Fatalf("zeta (not on PATH) must be marked not found; got:\n%s", got)
	}
}

func TestMCPEmptyConfigMessage(t *testing.T) {
	stubMCPConfig(t, mcpConfig{}, nil)

	var out, errBuf bytes.Buffer
	if code := runMCP(nil, &out, &errBuf); code != 0 {
		t.Fatalf("exit = %d, want 0; stderr=%q", code, errBuf.String())
	}
	if !strings.Contains(out.String(), "No MCP servers configured.") {
		t.Fatalf("empty config must show a friendly message; got:\n%s", out.String())
	}
}

func TestMCPListJSONFormat(t *testing.T) {
	cfg := mcpConfig{MCPServers: map[string]mcpServerConfig{
		"files": {Command: "mcp-server-filesystem", Args: []string{"."}, Transport: "stdio"},
	}}
	stubMCPConfig(t, cfg, nil, "mcp-server-filesystem")

	var out, errBuf bytes.Buffer
	if code := runMCP([]string{"list", "--format", "json"}, &out, &errBuf); code != 0 {
		t.Fatalf("exit = %d, want 0; stderr=%q", code, errBuf.String())
	}
	var decoded struct {
		Servers []mcpServerStatus `json:"servers"`
	}
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("json output must decode: %v; raw=%q", err, out.String())
	}
	if len(decoded.Servers) != 1 || decoded.Servers[0].Name != "files" || !decoded.Servers[0].Ready {
		t.Fatalf("unexpected json servers: %+v", decoded.Servers)
	}
}

func TestMCPUnknownSubcommand(t *testing.T) {
	stubMCPConfig(t, mcpConfig{}, nil)

	var out, errBuf bytes.Buffer
	if code := runMCP([]string{"bogus"}, &out, &errBuf); code != 1 {
		t.Fatalf("unknown subcommand must exit 1, got %d", code)
	}
	if !strings.Contains(errBuf.String(), "Unknown mcp subcommand") {
		t.Fatalf("expected friendly unknown-subcommand error; got:\n%s", errBuf.String())
	}
}

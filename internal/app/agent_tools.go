package app

// Native agentic loop tool layer — specs/native-agentic-loop-design.md. Each
// tool declares a MINIMUM posture; the loop only offers tools the resolved
// hand-off posture permits, so the same `jini trust` consent governs what the
// loop can do. All paths are confined to the working directory.

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var agentCommandTimeout = 120 * time.Second

type agentTool struct {
	Name        string
	MinPosture  handoffPosture
	Description string
	// Run executes the tool and returns an observation string. Errors are
	// returned to the loop, which feeds them back as observations so the model
	// can recover.
	Run func(workDir string, args map[string]string) (string, error)
}

const (
	agentMaxReadBytes   = 64 * 1024
	agentMaxSearchHits  = 40
	agentMaxListEntries = 200
)

// confinePath resolves p (relative to workDir) and rejects anything that
// escapes workDir, including via symlinks for paths that exist.
func confinePath(workDir, p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("empty path")
	}
	abs := p
	if !filepath.IsAbs(p) {
		abs = filepath.Join(workDir, p)
	}
	abs = filepath.Clean(abs)
	root := filepath.Clean(workDir)
	if r, err := filepath.EvalSymlinks(root); err == nil {
		root = r
	}
	// Resolve symlinks on the nearest EXISTING ancestor (the target file may
	// not exist yet), then compare — so a valid in-dir path whose parent is a
	// symlink (e.g. macOS /var → /private/var) is not falsely rejected.
	resolved := resolveExistingAncestor(abs)
	if resolved != root && !strings.HasPrefix(resolved, root+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes working directory: %s", p)
	}
	return abs, nil
}

func resolveExistingAncestor(abs string) string {
	cur := abs
	var tail []string
	for {
		if r, err := filepath.EvalSymlinks(cur); err == nil {
			full := r
			for i := len(tail) - 1; i >= 0; i-- {
				full = filepath.Join(full, tail[i])
			}
			return filepath.Clean(full)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return abs
		}
		tail = append(tail, filepath.Base(cur))
		cur = parent
	}
}

func toolReadFile() agentTool {
	return agentTool{
		Name:        "read_file",
		MinPosture:  posturePlan,
		Description: "read_file — read a file. arg: path",
		Run: func(workDir string, args map[string]string) (string, error) {
			path, err := confinePath(workDir, args["path"])
			if err != nil {
				return "", err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			if len(data) > agentMaxReadBytes {
				return string(data[:agentMaxReadBytes]) + "\n... (truncated)", nil
			}
			return string(data), nil
		},
	}
}

func toolListDir() agentTool {
	return agentTool{
		Name:        "list_dir",
		MinPosture:  posturePlan,
		Description: "list_dir — list a directory. arg: path (default .)",
		Run: func(workDir string, args map[string]string) (string, error) {
			rel := args["path"]
			if strings.TrimSpace(rel) == "" {
				rel = "."
			}
			path, err := confinePath(workDir, rel)
			if err != nil {
				return "", err
			}
			entries, err := os.ReadDir(path)
			if err != nil {
				return "", err
			}
			names := make([]string, 0, len(entries))
			for _, e := range entries {
				name := e.Name()
				if e.IsDir() {
					name += "/"
				}
				names = append(names, name)
			}
			sort.Strings(names)
			if len(names) > agentMaxListEntries {
				names = append(names[:agentMaxListEntries], "... (truncated)")
			}
			return strings.Join(names, "\n"), nil
		},
	}
}

func toolSearch() agentTool {
	return agentTool{
		Name:        "search",
		MinPosture:  posturePlan,
		Description: "search — find a substring across files. arg: query",
		Run: func(workDir string, args map[string]string) (string, error) {
			query := strings.TrimSpace(args["query"])
			if query == "" {
				return "", fmt.Errorf("empty query")
			}
			var hits []string
			root := filepath.Clean(workDir)
			_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
				if err != nil || d.IsDir() || len(hits) >= agentMaxSearchHits {
					if d != nil && d.IsDir() && strings.HasPrefix(d.Name(), ".") && path != root {
						return filepath.SkipDir
					}
					return nil
				}
				data, readErr := os.ReadFile(path)
				if readErr != nil || len(data) > agentMaxReadBytes {
					return nil
				}
				for i, line := range strings.Split(string(data), "\n") {
					if strings.Contains(line, query) {
						rel, _ := filepath.Rel(root, path)
						hits = append(hits, fmt.Sprintf("%s:%d: %s", rel, i+1, strings.TrimSpace(line)))
						if len(hits) >= agentMaxSearchHits {
							break
						}
					}
				}
				return nil
			})
			if len(hits) == 0 {
				return "no matches", nil
			}
			return strings.Join(hits, "\n"), nil
		},
	}
}

// toolEditFile writes a file's full new contents (body in a fenced block).
// Whole-file write is the most reliable edit primitive for weak local models;
// a surgical search/replace variant can come later.
func toolEditFile() agentTool {
	return agentTool{
		Name:        "edit_file",
		MinPosture:  postureSemi,
		Description: "edit_file — write a file's full new contents. arg: path; put the new contents in a ``` fenced block",
		Run: func(workDir string, args map[string]string) (string, error) {
			path, err := confinePath(workDir, args["path"])
			if err != nil {
				return "", err
			}
			body, ok := args["body"]
			if !ok {
				return "", fmt.Errorf("edit_file needs the new contents in a ``` fenced block")
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return "", err
			}
			if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
				return "", err
			}
			return fmt.Sprintf("wrote %d bytes to %s", len(body), strings.TrimSpace(args["path"])), nil
		},
	}
}

// readOnlyTools is the plan-posture toolset (available in any posture).
func readOnlyTools() []agentTool {
	return []agentTool{toolReadFile(), toolListDir(), toolSearch()}
}

// toolRunCommand runs a shell command in the working directory. Min posture is
// autonomous — the same level the user consented to with `jini trust
// --autonomous` ("apply edits and run commands"). The command can do anything
// the user's shell can; that IS the consented capability. Output is fed back
// so the loop can self-verify (edit → run tests → fix).
func toolRunCommand() agentTool {
	return agentTool{
		Name:        "run_command",
		MinPosture:  postureAutonomous,
		Description: "run_command — run a shell command in the working directory (e.g. run tests). arg: command",
		Run: func(workDir string, args map[string]string) (string, error) {
			cmdStr := strings.TrimSpace(args["command"])
			if cmdStr == "" {
				return "", fmt.Errorf("empty command")
			}
			ctx, cancel := context.WithTimeout(context.Background(), agentCommandTimeout)
			defer cancel()
			cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
			cmd.Dir = workDir
			out, err := cmd.CombinedOutput()
			result := string(out)
			if len(result) > agentMaxReadBytes {
				result = result[:agentMaxReadBytes] + "\n... (truncated)"
			}
			if err != nil {
				// Feed failures back as observations (not hard errors) so the
				// model can react and fix.
				return fmt.Sprintf("command exited with error: %v\n%s", err, result), nil
			}
			if strings.TrimSpace(result) == "" {
				return "(command produced no output)", nil
			}
			return result, nil
		},
	}
}

// allAgentTools is the full toolset; the loop offers only those the posture
// permits (read=plan, edit=semi, run=autonomous).
func allAgentTools() []agentTool {
	return append(readOnlyTools(), toolEditFile(), toolRunCommand())
}

// toolsForPosture returns the tools available at a posture (min posture <= p).
func toolsForPosture(all []agentTool, posture handoffPosture) []agentTool {
	out := make([]agentTool, 0, len(all))
	for _, t := range all {
		if t.MinPosture <= posture {
			out = append(out, t)
		}
	}
	return out
}

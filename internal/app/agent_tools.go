package app

// Native agentic loop tool layer — specs/native-agentic-loop-design.md. Each
// tool declares a MINIMUM posture; the loop only offers tools the resolved
// hand-off posture permits, so the same `jini trust` consent governs what the
// loop can do. All paths are confined to the working directory.

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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

// readOnlyTools is the plan-posture toolset (available in any posture).
func readOnlyTools() []agentTool {
	return []agentTool{toolReadFile(), toolListDir(), toolSearch()}
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

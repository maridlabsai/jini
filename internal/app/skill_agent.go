package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// On-the-fly skills and agents (PRD §Skills And Agents, free tier): create
// plain, portable markdown files with frontmatter — reviewable and diffable.
// This is the creation surface (`jini skill new` / `jini agent new`); content is
// secret-scrubbed before writing. Files live under .jini/skills and .jini/agents
// so they sit beside the repo and diff cleanly.

// isFreeSkillOrAgentCreation reports whether args are the free-tier creation
// surface (`jini skill new …` / `jini agent new …`). The PRD puts creating
// coding-focused skills and agents as plain reviewable files in the FREE tier;
// only the managed Skills OS and developer/tester agent fleets are commercial.
func isFreeSkillOrAgentCreation(args []string) bool {
	if len(args) < 2 {
		return false
	}
	switch exactCommandToken(args[0]) {
	case "skill", "agent":
		return exactCommandToken(args[1]) == "new"
	}
	return false
}

func runSkill(args []string, stdout, stderr io.Writer) int {
	return runSkillOrAgentNew("skill", "skills", args, stdout, stderr)
}

func runAgent(args []string, stdout, stderr io.Writer) int {
	return runSkillOrAgentNew("agent", "agents", args, stdout, stderr)
}

func runSkillOrAgentNew(kind, dirName string, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || exactCommandToken(args[0]) != "new" {
		fmt.Fprintf(stderr, "Usage: jini %s new <name> [description]\n", kind)
		return 1
	}
	rest := args[1:]
	if len(rest) == 0 || strings.TrimSpace(rest[0]) == "" {
		fmt.Fprintf(stderr, "Usage: jini %s new <name> [description]\n", kind)
		return 1
	}
	name := slugify(rest[0])
	if name == "" {
		fmt.Fprintf(stderr, "A %s name needs letters or digits: %q\n", kind, rest[0])
		return 1
	}
	description := strings.TrimSpace(strings.Join(rest[1:], " "))

	content := scaffoldSkillOrAgent(kind, name, description)
	// Secret scrub before writing — never persist a pasted credential to a file
	// that will be committed and shared.
	if leaked := firstLeakedSecret(content); leaked != "" {
		fmt.Fprintf(stderr, "Refusing to write: the %s text looks like it contains a secret (%s). Remove it and retry.\n", kind, leaked)
		return 1
	}

	base := filepath.Join(".jini", dirName)
	path := filepath.Join(base, name+".md")
	if _, err := os.Stat(path); err == nil {
		fmt.Fprintf(stderr, "%s already exists: %s — edit it directly.\n", titleCase(kind), path)
		return 1
	}
	if err := os.MkdirAll(base, 0o755); err != nil {
		fmt.Fprintf(stderr, "Could not create %s directory: %v\n", kind, err)
		return 1
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		fmt.Fprintf(stderr, "Could not write %s: %v\n", kind, err)
		return 1
	}

	fmt.Fprintf(stdout, "Created %s: %s\n", kind, path)
	fmt.Fprintln(stdout, "It's a plain, reviewable markdown file — edit the steps, commit it, and it travels with the repo.")
	return 0
}

func scaffoldSkillOrAgent(kind, name, description string) string {
	if description == "" {
		description = "Describe when to use this " + kind + " and what it does."
	}
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: " + name + "\n")
	b.WriteString("description: " + description + "\n")
	b.WriteString("kind: " + kind + "\n")
	b.WriteString("created: " + time.Now().UTC().Format("2006-01-02") + "\n")
	b.WriteString("---\n\n")
	b.WriteString("# " + name + "\n\n")
	if kind == "agent" {
		b.WriteString("## Role\n\n")
		b.WriteString("What this agent is responsible for. It inherits the approval matrix and\n")
		b.WriteString("cannot self-approve side effects.\n\n")
		b.WriteString("## Steps\n\n")
		b.WriteString("1. First step.\n2. Next step.\n")
		return b.String()
	}
	b.WriteString("## When to use\n\n")
	b.WriteString("The recurring situation this skill handles.\n\n")
	b.WriteString("## Steps\n\n")
	b.WriteString("1. First step.\n2. Next step.\n")
	return b.String()
}

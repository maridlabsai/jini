package app

// Native loop text protocol — specs/native-agentic-loop-design.md. The model
// drives tools by emitting a forgiving line-based ACTION block; this works on
// any model, including weak local ones. A native function-calling driver can
// replace this later for capable models.

import (
	"fmt"
	"strings"
)

type agentAction struct {
	Tool string
	Args map[string]string
}

// agentSystemPrompt describes the tools and the required response format.
func agentSystemPrompt(task string, tools []agentTool) string {
	var b strings.Builder
	b.WriteString("You are Jini's coding agent working in the current directory. ")
	b.WriteString("Work in small steps. Each reply MUST be exactly one action in this format:\n\n")
	b.WriteString("ACTION: <tool>\n<arg>: <value>\n\n")
	b.WriteString("Available tools:\n")
	for _, t := range tools {
		b.WriteString("- " + t.Description + "\n")
	}
	b.WriteString("- finish — end the task. arg: summary\n\n")
	b.WriteString("Emit ONE action per reply and nothing else. After each action you will receive an OBSERVATION. ")
	b.WriteString("When the task is done, use finish.\n\nTASK: ")
	b.WriteString(task)
	return b.String()
}

// parseAction extracts the last ACTION block from a model reply. ok=false when
// no well-formed action is present (the loop then repairs).
func parseAction(reply string) (agentAction, bool) {
	lines := strings.Split(reply, "\n")
	// Find the last "ACTION:" line so trailing chatter before it is ignored.
	actionIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "ACTION:") {
			actionIdx = i
		}
	}
	if actionIdx < 0 {
		return agentAction{}, false
	}
	tool := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[actionIdx]), "ACTION:"))
	if tool == "" {
		return agentAction{}, false
	}
	args := map[string]string{}
	for _, line := range lines[actionIdx+1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		args[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return agentAction{Tool: tool, Args: args}, true
}

// agentRepairPrompt is the re-prompt when a reply has no valid action.
func agentRepairPrompt(tools []agentTool) string {
	names := make([]string, 0, len(tools)+1)
	for _, t := range tools {
		names = append(names, t.Name)
	}
	names = append(names, "finish")
	return fmt.Sprintf("Your last reply had no valid action. Reply with exactly:\nACTION: <tool>\n<arg>: <value>\nAvailable tools: %s", strings.Join(names, ", "))
}

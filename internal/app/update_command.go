package app

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// `jini update [--channel <stable|beta|nightly>]` — self-update on your channel.
// It re-runs the public install.sh for the resolved channel (which verifies the
// checksum + macOS signature before replacing the binary), the standard self-
// update pattern (deno/rustup/deno). Channel resolution: an explicit --channel
// wins; else the channel this binary was built on; else the install receipt;
// else stable. Fully automatable — no prompts.

const defaultUpdateInstallURL = "https://raw.githubusercontent.com/maridlabsai/jini/main/install.sh"

func updateInstallURL() string {
	return firstNonEmpty(strings.TrimSpace(os.Getenv("JINI_UPDATE_INSTALL_URL")), defaultUpdateInstallURL)
}

// resolveUpdateChannel picks the channel to update on.
func resolveUpdateChannel(explicit string) string {
	if c := normalizeUpdateChannel(explicit); c != "" {
		return c
	}
	if c := jiniChannel(); c != "" && c != "source" {
		return c
	}
	if c := normalizeUpdateChannel(installReceiptChannel()); c != "" {
		return c
	}
	return "stable"
}

func normalizeUpdateChannel(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "stable", "beta", "nightly":
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

// installReceiptChannel reads channel= from the install receipt next to the
// running binary, if present.
func installReceiptChannel() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	// The receipt sits in the install dir alongside the resolved binary.
	data, err := os.ReadFile(strings.TrimSuffix(exe, "/jini") + "/install-receipt.txt")
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "channel="); ok {
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// updateInstallCommand builds the shell command that performs the update.
func updateInstallCommand(channel string) string {
	return fmt.Sprintf("curl -fsSL %s | sh -s -- --channel %s --force", updateInstallURL(), channel)
}

func runUpdate(args []string, stdout, stderr io.Writer) int {
	explicit := ""
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--channel" && i+1 < len(args):
			explicit = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--channel="):
			explicit = strings.TrimPrefix(args[i], "--channel=")
		default:
			fmt.Fprintf(stderr, "Unknown argument %q. Use `jini update [--channel <stable|beta|nightly>]`.\n", args[i])
			return 1
		}
	}
	if explicit != "" && normalizeUpdateChannel(explicit) == "" {
		fmt.Fprintf(stderr, "Unknown channel %q. Use stable, beta, or nightly.\n", explicit)
		return 1
	}

	channel := resolveUpdateChannel(explicit)
	command := updateInstallCommand(channel)
	fmt.Fprintf(stdout, "Current: jini %s (%s)\n", jiniVersion(), jiniChannel())
	fmt.Fprintf(stdout, "Updating on the %s channel …\n", channel)

	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(stderr, "Update failed: %v\n", err)
		fmt.Fprintf(stderr, "Run it manually: %s\n", command)
		return 1
	}
	fmt.Fprintln(stdout, "Updated. Run `jini version` to confirm.")
	return 0
}

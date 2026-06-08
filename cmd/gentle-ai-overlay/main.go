package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gentle-ai-custom/internal/overlay"
)

func main() {
	exitCode := run(os.Args[1:])
	os.Exit(exitCode)
}

func run(args []string) int {
	repoRoot, rest, err := parseGlobalArgs(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	if repoRoot == "" {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			fmt.Fprintf(os.Stderr, "ERROR: cannot determine working directory: %v\n", cwdErr)
			return 1
		}
		repoRoot = cwd
	}
	repoRoot, err = filepath.Abs(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot resolve repo root %q: %v\n", repoRoot, err)
		return 1
	}

	if len(rest) == 0 {
		printMainUsage(os.Stderr)
		return 1
	}

	switch rest[0] {
	case "apply-custom":
		return overlay.RunApplyCustom(repoRoot, rest[1:])
	case "apply-policy":
		return overlay.RunApplyPolicy(repoRoot, rest[1:])
	case "audit-upstream":
		return overlay.RunAuditUpstream(repoRoot, rest[1:])
	case "sync-upstream-assets":
		return overlay.RunSyncUpstreamAssets(repoRoot, rest[1:])
	case "-h", "--help", "help":
		printMainUsage(os.Stdout)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "ERROR: unknown subcommand %q\n", rest[0])
		printMainUsage(os.Stderr)
		return 1
	}
}

func parseGlobalArgs(args []string) (string, []string, error) {
	var repoRoot string
	idx := 0
	for idx < len(args) {
		switch args[idx] {
		case "--repo-root":
			if idx+1 >= len(args) {
				return "", nil, fmt.Errorf("--repo-root requires a value")
			}
			repoRoot = args[idx+1]
			idx += 2
		case "-h", "--help", "help":
			return repoRoot, []string{"help"}, nil
		default:
			if len(args[idx]) > 0 && args[idx][0] == '-' {
				return "", nil, fmt.Errorf("unknown global flag %q", args[idx])
			}
			return repoRoot, args[idx:], nil
		}
	}
	return repoRoot, nil, nil
}

func printMainUsage(out *os.File) {
	fmt.Fprintln(out, "Usage: gentle-ai-overlay [--repo-root <path>] <subcommand> [args]")
	fmt.Fprintln(out, "Subcommands:")
	fmt.Fprintln(out, "  apply-custom [--verbose] all | opencode")
	fmt.Fprintln(out, "  apply-policy [--verbose]")
	fmt.Fprintln(out, "  audit-upstream")
	fmt.Fprintln(out, "  sync-upstream-assets [--verbose]")
}

package overlay

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var gentleAIVersionPattern = regexp.MustCompile(`(?i)\bv?\d+(?:\.\d+)+(?:[-+][0-9A-Za-z.-]+)?\b`)

var errVersionHasMetadata = errors.New("version contains prerelease or build metadata")

func runVersionPreflight(repoRoot string, policy Policy, stdin *os.File, stderr io.Writer) error {
	statePath := filepath.Join(repoRoot, policy.Maintenance.StateFile)
	var state UpstreamState
	if err := readJSONFile(statePath, &state); err != nil {
		return fmt.Errorf("cannot read upstream state at %s: %w", statePath, err)
	}
	auditedVersion := normalizeVersionLabel(state.LastMaintainedVersion)
	if auditedVersion == "" {
		return fmt.Errorf("upstream state at %s is missing last_maintained_version", statePath)
	}

	installedVersion, detectErr := detectGentleAIVersion()
	interactive := isInteractiveInput(stdin)
	if detectErr != nil {
		return handleVersionUnknown(stdin, stderr, interactive, detectErr)
	}

	comparison, err := compareVersionStrings(installedVersion, auditedVersion)
	if err != nil {
		if errors.Is(err, errVersionHasMetadata) {
			return handleVersionUncertain(stdin, stderr, interactive, installedVersion, err)
		}
		return fmt.Errorf("cannot compare installed gentle-ai version %q against audited version %q from %s: %w", installedVersion, auditedVersion, statePath, err)
	}
	if comparison == 0 {
		return nil
	}

	if comparison < 0 {
		return handleVersionMismatch(stdin, stderr, interactive, installedVersion, auditedVersion, "older")
	}
	return handleVersionMismatch(stdin, stderr, interactive, installedVersion, auditedVersion, "newer")
}

func detectGentleAIVersion() (string, error) {
	cmd := exec.Command("gentle-ai", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("could not run `gentle-ai --version`: %w", err)
	}
	installedVersion, parseErr := extractGentleAIVersion(string(output))
	if parseErr == nil {
		return installedVersion, nil
	}
	return "", fmt.Errorf("version output %q did not contain a recognizable gentle-ai version", strings.TrimSpace(string(output)))
}

func extractGentleAIVersion(output string) (string, error) {
	match := gentleAIVersionPattern.FindString(output)
	if match == "" {
		return "", fmt.Errorf("no recognizable version token found")
	}
	return normalizeVersionLabel(match), nil
}

func normalizeVersionLabel(raw string) string {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "v")
	trimmed = strings.TrimPrefix(trimmed, "V")
	return strings.TrimSpace(trimmed)
}

func compareVersionStrings(installed, audited string) (int, error) {
	installedParts, installedHasMetadata, err := parseVersionParts(installed)
	if err != nil {
		return 0, err
	}
	if installedHasMetadata {
		return 0, fmt.Errorf("%w: installed version %q", errVersionHasMetadata, installed)
	}
	auditedParts, auditedHasMetadata, err := parseVersionParts(audited)
	if err != nil {
		return 0, err
	}
	if auditedHasMetadata {
		return 0, fmt.Errorf("%w: audited version %q", errVersionHasMetadata, audited)
	}
	maxParts := len(installedParts)
	if len(auditedParts) > maxParts {
		maxParts = len(auditedParts)
	}
	for idx := 0; idx < maxParts; idx++ {
		installedPart := 0
		if idx < len(installedParts) {
			installedPart = installedParts[idx]
		}
		auditedPart := 0
		if idx < len(auditedParts) {
			auditedPart = auditedParts[idx]
		}
		switch {
		case installedPart < auditedPart:
			return -1, nil
		case installedPart > auditedPart:
			return 1, nil
		}
	}
	return 0, nil
}

func parseVersionParts(version string) ([]int, bool, error) {
	normalized := normalizeVersionLabel(version)
	if normalized == "" {
		return nil, false, fmt.Errorf("version is empty")
	}
	core := normalized
	hasMetadata := false
	if idx := strings.IndexAny(normalized, "-+"); idx >= 0 {
		core = strings.TrimSpace(normalized[:idx])
		hasMetadata = true
	}
	if core == "" {
		return nil, false, fmt.Errorf("version is empty")
	}
	segments := strings.Split(core, ".")
	parts := make([]int, 0, len(segments))
	for _, segment := range segments {
		if segment == "" {
			return nil, false, fmt.Errorf("version %q contains an empty segment", version)
		}
		part, err := strconv.Atoi(segment)
		if err != nil {
			return nil, false, fmt.Errorf("version %q contains a non-numeric segment %q", version, segment)
		}
		parts = append(parts, part)
	}
	return parts, hasMetadata, nil
}

func isInteractiveInput(stdin *os.File) bool {
	if stdin == nil {
		return false
	}
	info, err := stdin.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func handleVersionUnknown(stdin *os.File, stderr io.Writer, interactive bool, cause error) error {
	message := fmt.Sprintf("WARNING: unable to detect the installed gentle-ai version: %v", cause)
	return warnAndMaybeContinue(stdin, stderr, interactive, message, "Continue anyway? [y/N] ", "installed gentle-ai version could not be detected in non-interactive mode")
}

func handleVersionUncertain(stdin *os.File, stderr io.Writer, interactive bool, installedVersion string, cause error) error {
	message := fmt.Sprintf("WARNING: installed gentle-ai version %s could not be compared safely against the audited baseline: %v", installedVersion, cause)
	message += " Prerelease or build-metadata versions should be treated as uncertain until the overlay is audited against them."
	return warnAndMaybeContinue(stdin, stderr, interactive, message, "Continue anyway? [y/N] ", fmt.Sprintf("installed gentle-ai version %s could not be compared safely in non-interactive mode", installedVersion))
}

func handleVersionMismatch(stdin *os.File, stderr io.Writer, interactive bool, installedVersion, auditedVersion, relation string) error {
	message := fmt.Sprintf("WARNING: installed gentle-ai version %s is %s than audited version %s.", installedVersion, relation, auditedVersion)
	if relation == "older" {
		message += " Upgrade gentle-ai before applying this overlay if you want to stay close to the audited baseline."
	} else {
		message += " This overlay has not been audited against that version yet."
	}
	message += " This is an advisory guardrail, not a compatibility guarantee."
	return warnAndMaybeContinue(stdin, stderr, interactive, message, "Continue anyway? [y/N] ", fmt.Sprintf("installed gentle-ai version %s is %s than audited version %s in non-interactive mode", installedVersion, relation, auditedVersion))
}

func warnAndMaybeContinue(stdin *os.File, stderr io.Writer, interactive bool, warning, prompt, nonInteractiveError string) error {
	if stderr == nil {
		stderr = os.Stderr
	}
	fmt.Fprintln(stderr, warning)
	if !interactive {
		return fmt.Errorf("%s", nonInteractiveError)
	}
	fmt.Fprint(stderr, prompt)
	if stdin == nil {
		return fmt.Errorf("cannot prompt for confirmation without an interactive input stream")
	}
	reader := bufio.NewReader(stdin)
	answer, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("could not read confirmation: %w", err)
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	if answer == "y" || answer == "yes" {
		return nil
	}
	return fmt.Errorf("version preflight aborted by user")
}

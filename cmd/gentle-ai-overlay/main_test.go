package main

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunSubcommandHelpDocumentsReleaseSelector(t *testing.T) {
	for _, command := range []string{"audit-upstream", "sync-upstream-assets"} {
		t.Run(command, func(t *testing.T) {
			output := captureStdout(t, func() {
				if exit := run([]string{command, "--help"}); exit != 0 {
					t.Fatalf("run help exit = %d, want 0", exit)
				}
			})
			if !strings.Contains(output, "--release v<V>") {
				t.Fatalf("help output %q does not document the exact release selector", output)
			}
		})
	}
}

func TestParseGlobalArgsRejectsUnknownFlag(t *testing.T) {
	if _, _, err := parseGlobalArgs([]string{"--unknown"}); err == nil {
		t.Fatal("parseGlobalArgs() accepted an unknown global flag")
	}
}

func TestMainHelpListsCodexApplyTarget(t *testing.T) {
	output := captureStdout(t, func() {
		if exit := run([]string{"--help"}); exit != 0 {
			t.Fatalf("run help exit = %d, want 0", exit)
		}
	})
	if !strings.Contains(output, "apply-custom [--verbose] all | codex | opencode") {
		t.Fatalf("main help output %q does not list the Codex apply target", output)
	}
}

func captureStdout(t *testing.T, run func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	previous := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = previous })
	run()
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	os.Stdout = previous
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read captured stdout: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	return string(output)
}

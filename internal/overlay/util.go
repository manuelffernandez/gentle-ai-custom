package overlay

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

func normalizeLF(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.ReplaceAll(text, "\r", "\n")
}

func normalizeLFTerminated(text string) string {
	normalized := normalizeLF(text)
	if strings.HasSuffix(normalized, "\n") {
		return normalized
	}
	return normalized + "\n"
}

func sha256Text(text string) string {
	sum := sha256.Sum256([]byte(normalizeLFTerminated(text)))
	return hex.EncodeToString(sum[:])
}

func expandUser(pathValue string) string {
	if pathValue == "" || pathValue[0] != '~' {
		return pathValue
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return pathValue
	}
	if pathValue == "~" {
		return home
	}
	if strings.HasPrefix(pathValue, "~/") || strings.HasPrefix(pathValue, "~\\") {
		return filepath.Join(home, pathValue[2:])
	}
	return pathValue
}

func readText(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func parseSimpleYAML(path string) (map[string]string, error) {
	text, err := readText(path)
	if err != nil {
		return nil, err
	}
	data := map[string]string{}
	for idx, rawLine := range strings.Split(text, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(rawLine, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid metadata line %d in %s: missing ':' separator", idx+1, path)
		}
		data[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return data, nil
}

func copyFile(src, dst string) error {
	raw, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return writeAtomicFile(dst, raw, 0o644)
}

func copyFileWithStatus(src, dst string) (string, error) {
	raw, err := os.ReadFile(src)
	if err != nil {
		return "", err
	}
	return writeAtomicFileWithStatus(dst, raw, 0o644)
}

func writeTextFile(path, content string) error {
	return writeAtomicFile(path, []byte(normalizeLFTerminated(content)), 0o644)
}

func writeTextFileWithStatus(path, content string) (string, error) {
	return writeAtomicFileWithStatus(path, []byte(normalizeLFTerminated(content)), 0o644)
}

func writeAtomicFile(path string, data []byte, defaultMode os.FileMode) error {
	_, err := writeAtomicFileWithStatus(path, data, defaultMode)
	return err
}

func writeAtomicFileWithStatus(path string, data []byte, defaultMode os.FileMode) (string, error) {
	status := "new"
	if existing, err := os.ReadFile(path); err == nil {
		if bytes.Equal(existing, data) {
			return "unchanged", nil
		} else {
			status = "changed"
		}
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	mode := defaultMode
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	cleanup := func() {
		_ = os.Remove(tmpPath)
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		cleanup()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return "", err
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		cleanup()
		return "", err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return "", err
	}
	return status, nil
}

func usageCommandName(subcommand string) string {
	if entrypoint := strings.TrimSpace(os.Getenv("GENTLE_AI_CUSTOM_ENTRYPOINT")); entrypoint != "" {
		return entrypoint
	}
	name := filepath.Base(os.Args[0])
	if strings.Contains(name, "gentle-ai-overlay") && subcommand != "" {
		return name + " " + subcommand
	}
	return name
}

func readJSONFile(path string, target any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

func readJSONAny(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func jsonString(value any) string {
	s, _ := value.(string)
	return s
}

func jsonObject(value any) (map[string]any, bool) {
	obj, ok := value.(map[string]any)
	return obj, ok
}

func jsonArray(value any) ([]any, bool) {
	arr, ok := value.([]any)
	return arr, ok
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func runGit(repo string, allowFailure bool, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	output, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(output))
	if err != nil {
		if allowFailure {
			return "", nil
		}
		if text == "" {
			text = err.Error()
		}
		return "", fmt.Errorf("%s: %s", repo, text)
	}
	return text, nil
}

func writeJSONIndented(path string, data any) error {
	_, err := writeJSONIndentedWithStatus(path, data)
	return err
}

func writeJSONIndentedWithStatus(path string, data any) (string, error) {
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	raw = append(raw, '\n')
	return writeAtomicFileWithStatus(path, raw, 0o644)
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

func writeString(w io.Writer, value string) {
	_, _ = io.WriteString(w, value)
}

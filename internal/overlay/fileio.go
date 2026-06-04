package overlay

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}


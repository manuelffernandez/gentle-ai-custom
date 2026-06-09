package overlay

import (
	"bytes"
	"os"
	"path/filepath"
)

func readText(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(raw), nil
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


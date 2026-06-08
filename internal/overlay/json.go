package overlay

import (
	"encoding/json"
	"os"
)

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

func writeJSONIndentedWithStatus(path string, data any) (string, error) {
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	raw = append(raw, '\n')
	return writeAtomicFileWithStatus(path, raw, 0o644)
}

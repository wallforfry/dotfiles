package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func replaceOnce(path, old, new string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !strings.Contains(string(content), old) {
		return fmt.Errorf("texte source absent")
	}
	return os.WriteFile(path, []byte(strings.Replace(string(content), old, new, 1)), 0o600)
}

func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o600)
}

func appendFile(path, content string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(content); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func removeLinePrefix(root, specification string) error {
	path, prefix, ok := strings.Cut(specification, "|")
	if !ok {
		return fmt.Errorf("suppression de ligne invalide")
	}
	content, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		return err
	}
	lines := strings.Split(string(content), "\n")
	kept := make([]string, 0, len(lines))
	removed := false
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			removed = true
			continue
		}
		kept = append(kept, line)
	}
	if !removed {
		return fmt.Errorf("ligne source absente")
	}
	return os.WriteFile(filepath.Join(root, path), []byte(strings.Join(kept, "\n")), 0o600)
}

func copyFirst(root, specification string) error {
	pattern, destination, ok := strings.Cut(specification, "|")
	if !ok {
		return fmt.Errorf("copie invalide")
	}
	matches, err := filepath.Glob(filepath.Join(root, pattern))
	if err != nil || len(matches) == 0 {
		return fmt.Errorf("source de copie absente")
	}
	content, err := os.ReadFile(matches[0])
	if err != nil {
		return err
	}
	return writeFile(filepath.Join(root, destination), string(content))
}

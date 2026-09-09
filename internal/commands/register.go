package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func RegisterClaudeHook(runtime Runtime, _ []string) int {
	home := runtime.env("HOME", "")
	settings := filepath.Join(home, ".claude", "settings.json")
	hook := filepath.Join(home, ".claude", "hooks", "agent-handoff")
	if info, err := os.Stat(hook); err != nil || info.Mode()&0o111 == 0 {
		fprintf(runtime.Stderr, "⚠️   %s absent ou non exécutable : rien à enregistrer\n", hook)
		return 0
	}
	content, mode, err := readSettings(settings)
	if err != nil {
		fprintf(runtime.Stderr, "⚠️   settings.json inchangé : %s\n", err)
		return 0
	}
	var document map[string]any
	if json.Unmarshal(content, &document) != nil {
		fprintf(runtime.Stderr, "⚠️   settings.json inchangé : JSON invalide\n")
		return 0
	}
	if hookRegistered(document, hook) {
		return 0
	}
	if err := appendStopHook(document, hook); err != nil {
		fprintf(runtime.Stderr, "⚠️   settings.json inchangé : %s\n", err)
		return 0
	}
	if err := replaceJSON(settings, content, document, mode); err != nil {
		fprintf(runtime.Stderr, "⚠️   settings.json inchangé : %s\n", err)
		return 0
	}
	fprintf(runtime.Stdout, "🪝  hook agent-handoff enregistré dans settings.json\n")
	return 0
}

func readSettings(path string) ([]byte, os.FileMode, error) {
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []byte("{}\n"), 0o600, nil
	}
	if err != nil {
		return nil, 0, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, 0, err
	}
	return content, info.Mode().Perm(), nil
}

func hookRegistered(document map[string]any, hook string) bool {
	hooks, _ := document["hooks"].(map[string]any)
	stops, _ := hooks["Stop"].([]any)
	for _, stop := range stops {
		group, _ := stop.(map[string]any)
		commands, _ := group["hooks"].([]any)
		for _, command := range commands {
			entry, _ := command.(map[string]any)
			if entry["command"] == hook {
				return true
			}
		}
	}
	return false
}

func appendStopHook(document map[string]any, hook string) error {
	hooks, present := document["hooks"]
	if !present {
		hooks = map[string]any{}
		document["hooks"] = hooks
	}
	hookMap, ok := hooks.(map[string]any)
	if !ok {
		return fmt.Errorf("champ hooks incompatible")
	}
	stops, present := hookMap["Stop"]
	if !present {
		stops = []any{}
	}
	stopList, ok := stops.([]any)
	if !ok {
		return fmt.Errorf("champ hooks.Stop incompatible")
	}
	entry := map[string]any{"hooks": []any{map[string]any{"type": "command", "command": hook}}}
	hookMap["Stop"] = append(stopList, entry)
	return nil
}

func replaceJSON(path string, original []byte, document map[string]any, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path+".bak", original, mode); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), "settings.*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporary.Write(append(encoded, '\n')); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

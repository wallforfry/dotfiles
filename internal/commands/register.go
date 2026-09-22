package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

var errSettingsChanged = errors.New("settings.json modifié pendant l'enregistrement")

type claudeHook struct {
	event      string
	matcher    string
	executable string
	command    string
	timeout    int
}

func claudeHooks(home string) []claudeHook {
	handoff := filepath.Join(home, ".claude", "hooks", "agent-handoff")
	wakeup := filepath.Join(home, ".local", "bin", "smartcard-wakeup")
	return []claudeHook{
		{event: "Stop", executable: handoff, command: handoff},
		{event: "PostToolUse", matcher: "^Bash$", executable: wakeup, command: wakeup + " --hook", timeout: 10},
	}
}

func RegisterClaudeHook(runtime Runtime, _ []string) int {
	home := runtime.env("HOME", "")
	settings := filepath.Join(home, ".claude", "settings.json")
	hooks := make([]claudeHook, 0, 2)
	for _, hook := range claudeHooks(home) {
		if info, err := os.Stat(hook.executable); err != nil || info.Mode()&0o111 == 0 {
			fprintf(runtime.Stderr, "⚠️   %s absent ou non exécutable : hook ignoré\n", hook.executable)
			continue
		}
		hooks = append(hooks, hook)
	}
	if len(hooks) == 0 {
		return 0
	}
	for range 3 {
		registered, err := registerHooks(settings, hooks)
		if errors.Is(err, errSettingsChanged) {
			continue
		}
		if err != nil {
			fprintf(runtime.Stderr, "⚠️   settings.json inchangé : %s\n", err)
			return 0
		}
		for _, hook := range registered {
			fprintf(runtime.Stdout, "🪝  hook %s enregistré dans settings.json\n", filepath.Base(hook.executable))
		}
		return 0
	}
	fprintf(runtime.Stderr, "⚠️   settings.json inchangé : modifications concurrentes répétées\n")
	return 0
}

// Une seule réécriture pour tous les hooks : la sauvegarde .bak reste celle de
// l'état d'origine, et une interruption ne laisse pas un enregistrement partiel.
func registerHooks(settings string, hooks []claudeHook) ([]claudeHook, error) {
	content, mode, existed, err := readSettings(settings)
	if err != nil {
		return nil, err
	}
	var document map[string]any
	if json.Unmarshal(content, &document) != nil {
		return nil, fmt.Errorf("JSON invalide")
	}
	missing := make([]claudeHook, 0, len(hooks))
	for _, hook := range hooks {
		if hookRegistered(document, hook) {
			continue
		}
		if err := appendHook(document, hook); err != nil {
			return nil, err
		}
		missing = append(missing, hook)
	}
	if len(missing) == 0 {
		return nil, nil
	}
	if err := replaceJSON(settings, content, document, mode, existed); err != nil {
		return nil, err
	}
	return missing, nil
}

func readSettings(path string) ([]byte, os.FileMode, bool, error) {
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []byte("{}\n"), 0o600, false, nil
	}
	if err != nil {
		return nil, 0, false, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, 0, false, err
	}
	return content, info.Mode().Perm(), true, nil
}

func hookRegistered(document map[string]any, hook claudeHook) bool {
	hooks, _ := document["hooks"].(map[string]any)
	groups, _ := hooks[hook.event].([]any)
	for _, raw := range groups {
		group, _ := raw.(map[string]any)
		entries, _ := group["hooks"].([]any)
		for _, entry := range entries {
			command, _ := entry.(map[string]any)
			if command["command"] == hook.command {
				return true
			}
		}
	}
	return false
}

func appendHook(document map[string]any, hook claudeHook) error {
	hooks, present := document["hooks"]
	if !present {
		hooks = map[string]any{}
		document["hooks"] = hooks
	}
	hookMap, ok := hooks.(map[string]any)
	if !ok {
		return fmt.Errorf("champ hooks incompatible")
	}
	groups, present := hookMap[hook.event]
	if !present {
		groups = []any{}
	}
	groupList, ok := groups.([]any)
	if !ok {
		return fmt.Errorf("champ hooks.%s incompatible", hook.event)
	}
	command := map[string]any{"type": "command", "command": hook.command}
	if hook.timeout > 0 {
		command["timeout"] = hook.timeout
	}
	entry := map[string]any{"hooks": []any{command}}
	if hook.matcher != "" {
		entry["matcher"] = hook.matcher
	}
	hookMap[hook.event] = append(groupList, entry)
	return nil
}

func replaceJSON(path string, original []byte, document map[string]any, mode os.FileMode, existed bool) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	temporaryPath, err := writeJSONCandidate(directory, document, mode)
	if err != nil {
		return err
	}
	defer os.Remove(temporaryPath)
	capturePath, err := captureSettings(path, directory, original, existed)
	if err != nil {
		return err
	}
	if capturePath != "" {
		defer os.Remove(capturePath)
	}
	return publishJSON(temporaryPath, capturePath, path, original, mode, existed)
}

func writeJSONCandidate(directory string, document map[string]any, mode os.FileMode) (string, error) {
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return "", err
	}
	temporary, err := os.CreateTemp(directory, "settings.next.*")
	if err != nil {
		return "", err
	}
	temporaryPath := temporary.Name()
	if _, err := temporary.Write(append(encoded, '\n')); err != nil {
		_ = temporary.Close()
		_ = os.Remove(temporaryPath)
		return "", err
	}
	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		_ = os.Remove(temporaryPath)
		return "", err
	}
	if err := temporary.Close(); err != nil {
		_ = os.Remove(temporaryPath)
		return "", err
	}
	return temporaryPath, nil
}

func captureSettings(path, directory string, original []byte, existed bool) (string, error) {
	capture, err := os.CreateTemp(directory, "settings.previous.*")
	if err != nil {
		return "", err
	}
	capturePath := capture.Name()
	if err := capture.Close(); err != nil {
		return "", err
	}
	if err := os.Remove(capturePath); err != nil {
		return "", err
	}
	if !existed {
		if _, err := os.Stat(path); err == nil {
			return "", errSettingsChanged
		} else if !os.IsNotExist(err) {
			return "", err
		}
		return "", nil
	}
	if err := os.Rename(path, capturePath); err != nil {
		if os.IsNotExist(err) {
			return "", errSettingsChanged
		}
		return "", err
	}
	captured, err := os.ReadFile(capturePath)
	if err != nil || !bytes.Equal(captured, original) {
		if restoreErr := restoreCapturedSettings(capturePath, path); restoreErr != nil {
			return "", restoreErr
		}
		if err != nil {
			return "", err
		}
		return "", errSettingsChanged
	}
	return capturePath, nil
}

func publishJSON(temporaryPath, capturePath, path string, original []byte, mode os.FileMode, existed bool) error {
	if err := os.WriteFile(path+".bak", original, mode); err != nil {
		if existed {
			if restoreErr := restoreCapturedSettings(capturePath, path); restoreErr != nil {
				return restoreErr
			}
		}
		return err
	}
	if err := os.Link(temporaryPath, path); err != nil {
		if existed {
			if restoreErr := restoreCapturedSettings(capturePath, path); restoreErr != nil {
				return restoreErr
			}
		}
		if errors.Is(err, fs.ErrExist) {
			return errSettingsChanged
		}
		return err
	}
	return nil
}

func restoreCapturedSettings(capture, path string) error {
	if err := os.Link(capture, path); err != nil && !errors.Is(err, fs.ErrExist) {
		return err
	}
	return nil
}

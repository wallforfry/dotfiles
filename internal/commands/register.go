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

func RegisterClaudeHook(runtime Runtime, _ []string) int {
	home := runtime.env("HOME", "")
	settings := filepath.Join(home, ".claude", "settings.json")
	hook := filepath.Join(home, ".claude", "hooks", "agent-handoff")
	if info, err := os.Stat(hook); err != nil || info.Mode()&0o111 == 0 {
		fprintf(runtime.Stderr, "⚠️   %s absent ou non exécutable : rien à enregistrer\n", hook)
		return 0
	}
	for range 3 {
		changed, err := registerHook(settings, hook)
		if errors.Is(err, errSettingsChanged) {
			continue
		}
		if err != nil {
			fprintf(runtime.Stderr, "⚠️   settings.json inchangé : %s\n", err)
			return 0
		}
		if changed {
			fprintf(runtime.Stdout, "🪝  hook agent-handoff enregistré dans settings.json\n")
		}
		return 0
	}
	fprintf(runtime.Stderr, "⚠️   settings.json inchangé : modifications concurrentes répétées\n")
	return 0
}

func registerHook(settings, hook string) (bool, error) {
	content, mode, existed, err := readSettings(settings)
	if err != nil {
		return false, err
	}
	var document map[string]any
	if json.Unmarshal(content, &document) != nil {
		return false, fmt.Errorf("JSON invalide")
	}
	if hookRegistered(document, hook) {
		return false, nil
	}
	if err := appendStopHook(document, hook); err != nil {
		return false, err
	}
	if err := replaceJSON(settings, content, document, mode, existed); err != nil {
		return false, err
	}
	return true, nil
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

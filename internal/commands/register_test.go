package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRegisterClaudeHookPreservesSettingsAndIsIdempotent(t *testing.T) {
	home := t.TempDir()
	hook := filepath.Join(home, ".claude", "hooks", "agent-handoff")
	settings := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(hook), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hook, []byte("binary"), 0o700); err != nil {
		t.Fatal(err)
	}
	initial := []byte(`{"statusLine":{"type":"command"},"hooks":{"Stop":[]}}`)
	if err := os.WriteFile(settings, initial, 0o600); err != nil {
		t.Fatal(err)
	}
	runtime, stdout, _ := testRuntime(&fakeExecutor{}, map[string]string{"HOME": home})
	if code := RegisterClaudeHook(runtime, nil); code != 0 {
		t.Fatalf("RegisterClaudeHook() = %d", code)
	}
	if code := RegisterClaudeHook(runtime, nil); code != 0 {
		t.Fatalf("second RegisterClaudeHook() = %d", code)
	}
	content, err := os.ReadFile(settings)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(content, &document); err != nil {
		t.Fatal(err)
	}
	if !hookRegistered(document, hook) || document["statusLine"] == nil {
		t.Fatalf("settings not preserved: %s", content)
	}
	if bytes.Count(stdout.Bytes(), []byte("enregistré")) != 1 {
		t.Fatalf("stdout = %q", stdout.String())
	}
	backup, err := os.ReadFile(settings + ".bak")
	if err != nil || !bytes.Equal(backup, initial) {
		t.Fatalf("backup = %q, err = %v", backup, err)
	}
}

func TestRegisterClaudeHookFailsClosedOnIncompatibleState(t *testing.T) {
	home := t.TempDir()
	hook := filepath.Join(home, ".claude", "hooks", "agent-handoff")
	settings := filepath.Join(home, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(hook), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(hook, nil, 0o700); err != nil {
		t.Fatal(err)
	}
	initial := []byte(`{"hooks":[]}`)
	if err := os.WriteFile(settings, initial, 0o600); err != nil {
		t.Fatal(err)
	}
	runtime, _, stderr := testRuntime(&fakeExecutor{}, map[string]string{"HOME": home})
	RegisterClaudeHook(runtime, nil)
	content, err := os.ReadFile(settings)
	if err != nil || !bytes.Equal(content, initial) {
		t.Fatalf("settings changed: %q, err = %v", content, err)
	}
	if stderr.Len() == 0 {
		t.Fatal("missing warning")
	}
}

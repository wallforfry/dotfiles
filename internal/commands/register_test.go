package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
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

func TestReplaceJSONRejectsConcurrentModification(t *testing.T) {
	directory := t.TempDir()
	settings := filepath.Join(directory, "settings.json")
	original := []byte(`{"theme":"dark"}`)
	concurrent := []byte(`{"theme":"light"}`)
	if err := os.WriteFile(settings, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(settings, concurrent, 0o600); err != nil {
		t.Fatal(err)
	}
	document := map[string]any{"theme": "dark", "hooks": map[string]any{}}
	if err := replaceJSON(settings, original, document, 0o600, true); !errors.Is(err, errSettingsChanged) {
		t.Fatalf("replaceJSON() error = %v", err)
	}
	content, err := os.ReadFile(settings)
	if err != nil || !bytes.Equal(content, concurrent) {
		t.Fatalf("concurrent settings changed: %q, err = %v", content, err)
	}
}

func TestReplaceJSONDoesNotOverwriteWriterDuringBackup(t *testing.T) {
	directory := t.TempDir()
	settings := filepath.Join(directory, "settings.json")
	original := []byte(`{"theme":"dark"}`)
	concurrent := []byte(`{"theme":"light"}`)
	if err := os.WriteFile(settings, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Mkfifo(settings+".bak", 0o600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		document := map[string]any{"theme": "dark", "hooks": map[string]any{}}
		done <- replaceJSON(settings, original, document, 0o600, true)
	}()
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(settings); os.IsNotExist(err) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("settings.json was not captured")
		}
		time.Sleep(time.Millisecond)
	}
	if err := os.WriteFile(settings, concurrent, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := os.ReadFile(settings + ".bak"); err != nil {
		t.Fatal(err)
	}
	if err := <-done; !errors.Is(err, errSettingsChanged) {
		t.Fatalf("replaceJSON() error = %v", err)
	}
	content, err := os.ReadFile(settings)
	if err != nil || !bytes.Equal(content, concurrent) {
		t.Fatalf("concurrent settings changed: %q, err = %v", content, err)
	}
}

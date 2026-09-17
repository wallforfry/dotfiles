package verify

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestHindsightBankSkillAddsAndRemovesMapping(t *testing.T) {
	home := t.TempDir()
	repository := filepath.Join(home, "repository")
	if err := os.MkdirAll(repository, 0o700); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(home, ".hindsight", "dotfiles.json")
	if err := os.MkdirAll(filepath.Dir(config), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"apiUrl":"https://memory.invalid","apiToken":"secret","registrations":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	log := filepath.Join(home, "chezmoi.log")
	fakeBin := filepath.Join(home, "bin")
	if err := os.MkdirAll(fakeBin, 0o700); err != nil {
		t.Fatal(err)
	}
	writeBankSkillCommand(t, fakeBin, "hindsight", "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$HINDSIGHT_LOG\"\n")
	writeBankSkillCommand(t, fakeBin, "chezmoi", "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$CHEZMOI_LOG\"\n")
	script := hindsightBankSkillScript(t)
	runBankSkill(t, script, home, fakeBin, log, nil, "add", repository, "bank")
	if registrations := bankSkillRegistrations(t, config); len(registrations) != 1 || registrations[0]["repository"] != canonicalBankSkillDirectory(t, repository) || registrations[0]["bank"] != "bank" {
		t.Fatalf("registrations after add = %#v", registrations)
	}
	runBankSkill(t, script, home, fakeBin, log, nil, "remove", repository)
	if registrations := bankSkillRegistrations(t, config); len(registrations) != 0 {
		t.Fatalf("registrations after remove = %#v", registrations)
	}
	content, err := os.ReadFile(log)
	if err != nil || strings.Count(string(content), "add --encrypt") != 2 || strings.Count(string(content), "apply --force") != 2 {
		t.Fatalf("chezmoi calls = %q, %v", content, err)
	}
}

func TestHindsightBankSkillMapsNonGitDirectory(t *testing.T) {
	home, config, fakeBin, log, script := setupBankSkill(t)
	directory := filepath.Join(home, "directory")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	runBankSkill(t, script, home, fakeBin, log, nil, "add", directory, "bank")
	if registrations := bankSkillRegistrations(t, config); len(registrations) != 1 || registrations[0]["repository"] != canonicalBankSkillDirectory(t, directory) {
		t.Fatalf("registrations = %#v", registrations)
	}
}

func TestHindsightBankSkillMapsWorktreeDirectory(t *testing.T) {
	home, config, fakeBin, log, script := setupBankSkill(t)
	mainRepository := filepath.Join(home, "main")
	worktree := filepath.Join(home, "worktree")
	for _, directory := range []string{mainRepository, worktree} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	runBankSkill(t, script, home, fakeBin, log, nil, "add", worktree, "bank")
	if registrations := bankSkillRegistrations(t, config); len(registrations) != 1 || registrations[0]["repository"] != canonicalBankSkillDirectory(t, worktree) {
		t.Fatalf("registrations = %#v", registrations)
	}
}

func TestHindsightBankSkillRejectsBlankBankBeforeMutation(t *testing.T) {
	home, config, fakeBin, log, script := setupBankSkill(t)
	directory := filepath.Join(home, "directory")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if output, err := runBankSkillError(script, home, fakeBin, log, nil, "add", directory, ""); err == nil {
		t.Fatalf("manage-bank succeeded with blank bank: %s", output)
	}
	if registrations := bankSkillRegistrations(t, config); len(registrations) != 0 {
		t.Fatalf("registrations = %#v", registrations)
	}
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatalf("chezmoi log error = %v", err)
	}
}

func TestHindsightBankSkillCreatesRemoteBank(t *testing.T) {
	home, _, fakeBin, log, script := setupBankSkill(t)
	hindsightLog := filepath.Join(home, "hindsight.log")
	runBankSkill(t, script, home, fakeBin, log, []string{"HINDSIGHT_LOG=" + hindsightLog}, "create", "bank")
	content, err := os.ReadFile(hindsightLog)
	if err != nil || string(content) != "bank create bank\n" {
		t.Fatalf("hindsight calls = %q, %v", content, err)
	}
}

func hindsightBankSkillScript(t *testing.T) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve test source path")
	}
	return filepath.Join(filepath.Dir(source), "..", "..", "dot_config", "agent-skills", "hindsight-bank", "scripts", "executable_manage-bank.sh")
}

func writeBankSkillCommand(t *testing.T, directory, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0o700); err != nil {
		t.Fatal(err)
	}
}

func setupBankSkill(t *testing.T) (string, string, string, string, string) {
	t.Helper()
	home := t.TempDir()
	config := filepath.Join(home, ".hindsight", "dotfiles.json")
	if err := os.MkdirAll(filepath.Dir(config), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, []byte(`{"apiUrl":"https://memory.invalid","apiToken":"secret","registrations":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	fakeBin := filepath.Join(home, "bin")
	if err := os.MkdirAll(fakeBin, 0o700); err != nil {
		t.Fatal(err)
	}
	writeBankSkillCommand(t, fakeBin, "hindsight", "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$HINDSIGHT_LOG\"\n")
	writeBankSkillCommand(t, fakeBin, "chezmoi", "#!/bin/sh\nprintf '%s\\n' \"$*\" >> \"$CHEZMOI_LOG\"\n")
	return home, config, fakeBin, filepath.Join(home, "chezmoi.log"), hindsightBankSkillScript(t)
}

func runBankSkill(t *testing.T, script, home, fakeBin, log string, environment []string, args ...string) {
	t.Helper()
	if output, err := runBankSkillError(script, home, fakeBin, log, environment, args...); err != nil {
		t.Fatalf("manage-bank %v: %s: %v", args, output, err)
	}
}

func runBankSkillError(script, home, fakeBin, log string, environment []string, args ...string) ([]byte, error) {
	command := exec.Command("sh", append([]string{script}, args...)...)
	command.Env = append(os.Environ(), append([]string{"HOME=" + home, "PATH=" + fakeBin + ":" + os.Getenv("PATH"), "CHEZMOI_LOG=" + log}, environment...)...)
	return command.CombinedOutput()
}

func bankSkillRegistrations(t *testing.T, path string) []map[string]string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		Registrations []map[string]string `json:"registrations"`
	}
	if err := json.Unmarshal(content, &config); err != nil {
		t.Fatal(err)
	}
	return config.Registrations
}

func canonicalBankSkillDirectory(t *testing.T, directory string) string {
	t.Helper()
	canonical, err := filepath.EvalSymlinks(directory)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}

package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/wallforfry/dotfiles/internal/commands"
)

func TestUnknownCommandUsesUsageExit(t *testing.T) {
	stderr := &bytes.Buffer{}
	runtime := commands.NewRuntime()
	runtime.Stdout = &bytes.Buffer{}
	runtime.Stderr = stderr
	if code := run([]string{"dotfiles", "unknown"}, runtime); code != commands.ExitUsage {
		t.Fatalf("run() = %d", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("missing usage")
	}
}

func TestSymlinkNameDispatchesDirectly(t *testing.T) {
	runtime := commands.NewRuntime()
	stdout := &bytes.Buffer{}
	runtime.Stdout = stdout
	runtime.Stderr = &bytes.Buffer{}
	if code := run([]string{"/home/user/bin/cloak", "--url"}, runtime); code != 0 {
		t.Fatalf("run() = %d", code)
	}
	if stdout.String() != "http://host.docker.internal:9222\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestRepositoryRootAcceptsExplicitPath(t *testing.T) {
	runtime := commands.NewRuntime()
	runtime.Stderr = &bytes.Buffer{}
	want := t.TempDir()
	got, code := repositoryRoot([]string{"--repository", want}, runtime)
	if code != 0 || got != want {
		t.Fatalf("repositoryRoot() = %q, %d", got, code)
	}
	if _, err := os.Stat(got); err != nil {
		t.Fatal(err)
	}
}

func TestHarnessAuditRequiresRepository(t *testing.T) {
	runtime := commands.NewRuntime()
	runtime.Stdout = &bytes.Buffer{}
	runtime.Stderr = &bytes.Buffer{}
	if code := run([]string{"dotfiles", "harness-audit"}, runtime); code != commands.ExitUsage {
		t.Fatalf("run() = %d", code)
	}
}

func TestAuditConfigReadsOverrides(t *testing.T) {
	t.Setenv("CLAUDE_PROJECTS", "/claude")
	t.Setenv("CODEX_SESSIONS", "/codex")
	t.Setenv("HARNESS_RULES_SINCE", "2026-09-01")
	config := auditConfig("/repository")
	if config.Root != "/repository" || config.ClaudeProjects != "/claude" || config.CodexSessions != "/codex" || config.Since != "2026-09-01" {
		t.Fatalf("auditConfig() = %#v", config)
	}
}

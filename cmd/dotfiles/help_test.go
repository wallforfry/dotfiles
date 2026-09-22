package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wallforfry/dotfiles/internal/commands"
)

func newTestRuntime() (commands.Runtime, *bytes.Buffer, *bytes.Buffer) {
	runtime := commands.NewRuntime()
	stdout, stderr := &bytes.Buffer{}, &bytes.Buffer{}
	runtime.Stdout, runtime.Stderr = stdout, stderr
	return runtime, stdout, stderr
}

// Le catalogue est la source unique de l'aide et de la complétion : une
// commande routée sans entrée ne serait ni documentée ni complétée.
func TestEveryDispatchedCommandIsCatalogued(t *testing.T) {
	routed := []string{"completion", "verify", "harness-audit", "validate-skill-routing", "register-claude-hook", "register-hindsight", "hindsight"}
	for name := range directCommands {
		routed = append(routed, name)
	}
	for _, name := range routed {
		if _, exists := lookup(name); !exists {
			t.Errorf("commande %q absente du catalogue", name)
		}
	}
}

func TestCataloguedCommandsHaveSummaryAndUsage(t *testing.T) {
	for _, command := range catalog {
		if command.Summary == "" || command.Usage == "" {
			t.Errorf("commande %q incomplète : %#v", command.Name, command)
		}
	}
}

func TestHelpListsEveryCommand(t *testing.T) {
	runtime, stdout, _ := newTestRuntime()
	if code := run([]string{"dotfiles", "help"}, runtime); code != 0 {
		t.Fatalf("run() = %d", code)
	}
	for _, command := range catalog {
		if !strings.Contains(stdout.String(), command.Name) {
			t.Errorf("aide sans %q", command.Name)
		}
	}
}

func TestHelpFlagOnACommandPrintsItsUsage(t *testing.T) {
	runtime, stdout, _ := newTestRuntime()
	if code := run([]string{"dotfiles", "verify", "--help"}, runtime); code != 0 {
		t.Fatalf("run() = %d", code)
	}
	if !strings.Contains(stdout.String(), "dotfiles verify [--repository") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

// Un lien de ~/.local/bin doit décrire sa propre commande, pas le CLI entier.
func TestHelpFlagOnALinkedNamePrintsItsUsage(t *testing.T) {
	runtime, stdout, _ := newTestRuntime()
	if code := run([]string{"/home/user/.local/bin/cloak", "--help"}, runtime); code != 0 {
		t.Fatalf("run() = %d", code)
	}
	if !strings.Contains(stdout.String(), "cloak --start") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestHelpRejectsAnUnknownCommand(t *testing.T) {
	runtime, _, stderr := newTestRuntime()
	if code := run([]string{"dotfiles", "help", "inconnue"}, runtime); code != commands.ExitUsage {
		t.Fatalf("run() = %d", code)
	}
	if !strings.Contains(stderr.String(), "inconnue") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

// « help » n'est pas un drapeau : une commande liée doit garder cet argument.
func TestBareHelpWordReachesALinkedCommand(t *testing.T) {
	runtime, _, stderr := newTestRuntime()
	if code := run([]string{"/home/user/.local/bin/smartcard-wakeup", "help"}, runtime); code != commands.ExitUsage {
		t.Fatalf("run() = %d", code)
	}
	if !strings.HasPrefix(stderr.String(), "smartcard-wakeup: usage") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestHelpFlagAfterSubcommandsPrintsTheCommandUsage(t *testing.T) {
	runtime, stdout, _ := newTestRuntime()
	if code := run([]string{"dotfiles", "hindsight", "bank", "--help"}, runtime); code != 0 {
		t.Fatalf("run() = %d", code)
	}
	if !strings.Contains(stdout.String(), "dotfiles hindsight bank create") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestLinkedFlagMatchesTheDeployedSymlinks(t *testing.T) {
	for _, command := range catalog {
		_, err := os.Stat(filepath.Join("..", "..", "dot_local", "bin", "symlink_"+command.Name+".tmpl"))
		if deployed := err == nil; deployed != command.Linked {
			t.Errorf("%q : Linked = %t, lien déployé = %t", command.Name, command.Linked, deployed)
		}
	}
}

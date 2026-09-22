package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wallforfry/dotfiles/internal/commands"
)

func TestCompletionRequiresASupportedShell(t *testing.T) {
	for _, args := range [][]string{{}, {"fish"}, {"zsh", "bash"}} {
		runtime, _, stderr := newTestRuntime()
		if code := Completion(runtime, args); code != commands.ExitUsage {
			t.Fatalf("Completion(%v) = %d", args, code)
		}
		if stderr.Len() == 0 {
			t.Fatalf("Completion(%v) sans diagnostic", args)
		}
	}
}

func TestCompletionScriptsCoverEveryCommand(t *testing.T) {
	for _, shell := range []string{"zsh", "bash"} {
		runtime, stdout, _ := newTestRuntime()
		if code := run([]string{"dotfiles", "completion", shell}, runtime); code != 0 {
			t.Fatalf("completion %s = %d", shell, code)
		}
		for _, command := range commandList() {
			if !strings.Contains(stdout.String(), command.Name) {
				t.Errorf("complétion %s sans %q", shell, command.Name)
			}
		}
	}
}

// Une apostrophe dans un résumé casserait un script mal échappé : l'interpréteur
// est le seul juge, un test de sous-chaîne ne prouve rien.
func TestCompletionScriptsAreSyntacticallyValid(t *testing.T) {
	for _, shell := range []string{"zsh", "bash"} {
		binary, err := exec.LookPath(shell)
		if err != nil {
			t.Skipf("%s absent de la machine", shell)
		}
		runtime, stdout, _ := newTestRuntime()
		if code := run([]string{"dotfiles", "completion", shell}, runtime); code != 0 {
			t.Fatalf("completion %s = %d", shell, code)
		}
		script := filepath.Join(t.TempDir(), "completion."+shell)
		if err := os.WriteFile(script, stdout.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
		if output, err := exec.Command(binary, "-n", script).CombinedOutput(); err != nil {
			t.Fatalf("%s -n : %v\n%s", shell, err, output)
		}
	}
}

// Un contrôle de syntaxe ne prouve rien de l'échappement : des apostrophes non
// échappées se ré-apparient et laissent un script valide mais faux. Seule la
// valeur rendue par l'interpréteur le prouve.
func TestQuotedSummariesSurviveTheShell(t *testing.T) {
	for _, shell := range []string{"zsh", "bash"} {
		binary, err := exec.LookPath(shell)
		if err != nil {
			t.Skipf("%s absent de la machine", shell)
		}
		for _, command := range commandList() {
			output, err := exec.Command(binary, "-c", "printf %s "+shellQuote(command.Summary)).Output()
			if err != nil {
				t.Fatalf("%s : %v", shell, err)
			}
			if string(output) != command.Summary {
				t.Errorf("%s rend %q pour %q", shell, output, command.Summary)
			}
		}
	}
}

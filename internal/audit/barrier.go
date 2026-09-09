package audit

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/wallforfry/dotfiles/internal/verify"
)

func (a *auditor) measureBarrier() {
	a.section("Pouvoir de détection de la barrière Go", func() error {
		cache, err := os.UserCacheDir()
		if err != nil {
			return fmt.Errorf("cache utilisateur introuvable : détection non mesurée")
		}
		scratch, err := os.MkdirTemp(cache, "harness-audit.")
		if err != nil {
			return fmt.Errorf("clone temporaire impossible : détection non mesurée")
		}
		defer os.RemoveAll(scratch)
		if err := probeCaptureContract(scratch); err != nil {
			return fmt.Errorf("le contrat de copie du worktree n'est pas respecté")
		}
		return a.runMutationMatrix(scratch)
	})
}

func (a *auditor) runMutationMatrix(scratch string) error {
	repository := filepath.Join(scratch, "rep")
	if output, err := runCommand("", "git", "clone", "-q", "--no-hardlinks", a.config.Root, repository); err != nil {
		return fmt.Errorf("clone impossible : détection non mesurée: %s", output)
	}
	baseline, err := captureWorktree(a.config.Root, repository, scratch)
	if err != nil {
		return fmt.Errorf("capture du worktree interrompue : détection non mesurable")
	}
	marker := "harness-audit-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	list := filepath.Join(scratch, "sensible.txt")
	if err := os.WriteFile(list, []byte(marker+"\n"), 0o600); err != nil {
		return fmt.Errorf("liste sensible temporaire impossible")
	}
	if runVerify(repository, list) == 0 {
		return a.evaluateMutations(repository, baseline, list, marker)
	}
	return fmt.Errorf("le clone est rouge avant toute mutation : détection non mesurable")
}

func (a *auditor) evaluateMutations(repository, baseline, list, marker string) error {
	passed, rejected, accepted, observed := 0, 0, 0, 0
	expected := 0
	for _, item := range mutationCases() {
		if item.expectation == "observe" {
			observed++
			fmt.Fprintf(a.stdout, "  %-22s %-24s observation\n", item.promise, item.control)
			continue
		}
		expected++
		if err := restore(repository, baseline, list, marker); err != nil {
			return fmt.Errorf("restauration du clone interrompue : matrice abandonnée")
		}
		if err := applyMutation(repository, list, marker, item.steps); err != nil {
			return fmt.Errorf("mutation inapplicable, à réécrire : %s", item.label)
		}
		result := "reject"
		if runControl(repository, list, item.control) == 0 {
			result = "accept"
		}
		if result != item.expectation {
			return fmt.Errorf("%s ne respecte pas %s : %s devait être %s", item.control, item.promise, item.label, item.expectation)
		}
		passed++
		if result == "reject" {
			rejected++
		} else {
			accepted++
		}
	}
	if passed != expected {
		return fmt.Errorf("%d/%d comportements de barrière conformes", passed, expected)
	}
	a.ok("%d défauts rejetés, %d anti-mutants acceptés, %d promesses observées", rejected, accepted, observed)
	return nil
}

func restore(repository, baseline, list, marker string) error {
	if err := os.WriteFile(list, []byte(marker+"\n"), 0o600); err != nil {
		return err
	}
	commands := [][]string{{"switch", "-q", "--detach", baseline}, {"restore", "--source=" + baseline, "--staged", "--worktree", "."}, {"clean", "-qfd"}}
	for _, arguments := range commands {
		if _, err := gitOutput(repository, arguments...); err != nil {
			return err
		}
	}
	return nil
}

func runVerify(repository, sensitiveList string) int {
	return withSensitiveList(sensitiveList, func() int {
		return verify.Run(repository, io.Discard, io.Discard)
	})
}

func runControl(repository, sensitiveList string, control verify.Control) int {
	return withSensitiveList(sensitiveList, func() int {
		return verify.RunControl(repository, control, io.Discard, io.Discard)
	})
}

func withSensitiveList(sensitiveList string, run func() int) int {
	previous, existed := os.LookupEnv("SENSIBLE_LIST")
	os.Setenv("SENSIBLE_LIST", sensitiveList)
	defer func() {
		if existed {
			os.Setenv("SENSIBLE_LIST", previous)
		} else {
			os.Unsetenv("SENSIBLE_LIST")
		}
	}()
	return run()
}

func runCommand(directory, name string, arguments ...string) ([]byte, error) {
	command := exec.Command(name, arguments...)
	command.Dir = directory
	return command.CombinedOutput()
}

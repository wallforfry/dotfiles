package commands

import (
	"errors"
	"strings"
	"testing"
)

func TestSmartcardWakeupProbesBeforeRestartingScdaemon(t *testing.T) {
	reachable := &fakeExecutor{}
	runtime, stdout, _ := testRuntime(reachable, map[string]string{})
	if code := SmartcardWakeup(runtime, nil); code != 0 {
		t.Fatalf("SmartcardWakeup() = %d", code)
	}
	assertArgs(t, reachable.processes, [][]string{{"--card-status"}})
	if !strings.Contains(stdout.String(), "déjà disponible") {
		t.Fatalf("stdout = %q", stdout.String())
	}

	repaired := &fakeExecutor{errors: []error{errors.New("no such device"), nil, nil}}
	runtime, stdout, _ = testRuntime(repaired, map[string]string{})
	if code := SmartcardWakeup(runtime, nil); code != 0 {
		t.Fatalf("SmartcardWakeup() after repair = %d", code)
	}
	assertArgs(t, repaired.processes, [][]string{{"--card-status"}, {"--kill", "scdaemon"}, {"--card-status"}})
	if !strings.Contains(stdout.String(), "réveillée") {
		t.Fatalf("stdout = %q", stdout.String())
	}

	absent := &fakeExecutor{errors: []error{errors.New("down"), nil, errors.New("still down")}}
	runtime, _, stderr := testRuntime(absent, map[string]string{})
	if code := SmartcardWakeup(runtime, nil); code != ExitUnavailable {
		t.Fatalf("SmartcardWakeup() without card = %d", code)
	}
	if stderr.Len() == 0 {
		t.Fatal("missing warning")
	}
	if code := SmartcardWakeup(runtime, []string{"--status"}); code != ExitUsage {
		t.Fatalf("usage code = %d", code)
	}
}

func TestSmartcardHookActsOnlyOnCardFailures(t *testing.T) {
	unrelated := &fakeExecutor{}
	runtime, _, stderr := testRuntime(unrelated, map[string]string{})
	runtime.Stdin = strings.NewReader(`{"tool_response":{"stderr":"npm ERR! 404"}}`)
	if code := SmartcardWakeup(runtime, []string{"--hook"}); code != 0 || len(unrelated.processes) != 0 {
		t.Fatalf("code = %d, processes = %d", code, len(unrelated.processes))
	}

	// Une signature sans panne réelle ne doit rien redémarrer.
	healthy := &fakeExecutor{}
	runtime, _, stderr = testRuntime(healthy, map[string]string{})
	runtime.Stdin = strings.NewReader(`{"tool_response":{"stderr":"gpg: no such device"}}`)
	if code := SmartcardWakeup(runtime, []string{"--hook"}); code != 0 {
		t.Fatalf("code = %d", code)
	}
	assertArgs(t, healthy.processes, [][]string{{"--card-status"}})

	broken := &fakeExecutor{errors: []error{errors.New("down"), nil, nil}}
	runtime, _, stderr = testRuntime(broken, map[string]string{})
	runtime.Stdin = strings.NewReader(`{"tool_response":{"stderr":"gpg: échec de la sélection de la carte"}}`)
	if code := SmartcardWakeup(runtime, []string{"--hook"}); code != ExitFeedback {
		t.Fatalf("code = %d, want %d", code, ExitFeedback)
	}
	if !strings.Contains(stderr.String(), "relance la commande") {
		t.Fatalf("stderr = %q", stderr.String())
	}

	// Sans carte - Linux, DSM - le hook se tait au lieu d'échouer à chaque appel.
	absent := &fakeExecutor{errors: []error{errors.New("down"), nil, errors.New("down")}}
	runtime, _, stderr = testRuntime(absent, map[string]string{})
	runtime.Stdin = strings.NewReader(`{"tool_response":{"stderr":"gpg: card error"}}`)
	if code := SmartcardWakeup(runtime, []string{"--hook"}); code != 0 || stderr.Len() != 0 {
		t.Fatalf("code = %d, stderr = %q", code, stderr.String())
	}

	// La commande soumise porte les signatures dès qu'on lit ce fichier.
	quoted := &fakeExecutor{}
	runtime, _, _ = testRuntime(quoted, map[string]string{})
	runtime.Stdin = strings.NewReader(`{"tool_input":{"command":"grep -rn \"card not available\" ."},"tool_response":{"stdout":""}}`)
	if code := SmartcardWakeup(runtime, []string{"--hook"}); code != 0 || len(quoted.processes) != 0 {
		t.Fatalf("code = %d, processes = %d", code, len(quoted.processes))
	}
}

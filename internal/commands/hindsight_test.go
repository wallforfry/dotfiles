package commands

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRegisterHindsightMergesMultiplePrivateConfigurations(t *testing.T) {
	home := t.TempDir()
	first := makeHindsightRepository(t, home, "first")
	second := makeHindsightRepository(t, home, "second")
	configuration := writeHindsightConfiguration(t, home, "https://memory.invalid", "secret", []hindsightRegistration{{Repository: first, Bank: "personal"}, {Repository: second, Bank: "personal"}})
	executor := &fakeExecutor{}
	runtime, _, _ := testRuntime(executor, map[string]string{"HOME": home})
	if code := RegisterHindsight(runtime, []string{"--config", configuration}); code != 0 {
		t.Fatalf("RegisterHindsight() = %d", code)
	}
	config := readJSON(t, filepath.Join(home, ".hindsight", "coding-agent.json"))
	paths := config["mapPathToBank"].(map[string]any)
	if paths[first] != "personal" || paths[second] != "personal" {
		t.Fatalf("paths = %#v", paths)
	}
	cursor := readJSON(t, filepath.Join(home, ".cursor", "mcp.json"))
	servers := cursor["mcpServers"].(map[string]any)
	if len(servers) != 1 || servers["hindsight-memory-personal"].(map[string]any)["url"] != "https://memory.invalid/mcp/personal/" {
		t.Fatalf("servers = %#v", servers)
	}
	if len(executor.processes) < 1 || executor.processes[0].Name != "bunx" {
		t.Fatalf("processes = %#v", executor.processes)
	}
	for _, path := range []string{filepath.Join(home, ".hindsight", "coding-agent.json"), filepath.Join(home, ".cursor", "mcp.json")} {
		info, err := os.Stat(path)
		if err != nil || info.Mode().Perm() != 0o600 {
			t.Fatalf("permissions %s = %v, %v", path, info.Mode().Perm(), err)
		}
	}
}

func TestRegisterHindsightRemovesDeletedRegistrations(t *testing.T) {
	home := t.TempDir()
	first := makeHindsightRepository(t, home, "first")
	second := makeHindsightRepository(t, home, "second")
	configuration := writeHindsightConfiguration(t, home, "https://memory.invalid", "secret", []hindsightRegistration{{Repository: first, Bank: "first"}, {Repository: second, Bank: "second"}})
	runtime, _, _ := testRuntime(&fakeExecutor{}, map[string]string{"HOME": home})
	if code := RegisterHindsight(runtime, []string{"--config", configuration}); code != 0 {
		t.Fatalf("initial RegisterHindsight() = %d", code)
	}
	configuration = writeHindsightConfiguration(t, home, "https://memory.invalid", "secret", []hindsightRegistration{{Repository: second, Bank: "second"}})
	if code := RegisterHindsight(runtime, []string{"--config", configuration}); code != 0 {
		t.Fatalf("updated RegisterHindsight() = %d", code)
	}
	paths := readJSON(t, filepath.Join(home, ".hindsight", "coding-agent.json"))["mapPathToBank"].(map[string]any)
	if _, exists := paths[first]; exists {
		t.Fatalf("removed repository is still registered: %#v", paths)
	}
	servers := readJSON(t, filepath.Join(home, ".cursor", "mcp.json"))["mcpServers"].(map[string]any)
	if _, exists := servers["hindsight-memory-first"]; exists {
		t.Fatalf("removed bank is still registered: %#v", servers)
	}
	configuration = writeHindsightConfiguration(t, home, "https://memory.invalid", "secret", nil)
	if code := RegisterHindsight(runtime, []string{"--config", configuration}); code != 0 {
		t.Fatalf("empty RegisterHindsight() = %d", code)
	}
	paths = readJSON(t, filepath.Join(home, ".hindsight", "coding-agent.json"))["mapPathToBank"].(map[string]any)
	if len(paths) != 0 {
		t.Fatalf("all repositories should be removed: %#v", paths)
	}
	servers = readJSON(t, filepath.Join(home, ".cursor", "mcp.json"))["mcpServers"].(map[string]any)
	if len(servers) != 0 {
		t.Fatalf("all banks should be removed: %#v", servers)
	}
}

func TestRegisterHindsightRejectsMissingCredentialsWithoutWriting(t *testing.T) {
	home := t.TempDir()
	configuration := writeHindsightConfiguration(t, home, "https://memory.invalid", "", nil)
	runtime, _, _ := testRuntime(&fakeExecutor{}, map[string]string{"HOME": home})
	if code := RegisterHindsight(runtime, []string{"--config", configuration}); code != ExitUsage {
		t.Fatalf("RegisterHindsight() = %d", code)
	}
	if _, err := os.Stat(filepath.Join(home, ".hindsight", "coding-agent.json")); !os.IsNotExist(err) {
		t.Fatalf("configuration created: %v", err)
	}
}

func TestHindsightBankAddsAndRemovesDirectory(t *testing.T) {
	home := t.TempDir()
	directory := makeHindsightRepository(t, home, "directory")
	configuration := filepath.Join(home, ".hindsight", "dotfiles.json")
	if err := os.MkdirAll(filepath.Dir(configuration), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configuration, []byte(`{"apiUrl":"https://memory.invalid","apiToken":"secret","registrations":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(configuration, 0o644); err != nil {
		t.Fatal(err)
	}
	executor := &fakeExecutor{}
	runtime, _, _ := testRuntime(executor, map[string]string{"HOME": home})
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relativeDirectory, err := filepath.Rel(workingDirectory, directory)
	if err != nil {
		t.Fatal(err)
	}
	if code := HindsightBank(runtime, []string{"bank", "add", relativeDirectory, "bank"}); code != 0 {
		t.Fatalf("HindsightBank(add) = %d", code)
	}
	updated := readHindsightConfiguration(t, configuration)
	if len(updated.Registrations) != 1 || updated.Registrations[0] != (hindsightRegistration{Repository: canonicalHindsightDirectory(t, directory), Bank: "bank"}) {
		t.Fatalf("registrations after add = %#v", updated.Registrations)
	}
	info, err := os.Stat(configuration)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("configuration permissions = %v, %v", info, err)
	}
	assertArgs(t, executor.processes, [][]string{{"add", "--encrypt", configuration}, {"apply", "--force"}})
	if executor.processes[0].Name != "chezmoi" || executor.processes[1].Name != "chezmoi" {
		t.Fatalf("processes = %#v", executor.processes)
	}
	executor = &fakeExecutor{}
	runtime, _, _ = testRuntime(executor, map[string]string{"HOME": home})
	if code := HindsightBank(runtime, []string{"bank", "remove", directory}); code != 0 {
		t.Fatalf("HindsightBank(remove) = %d", code)
	}
	if registrations := readHindsightConfiguration(t, configuration).Registrations; len(registrations) != 0 {
		t.Fatalf("registrations after remove = %#v", registrations)
	}
	assertArgs(t, executor.processes, [][]string{{"add", "--encrypt", configuration}, {"apply", "--force"}})
}

func TestHindsightBankCreatesRemoteBank(t *testing.T) {
	executor := &fakeExecutor{}
	runtime, _, _ := testRuntime(executor, map[string]string{})
	if code := HindsightBank(runtime, []string{"bank", "create", "bank"}); code != 0 {
		t.Fatalf("HindsightBank(create) = %d", code)
	}
	assertArgs(t, executor.processes, [][]string{{"bank", "create", "bank"}})
	if executor.processes[0].Name != "hindsight" {
		t.Fatalf("process = %#v", executor.processes[0])
	}
}

func TestHindsightBankRejectsBlankBankWithoutWriting(t *testing.T) {
	home := t.TempDir()
	directory := makeHindsightRepository(t, home, "directory")
	configuration := filepath.Join(home, ".hindsight", "dotfiles.json")
	if err := os.MkdirAll(filepath.Dir(configuration), 0o700); err != nil {
		t.Fatal(err)
	}
	original := []byte(`{"apiUrl":"https://memory.invalid","apiToken":"secret","registrations":[]}`)
	if err := os.WriteFile(configuration, original, 0o600); err != nil {
		t.Fatal(err)
	}
	executor := &fakeExecutor{}
	runtime, _, _ := testRuntime(executor, map[string]string{"HOME": home})
	if code := HindsightBank(runtime, []string{"bank", "add", directory, ""}); code != ExitUsage {
		t.Fatalf("HindsightBank(blank add) = %d", code)
	}
	content, err := os.ReadFile(configuration)
	if err != nil || string(content) != string(original) || len(executor.processes) != 0 {
		t.Fatalf("configuration or processes changed: %q, %#v, %v", content, executor.processes, err)
	}
}

func TestHindsightBankRestoresConfigurationWhenChezmoiFails(t *testing.T) {
	for _, test := range []struct {
		name          string
		errors        []error
		processesWant int
		restored      bool
	}{
		{name: "encrypt", errors: []error{errors.New("encrypt")}, processesWant: 1, restored: true},
		{name: "apply", errors: []error{nil, errors.New("apply")}, processesWant: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			directory := makeHindsightRepository(t, home, "directory")
			configuration := filepath.Join(home, ".hindsight", "dotfiles.json")
			if err := os.MkdirAll(filepath.Dir(configuration), 0o700); err != nil {
				t.Fatal(err)
			}
			original := []byte(`{"apiUrl":"https://memory.invalid","apiToken":"secret","registrations":[]}`)
			if err := os.WriteFile(configuration, original, 0o600); err != nil {
				t.Fatal(err)
			}
			executor := &fakeExecutor{errors: test.errors}
			runtime, _, _ := testRuntime(executor, map[string]string{"HOME": home})
			if code := HindsightBank(runtime, []string{"bank", "add", directory, "bank"}); code == 0 {
				t.Fatal("HindsightBank(add) succeeded")
			}
			content, err := os.ReadFile(configuration)
			if err != nil {
				t.Fatalf("configuration = %q, %v", content, err)
			}
			if test.restored && string(content) != string(original) {
				t.Fatalf("configuration after rollback = %q", content)
			}
			if !test.restored && len(readHindsightConfiguration(t, configuration).Registrations) != 1 {
				t.Fatalf("configuration after apply failure = %q", content)
			}
			info, err := os.Stat(configuration)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode().Perm() != 0o600 {
				t.Fatalf("configuration permissions = %v", info.Mode().Perm())
			}
			if len(executor.processes) != test.processesWant {
				t.Fatalf("processes = %#v", executor.processes)
			}
			if _, err := os.Stat(configuration + ".bak"); !os.IsNotExist(err) {
				t.Fatalf("clear backup error = %v", err)
			}
		})
	}
}

func TestExpandHindsightRepositoryExpandsHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	path, err := expandHindsightRepository("~/repository")
	if err != nil || path != filepath.Join(home, "repository") {
		t.Fatalf("expandHindsightRepository() = %q, %v", path, err)
	}
}

func makeHindsightRepository(t *testing.T, home, name string) string {
	t.Helper()
	repository := filepath.Join(home, name)
	if err := os.MkdirAll(repository, 0o700); err != nil {
		t.Fatal(err)
	}
	return repository
}

func writeHindsightConfiguration(t *testing.T, home, apiURL, token string, registrations []hindsightRegistration) string {
	t.Helper()
	path := filepath.Join(home, "hindsight.json")
	content, err := json.Marshal(hindsightConfiguration{APIURL: apiURL, APIToken: token, Registrations: registrations})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func readHindsightConfiguration(t *testing.T, path string) hindsightConfiguration {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	configuration, _, err := parseHindsightConfigurationContent(content)
	if err != nil {
		t.Fatal(err)
	}
	return configuration
}

func canonicalHindsightDirectory(t *testing.T, directory string) string {
	t.Helper()
	canonical, err := filepath.EvalSymlinks(directory)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(content, &document); err != nil {
		t.Fatal(err)
	}
	return document
}

package commands

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRegisterHindsightSynchronizesCursorFromNativeConfiguration(t *testing.T) {
	home := t.TempDir()
	first := makeHindsightRepository(t, home, "first")
	second := makeHindsightRepository(t, home, "second")
	configuration := writeNativeHindsightConfiguration(t, home, map[string]string{first: "personal", second: "personal"})
	originalConfiguration, err := os.ReadFile(configuration)
	if err != nil {
		t.Fatal(err)
	}
	cursorPath := filepath.Join(home, ".cursor", "mcp.json")
	writeJSON(t, cursorPath, map[string]any{"mcpServers": map[string]any{
		"other": map[string]any{"url": "https://other.invalid"}, "hindsight-memory-obsolete": map[string]any{"url": "https://obsolete.invalid"},
	}})
	executor := &fakeExecutor{}
	runtime, _, _ := testRuntime(executor, map[string]string{"HOME": home})
	if code := RegisterHindsight(runtime, []string{"--config", configuration}); code != 0 {
		t.Fatalf("RegisterHindsight() = %d", code)
	}
	servers := readJSON(t, cursorPath)["mcpServers"].(map[string]any)
	if len(servers) != 2 || servers["other"] == nil || servers["hindsight-memory-personal"].(map[string]any)["url"] != "https://memory.invalid/mcp/personal/" {
		t.Fatalf("servers = %#v", servers)
	}
	if _, exists := servers["hindsight-memory-obsolete"]; exists {
		t.Fatalf("obsolete server remains: %#v", servers)
	}
	if content, err := os.ReadFile(configuration); err != nil || string(content) != string(originalConfiguration) {
		t.Fatalf("native configuration changed: %q, %v", content, err)
	}
	if len(executor.processes) < 1 || executor.processes[0].Name != "bunx" {
		t.Fatalf("processes = %#v", executor.processes)
	}
	info, err := os.Stat(cursorPath)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("permissions %s = %v, %v", cursorPath, info.Mode().Perm(), err)
	}
}

func TestRegisterHindsightRejectsMissingCredentialsWithoutWriting(t *testing.T) {
	home := t.TempDir()
	configuration := writeHindsightConfiguration(t, home, "https://memory.invalid", "", map[string]string{})
	runtime, _, _ := testRuntime(&fakeExecutor{}, map[string]string{"HOME": home})
	if code := RegisterHindsight(runtime, []string{"--config", configuration}); code != ExitUsage {
		t.Fatalf("RegisterHindsight() = %d", code)
	}
	if _, err := os.Stat(filepath.Join(home, ".cursor", "mcp.json")); !os.IsNotExist(err) {
		t.Fatalf("cursor configuration created: %v", err)
	}
}

func TestHindsightBankAddsAndRemovesDirectoryFromNativeConfiguration(t *testing.T) {
	home := t.TempDir()
	directory := makeHindsightRepository(t, home, "directory")
	configuration := writeNativeHindsightConfiguration(t, home, map[string]string{})
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
	if updated.MapPathToBank[canonicalHindsightDirectory(t, directory)] != "bank" || len(updated.MapPathToBank) != 1 {
		t.Fatalf("mapPathToBank after add = %#v", updated.MapPathToBank)
	}
	info, err := os.Stat(configuration)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("configuration permissions = %v, %v", info, err)
	}
	assertArgs(t, executor.processes, [][]string{{"add", "--encrypt", configuration}, {"apply", "--force"}})
	executor = &fakeExecutor{}
	runtime, _, _ = testRuntime(executor, map[string]string{"HOME": home})
	if code := HindsightBank(runtime, []string{"bank", "remove", directory}); code != 0 {
		t.Fatalf("HindsightBank(remove) = %d", code)
	}
	if paths := readHindsightConfiguration(t, configuration).MapPathToBank; len(paths) != 0 {
		t.Fatalf("mapPathToBank after remove = %#v", paths)
	}
	assertArgs(t, executor.processes, [][]string{{"add", "--encrypt", configuration}, {"apply", "--force"}})
}

func TestHindsightBankReplacesTildeMappedDirectory(t *testing.T) {
	home := t.TempDir()
	userHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(userHome, ".hindsight")
	configuration := writeNativeHindsightConfiguration(t, home, map[string]string{"~/.hindsight": "old"})
	runtime, _, _ := testRuntime(&fakeExecutor{}, map[string]string{"HOME": home})
	if code := HindsightBank(runtime, []string{"bank", "add", directory, "new"}); code != 0 {
		t.Fatalf("HindsightBank(add) = %d", code)
	}
	paths := readHindsightConfiguration(t, configuration).MapPathToBank
	canonical := canonicalHindsightDirectory(t, directory)
	if len(paths) != 1 || paths[canonical] != "new" {
		t.Fatalf("mapPathToBank after add = %#v", paths)
	}
	if code := HindsightBank(runtime, []string{"bank", "remove", directory}); code != 0 {
		t.Fatalf("HindsightBank(remove) = %d", code)
	}
	if paths := readHindsightConfiguration(t, configuration).MapPathToBank; len(paths) != 0 {
		t.Fatalf("mapPathToBank after remove = %#v", paths)
	}
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

func TestHindsightBankRejectsBlankRemoteBankWithoutRunningCLI(t *testing.T) {
	executor := &fakeExecutor{}
	runtime, _, _ := testRuntime(executor, map[string]string{})
	if code := HindsightBank(runtime, []string{"bank", "create", "   "}); code != ExitUsage {
		t.Fatalf("HindsightBank(blank create) = %d", code)
	}
	if len(executor.processes) != 0 {
		t.Fatalf("processes = %#v", executor.processes)
	}
}

func TestHindsightBankPropagatesRemoteBankCreationFailure(t *testing.T) {
	executor := &fakeExecutor{errors: []error{errors.New("remote failure")}}
	runtime, _, _ := testRuntime(executor, map[string]string{})
	if code := HindsightBank(runtime, []string{"bank", "create", "bank"}); code == 0 {
		t.Fatal("HindsightBank(create) succeeded")
	}
	assertArgs(t, executor.processes, [][]string{{"bank", "create", "bank"}})
}

func TestHindsightBankRejectsBlankBankWithoutWriting(t *testing.T) {
	home := t.TempDir()
	directory := makeHindsightRepository(t, home, "directory")
	configuration := writeNativeHindsightConfiguration(t, home, map[string]string{})
	original, err := os.ReadFile(configuration)
	if err != nil {
		t.Fatal(err)
	}
	executor := &fakeExecutor{}
	runtime, _, _ := testRuntime(executor, map[string]string{"HOME": home})
	if code := HindsightBank(runtime, []string{"bank", "add", directory, "   "}); code != ExitUsage {
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
			configuration := writeNativeHindsightConfiguration(t, home, map[string]string{})
			original, err := os.ReadFile(configuration)
			if err != nil {
				t.Fatal(err)
			}
			executor := &fakeExecutor{errors: test.errors}
			runtime, _, _ := testRuntime(executor, map[string]string{"HOME": home})
			if code := HindsightBank(runtime, []string{"bank", "add", directory, "bank"}); code == 0 {
				t.Fatal("HindsightBank(add) succeeded")
			}
			content, err := os.ReadFile(configuration)
			if err != nil {
				t.Fatal(err)
			}
			if test.restored && string(content) != string(original) {
				t.Fatalf("configuration after rollback = %q", content)
			}
			if !test.restored && len(readHindsightConfiguration(t, configuration).MapPathToBank) != 1 {
				t.Fatalf("configuration after apply failure = %q", content)
			}
			info, err := os.Stat(configuration)
			if err != nil || info.Mode().Perm() != 0o600 {
				t.Fatalf("configuration permissions = %v, %v", info, err)
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
func writeHindsightConfiguration(t *testing.T, home, apiURL, token string, paths map[string]string) string {
	t.Helper()
	path := filepath.Join(home, "hindsight.json")
	writeHindsightConfigurationAt(t, path, apiURL, token, paths)
	return path
}
func writeNativeHindsightConfiguration(t *testing.T, home string, paths map[string]string) string {
	t.Helper()
	path := filepath.Join(home, ".hindsight", "coding-agent.json")
	writeHindsightConfigurationAt(t, path, "https://memory.invalid", "secret", paths)
	return path
}
func writeHindsightConfigurationAt(t *testing.T, path, apiURL, token string, paths map[string]string) {
	t.Helper()
	content, err := json.Marshal(hindsightConfiguration{APIURL: apiURL, APIToken: token, MapPathToBank: paths})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
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
func writeJSON(t *testing.T, path string, document map[string]any) {
	t.Helper()
	content, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
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

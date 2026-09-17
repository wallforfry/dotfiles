package commands

import (
	"encoding/json"
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
}

func TestRegisterHindsightRejectsInvalidConfigurationWithoutWriting(t *testing.T) {
	home := t.TempDir()
	configuration := writeHindsightConfiguration(t, home, "https://memory.invalid", "secret", nil)
	runtime, _, _ := testRuntime(&fakeExecutor{}, map[string]string{"HOME": home})
	if code := RegisterHindsight(runtime, []string{"--config", configuration}); code != ExitUsage {
		t.Fatalf("RegisterHindsight() = %d", code)
	}
	if _, err := os.Stat(filepath.Join(home, ".hindsight", "coding-agent.json")); !os.IsNotExist(err) {
		t.Fatalf("configuration created: %v", err)
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

package commands

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type fakeExecutor struct {
	processes []Process
	errors    []error
}

func (executor *fakeExecutor) Run(process Process) error {
	executor.processes = append(executor.processes, process)
	index := len(executor.processes) - 1
	if index < len(executor.errors) {
		return executor.errors[index]
	}
	return nil
}

func testRuntime(executor Executor, environment map[string]string) (Runtime, *bytes.Buffer, *bytes.Buffer) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	entries := make([]string, 0, len(environment))
	for key, value := range environment {
		entries = append(entries, key+"="+value)
	}
	return Runtime{
		Executor: executor,
		LookupEnv: func(key string) (string, bool) {
			value, present := environment[key]
			return value, present
		},
		Environ: func() []string { return entries },
		Stdin:   strings.NewReader(""),
		Stdout:  stdout,
		Stderr:  stderr,
	}, stdout, stderr
}

func TestCloakOptionsAndContainerRace(t *testing.T) {
	executor := &fakeExecutor{errors: []error{nil, errors.New("absent"), errors.New("race"), nil}}
	runtime, stdout, _ := testRuntime(executor, map[string]string{})
	if code := Cloak(runtime, []string{"--start"}); code != 0 {
		t.Fatalf("Cloak() = %d, want 0", code)
	}
	want := [][]string{{"info"}, {"start", "cloak"}, {"run", "--detach", "--name", "cloak", "--publish", "127.0.0.1:9222:9222", "cloakhq/cloakbrowser:0.5.3", "cloakserve", "--idle-timeout=300"}, {"start", "cloak"}}
	assertArgs(t, executor.processes, want)
	if !strings.Contains(stdout.String(), "http://host.docker.internal:9222") {
		t.Fatalf("stdout = %q", stdout.String())
	}

	noDocker := &fakeExecutor{errors: []error{errors.New("down")}}
	runtime, _, stderr := testRuntime(noDocker, map[string]string{})
	if code := Cloak(runtime, []string{"--status"}); code != ExitUnavailable || stderr.String() != "cloak: daemon docker indisponible\n" {
		t.Fatalf("code = %d, stderr = %q", code, stderr.String())
	}
	if code := Cloak(runtime, nil); code != ExitUsage {
		t.Fatalf("usage code = %d", code)
	}
}

func TestCommandsUseSysexitsForUsageAndUnavailableDocker(t *testing.T) {
	compose := filepath.Join(t.TempDir(), "compose.yml")
	if err := os.WriteFile(compose, []byte("services: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	usageCases := []struct {
		name        string
		command     func(Runtime, []string) int
		environment map[string]string
	}{
		{name: "cloak", command: Cloak, environment: map[string]string{}},
		{name: "firecrawl", command: Firecrawl, environment: map[string]string{"FIRECRAWL_COMPOSE": compose}},
		{name: "postgres", command: Postgres, environment: map[string]string{}},
		{name: "scrapling", command: Scrapling, environment: map[string]string{}},
	}
	for _, test := range usageCases {
		t.Run(test.name+" usage", func(t *testing.T) {
			runtime, _, _ := testRuntime(&fakeExecutor{}, test.environment)
			if code := test.command(runtime, []string{"--unknown"}); code != ExitUsage {
				t.Fatalf("code = %d, want %d", code, ExitUsage)
			}
		})
		t.Run(test.name+" unavailable", func(t *testing.T) {
			runtime, _, _ := testRuntime(&fakeExecutor{errors: []error{errors.New("down")}}, test.environment)
			args := []string{"--status"}
			if test.name == "cloak" {
				args = []string{"--start"}
			}
			if code := test.command(runtime, args); code != ExitUnavailable {
				t.Fatalf("code = %d, want %d", code, ExitUnavailable)
			}
		})
	}
}

func TestReplacingCommandPreservesShellResolutionExitCodes(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	missing := exitCode((OSExecutor{}).Run(Process{Name: "absent-command", Replace: true}))
	if missing != 127 {
		t.Fatalf("missing command code = %d", missing)
	}

	directory := t.TempDir()
	t.Setenv("PATH", directory)
	blocked := filepath.Join(directory, "blocked-command")
	if err := os.WriteFile(blocked, []byte("#!/bin/sh\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	notExecutable := exitCode((OSExecutor{}).Run(Process{Name: "blocked-command", Replace: true}))
	if notExecutable != 126 {
		t.Fatalf("non-executable command code = %d", notExecutable)
	}
}

func TestPostgresKeepsSecretOutOfArguments(t *testing.T) {
	const uri = "postgres://user:secret@host.docker.internal:5432/db"
	executor := &fakeExecutor{errors: []error{nil, errors.New("absent"), nil, nil}}
	runtime, _, _ := testRuntime(executor, map[string]string{"DATABASE_URI": uri})
	if code := Postgres(runtime, nil); code != 0 {
		t.Fatalf("Postgres() = %d", code)
	}
	container := postgresContainer(uri)
	want := [][]string{
		{"info"},
		{"start", container},
		{"run", "--detach", "--name", container, "--add-host=host.docker.internal:host-gateway", "--env", "DATABASE_URI", "--entrypoint", "sleep", "crystaldba/postgres-mcp", "infinity"},
		{"exec", "--interactive", container, "postgres-mcp", "--access-mode=restricted"},
	}
	assertArgs(t, executor.processes, want)
	for _, process := range executor.processes {
		if strings.Contains(strings.Join(process.Args, " "), uri) {
			t.Fatalf("DATABASE_URI leaked in arguments: %q", process.Args)
		}
	}
	if !contains(executor.processes[2].Env, "DATABASE_URI="+uri) {
		t.Fatalf("DATABASE_URI missing from docker environment: %q", executor.processes[2].Env)
	}
	if !executor.processes[3].Replace {
		t.Fatal("postgres-mcp must replace the wrapper process")
	}
	if got := postgresContainer(uri); got != "postgres-mcp-577f7f27" {
		t.Fatalf("postgresContainer() = %q", got)
	}
}

func TestPostgresValidationAndStatus(t *testing.T) {
	executor := &fakeExecutor{}
	runtime, _, stderr := testRuntime(executor, map[string]string{})
	if code := Postgres(runtime, nil); code != ExitUsage {
		t.Fatalf("missing URI code = %d", code)
	}
	if !strings.Contains(stderr.String(), "DATABASE_URI absente") {
		t.Fatalf("stderr = %q", stderr.String())
	}

	executor = &fakeExecutor{}
	runtime, _, _ = testRuntime(executor, map[string]string{})
	if code := Postgres(runtime, []string{"--status"}); code != 0 {
		t.Fatalf("status code = %d", code)
	}
	assertArgs(t, executor.processes, [][]string{{"info"}, {"ps", "--all", "--filter", "name=^postgres-mcp-", "--format", "{{.Names}}\t{{.Image}}\t{{.Status}}"}})
}

func TestScraplingRetriesStartAfterCreationRace(t *testing.T) {
	executor := &fakeExecutor{errors: []error{nil, errors.New("absent"), errors.New("race"), nil, nil}}
	runtime, _, _ := testRuntime(executor, map[string]string{})
	if code := Scrapling(runtime, nil); code != 0 {
		t.Fatalf("Scrapling() = %d", code)
	}
	want := [][]string{
		{"info"},
		{"start", "scrapling-mcp"},
		{"run", "--detach", "--name", "scrapling-mcp", "--add-host=host.docker.internal:host-gateway", "--volume", "scrapling-profiles:/profiles", "--entrypoint", "sleep", "pyd4vinci/scrapling:latest", "infinity"},
		{"start", "scrapling-mcp"},
		{"exec", "--interactive", "scrapling-mcp", "uv", "run", "scrapling", "mcp"},
	}
	assertArgs(t, executor.processes, want)
	if !executor.processes[4].Replace {
		t.Fatal("scrapling-mcp must replace the wrapper process")
	}
}

func TestFirecrawlStartsOnlyWhenHealthCheckFails(t *testing.T) {
	compose := filepath.Join(t.TempDir(), "compose.yml")
	if err := os.WriteFile(compose, []byte("services: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	executor := &fakeExecutor{errors: []error{nil, errors.New("unhealthy"), nil, nil}}
	runtime, _, _ := testRuntime(executor, map[string]string{"FIRECRAWL_COMPOSE": compose})
	if code := Firecrawl(runtime, nil); code != 0 {
		t.Fatalf("Firecrawl() = %d", code)
	}
	assertArgs(t, executor.processes, [][]string{
		{"info"},
		{"-fsS", "--max-time", "2", "http://localhost:3002/v0/health/liveness"},
		{"compose", "--file", compose, "up", "--detach", "--wait", "--wait-timeout", "90"},
		{"--yes", "firecrawl-mcp"},
	})
	if process := executor.processes[3]; process.Name != "npx" || !contains(process.Env, "FIRECRAWL_API_URL=http://localhost:3002") {
		t.Fatalf("npx process = %#v", process)
	} else if !process.Replace {
		t.Fatal("firecrawl-mcp must replace the wrapper process")
	}
}

func TestFirecrawlMissingComposeIsUnavailable(t *testing.T) {
	executor := &fakeExecutor{}
	runtime, _, stderr := testRuntime(executor, map[string]string{"FIRECRAWL_COMPOSE": filepath.Join(t.TempDir(), "missing")})
	if code := Firecrawl(runtime, []string{"--start"}); code != ExitUnavailable {
		t.Fatalf("Firecrawl() = %d", code)
	}
	if !strings.Contains(stderr.String(), "composition absente") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func assertArgs(t *testing.T, processes []Process, want [][]string) {
	t.Helper()
	got := make([][]string, len(processes))
	for index, process := range processes {
		got[index] = process.Args
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args = %#v, want %#v", got, want)
	}
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

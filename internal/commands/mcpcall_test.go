package commands

import (
	"bufio"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// mcpServerExecutor answers the docker exec of an MCP server like a stdio
// server would, and records every message it received.
type mcpServerExecutor struct {
	fakeExecutor
	respond  bool
	result   string
	received []map[string]any
}

func (executor *mcpServerExecutor) Run(process Process) error {
	if err := executor.fakeExecutor.Run(process); err != nil || firstArg(process.Args) != "exec" || !executor.respond {
		return err
	}
	scanner := bufio.NewScanner(process.Stdin)
	for scanner.Scan() {
		var message map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &message); err != nil {
			return err
		}
		executor.received = append(executor.received, message)
		switch message["method"] {
		case "initialize":
			fprintf(process.Stdout, `{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-06-18","capabilities":{"tools":{}}}}`+"\n")
		case "tools/call":
			fprintf(process.Stdout, `{"jsonrpc":"2.0","method":"notifications/message","params":{"level":"info"}}`+"\n")
			fprintf(process.Stdout, `{"jsonrpc":"2.0","id":2,"method":"roots/list"}`+"\n")
			fprintf(process.Stdout, `{"jsonrpc":"2.0","id":2,"result":%s}`+"\n", executor.result)
		}
	}
	return scanner.Err()
}

func TestScraplingCallsOneToolThenClosesTheServer(t *testing.T) {
	executor := &mcpServerExecutor{respond: true, result: `{"content":[{"type":"text","text":"# Example\n"}]}`}
	runtime, stdout, _ := testRuntime(executor, map[string]string{})
	if code := Scrapling(runtime, []string{"fetch", `{"url":"https://example.com"}`}); code != 0 {
		t.Fatalf("Scrapling() = %d", code)
	}
	assertArgs(t, executor.processes, [][]string{
		{"info"},
		{"start", "scrapling-mcp"},
		{"exec", "--interactive", "scrapling-mcp", "uv", "run", "scrapling", "mcp"},
	})
	methods := make([]any, len(executor.received))
	for index, message := range executor.received {
		methods[index] = message["method"]
	}
	if want := []any{"initialize", "notifications/initialized", "tools/call"}; !reflect.DeepEqual(methods, want) {
		t.Fatalf("methods = %v, want %v", methods, want)
	}
	params := executor.received[2]["params"].(map[string]any)
	if params["name"] != "fetch" || !reflect.DeepEqual(params["arguments"], map[string]any{"url": "https://example.com"}) {
		t.Fatalf("tools/call params = %v", params)
	}
	if stdout.String() != "# Example\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestScraplingReportsToolErrorsOnStderr(t *testing.T) {
	executor := &mcpServerExecutor{respond: true, result: `{"content":[{"type":"text","text":"timeout"}],"isError":true}`}
	runtime, stdout, stderr := testRuntime(executor, map[string]string{})
	if code := Scrapling(runtime, []string{"get"}); code != 1 {
		t.Fatalf("Scrapling() = %d, want 1", code)
	}
	if stdout.String() != "" || stderr.String() != "timeout\n" {
		t.Fatalf("stdout = %q, stderr = %q", stdout.String(), stderr.String())
	}
}

func TestScraplingSavesImagesToFiles(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	executor := &mcpServerExecutor{respond: true, result: `{"content":[{"type":"image","data":"aW1hZ2U=","mimeType":"image/png"}]}`}
	runtime, stdout, _ := testRuntime(executor, map[string]string{})
	if code := Scrapling(runtime, []string{"screenshot", `{"url":"https://example.com"}`}); code != 0 {
		t.Fatalf("Scrapling() = %d", code)
	}
	path := strings.TrimSpace(stdout.String())
	if !strings.HasSuffix(path, ".png") {
		t.Fatalf("path = %q", path)
	}
	if content, err := os.ReadFile(path); err != nil || string(content) != "image" {
		t.Fatalf("content = %q, err = %v", content, err)
	}
}

func TestScraplingFailsWhenTheServerDiesWithoutAnswering(t *testing.T) {
	executor := &mcpServerExecutor{}
	runtime, _, stderr := testRuntime(executor, map[string]string{})
	if code := Scrapling(runtime, []string{"get"}); code != 1 {
		t.Fatalf("Scrapling() = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "sans répondre") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestScraplingRejectsArgumentsThatAreNotAnObject(t *testing.T) {
	for _, args := range [][]string{nil, {"--start"}, {"fetch", "[]"}, {"fetch", "null"}, {"fetch", "{}", "extra"}} {
		executor := &mcpServerExecutor{respond: true}
		runtime, _, _ := testRuntime(executor, map[string]string{})
		if code := Scrapling(runtime, args); code != ExitUsage {
			t.Fatalf("Scrapling(%q) = %d, want %d", args, code, ExitUsage)
		}
		assertArgs(t, executor.processes, [][]string{{"info"}})
	}
}

func TestCallToolReturnsWhenARealServerDiesWithoutAnswering(t *testing.T) {
	for _, script := range []string{"exit 3", "read line; exit 0"} {
		runtime, _, stderr := testRuntime(OSExecutor{}, map[string]string{})
		done := make(chan int, 1)
		go func() {
			done <- callTool(runtime, "test", runtime.process("sh", "-c", script), "get", map[string]any{})
		}()
		select {
		case code := <-done:
			if code == 0 || !strings.Contains(stderr.String(), "sans répondre") {
				t.Fatalf("%q: code = %d, stderr = %q", script, code, stderr.String())
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("%q: callTool still waiting after the server exited", script)
		}
	}
}

func TestCallToolTreatsAnErrorWithoutIDAsFinal(t *testing.T) {
	runtime, _, stderr := testRuntime(OSExecutor{}, map[string]string{})
	server := runtime.process("sh", "-c", `read line; echo '{"jsonrpc":"2.0","id":null,"error":{"message":"parse error"}}'; cat >/dev/null`)
	done := make(chan int, 1)
	go func() { done <- callTool(runtime, "test", server, "get", map[string]any{}) }()
	select {
	case code := <-done:
		if code != 1 || stderr.String() != "test: parse error\n" {
			t.Fatalf("code = %d, stderr = %q", code, stderr.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("callTool still waiting after an error without id")
	}
}

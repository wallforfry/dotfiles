package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAgentHandoffBlocksOnceAtDerivedThreshold(t *testing.T) {
	temporary := t.TempDir()
	transcript := filepath.Join(temporary, "transcript.jsonl")
	lines := make([]string, 501)
	for index := range lines {
		lines[index] = `{ "type": "user" }`
	}
	lines[0] = usageLine(90_000)
	lines[500] = usageLine(85_000)
	if err := os.WriteFile(transcript, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	input := fmt.Sprintf(`{"transcript_path":%q,"session_id":"session","stop_hook_active":false}`, transcript)
	runtime, stdout, _ := testRuntime(&fakeExecutor{}, map[string]string{
		"XDG_STATE_HOME":                  temporary,
		"CLAUDE_CODE_AUTO_COMPACT_WINDOW": "100000",
	})
	runtime.Stdin = strings.NewReader(input)
	if code := AgentHandoff(runtime, nil); code != 0 {
		t.Fatalf("AgentHandoff() = %d", code)
	}
	if !strings.Contains(stdout.String(), `"decision":"block"`) || !strings.Contains(stdout.String(), "85k handoff threshold") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(temporary, "claude", "handoff", "session")); err != nil {
		t.Fatalf("sentinel: %v", err)
	}

	runtime.Stdin = strings.NewReader(input)
	runtime.Stdout = &bytes.Buffer{}
	if code := AgentHandoff(runtime, nil); code != 0 || runtime.Stdout.(*bytes.Buffer).Len() != 0 {
		t.Fatalf("second invocation code = %d, stdout = %q", code, runtime.Stdout)
	}
}

func TestAgentHandoffUsesExplicitThresholdAndIgnoresSidechains(t *testing.T) {
	temporary := t.TempDir()
	transcript := filepath.Join(temporary, "transcript.jsonl")
	contents := usageLine(120_000) + "\n" + `{"type":"assistant","isSidechain":true,"message":{"usage":{"input_tokens":999999}}}` + "\n"
	if err := os.WriteFile(transcript, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	input := fmt.Sprintf(`{"transcript_path":%q,"session_id":"explicit"}`, transcript)
	runtime, stdout, _ := testRuntime(&fakeExecutor{}, map[string]string{
		"XDG_STATE_HOME":          temporary,
		"HANDOFF_TOKEN_THRESHOLD": "120000",
	})
	runtime.Stdin = strings.NewReader(input)
	if code := AgentHandoff(runtime, nil); code != 0 || !strings.Contains(stdout.String(), "120k tokens") {
		t.Fatalf("code = %d, stdout = %q", code, stdout.String())
	}
}

func TestAgentHandoffFailsOpenOnInvalidInput(t *testing.T) {
	runtime, stdout, _ := testRuntime(&fakeExecutor{}, map[string]string{"HANDOFF_TOKEN_THRESHOLD": "invalid"})
	runtime.Stdin = strings.NewReader("not-json")
	if code := AgentHandoff(runtime, nil); code != 0 || stdout.Len() != 0 {
		t.Fatalf("code = %d, stdout = %q", code, stdout.String())
	}
}

func TestReadLastLinesDoesNotKeepOlderEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "transcript.jsonl")
	lines := make([]string, 700)
	for index := range lines {
		lines[index] = fmt.Sprintf("line-%03d", index)
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	contents, err := readLastLines(path, 500)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(contents, []byte("line-199")) || !bytes.Contains(contents, []byte("line-200")) || !bytes.Contains(contents, []byte("line-699")) {
		t.Fatalf("unexpected tail boundaries")
	}
}

func usageLine(inputTokens int) string {
	return fmt.Sprintf(`{"type":"assistant","message":{"usage":{"input_tokens":%d,"cache_read_input_tokens":0,"cache_creation_input_tokens":0}}}`, inputTokens)
}

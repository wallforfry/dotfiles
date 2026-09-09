package telemetry

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestClaudeNormalizesTextActivationAndAllEditors(t *testing.T) {
	records := []map[string]any{
		{
			"type": "assistant", "timestamp": "2026-08-18T00:00:00Z",
			"message": map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "text", "text": "a—b·c"}}},
		},
		{
			"type": "assistant", "timestamp": "2026-08-18T00:00:01Z",
			"message": map[string]any{"role": "assistant", "content": []any{
				map[string]any{"type": "tool_use", "name": "Skill", "input": map[string]any{"skill": nil}},
				map[string]any{"type": "tool_use", "name": "Agent", "input": map[string]any{}},
				map[string]any{"type": "tool_use", "name": "Write", "input": map[string]any{"file_path": "a.py", "content": "# x\ny = 1"}},
				map[string]any{"type": "tool_use", "name": "Edit", "input": map[string]any{"file_path": "a.ts", "new_string": "// x\ny()"}},
				map[string]any{"type": "tool_use", "name": "MultiEdit", "input": map[string]any{"file_path": "a.js", "edits": []any{map[string]any{"new_string": "/* x */\ny()"}}}},
				map[string]any{"type": "tool_use", "name": "NotebookEdit", "input": map[string]any{"notebook_path": "a.ipynb", "new_source": "# x\ny = 1"}},
				map[string]any{"type": "tool_use", "name": "Bash", "input": map[string]any{"command": "git commit -m test"}},
			}},
		},
	}
	summary := summarizeRecords(t, "claude", records)
	assertCount(t, summary.Dash, "depuis", 1)
	assertCount(t, summary.MiddleDot, "depuis", 1)
	assertCount(t, summary.Skills, None, 1)
	assertCount(t, summary.Agents, "agent", 1)
	assertCount(t, summary.Lines, "depuis", 8)
	assertCount(t, summary.Comments, "depuis", 4)
	assertCount(t, summary.UninspectableWrites, "bash", 1)
}

func TestCodexNormalizesMessagesWithoutCompletedProjection(t *testing.T) {
	records := []map[string]any{
		{"type": "response_item", "timestamp": "2026-08-16T00:00:00Z", "payload": map[string]any{
			"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "output_text", "text": "a—b·c"}},
		}},
		{"type": "response_item", "timestamp": "2026-08-18T00:00:02Z", "payload": map[string]any{
			"type": "item_completed", "item": map[string]any{"type": "AgentMessage", "content": []any{map[string]any{"type": "Text", "text": "a—b·c"}}},
		}},
	}
	summary := summarizeRecords(t, "codex", records)
	assertCount(t, summary.Blocks, "avant", 1)
	assertCount(t, summary.Dash, "avant", 1)
	assertCount(t, summary.MiddleDot, "avant", 1)
}

func TestCodexNormalizesCompletedFileAndCommandTools(t *testing.T) {
	records := []map[string]any{
		{"type": "response_item", "timestamp": "2026-08-18T00:00:03Z", "payload": map[string]any{
			"type": "item_completed", "item": map[string]any{"type": "FileChange", "changes": map[string]any{
				"safe.py": map[string]any{"type": "update", "unified_diff": "@@\n+# x\n+y = 1\n"},
			}},
		}},
		{"type": "response_item", "payload": map[string]any{
			"type": "item_completed", "item": map[string]any{"type": "CommandExecution", "command": []any{"git", "commit", "-m", "test"}},
		}},
	}
	summary := summarizeRecords(t, "codex", records)
	assertCount(t, summary.Lines, "depuis", 2)
	assertCount(t, summary.Comments, "depuis", 1)
	assertCount(t, summary.WriteTools, "filechange", 1)
	assertCount(t, summary.UninspectableWrites, "commandexecution", 1)
}

func TestCodexNormalizesDirectAndNestedToolsWithoutCompletedDuplicates(t *testing.T) {
	records := []map[string]any{
		{"type": "response_item", "timestamp": "2026-08-18T00:00:00Z", "payload": map[string]any{
			"type": "function_call", "name": "spawn_agent", "arguments": `{"task_name":"reader"}`,
		}},
		{"type": "response_item", "payload": map[string]any{
			"type": "custom_tool_call", "name": "exec", "input": "await tools.apply_patch(`*** Begin Patch\\n*** Update File: a.py\\n+# x\\n+y = 1\\n*** End Patch`)",
		}},
		{"type": "response_item", "payload": map[string]any{
			"type": "custom_tool_call", "name": "exec", "input": "await tools.write_stdin({session_id: 1, chars: 'x'})",
		}},
		{"type": "response_item", "payload": map[string]any{
			"type": "item_completed", "item": map[string]any{"type": "SubAgentActivity", "kind": "started"},
		}},
		{"type": "response_item", "payload": map[string]any{
			"type": "item_completed", "item": map[string]any{"type": "CollabAgentToolCall", "tool": "spawn_agent"},
		}},
	}
	summary := summarizeRecords(t, "codex", records)
	assertCount(t, summary.Agents, "spawn_agent", 1)
	assertCount(t, summary.Lines, "inconnu", 2)
	assertCount(t, summary.Comments, "inconnu", 1)
	assertCount(t, summary.WriteTools, "apply_patch", 1)
	assertCount(t, summary.UninspectableWrites, "write_stdin", 1)
	encoded, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "reader") {
		t.Fatal("summary leaked the subagent task name")
	}
}

func TestZeroDenominatorIsUnknown(t *testing.T) {
	if got := Rate(0, 0); got != "inconnu" {
		t.Fatalf("Rate(0, 0) = %q, want inconnu", got)
	}
}

func TestCacheRejectsAnotherVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	content := `{"version":13,"since":"x","files":{}}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := LoadCache(path, "x"); len(got) != 0 {
		t.Fatalf("LoadCache returned %d entries, want 0", len(got))
	}
}

func TestCacheRejectsAnotherPeriod(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cache.json")
	content := `{"version":14,"since":"x","files":{}}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if got := LoadCache(path, "y"); len(got) != 0 {
		t.Fatalf("LoadCache returned %d entries, want 0", len(got))
	}
}

func TestCacheContainsOnlyHashesAndAggregates(t *testing.T) {
	directory := t.TempDir()
	transcript := filepath.Join(directory, "private-session.jsonl")
	writeRecords(t, transcript, []map[string]any{{"type": "compacted", "secret": "do-not-cache"}})
	entries, _, _, err := CollectSource("codex", directory, "2026-08-17", nil)
	if err != nil {
		t.Fatal(err)
	}
	cachePath := filepath.Join(directory, "cache", "telemetry.json")
	if err := SaveCache(cachePath, "2026-08-17", entries); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"private-session", "do-not-cache", transcript} {
		if bytes.Contains(content, []byte(forbidden)) {
			t.Fatalf("cache contains private value %q", forbidden)
		}
	}
	info, err := os.Stat(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("cache mode = %o, want 600", got)
	}
}

func TestActivatedBodyCostWeightsLocalSkillsOnly(t *testing.T) {
	summary := EmptySummary("claude")
	summary.Skills = map[string]int{"adr": 2, "external": 9}
	activations, cost := ActivatedBodyCost(summary, map[string]int64{"adr": 100})
	if activations != 2 || cost != 200 {
		t.Fatalf("ActivatedBodyCost = (%d, %d), want (2, 200)", activations, cost)
	}
}

func TestUnknownAndInvalidRecordsAreNotSilentlyAccepted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.jsonl")
	if err := os.WriteFile(path, []byte("{\"type\":\"future\"}\nnot-json\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	summary, err := SummarizeFile("codex", path, "2026-08-17")
	if err != nil {
		t.Fatal(err)
	}
	assertCount(t, summary.UnknownRecords, "codex", 1)
	assertCount(t, summary.InvalidRecords, "codex", 1)
	assertCount(t, summary.ReadRecords, "codex", 2)
	if len(summary.RecognizedRecords) != 0 {
		t.Fatalf("recognized records = %v, want none", summary.RecognizedRecords)
	}
}

func TestMalformedKnownCodexRecordIsNotSilentlyAccepted(t *testing.T) {
	records := []map[string]any{
		{"type": "response_item", "payload": map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "future"}}}},
		{"type": "response_item", "payload": map[string]any{"type": "message", "role": "future", "content": []any{}}},
		{"type": "event_msg", "payload": map[string]any{"type": "item_completed", "item": map[string]any{"type": "CommandExecution", "command": "git commit"}}},
	}
	summary := summarizeRecords(t, "codex", records)
	assertCount(t, summary.UnknownRecords, "codex", 3)
	if len(summary.Blocks) != 0 || len(summary.WriteTools) != 0 {
		t.Fatal("malformed records produced events")
	}
}

func TestCodexRecognizesCompactionRecordsWithoutInventingEvents(t *testing.T) {
	records := []map[string]any{
		{"type": "compacted"},
		{"type": "event_msg", "payload": map[string]any{"type": "item_completed", "item": map[string]any{"type": "ContextCompaction"}}},
	}
	summary := summarizeRecords(t, "codex", records)
	assertCount(t, summary.RecognizedRecords, "codex", 2)
	if len(summary.UnknownRecords) != 0 || len(summary.Blocks) != 0 || len(summary.Agents) != 0 {
		t.Fatal("compaction records were rejected or produced events")
	}
}

func summarizeRecords(t *testing.T, source string, records []map[string]any) Summary {
	t.Helper()
	path := filepath.Join(t.TempDir(), "session.jsonl")
	writeRecords(t, path, records)
	summary, err := SummarizeFile(source, path, "2026-08-17")
	if err != nil {
		t.Fatal(err)
	}
	return summary
}

func writeRecords(t *testing.T, path string, records []map[string]any) {
	t.Helper()
	var content bytes.Buffer
	encoder := json.NewEncoder(&content)
	encoder.SetEscapeHTML(false)
	for _, record := range records {
		if err := encoder.Encode(record); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(path, content.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertCount(t *testing.T, values map[string]int, key string, want int) {
	t.Helper()
	if got := values[key]; got != want {
		t.Fatalf("count[%q] = %d, want %d", key, got, want)
	}
}

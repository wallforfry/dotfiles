package telemetry

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestRealtimeFixtureIsFullyRecognized(t *testing.T) {
	summary, err := SummarizeFile("codex", filepath.Join("testdata", "codex-realtime.jsonl"), "2026-08-17")
	if err != nil {
		t.Fatal(err)
	}
	assertCount(t, summary.ReadRecords, "codex", 4)
	assertCount(t, summary.RecognizedRecords, "codex", 4)
	assertCount(t, summary.Blocks, "depuis", 1)
	assertCount(t, summary.Dash, "depuis", 1)
	assertCount(t, summary.MiddleDot, "depuis", 1)
	if len(summary.UnknownRecords) != 0 || len(summary.InvalidRecords) != 0 || len(summary.Agents) != 0 {
		t.Fatalf("unexpected counters: unknown=%v invalid=%v agents=%v", summary.UnknownRecords, summary.InvalidRecords, summary.Agents)
	}
}

func TestMalformedRealtimeRecordsAreUnknown(t *testing.T) {
	records := []map[string]any{
		{"type": "realtime_item", "payload": map[string]any{"type": "transcript_segment", "id": "event-1", "realtime_session_id": "session-1", "role": "speaker", "text": "future role"}},
		{"type": "realtime_item", "payload": map[string]any{"type": "future", "id": "event-2", "realtime_session_id": "session-1"}},
		{"type": "event_msg", "payload": map[string]any{"type": "transcript_segment", "id": "event-3", "realtime_session_id": "session-1", "role": "assistant", "text": "wrong envelope"}},
	}
	summary := summarizeRecords(t, "codex", records)
	assertCount(t, summary.ReadRecords, "codex", 3)
	assertCount(t, summary.UnknownRecords, "codex", 3)
	if len(summary.RecognizedRecords) != 0 {
		t.Fatalf("recognized records = %v, want none", summary.RecognizedRecords)
	}
}

func TestTurnAbortedRequiresCompleteShapeAndProducesNoEvent(t *testing.T) {
	records := []map[string]any{
		{"type": "event_msg", "payload": map[string]any{"type": "turn_aborted", "turn_id": "turn-1", "reason": "interrupted", "started_at": 1000, "completed_at": 2000, "duration_ms": 1000}},
		{"type": "event_msg", "payload": map[string]any{"type": "turn_aborted", "turn_id": "turn-2", "started_at": 1000, "completed_at": 2000, "duration_ms": 1000}},
		{"type": "event_msg", "payload": map[string]any{"type": "turn_aborted", "turn_id": "turn-3", "reason": "interrupted", "started_at": 1.5, "completed_at": 2000, "duration_ms": 1000}},
	}
	summary := summarizeRecords(t, "codex", records)
	assertCount(t, summary.RecognizedRecords, "codex", 1)
	assertCount(t, summary.UnknownRecords, "codex", 2)
	if len(summary.Blocks) != 0 || len(summary.Agents) != 0 {
		t.Fatal("turn_aborted produced an event")
	}
}

func TestAgentLabelsDoNotUseTaskNameOrDescription(t *testing.T) {
	records := []map[string]any{
		{"type": "response_item", "payload": map[string]any{"type": "function_call", "name": "spawn_agent", "arguments": `{"task_name":"private-label","message":"private"}`}},
		{"type": "response_item", "payload": map[string]any{"type": "function_call", "name": "spawn_agent", "arguments": `{"agent_type":"reader","task_name":"private-label"}`}},
	}
	summary := summarizeRecords(t, "codex", records)
	assertCount(t, summary.Agents, "spawn_agent", 1)
	assertCount(t, summary.Agents, "reader", 1)
	if strings.Contains(mustJSON(t, summary), "private-label") {
		t.Fatal("summary leaked the task name")
	}

	claude := summarizeRecords(t, "claude", []map[string]any{{
		"type": "assistant", "message": map[string]any{"role": "assistant", "content": []any{
			map[string]any{"type": "tool_use", "name": "Agent", "input": map[string]any{"description": "private-description"}},
		}},
	}})
	assertCount(t, claude.Agents, "agent", 1)
	if strings.Contains(mustJSON(t, claude), "private-description") {
		t.Fatal("summary leaked the description")
	}
}

func TestTelemetryErrorsAreTypedAndSanitized(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "private-transcript.jsonl")
	_, err := SummarizeFile("codex", missing, "2026-08-17")
	var readError *ReadInterruptedError
	if !errors.As(err, &readError) {
		t.Fatalf("error %T, want ReadInterruptedError", err)
	}
	if strings.Contains(err.Error(), missing) {
		t.Fatal("read error exposed transcript path")
	}
	if err.Error() != "codex : lecture interrompue" {
		t.Fatalf("read error = %q", err.Error())
	}

	summary := EmptySummary("codex")
	summary.UnknownRecords["codex"] = 2
	summary.InvalidRecords["codex"] = 1
	var driftError *FormatDriftError
	drift := ValidateFormats(summary)
	if !errors.As(drift, &driftError) || driftError.Unknown != 2 || driftError.Invalid != 1 {
		t.Fatalf("unexpected format drift error: %#v", driftError)
	}
	if drift.Error() != "formats non mesurés : 2 inconnus, 1 invalides" || drift.Error() == err.Error() {
		t.Fatalf("format drift error = %q", drift.Error())
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	content, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

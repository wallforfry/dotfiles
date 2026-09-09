package telemetry

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

const (
	Unknown = "<inconnu>"
	None    = "<aucun>"
)

var Eras = []string{"avant", "depuis", "inconnu"}

var codeExtensions = []string{
	".c", ".cc", ".cpp", ".cs", ".css", ".cts", ".go", ".h", ".hpp", ".html", ".java",
	".js", ".jsx", ".kt", ".kts", ".lua", ".mjs", ".mts", ".php", ".py", ".rb", ".rs",
	".scss", ".sh", ".sql", ".svelte", ".swift", ".toml", ".ts", ".tsx", ".vue", ".yaml",
	".yml",
}

var explicitWriters = []string{"write", "edit", "multiedit", "notebookedit", "apply_patch", "filechange"}
var shellWriters = []string{"bash", "exec_command", "commandexecution"}
var nestedToolPattern = regexp.MustCompile(`\btools\.([A-Za-z][A-Za-z0-9_]*)\s*\(`)
var redirectPattern = regexp.MustCompile(`>{1,2}\s*([^&\s]+)`)
var shellWritePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(?:tee|touch|install|cp|mv)\s`),
	regexp.MustCompile(`(?i)\bsed\s+[^\n]*\s-i(?:\s|$)`),
	regexp.MustCompile(`(?i)\bgit\s+(?:commit|apply|am|cherry-pick|merge|rebase)\b`),
	regexp.MustCompile(`(?i)\.write_(?:text|bytes)\s*\(`),
	regexp.MustCompile(`(?i)\bopen\s*\([^\n]*["'][wax][bt+]?["']`),
}

type event struct {
	source    string
	kind      string
	timestamp any
	role      any
	text      any
	name      any
	payload   any
}

func normalizedEvents(source string, record map[string]any) []event {
	if source == "claude" {
		return claudeEvents(record)
	}
	return codexEvents(record)
}

func claudeEvents(record map[string]any) []event {
	message, ok := object(record["message"])
	if !ok {
		return nil
	}
	content, ok := message["content"].([]any)
	if !ok {
		return nil
	}
	events := make([]event, 0, len(content))
	for _, rawBlock := range content {
		block, ok := object(rawBlock)
		if !ok {
			continue
		}
		switch block["type"] {
		case "text":
			events = append(events, event{source: "claude", kind: "text", timestamp: record["timestamp"], role: message["role"], text: block["text"]})
		case "tool_use":
			events = append(events, event{source: "claude", kind: "tool", timestamp: record["timestamp"], name: block["name"], payload: block["input"]})
		}
	}
	return events
}

func codexEvents(record map[string]any) []event {
	payload, ok := object(record["payload"])
	if !ok {
		return nil
	}
	switch payload["type"] {
	case "item_completed":
		return codexItemEvents(record, payload)
	case "message":
		return codexMessageEvents(record, payload)
	case "transcript_segment":
		return []event{{source: "codex", kind: "text", timestamp: record["timestamp"], role: payload["role"], text: payload["text"]}}
	case "function_call":
		return []event{{source: "codex", kind: "tool", timestamp: record["timestamp"], name: payload["name"], payload: decodeJSON(payload["arguments"])}}
	case "custom_tool_call":
		input, ok := payload["input"].(string)
		if !ok {
			return []event{{source: "codex", kind: "tool", timestamp: record["timestamp"], name: payload["name"], payload: payload["input"]}}
		}
		matches := nestedToolPattern.FindAllStringSubmatch(input, -1)
		if len(matches) == 0 {
			return []event{{source: "codex", kind: "tool", timestamp: record["timestamp"], name: payload["name"], payload: input}}
		}
		events := make([]event, 0, len(matches))
		for _, match := range matches {
			events = append(events, event{source: "codex", kind: "tool", timestamp: record["timestamp"], name: match[1], payload: input})
		}
		return events
	default:
		return nil
	}
}

func codexItemEvents(record, payload map[string]any) []event {
	item, ok := object(payload["item"])
	if !ok {
		return nil
	}
	switch item["type"] {
	case "FileChange":
		return []event{{source: "codex", kind: "tool", timestamp: record["timestamp"], name: "filechange", payload: item}}
	case "CommandExecution":
		return []event{{source: "codex", kind: "tool", timestamp: record["timestamp"], name: "commandexecution", payload: item}}
	default:
		return nil
	}
}

func codexMessageEvents(record, payload map[string]any) []event {
	content, _ := payload["content"].([]any)
	events := make([]event, 0, len(content))
	for _, rawBlock := range content {
		block, ok := object(rawBlock)
		if ok && slices.Contains([]string{"input_text", "output_text"}, stringValue(block["type"])) {
			events = append(events, event{source: "codex", kind: "text", timestamp: record["timestamp"], role: payload["role"], text: block["text"]})
		}
	}
	return events
}

func decodeJSON(value any) any {
	text, ok := value.(string)
	if !ok {
		return value
	}
	var decoded any
	if json.Unmarshal([]byte(text), &decoded) != nil {
		return value
	}
	return decoded
}

func labelled(payload any, keys ...string) string {
	values, ok := object(payload)
	if !ok {
		return Unknown
	}
	for _, key := range keys {
		value, exists := values[key]
		if !exists {
			continue
		}
		if value == nil {
			return None
		}
		return fmt.Sprint(value)
	}
	return Unknown
}

func era(timestamp any, since string) string {
	value, ok := timestamp.(string)
	if !ok || !datePrefix(value) {
		return "inconnu"
	}
	if value[:10] < since {
		return "avant"
	}
	return "depuis"
}

func datePrefix(value string) bool {
	if len(value) < 10 || value[4] != '-' || value[7] != '-' {
		return false
	}
	for index := range 10 {
		if index == 4 || index == 7 {
			continue
		}
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}
	return true
}

func isPotentialWrite(value event) bool {
	name := strings.ToLower(stringValue(value.name))
	if name == "write_stdin" {
		return true
	}
	if !slices.Contains(shellWriters, name) {
		return false
	}
	payload := value.payload
	if objectPayload, ok := object(payload); ok {
		payload = objectPayload["command"]
		if payload == nil {
			payload = objectPayload["cmd"]
		}
	}
	command := commandText(payload)
	for _, match := range redirectPattern.FindAllStringSubmatch(command, -1) {
		if !strings.EqualFold(match[1], "/dev/null") {
			return true
		}
	}
	for _, pattern := range shellWritePatterns {
		if pattern.MatchString(command) {
			return true
		}
	}
	return false
}

func commandText(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			if part := commandText(item); part != "" {
				parts = append(parts, part)
			}
		}
		return strings.Join(parts, " ")
	default:
		return ""
	}
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

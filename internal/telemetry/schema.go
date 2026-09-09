package telemetry

import "slices"

var claudeRecordTypes = []string{
	"assistant", "atis-latch", "artifact-autoreact-ledger", "artifact-comment-monitor",
	"attachment", "bridge-session", "custom-title", "frame-link", "last-prompt", "mode",
	"pr-link", "queue-operation", "relocated", "result", "started", "system", "user",
	"worktree-state",
}

var codexRecordTypes = []string{
	"compacted", "event_msg", "inter_agent_communication_metadata", "response_item", "session_meta",
	"realtime_item", "token_usage_record", "turn_context", "world_state",
}

var codexPayloadTypes = []any{
	"agent_message", "custom_tool_call", "custom_tool_call_output", "function_call",
	"function_call_output", "item_completed", "message", "reasoning", "task_complete",
	"task_started", "thread_settings_applied", "token_count", "turn_aborted",
	nil,
}

var codexItemTypes = []string{
	"AgentMessage", "CollabAgentToolCall", "CommandExecution", "Extension", "FileChange",
	"ContextCompaction", "McpToolCall", "Reasoning", "SubAgentActivity", "UserMessage",
}

func validBlocks(value any, textTypes, passiveTypes []string) bool {
	blocks, ok := value.([]any)
	if !ok {
		return false
	}
	allowed := append(slices.Clone(textTypes), passiveTypes...)
	for _, rawBlock := range blocks {
		block, ok := object(rawBlock)
		if !ok {
			return false
		}
		blockType, ok := block["type"].(string)
		if !ok || !slices.Contains(allowed, blockType) {
			return false
		}
		if slices.Contains(textTypes, blockType) {
			if _, ok := block["text"].(string); !ok {
				return false
			}
		}
	}
	return true
}

func validClaudeRecord(record map[string]any) bool {
	recordType, ok := record["type"].(string)
	if !ok || !slices.Contains(claudeRecordTypes, recordType) {
		return false
	}
	if recordType != "assistant" && recordType != "user" {
		return true
	}
	message, ok := object(record["message"])
	if !ok || message["role"] != recordType {
		return false
	}
	content := message["content"]
	if recordType == "user" {
		if _, ok := content.(string); ok {
			return true
		}
		blocks, ok := content.([]any)
		if !ok {
			return false
		}
		for _, block := range blocks {
			if _, ok := object(block); !ok {
				return false
			}
		}
		return true
	}
	if !validBlocks(content, []string{"text"}, []string{"thinking", "tool_use"}) {
		return false
	}
	for _, rawBlock := range content.([]any) {
		block, _ := object(rawBlock)
		if block["type"] == "tool_use" {
			if _, ok := block["name"].(string); !ok {
				return false
			}
		}
	}
	return true
}

func validFileChanges(value any) bool {
	changes, ok := object(value)
	if !ok {
		return false
	}
	for _, rawChange := range changes {
		change, ok := object(rawChange)
		if !ok {
			return false
		}
		changeType, ok := change["type"].(string)
		if !ok || !slices.Contains([]string{"add", "update", "delete"}, changeType) {
			return false
		}
		contentKey := "content"
		if changeType == "update" {
			contentKey = "unified_diff"
		}
		if _, ok := change[contentKey].(string); !ok {
			return false
		}
	}
	return true
}

func validItem(value any) bool {
	item, ok := object(value)
	if !ok {
		return false
	}
	itemType, ok := item["type"].(string)
	if !ok || !slices.Contains(codexItemTypes, itemType) {
		return false
	}
	switch itemType {
	case "AgentMessage":
		return validBlocks(item["content"], []string{"Text"}, nil)
	case "UserMessage":
		return validBlocks(item["content"], []string{"text"}, nil)
	case "FileChange":
		return validFileChanges(item["changes"])
	case "CommandExecution":
		command, ok := item["command"].([]any)
		if !ok || len(command) == 0 {
			return false
		}
		for _, part := range command {
			if _, ok := part.(string); !ok {
				return false
			}
		}
	}
	return true
}

func validCodexPayload(payload map[string]any) bool {
	payloadType := payload["type"]
	if payloadType != nil {
		if _, ok := payloadType.(string); !ok {
			return false
		}
	}
	if !slices.Contains(codexPayloadTypes, payloadType) {
		return false
	}
	switch payloadType {
	case "message":
		role, ok := payload["role"].(string)
		return ok && slices.Contains([]string{"assistant", "developer", "user"}, role) &&
			validBlocks(payload["content"], []string{"input_text", "output_text"}, nil)
	case "agent_message":
		return validBlocks(payload["content"], []string{"input_text"}, []string{"encrypted_content"})
	case "function_call":
		_, nameOK := payload["name"].(string)
		_, argumentsOK := payload["arguments"].(string)
		return nameOK && argumentsOK
	case "custom_tool_call":
		_, nameOK := payload["name"].(string)
		_, inputOK := payload["input"].(string)
		return nameOK && inputOK
	case "item_completed":
		return validItem(payload["item"])
	case "turn_aborted":
		return stringFields(payload, "turn_id", "reason") &&
			integerFields(payload, "started_at", "completed_at", "duration_ms")
	default:
		return true
	}
}

func stringFields(object map[string]any, keys ...string) bool {
	for _, key := range keys {
		if _, ok := object[key].(string); !ok {
			return false
		}
	}
	return true
}

func recognizedRecord(source string, record map[string]any) bool {
	if source == "claude" {
		return validClaudeRecord(record)
	}
	recordType, ok := record["type"].(string)
	if !ok || !slices.Contains(codexRecordTypes, recordType) {
		return false
	}
	payload, ok := object(record["payload"])
	if !ok {
		return slices.Contains([]string{"compacted", "inter_agent_communication_metadata", "world_state"}, recordType)
	}
	if recordType == "realtime_item" {
		return validRealtimePayload(payload)
	}
	return validCodexPayload(payload)
}

func object(value any) (map[string]any, bool) {
	object, ok := value.(map[string]any)
	return object, ok
}

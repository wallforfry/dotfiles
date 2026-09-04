CLAUDE_RECORD_TYPES = {
    "assistant", "atis-latch", "artifact-autoreact-ledger", "artifact-comment-monitor",
    "attachment", "bridge-session", "custom-title", "frame-link", "last-prompt", "mode",
    "pr-link", "queue-operation", "relocated", "result", "started", "system", "user",
    "worktree-state",
}
CODEX_RECORD_TYPES = {
    "compacted", "event_msg", "inter_agent_communication_metadata", "response_item", "session_meta",
    "token_usage_record", "turn_context", "world_state",
}
CODEX_PAYLOAD_TYPES = {
    "agent_message", "custom_tool_call", "custom_tool_call_output", "function_call",
    "function_call_output", "item_completed", "message", "reasoning", "task_complete",
    "task_started", "thread_settings_applied", "token_count", None,
}
CODEX_ITEM_TYPES = {
    "AgentMessage", "CollabAgentToolCall", "CommandExecution", "Extension", "FileChange",
    "ContextCompaction", "McpToolCall", "Reasoning", "SubAgentActivity", "UserMessage",
}


def valid_blocks(value, text_types, passive_types=()):
    if not isinstance(value, list):
        return False
    allowed = set(text_types) | set(passive_types)
    return all(
        isinstance(block, dict)
        and block.get("type") in allowed
        and (block.get("type") not in text_types or isinstance(block.get("text"), str))
        for block in value
    )


def valid_claude_record(record):
    record_type = record.get("type")
    if record_type not in CLAUDE_RECORD_TYPES:
        return False
    if record_type not in {"assistant", "user"}:
        return True
    message = record.get("message")
    if not isinstance(message, dict) or message.get("role") != record_type:
        return False
    content = message.get("content")
    if record_type == "user" and isinstance(content, str):
        return True
    if record_type == "user":
        return isinstance(content, list) and all(isinstance(block, dict) for block in content)
    return valid_blocks(content, {"text"}, {"thinking", "tool_use"}) and all(
        block.get("type") != "tool_use" or isinstance(block.get("name"), str)
        for block in content
    )


def valid_file_changes(changes):
    if not isinstance(changes, dict):
        return False
    for path, change in changes.items():
        if not isinstance(path, str) or not isinstance(change, dict):
            return False
        change_type = change.get("type")
        content = change.get("unified_diff") if change_type == "update" else change.get("content")
        if change_type not in {"add", "update", "delete"} or not isinstance(content, str):
            return False
    return True


def valid_item(item):
    if not isinstance(item, dict) or item.get("type") not in CODEX_ITEM_TYPES:
        return False
    item_type = item.get("type")
    if item_type == "AgentMessage":
        return valid_blocks(item.get("content"), {"Text"})
    if item_type == "UserMessage":
        return valid_blocks(item.get("content"), {"text"})
    if item_type == "FileChange":
        return valid_file_changes(item.get("changes"))
    if item_type == "CommandExecution":
        command = item.get("command")
        return isinstance(command, list) and bool(command) and all(isinstance(part, str) for part in command)
    return True


def valid_codex_payload(payload):
    payload_type = payload.get("type")
    if payload_type not in CODEX_PAYLOAD_TYPES:
        return False
    if payload_type == "message":
        return isinstance(payload.get("role"), str) and valid_blocks(
            payload.get("content"), {"input_text", "output_text"}
        )
    if payload_type == "agent_message":
        return valid_blocks(payload.get("content"), {"input_text"}, {"encrypted_content"})
    if payload_type == "function_call":
        return isinstance(payload.get("name"), str) and isinstance(payload.get("arguments"), str)
    if payload_type == "custom_tool_call":
        return isinstance(payload.get("name"), str) and isinstance(payload.get("input"), str)
    if payload_type == "item_completed":
        return valid_item(payload.get("item"))
    return True


def recognized_record(source, record):
    if source == "claude":
        return valid_claude_record(record)
    record_type = record.get("type")
    if record_type not in CODEX_RECORD_TYPES:
        return False
    payload = record.get("payload")
    if not isinstance(payload, dict):
        return record_type in {"compacted", "inter_agent_communication_metadata", "world_state"}
    return valid_codex_payload(payload)

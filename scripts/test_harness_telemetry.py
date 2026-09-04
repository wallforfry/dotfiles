#!/usr/bin/env python3

import importlib.util
import json
import pathlib
import sys
import tempfile
import unittest


MODULE_PATH = pathlib.Path(__file__).with_name("harness_telemetry.py")
sys.path.insert(0, str(MODULE_PATH.parent))
SPEC = importlib.util.spec_from_file_location("harness_telemetry", MODULE_PATH)
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class HarnessTelemetryTest(unittest.TestCase):
    def summarize(self, source, records):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory, "session.jsonl")
            path.write_text("".join(json.dumps(record) + "\n" for record in records), encoding="utf-8")
            return MODULE.summarize_file(source, path, "2026-08-17")

    def test_claude_normalizes_text_activation_and_all_editors(self):
        records = [
            {
                "type": "assistant",
                "timestamp": "2026-08-18T00:00:00Z",
                "message": {
                    "role": "assistant",
                    "content": [{"type": "text", "text": "a—b·c"}],
                },
            },
            {
                "type": "assistant",
                "timestamp": "2026-08-18T00:00:01Z",
                "message": {
                    "role": "assistant",
                    "content": [
                        {"type": "tool_use", "name": "Skill", "input": {"skill": None}},
                        {"type": "tool_use", "name": "Agent", "input": {}},
                        {"type": "tool_use", "name": "Write", "input": {"file_path": "a.py", "content": "# x\ny = 1"}},
                        {"type": "tool_use", "name": "Edit", "input": {"file_path": "a.ts", "new_string": "// x\ny()"}},
                        {"type": "tool_use", "name": "MultiEdit", "input": {"file_path": "a.js", "edits": [{"new_string": "/* x */\ny()"}]}},
                        {"type": "tool_use", "name": "NotebookEdit", "input": {"notebook_path": "a.ipynb", "new_source": "# x\ny = 1"}},
                        {"type": "tool_use", "name": "Bash", "input": {"command": "git commit -m test"}},
                    ],
                },
            },
        ]
        summary = self.summarize("claude", records)
        self.assertEqual(summary["dash"]["depuis"], 1)
        self.assertEqual(summary["middle_dot"]["depuis"], 1)
        self.assertEqual(summary["skills"][MODULE.NONE], 1)
        self.assertEqual(summary["agents"][MODULE.UNKNOWN], 1)
        self.assertEqual(summary["lines"]["depuis"], 8)
        self.assertEqual(summary["comments"]["depuis"], 4)
        self.assertEqual(summary["uninspectable_writes"]["bash"], 1)

    def test_codex_normalizes_messages_without_counting_completed_projection(self):
        records = [
            {
                "type": "response_item",
                "timestamp": "2026-08-16T00:00:00Z",
                "payload": {
                    "type": "message",
                    "role": "assistant",
                    "content": [{"type": "output_text", "text": "a—b·c"}],
                },
            },
            {
                "type": "response_item",
                "timestamp": "2026-08-18T00:00:02Z",
                "payload": {
                    "type": "item_completed",
                    "item": {
                        "type": "AgentMessage",
                        "content": [{"type": "Text", "text": "a—b·c"}],
                    },
                },
            },
        ]
        summary = self.summarize("codex", records)
        self.assertEqual(summary["blocks"]["avant"], 1)
        self.assertEqual(summary["dash"]["avant"], 1)
        self.assertEqual(summary["middle_dot"]["avant"], 1)

    def test_codex_normalizes_completed_file_and_command_tools(self):
        records = [
            {
                "type": "response_item",
                "timestamp": "2026-08-18T00:00:03Z",
                "payload": {
                    "type": "item_completed",
                    "item": {
                        "type": "FileChange",
                        "changes": {
                            "safe.py": {
                                "type": "update",
                                "unified_diff": "@@\n+# x\n+y = 1\n",
                            }
                        },
                    },
                },
            },
            {
                "type": "response_item",
                "payload": {
                    "type": "item_completed",
                    "item": {
                        "type": "CommandExecution",
                        "command": ["git", "commit", "-m", "test"],
                    },
                },
            },
        ]
        summary = self.summarize("codex", records)
        self.assertEqual(summary["lines"]["depuis"], 2)
        self.assertEqual(summary["comments"]["depuis"], 1)
        self.assertEqual(summary["write_tools"]["filechange"], 1)
        self.assertEqual(summary["uninspectable_writes"]["commandexecution"], 1)

    def test_codex_normalizes_direct_and_nested_tools_without_completed_duplicates(self):
        records = [
            {
                "type": "response_item",
                "timestamp": "2026-08-18T00:00:00Z",
                "payload": {
                    "type": "function_call",
                    "name": "spawn_agent",
                    "arguments": json.dumps({"task_name": "reader"}),
                },
            },
            {
                "type": "response_item",
                "payload": {
                    "type": "custom_tool_call",
                    "name": "exec",
                    "input": "await tools.apply_patch(`*** Begin Patch\\n*** Update File: a.py\\n+# x\\n+y = 1\\n*** End Patch`)",
                },
            },
            {
                "type": "response_item",
                "payload": {
                    "type": "custom_tool_call",
                    "name": "exec",
                    "input": "await tools.write_stdin({session_id: 1, chars: 'x'})",
                },
            },
            {
                "type": "response_item",
                "payload": {
                    "type": "item_completed",
                    "item": {"type": "SubAgentActivity", "kind": "started"},
                },
            },
            {
                "type": "response_item",
                "payload": {
                    "type": "item_completed",
                    "item": {"type": "CollabAgentToolCall", "tool": "spawn_agent"},
                },
            },
        ]
        summary = self.summarize("codex", records)
        self.assertEqual(summary["agents"][MODULE.UNKNOWN], 1)
        self.assertEqual(summary["lines"]["inconnu"], 2)
        self.assertEqual(summary["comments"]["inconnu"], 1)
        self.assertEqual(summary["write_tools"]["apply_patch"], 1)
        self.assertEqual(summary["uninspectable_writes"]["write_stdin"], 1)
        self.assertNotIn("reader", json.dumps(summary))

    def test_zero_denominator_is_unknown(self):
        self.assertEqual(MODULE.rate(0, 0), "inconnu")

    def test_cache_rejects_another_version_or_period(self):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory, "cache.json")
            path.write_text(json.dumps({"version": MODULE.CACHE_VERSION, "since": "x", "files": {"a": {}}}), encoding="utf-8")
            self.assertEqual(MODULE.load_cache(path, "y"), {})

    def test_activated_body_cost_weights_local_skills_only(self):
        summary = MODULE.empty_summary("claude")
        summary["skills"] = {"adr": 2, "external": 9}
        self.assertEqual(MODULE.activated_body_cost(summary, {"adr": 100}), (2, 200))

    def test_unknown_and_invalid_records_are_not_silently_accepted(self):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory, "session.jsonl")
            path.write_text('{"type":"future"}\nnot-json\n', encoding="utf-8")
            summary = MODULE.summarize_file("codex", path, "2026-08-17")
        self.assertEqual(summary["unknown_records"]["codex"], 1)
        self.assertEqual(summary["invalid_records"]["codex"], 1)

    def test_malformed_known_codex_record_is_not_silently_accepted(self):
        records = [
            {
                "type": "response_item",
                "payload": {"type": "message", "role": "assistant", "content": [{"type": "future"}]},
            },
            {
                "type": "response_item",
                "payload": {"type": "message", "role": "future", "content": []},
            },
            {
                "type": "event_msg",
                "payload": {
                    "type": "item_completed",
                    "item": {"type": "CommandExecution", "command": "git commit"},
                },
            },
        ]
        summary = self.summarize("codex", records)
        self.assertEqual(summary["unknown_records"]["codex"], 3)
        self.assertEqual(summary["blocks"], {})
        self.assertEqual(summary["write_tools"], {})

    def test_codex_recognizes_compaction_records_without_inventing_events(self):
        records = [
            {"type": "compacted"},
            {
                "type": "event_msg",
                "payload": {
                    "type": "item_completed",
                    "item": {"type": "ContextCompaction"},
                },
            },
        ]
        summary = self.summarize("codex", records)
        self.assertEqual(summary["records"]["codex"], 2)
        self.assertEqual(summary["unknown_records"], {})
        self.assertEqual(summary["blocks"], {})
        self.assertEqual(summary["agents"], {})


if __name__ == "__main__":
    unittest.main()

#!/usr/bin/env python3

import importlib.util
import json
import pathlib
import sys
import tempfile
import unittest


MODULE_PATH = pathlib.Path(__file__).with_name("harness_telemetry.py")
FIXTURES = pathlib.Path(__file__).with_name("fixtures") / "harness-telemetry"
sys.path.insert(0, str(MODULE_PATH.parent))
SPEC = importlib.util.spec_from_file_location("harness_telemetry", MODULE_PATH)
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class HarnessTelemetryFormatTest(unittest.TestCase):
    def summarize(self, source, records):
        with tempfile.TemporaryDirectory() as directory:
            path = pathlib.Path(directory, "session.jsonl")
            path.write_text("".join(json.dumps(record) + "\n" for record in records), encoding="utf-8")
            return MODULE.summarize_file(source, path, "2026-08-17")

    def test_codex_normalizes_assistant_realtime_segment_and_keeps_other_items_passive(self):
        summary = MODULE.summarize_file(
            "codex", FIXTURES / "codex-realtime.jsonl", "2026-08-17"
        )
        self.assertEqual(summary["read_records"]["codex"], 4)
        self.assertEqual(summary["recognized_records"]["codex"], 4)
        self.assertEqual(summary["unknown_records"], {})
        self.assertEqual(summary["invalid_records"], {})
        self.assertEqual(summary["blocks"], {"depuis": 1})
        self.assertEqual(summary["dash"], {"depuis": 1})
        self.assertEqual(summary["middle_dot"], {"depuis": 1})
        self.assertEqual(summary["agents"], {})

    def test_malformed_realtime_records_are_unknown(self):
        records = [
            {
                "type": "realtime_item",
                "payload": {
                    "type": "transcript_segment",
                    "id": "event-1",
                    "realtime_session_id": "session-1",
                    "role": "speaker",
                    "text": "future role",
                },
            },
            {
                "type": "realtime_item",
                "payload": {
                    "type": "future",
                    "id": "event-2",
                    "realtime_session_id": "session-1",
                },
            },
        ]
        summary = self.summarize("codex", records)
        self.assertEqual(summary["read_records"]["codex"], 2)
        self.assertEqual(summary["unknown_records"]["codex"], 2)
        self.assertEqual(summary["recognized_records"], {})

    def test_codex_spawn_agent_uses_stable_tool_label_without_task_name(self):
        records = [
            {
                "type": "response_item",
                "payload": {
                    "type": "function_call",
                    "name": "spawn_agent",
                    "arguments": json.dumps({"task_name": "private-label", "message": "private"}),
                },
            },
            {
                "type": "response_item",
                "payload": {
                    "type": "function_call",
                    "name": "spawn_agent",
                    "arguments": json.dumps({"agent_type": "reader", "task_name": "private-label"}),
                },
            },
        ]
        summary = self.summarize("codex", records)
        self.assertEqual(summary["agents"], {"spawn_agent": 1, "reader": 1})
        self.assertNotIn("private-label", json.dumps(summary))

    def test_claude_agent_uses_stable_tool_label_without_description(self):
        records = [
            {
                "type": "assistant",
                "message": {
                    "role": "assistant",
                    "content": [
                        {
                            "type": "tool_use",
                            "name": "Agent",
                            "input": {"description": "fixture", "prompt": "fixture"},
                        }
                    ],
                },
            }
        ]
        summary = self.summarize("claude", records)
        self.assertEqual(summary["agents"], {"agent": 1})
        self.assertNotIn("fixture", json.dumps(summary))

    def test_unqualified_generic_task_stays_unknown(self):
        records = [
            {
                "type": "assistant",
                "message": {
                    "role": "assistant",
                    "content": [{"type": "tool_use", "name": "Task", "input": {}}],
                },
            }
        ]
        summary = self.summarize("claude", records)
        self.assertEqual(summary["agents"], {MODULE.UNKNOWN: 1})

    def test_codex_turn_aborted_is_passive(self):
        records = [
            {
                "type": "event_msg",
                "payload": {
                    "type": "turn_aborted",
                    "turn_id": "turn-1",
                    "reason": "interrupted",
                    "started_at": 1000,
                    "completed_at": 2000,
                    "duration_ms": 1000,
                },
            },
            {
                "type": "event_msg",
                "payload": {
                    "type": "turn_aborted",
                    "turn_id": "turn-2",
                    "started_at": 1000,
                    "completed_at": 2000,
                    "duration_ms": 1000,
                },
            },
        ]
        summary = self.summarize("codex", records)
        self.assertEqual(summary["recognized_records"], {"codex": 1})
        self.assertEqual(summary["unknown_records"], {"codex": 1})
        self.assertEqual(summary["blocks"], {})
        self.assertEqual(summary["agents"], {})


if __name__ == "__main__":
    unittest.main()

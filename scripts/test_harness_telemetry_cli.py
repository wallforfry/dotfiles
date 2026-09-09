#!/usr/bin/env python3

import contextlib
import importlib.util
import io
import json
import pathlib
import subprocess
import sys
import tempfile
import unittest
from unittest import mock


MODULE_PATH = pathlib.Path(__file__).with_name("harness_telemetry.py")
sys.path.insert(0, str(MODULE_PATH.parent))
SPEC = importlib.util.spec_from_file_location("harness_telemetry", MODULE_PATH)
MODULE = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(MODULE)


class HarnessTelemetryCliTest(unittest.TestCase):
    def run_main(self, root, cache):
        arguments = [
            str(MODULE_PATH),
            "--codex-root",
            str(root),
            "--since",
            "2026-08-17",
            "--cache",
            str(cache),
        ]
        stdout = io.StringIO()
        stderr = io.StringIO()
        with mock.patch.object(sys, "argv", arguments):
            with contextlib.redirect_stdout(stdout), contextlib.redirect_stderr(stderr):
                status = MODULE.main()
        return status, stdout.getvalue(), stderr.getvalue()

    def test_format_drift_returns_dedicated_status(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory, "sessions")
            root.mkdir()
            pathlib.Path(root, "session.jsonl").write_text(
                json.dumps({"type": "future"}) + "\n", encoding="utf-8"
            )
            status, _, stderr = self.run_main(root, pathlib.Path(directory, "cache.json"))
        self.assertEqual(status, 2)
        self.assertIn("formats non mesurés", stderr)

    def test_interrupted_read_returns_dedicated_status_without_path(self):
        with tempfile.TemporaryDirectory() as directory:
            root = pathlib.Path(directory, "sessions")
            root.mkdir()
            pathlib.Path(root, "session.jsonl").symlink_to(pathlib.Path(directory, "missing"))
            status, _, stderr = self.run_main(root, pathlib.Path(directory, "cache.json"))
            self.assertNotIn(str(root), stderr)
        self.assertEqual(status, 3)
        self.assertIn("lecture interrompue", stderr)

    def test_shell_wrapper_translates_telemetry_statuses(self):
        script = MODULE_PATH.with_name("harness-telemetry-status.sh")
        messages = {
            status: subprocess.run(
                ["bash", "-c", 'source "$1"; telemetry_failure_message "$2"', "_", str(script), str(status)],
                check=True,
                capture_output=True,
                text=True,
            ).stdout
            for status in (2, 3)
        }
        self.assertEqual(
            messages[2],
            "dérive de format des transcripts : activation et adhérence non mesurées\n",
        )
        self.assertEqual(
            messages[3],
            "lecture des transcripts interrompue : activation et adhérence non mesurées\n",
        )
        self.assertNotEqual(messages[2], messages[3])


if __name__ == "__main__":
    unittest.main()

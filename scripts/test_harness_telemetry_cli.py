#!/usr/bin/env python3

import contextlib
import importlib.util
import io
import json
import pathlib
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


if __name__ == "__main__":
    unittest.main()

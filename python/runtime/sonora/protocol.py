"""The JSON-lines channel between the worker and the app."""

import json
import os
import sys
import threading


class Channel:
    def __init__(self, out):
        self._out = out
        self._lock = threading.Lock()

    @classmethod
    def take_stdout(cls):
        """Keep the real stdout for protocol messages and point fd 1 (and
        sys.stdout) at stderr, so stray prints can't corrupt the stream."""
        out = os.fdopen(os.dup(1), "w", encoding="utf-8", buffering=1)
        os.dup2(2, 1)
        sys.stdout = sys.stderr
        return cls(out)

    def send(self, **message):
        line = json.dumps(message, ensure_ascii=False)
        with self._lock:
            self._out.write(line + "\n")
            self._out.flush()

    def requests(self):
        for line in sys.stdin:
            line = line.strip()
            if line:
                yield json.loads(line)

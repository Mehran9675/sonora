"""Helpers shared by Sonora engine adapters.

An adapter is a Python file defining ``Engine``::

    class Engine:
        def __init__(self, ctx): ...
        def load(self, options) -> dict: ...         # returns {"voices": [...]} (optional)
        def synthesize(self, request) -> dict: ...   # writes request["output"]

Speech-to-text engines ("task": "transcribe" in engine.json) define
``transcribe`` instead of ``synthesize``::

        def transcribe(self, request) -> dict: ...
            # request: {"audio": 16 kHz mono WAV path, "language": "en" or "",
            #           "translate": bool, "options": {...}}
            # returns {"text": ..., "language": ..., "segments": [{"start", "end", "text"}]}

See python/engines/*/adapter.py for real examples.
"""

from .audio import save_wav, read_wav, concat  # noqa: F401
from .text import split_text  # noqa: F401

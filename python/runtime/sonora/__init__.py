"""Helpers shared by Sonora engine adapters.

An adapter is a Python file defining ``Engine``::

    class Engine:
        def __init__(self, ctx): ...
        def load(self, options) -> dict: ...         # returns {"voices": [...]} (optional)
        def synthesize(self, request) -> dict: ...   # writes request["output"]

See python/engines/*/adapter.py for real examples.
"""

from .audio import save_wav, read_wav, concat  # noqa: F401
from .text import split_text  # noqa: F401

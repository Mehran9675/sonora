"""Imports an adapter file and instantiates its Engine class."""

import importlib.util
import os
import sys


def load_adapter(path, ctx):
    folder = os.path.dirname(os.path.abspath(path))
    if folder not in sys.path:
        sys.path.insert(0, folder)
    spec = importlib.util.spec_from_file_location("sonora_adapter", path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    if not hasattr(module, "Engine"):
        raise RuntimeError(f"{path} does not define an Engine class")
    return module.Engine(ctx)

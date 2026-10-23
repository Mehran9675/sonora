"""Sonora engine worker.

Runs one engine adapter in a long-lived process so the model stays
loaded between requests. Talks JSON lines with the app: requests arrive
on stdin, replies and events leave on the original stdout. Anything a
library prints is redirected to stderr, which the app shows as the
engine log.
"""

import argparse
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from sonora.protocol import Channel  # noqa: E402
from sonora.context import Context  # noqa: E402
from sonora.loader import load_adapter  # noqa: E402
from sonora.server import serve  # noqa: E402


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--adapter", required=True, help="path to the engine's adapter.py")
    parser.add_argument("--engine-dir", required=True, help="the engine's install folder")
    args = parser.parse_args()

    channel = Channel.take_stdout()
    ctx = Context(channel, engine_dir=args.engine_dir)
    try:
        adapter = load_adapter(args.adapter, ctx)
    except Exception as exc:  # report import errors as a failed handshake
        channel.send(event="fatal", message=f"{type(exc).__name__}: {exc}")
        raise
    channel.send(event="ready")
    serve(channel, adapter)


if __name__ == "__main__":
    main()

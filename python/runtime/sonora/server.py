"""Dispatches requests from the app to the adapter."""

import traceback


def serve(channel, adapter):
    for req in channel.requests():
        rid = req.get("id")
        cmd = req.get("cmd")
        if cmd == "quit":
            break
        try:
            if cmd == "ping":
                result = {"pong": True}
            elif cmd == "load":
                result = adapter.load(req.get("options") or {}) or {}
            elif cmd == "synthesize":
                result = adapter.synthesize(req.get("request") or {}) or {}
            else:
                raise ValueError(f"unknown command {cmd!r}")
            channel.send(id=rid, ok=True, result=result)
        except Exception as exc:
            traceback.print_exc()
            channel.send(id=rid, ok=False, error=f"{type(exc).__name__}: {exc}")

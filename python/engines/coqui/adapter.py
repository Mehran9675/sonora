"""Coqui TTS (maintained fork: https://github.com/idiap/coqui-ai-TTS)."""

import os

from sonora import save_wav


class Engine:
    def __init__(self, ctx):
        self.ctx = ctx
        self.tts = None

    def load(self, options):
        name = options.get("model") or "tts_models/multilingual/multi-dataset/xtts_v2"
        if "xtts" in name:
            if not options.get("agree_cpml"):
                raise RuntimeError("XTTS needs the Coqui Public Model License accepted (see the engine options)")
            os.environ["COQUI_TOS_AGREED"] = "1"
        from TTS.api import TTS

        self.ctx.log(f"Loading {name} (the first run downloads the model)…")
        self.tts = TTS(name).to(self.ctx.device())
        speakers = list(getattr(self.tts, "speakers", None) or [])
        return {"voices": [{"id": s.strip(), "name": s.strip()} for s in speakers]}

    def synthesize(self, req):
        opts = req.get("options") or {}
        kwargs = {"text": req["text"], "file_path": req["output"]}
        if req.get("reference"):
            kwargs["speaker_wav"] = req["reference"]
        elif req.get("voice"):
            kwargs["speaker"] = req["voice"]
        if getattr(self.tts, "is_multi_lingual", False):
            kwargs["language"] = opts.get("language") or "en"
        speed = float(req.get("speed") or 1.0)
        if speed != 1.0:
            kwargs["speed"] = speed
        self.tts.tts_to_file(**kwargs)
        return {}

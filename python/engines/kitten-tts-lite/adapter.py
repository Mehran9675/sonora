"""Kitten TTS 0.8 ONNX models."""

from sonora import save_wav


class Engine:
    def __init__(self, ctx):
        self.ctx = ctx
        self.model = None

    def load(self, options):
        from kittenml import KittenTTS

        name = options.get("model") or "KittenML/kitten-tts-mini-0.8"
        self.ctx.log(f"Loading {name}…")
        self.model = KittenTTS(name)
        return {"voices": [{"id": v, "name": v} for v in self.model.available_voices]}

    def synthesize(self, req):
        opts = req.get("options") or {}
        audio = self.model.generate(
            req["text"],
            voice=req.get("voice") or "Jasper",
            speed=float(req.get("speed") or 1.0),
            clean_text=bool(opts.get("clean_text", True)),
        )
        return save_wav(req["output"], audio, 24000)

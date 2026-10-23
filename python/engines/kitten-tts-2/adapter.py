"""Kitten TTS 2 (https://github.com/KittenML/KittenTTS)."""

from sonora import save_wav

MODEL = "KittenML/kitten-tts-2"


class Engine:
    def __init__(self, ctx):
        self.ctx = ctx
        self.model = None

    def load(self, options):
        from kittenml import KittenTTS

        kwargs = {"device": self.ctx.device()}
        if options.get("weights") and options["weights"] != "packed":
            kwargs["weights"] = options["weights"]
        self.ctx.log("Loading Kitten TTS 2 (the first run downloads the model)…")
        self.model = KittenTTS(MODEL, **kwargs)
        return {"voices": [{"id": v, "name": v} for v in self.model.available_voices]}

    def synthesize(self, req):
        opts = req.get("options") or {}
        kwargs = {"preset": opts.get("preset") or "stable", "normalize": bool(opts.get("normalize", True))}
        if req.get("reference"):
            kwargs["reference"] = req["reference"]
            if req.get("reference_text"):
                kwargs["reference_text"] = req["reference_text"]
        elif req.get("voice"):
            kwargs["voice"] = req["voice"]
        if float(opts.get("temperature") or 0) > 0:
            kwargs["temperature"] = float(opts["temperature"])
        seed = str(opts.get("seed") or "").strip()
        if seed.isdigit():
            kwargs["advanced"] = {"seed": int(seed)}
        audio = self.model.generate(req["text"], **kwargs)
        return save_wav(req["output"], audio, self.model.sample_rate)

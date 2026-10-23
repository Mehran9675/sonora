"""Chatterbox by Resemble AI (https://github.com/resemble-ai/chatterbox)."""

from sonora import concat, save_wav, split_text


class Engine:
    def __init__(self, ctx):
        self.ctx = ctx
        self.model = None

    def load(self, options):
        from chatterbox.tts import ChatterboxTTS

        self.ctx.log("Loading Chatterbox (the first run downloads the model)…")
        self.model = ChatterboxTTS.from_pretrained(device=self.ctx.device())
        return {}

    def synthesize(self, req):
        opts = req.get("options") or {}
        kwargs = {
            "exaggeration": float(opts.get("exaggeration", 0.5)),
            "cfg_weight": float(opts.get("cfg_weight", 0.5)),
            "temperature": float(opts.get("temperature", 0.8)),
        }
        if req.get("reference"):
            kwargs["audio_prompt_path"] = req["reference"]
        chunks = [self.model.generate(chunk, **kwargs) for chunk in split_text(req["text"], 300)]
        return save_wav(req["output"], concat(chunks, self.model.sr), self.model.sr)

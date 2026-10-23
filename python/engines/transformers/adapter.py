"""Any Hugging Face text-to-speech model via the Transformers pipeline."""

from sonora import concat, save_wav, split_text


class Engine:
    def __init__(self, ctx):
        self.ctx = ctx
        self.pipe = None

    def load(self, options):
        from transformers import pipeline

        name = options.get("model") or "facebook/mms-tts-eng"
        self.ctx.log(f"Loading {name} (the first run downloads the model)…")
        self.pipe = pipeline("text-to-speech", model=name, device=self.ctx.device())
        return {}

    def synthesize(self, req):
        opts = req.get("options") or {}
        kwargs = {}
        preset = (opts.get("voice_preset") or "").strip()
        if preset:
            kwargs["voice_preset"] = preset
        chunks, rate = [], 16000
        for chunk in split_text(req["text"], 250):
            out = self.pipe(chunk, **kwargs)
            chunks.append(out["audio"])
            rate = out["sampling_rate"]
        return save_wav(req["output"], concat(chunks, rate), rate)

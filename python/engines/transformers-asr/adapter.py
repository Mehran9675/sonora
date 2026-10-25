"""Any Hugging Face speech recognition model via the Transformers
pipeline."""

from sonora import read_wav


class Engine:
    def __init__(self, ctx):
        self.ctx = ctx
        self.pipe = None

    def load(self, options):
        from transformers import pipeline

        name = options.get("model") or "openai/whisper-large-v3-turbo"
        self.ctx.log(f"Loading {name} (the first run downloads the model)…")
        self.pipe = pipeline("automatic-speech-recognition", model=name, device=self.ctx.device())
        self.whisper = "whisper" in self.pipe.model.config.model_type
        return {}

    def transcribe(self, req):
        audio, rate = read_wav(req["audio"])
        kwargs = {"chunk_length_s": 30, "batch_size": 4}
        generate = {}
        if self.whisper:
            if req.get("language"):
                generate["language"] = req["language"]
            if req.get("translate"):
                generate["task"] = "translate"
        if generate:
            kwargs["generate_kwargs"] = generate
        self.ctx.log("Transcribing…")
        try:
            out = self.pipe({"raw": audio, "sampling_rate": rate}, return_timestamps=True, **kwargs)
        except ValueError:
            # Models without timestamp support (some CTC models).
            out = self.pipe({"raw": audio, "sampling_rate": rate}, **kwargs)
        segments = []
        for c in out.get("chunks") or []:
            start, end = c.get("timestamp") or (None, None)
            if start is None:
                continue
            segments.append({"start": float(start), "end": float(end if end is not None else start), "text": c["text"].strip()})
        return {"text": out["text"].strip(), "language": req.get("language") or "", "segments": segments}

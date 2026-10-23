"""Kokoro-82M (https://github.com/hexgrad/kokoro)."""

from sonora import concat, save_wav

REPO = "hexgrad/Kokoro-82M"


class Engine:
    def __init__(self, ctx):
        self.ctx = ctx
        self.pipelines = {}
        self.model = None

    def load(self, options):
        self._pipeline("a")  # load the model now so the first request is quick
        return {}

    def _pipeline(self, lang):
        if lang not in self.pipelines:
            from kokoro import KModel, KPipeline

            if self.model is None:
                self.ctx.log("Loading Kokoro (the first run downloads the model)…")
                self.model = KModel(repo_id=REPO).to(self.ctx.device()).eval()
            self.pipelines[lang] = KPipeline(lang_code=lang, repo_id=REPO, model=self.model)
        return self.pipelines[lang]

    def synthesize(self, req):
        voice = req.get("voice") or "af_heart"
        pipeline = self._pipeline(voice[0])
        results = pipeline(req["text"], voice=voice, speed=float(req.get("speed") or 1.0))
        chunks = [audio for _, _, audio in results if audio is not None]
        if not chunks:
            raise RuntimeError("Kokoro returned no audio")
        return save_wav(req["output"], concat(chunks, 24000, gap=0.05), 24000)

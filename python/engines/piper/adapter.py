"""Piper voices (https://github.com/OHF-Voice/piper1-gpl)."""

import os
import wave

VOICES_REPO = "rhasspy/piper-voices"


class Engine:
    def __init__(self, ctx):
        self.ctx = ctx
        self.voices = {}

    def load(self, options):
        import piper  # noqa: F401  fail early if the install is broken

        return {}

    def _voice_path(self, name):
        if name.endswith(".onnx") and os.path.exists(name):
            return name
        locale, speaker, quality = name.split("-", 2)
        lang = locale.split("_")[0]
        base = f"{lang}/{locale}/{speaker}/{quality}/{name}.onnx"
        from huggingface_hub import hf_hub_download

        self.ctx.log(f"Fetching voice {name}…")
        path = hf_hub_download(VOICES_REPO, base, cache_dir=self.ctx.models_dir)
        hf_hub_download(VOICES_REPO, base + ".json", cache_dir=self.ctx.models_dir)
        return path

    def _voice(self, name):
        if name not in self.voices:
            from piper import PiperVoice

            self.voices[name] = PiperVoice.load(self._voice_path(name))
        return self.voices[name]

    def synthesize(self, req):
        opts = req.get("options") or {}
        name = (opts.get("custom_voice") or "").strip() or req.get("voice") or "en_US-lessac-medium"
        voice = self._voice(name)
        length_scale = 1.0 / max(0.25, float(req.get("speed") or 1.0))
        noise = float(opts.get("variation") if opts.get("variation") is not None else 0.667)
        with wave.open(req["output"], "wb") as wav:
            try:  # piper-tts >= 1.3
                from piper import SynthesisConfig

                cfg = SynthesisConfig(length_scale=length_scale, noise_scale=noise)
                voice.synthesize_wav(req["text"], wav, syn_config=cfg)
            except ImportError:  # older releases
                voice.synthesize(req["text"], wav, length_scale=length_scale, noise_scale=noise)
        return {"sample_rate": voice.config.sample_rate}

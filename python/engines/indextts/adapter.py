"""IndexTTS by Bilibili (https://github.com/index-tts/index-tts).

The repository (downloaded at install time into ctx.source_dir) is
installed as the indextts package. Checkpoints are downloaded from
Hugging Face into ctx.models_dir on first load; IndexTTS fetches its
smaller helper models (w2v-BERT, BigVGAN…) into the same folder.
"""

import os

from sonora import read_wav, save_wav

REPOS = {"2.5": "IndexTeam/IndexTTS-2.5", "2": "IndexTeam/IndexTTS-2"}
# Order of IndexTTS's emotion vector.
EMOTIONS = ["happy", "angry", "sad", "afraid", "disgusted", "melancholic", "surprised", "calm"]


class Engine:
    def __init__(self, ctx):
        self.ctx = ctx
        self.tts = None
        self.version = "2.5"

    def load(self, options):
        self.version = options.get("version") if options.get("version") in REPOS else "2.5"
        model_dir = os.path.join(self.ctx.models_dir, REPOS[self.version].split("/")[1])
        if not os.path.exists(os.path.join(model_dir, "config.yaml")):
            from huggingface_hub import snapshot_download

            self.ctx.log(f"Downloading IndexTTS-{self.version} (≈ 6 GB, first run only)…")
            snapshot_download(REPOS[self.version], local_dir=model_dir)
        self.ctx.log(f"Loading IndexTTS-{self.version}…")
        # None lets IndexTTS pick the GPU (and whether half precision suits it).
        device = "cpu" if self.ctx.device() == "cpu" else None
        half = bool(options.get("half", True))
        cfg = os.path.join(model_dir, "config.yaml")
        if self.version == "2.5":
            from indextts.infer_v2_5 import IndexTTS2

            self.tts = IndexTTS2(cfg_path=cfg, model_dir=model_dir, use_bf16=half, device=device, use_qwen_emo=False)
        else:
            from indextts.infer_v2 import IndexTTS2

            self.tts = IndexTTS2(cfg_path=cfg, model_dir=model_dir, use_fp16=half, device=device, use_qwen_emo=False)
        return {}

    def synthesize(self, req):
        if not req.get("reference"):
            raise ValueError("IndexTTS needs a recording of the voice to speak with")
        opts = req.get("options") or {}
        kwargs = {"use_random": bool(opts.get("random", False))}
        emotion = opts.get("emotion") or ""
        if emotion in EMOTIONS:
            vec = [0.0] * len(EMOTIONS)
            vec[EMOTIONS.index(emotion)] = float(opts.get("intensity") or 0.6)
            kwargs["emo_vector"] = vec
        if self.version == "2.5":
            kwargs["lang"] = opts.get("language") or "en"
            speed = float(req.get("speed") or 1.0)
            # duration_factor stretches time: 2× speed is half the duration.
            kwargs["duration_factor"] = min(2.0, max(0.5, 1.0 / speed))
        self.tts.infer(spk_audio_prompt=req["reference"], text=req["text"], output_path=req["output"], **kwargs)
        if not os.path.exists(req["output"]):
            raise RuntimeError("IndexTTS produced no audio; see the engine log")
        samples, rate = read_wav(req["output"])
        return save_wav(req["output"], samples, rate)

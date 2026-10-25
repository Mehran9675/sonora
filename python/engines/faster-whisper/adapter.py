"""Whisper speech recognition through faster-whisper
(https://github.com/SYSTRAN/faster-whisper), which runs on CTranslate2:
int8 on the CPU, float16 on an NVIDIA GPU."""

import os


class Engine:
    def __init__(self, ctx):
        self.ctx = ctx
        self.model = None
        self.name = None

    def load(self, options):
        self.name = options.get("model") or "large-v3-turbo"
        device, compute = "cpu", "int8"
        if os.environ.get("SONORA_DEVICE") != "cpu":
            try:
                import ctranslate2

                if ctranslate2.get_cuda_device_count() > 0:
                    device, compute = "cuda", "float16"
            except Exception:
                pass
        self._open(device, compute)
        return {}

    def _open(self, device, compute):
        from faster_whisper import WhisperModel

        self.ctx.log(f"Loading Whisper {self.name} on the {device.upper()} (the first run downloads it)…")
        self.model = WhisperModel(self.name, device=device, compute_type=compute, download_root=self.ctx.models_dir)
        self.device = device

    def transcribe(self, req):
        try:
            return self._transcribe(req)
        except RuntimeError as exc:
            # CUDA found but its libraries (cuBLAS, cuDNN) are missing:
            # carry on with the CPU.
            if self.device != "cuda" or not any(k in str(exc).lower() for k in ("cuda", "cublas", "cudnn")):
                raise
            self.ctx.log(f"The GPU isn't usable ({exc}); using the CPU instead.")
            self._open("cpu", "int8")
            return self._transcribe(req)

    def _transcribe(self, req):
        opts = req.get("options") or {}
        segments, info = self.model.transcribe(
            req["audio"],
            language=req.get("language") or None,
            task="translate" if req.get("translate") else "transcribe",
            vad_filter=bool(opts.get("vad", True)),
            beam_size=max(1, int(float(opts.get("beam_size") or 5))),
            initial_prompt=(opts.get("prompt") or "").strip() or None,
        )
        out, text = [], ""
        for seg in segments:
            out.append({"start": seg.start, "end": seg.end, "text": seg.text.strip()})
            text += seg.text
            if info.duration:
                self.ctx.progress(min(1.0, seg.end / info.duration), "Transcribing…")
        return {"text": text.strip(), "language": info.language, "segments": out}

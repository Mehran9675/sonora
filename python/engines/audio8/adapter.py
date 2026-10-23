"""Audio8 TTS via the project's own inference script.

The repository (downloaded at install time into ctx.source_dir) ships
audio8_tts_infer.py, whose batch mode loads the model once for all
chunks of a request. Inputs over ~150 characters lose quality, so text
is split into sentences first and the pieces are joined afterwards.
"""

import json
import os
import shutil
import subprocess
import sys
import tempfile

from sonora import concat, read_wav, save_wav, split_text


class Engine:
    def __init__(self, ctx):
        self.ctx = ctx
        self.model = None

    def load(self, options):
        self.model = options.get("model") or "Audio8/Audio8-TTS-Preview-0.6b"
        script = os.path.join(self.ctx.source_dir, "audio8_tts_infer.py")
        if not os.path.exists(script):
            raise RuntimeError("Audio8 sources are missing; reinstall the engine")
        return {}

    def synthesize(self, req):
        chunks = split_text(req["text"], 150)
        if not chunks:
            raise ValueError("nothing to say")
        if req.get("reference") and not (req.get("reference_text") or "").strip():
            raise ValueError("Audio8 needs the transcript of the reference recording")
        opts = req.get("options") or {}
        work = tempfile.mkdtemp(prefix="audio8-")
        manifest = os.path.join(work, "prompts.jsonl")
        out_dir = os.path.join(work, "out")
        with open(manifest, "w", encoding="utf-8") as f:
            for i, chunk in enumerate(chunks):
                row = {"id": f"chunk_{i:04d}", "text": chunk}
                if req.get("reference"):
                    row["reference_audio"] = os.path.abspath(req["reference"])
                    row["reference_text"] = req["reference_text"].strip()
                f.write(json.dumps(row, ensure_ascii=False) + "\n")
        cmd = [
            sys.executable, "audio8_tts_infer.py",
            "--model", self.model,
            "--input-jsonl", manifest,
            "--output-dir", out_dir,
            "--batch-size", "1",
            "--device", self.ctx.device(),
            "--temperature", str(float(opts.get("temperature") or 0.8)),
            "--top-p", str(float(opts.get("top_p") or 0.95)),
        ]
        seed = str(opts.get("seed") or "").strip()
        if seed.isdigit():
            cmd += ["--seed", seed]
        self._run(cmd)
        parts = []
        for i in range(len(chunks)):
            path = os.path.join(out_dir, f"chunk_{i:04d}.wav")
            if not os.path.exists(path):
                raise RuntimeError(f"Audio8 produced no audio for chunk {i + 1}; see the engine log")
            parts.append(read_wav(path))
        rate = parts[0][1]
        shutil.rmtree(work, ignore_errors=True)
        return save_wav(req["output"], concat([p[0] for p in parts], rate), rate)

    def _run(self, cmd):
        proc = subprocess.Popen(
            cmd, cwd=self.ctx.source_dir, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
            text=True, encoding="utf-8", errors="replace",
        )
        tail = []
        for line in proc.stdout:
            line = line.rstrip()
            if line:
                tail = (tail + [line])[-10:]
                self.ctx.log(line)
        if proc.wait() != 0:
            raise RuntimeError("Audio8 inference failed:\n" + "\n".join(tail))

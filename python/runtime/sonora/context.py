"""What an adapter gets from the host: logging, folders and settings."""

import os


class Context:
    def __init__(self, channel, engine_dir):
        self._channel = channel
        self.engine_dir = engine_dir
        self.source_dir = os.path.join(engine_dir, "src")
        self.models_dir = os.path.join(engine_dir, "models")
        os.makedirs(self.models_dir, exist_ok=True)

    def log(self, message):
        """Show a line in the engine's log and status bar."""
        self._channel.send(event="log", message=str(message))

    def progress(self, fraction, message=""):
        """Report progress in [0, 1] for long steps such as downloads."""
        self._channel.send(event="progress", value=float(fraction), message=str(message))

    def device(self):
        """The compute device the user picked: "cuda", "mps" or "cpu"."""
        want = os.environ.get("SONORA_DEVICE", "auto")
        if want == "cpu":
            return "cpu"
        try:
            import torch

            if torch.cuda.is_available():
                return "cuda"
            if getattr(torch.backends, "mps", None) and torch.backends.mps.is_available():
                return "mps"
        except Exception:
            pass
        return "cpu"

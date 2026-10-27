# Writing an engine plugin

Any Python speech model - text to speech or speech to text - can become a Sonora engine
without touching the app's code. A plugin is a folder with two files:

```
<data folder>/plugins/my-engine/
├── engine.json   # what the engine is, how to install it, which settings it has
└── adapter.py    # ~20 lines of Python that load the model and synthesize
```

Find the plugins folder from **Engines → Add engine → Python plugin → Open plugins
folder**, drop your folder in, and press **Reload engines**. Your engine appears with an
**Install** button, exactly like the built-in ones. The built-in engines in
[`python/engines/`](../python/engines) use the same format and are complete examples.

## engine.json

```json
{
  "id": "my-engine",
  "name": "My Engine",
  "vendor": "Me",
  "description": "One or two sentences shown on the engine card.",
  "tags": ["Voice cloning", "CPU"],
  "license": "Apache-2.0",
  "homepage": "https://example.com",
  "size": "≈ 2 GB",
  "accent": ["#f97316", "#ec4899"],

  "python": "3.11",
  "install": [["my-tts-package>=1.0", "soundfile"]],

  "speed": true,
  "clone": "optional",
  "transcript": "optional",
  "default_voice": "alice",
  "voices": [{"id": "alice", "name": "Alice", "group": "English", "description": "Female"}],

  "options": [
    {"id": "model", "label": "Model", "type": "select", "scope": "load", "default": "small",
     "choices": [{"value": "small", "label": "Small"}, {"value": "large", "label": "Large"}]},
    {"id": "temperature", "label": "Temperature", "type": "number", "min": 0, "max": 1.5, "step": 0.05, "default": 0.7}
  ]
}
```

| Field | Meaning |
| --- | --- |
| `install` | pip steps run in order inside the engine's private environment. Each step is a list of `pip install` arguments (`-r file`, `--no-deps`, …). `{src}` expands to the unpacked `source`. PyTorch is matched to the user's GPU automatically. |
| `source.url` | Optional `.zip`/`.tar.gz` of code the adapter needs (e.g. a GitHub archive). Its top folder is stripped and it lands in `src/`. |
| `python` | Python version; uv downloads it if missing. Default `3.11`. |
| `env` | Extra environment variables for install and runtime. |
| `speed` | `true` shows a speed slider and sends `speed`. |
| `clone` | `"optional"` or `"required"` shows the voice-cloning section. |
| `transcript` | Whether the reference recording's transcript is `"optional"` or `"required"`. |
| `voices` | A fixed voice list. Leave it out to report voices from `load()` instead. |
| `hint` | A tip shown under the editor. |
| `model_option` | The `select` option that picks the model, when it isn't called `model`. The engine card then asks which model to download on install (it's loaded once, so its files download), and the Studio defaults to a model that has loaded before. |
| `task` | `"transcribe"` for a speech-to-text engine (see below); leave it out for text to speech. |
| `translate` | Speech-to-text engines: `true` offers *Translate into English*. |
| `languages` | Speech-to-text engines: the language codes it understands, `""` meaning it can detect the language. Leave it out for "any". |
| `options` | Settings rendered as controls: `select` (segments for ≤ 3 choices, a searchable list beyond), `number` (slider over `min`–`max`), `text` (with optional `suggestions`, and `keyboard`: `url`, `number`, `email` or `phone` for the on-screen keyboard on touch screens), `secret`, `bool`. `scope: "load"` means changing it reloads the model. |

## adapter.py

```python
from sonora import save_wav, split_text, concat


class Engine:
    def __init__(self, ctx):
        self.ctx = ctx          # ctx.log(), ctx.progress(), ctx.device(), ctx.models_dir, ctx.source_dir
        self.model = None

    def load(self, options):
        """Load the model. options holds the "load"-scope settings.
        Optionally return {"voices": [{"id": ..., "name": ...}, ...]}."""
        from my_tts import Model

        self.ctx.log("Loading model…")
        self.model = Model(options["model"], device=self.ctx.device())

    def synthesize(self, req):
        """Write req["output"] as a WAV file.

        req has: text, voice, speed, reference, reference_text, options, output."""
        chunks = [self.model.speak(c, voice=req["voice"]) for c in split_text(req["text"], 300)]
        return save_wav(req["output"], concat(chunks, 24000), 24000)
```

The adapter runs in a long-lived worker process, so the model stays in memory between
requests. Anything it prints is shown in the engine's log; exceptions are shown to the user
(the first line in the UI, the full traceback in the log).

Helpers from the `sonora` package:

| Helper | Use |
| --- | --- |
| `save_wav(path, audio, rate)` | Writes numpy / torch / list audio as 16-bit WAV. |
| `read_wav(path)` | Returns `(float32 array, rate)`. |
| `concat(chunks, rate, gap=0.15)` | Joins chunks with short pauses. |
| `split_text(text, max_chars)` | Splits on sentences for models that prefer short inputs. |
| `ctx.device()` | `"cuda"`, `"mps"` or `"cpu"`, following the user's setting. |
| `ctx.log(msg)` / `ctx.progress(fraction, msg)` | Status shown in the app. |

## Engines that aren't Python

Two more kinds need no code at all and are added from **Engines → Add engine**:

- **API server** – anything implementing OpenAI's `POST /v1/audio/speech` (OpenAI,
  Kokoro-FastAPI, LocalAI, openedai-speech, Audio8's server, KittenTTS's Docker image…).
  If the server offers `GET /v1/audio/voices`, its voice list is used.
- **Command line** – any program that can write a WAV file, through a template such as
  `espeak-ng -w {output} -v {voice} -f {text_file}`. Placeholders: `{text}`, `{text_file}`,
  `{output}`, `{voice}`, `{speed}`, `{length_scale}`, `{wpm}`. Without `{output}` the
  program's standard output is read as audio (WAV, or anything ffmpeg can convert).

## Speech-to-text engines

Set `"task": "transcribe"` in engine.json and define `transcribe` instead of
`synthesize`:

```python
class Engine:
    def __init__(self, ctx):
        self.ctx = ctx

    def load(self, options):
        ...  # load the model
        return {}

    def transcribe(self, req):
        # req["audio"] is a 16 kHz mono WAV, req["language"] an ISO code or "" to
        # detect, req["translate"] a bool, req["options"] the engine's settings.
        self.ctx.progress(0.5, "Transcribing…")
        return {"text": "Hello world.", "language": "en",
                "segments": [{"start": 0.0, "end": 1.2, "text": "Hello world."}]}
```

Segments are optional; they give SRT exports their timings. See
[`python/engines/faster-whisper`](../python/engines/faster-whisper) for a complete
engine.

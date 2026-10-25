# Architecture

Sonora is a Go desktop app built with [giode](https://github.com/Mehran9675/giode) (a
retained-style UI library on [Gio](https://gioui.org)). Models run out of process, so the
app stays small, starts instantly, and one misbehaving model can't take the UI down.

```
┌───────────────────────────── Sonora (Go) ─────────────────────────────┐
│ ui/        pages → components → giode                                 │
│ manager/   engine state for the UI, background jobs, friendly errors  │
│ engines/   Engine interface ─┬─ python   → worker.py ⇄ adapter.py     │
│                              ├─ openai   → POST /v1/audio/speech      │
│                              ├─ command  → any CLI writing a WAV      │
│                              ├─ system   → SAPI / say / eSpeak NG     │
│                              └─ onnx     → sherpa-onnx, in process    │
│ pyenv/     uv: downloads Python, one virtualenv per engine            │
│ audio/     WAV codec, peaks, resampling, playback (oto)               │
└───────────────────────────────────────────────────────────────────────┘
```

## Python engines

1. **Install** (`internal/engines/python/install.go`): `pyenv.EnsureUV` downloads
   [uv](https://github.com/astral-sh/uv) on first use; uv creates
   `<data>/engines/<id>/venv` with the requested Python (downloading it if needed) and
   installs the manifest's packages. `UV_TORCH_BACKEND=auto` picks the PyTorch build that
   matches the GPU; if PyTorch's index is unreachable, the step is retried from PyPI.
2. **Run** (`worker.go`): the venv's Python runs `python/runtime/worker.py` with the
   engine's `adapter.py`. They talk JSON lines: `load` (with the load-scope options) and
   `synthesize`; the worker sends `log`/`progress` events. The worker moves stdout to
   stderr at startup so library prints can't corrupt the protocol.
3. The model stays loaded until a load-scope option changes, the device setting changes,
   or the user presses Stop (which kills the worker - a generating model can't be
   interrupted any other way).

The Python sources and built-in engine definitions are embedded in the binary
(`python/embed.go`) and unpacked to `<data>/runtime` at startup, so updates take effect
on the next launch. User plugins live in `<data>/plugins`.

## On-device engines

`internal/engines/onnx` runs Kitten, Kokoro, Piper and ZipVoice models inside the app through
[sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx)'s C API (`onnx/sherpa`, cgo). It is
compiled in only with the `sherpa` build tag, which the Android build uses; without it the
package reports `Available = false` and the engines are hidden. `tools/fetchsherpa`
downloads the prebuilt libraries into `third_party/sherpa-onnx/`.

Models are sherpa-onnx's self-contained packages (model, tokens, espeak-ng data, voices),
downloaded on first use to `<data>/models/ondevice/<package>/`. Each Piper voice is its own
package; Kitten, Kokoro and ZipVoice offer their packages through a load-scope "model"
option. ZipVoice also needs the Vocos vocoder, downloaded once next to the packages, and
gets the reference recording (WAV) and its transcript per request. The model stays in
memory until Eject, and Stop cancels generation through sherpa-onnx's progress callback.

## UI

- `ui/app` wires services and pages; `ui/shell` is the frame (title bar, sidebar).
- `ui/pages/*` are the four screens, each split into one file per part.
- `ui/components/*` are reusable pieces, one per package: picker, waveform, player bar,
  option fields (which render any engine's settings from its manifest), segmented
  control, toggle, badges, chips, log view.
- Background work never touches UI state directly: the manager applies results with
  `App.Post`, so everything the view reads is changed on the UI goroutine.

## Data folder

| OS | Location |
| --- | --- |
| Windows | `%LOCALAPPDATA%\Sonora` |
| macOS | `~/Library/Application Support/Sonora` |
| Linux | `$XDG_DATA_HOME/sonora` (`~/.local/share/sonora`) |

`SONORA_HOME` overrides it. Inside: `settings.json`, `library/` (clips + `index.json`),
`engines/<id>/` (venv, sources, models), `plugins/`, `python/` (uv-managed Pythons),
`bin/uv`. Hugging Face models use the standard Hugging Face cache, so models already
downloaded by other tools are reused.

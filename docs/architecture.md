# Architecture

Sonora is a Go desktop app built with [giode](https://github.com/Mehran9675/giode) (a
retained-style UI library on [Gio](https://gioui.org)). Models run out of process, so the
app stays small, starts instantly, and one misbehaving model can't take the UI down.

```
┌───────────────────────────── Sonora (Go) ─────────────────────────────┐
│ ui/        pages → components → giode                                 │
│ manager/   engine state for the UI, background jobs, friendly errors  │
│ engines/   Engine interface ─┬─ python   → worker.py ⇄ adapter.py     │
│            (+ Transcriber)   ├─ openai   → /v1/audio/speech, …/trans… │
│                              ├─ command  → any CLI writing a WAV      │
│                              ├─ system   → SAPI / say / eSpeak NG     │
│                              └─ onnx     → sherpa-onnx, in process    │
│ pyenv/     uv: downloads Python, one virtualenv per engine            │
│ audio/     WAV + Opus codecs, decoding (MP3/FLAC/Ogg/ffmpeg), playback│
│ textfile/  documents → script text;  transcripts/ the transcript store│
└───────────────────────────────────────────────────────────────────────┘
```

## Python engines

1. **Install** (`internal/engines/python/install.go`): `pyenv.EnsureUV` downloads
   [uv](https://github.com/astral-sh/uv) on first use; uv creates
   `<data>/engines/<id>/venv` with the requested Python (downloading it if needed) and
   installs the manifest's packages. `UV_TORCH_BACKEND=auto` picks the PyTorch build that
   matches the GPU; if PyTorch's index is unreachable, the step is retried from PyPI.
2. **Run** (`worker.go`): the venv's Python runs `python/runtime/worker.py` with the
   engine's `adapter.py`. They talk JSON lines: `load` (with the load-scope options),
   `synthesize`, and `transcribe` for speech-to-text engines; the worker sends
   `log`/`progress` events. The worker moves stdout to
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
compiled in on Android, and on desktops with the `sherpa` build tag; without it the
package reports `Available = false` and the engines are hidden. `tools/fetchsherpa`
downloads the prebuilt libraries into `third_party/sherpa-onnx/`.

Models are sherpa-onnx's self-contained packages (model, tokens, espeak-ng data, voices),
downloaded on first use to `<data>/models/ondevice/<package>/`. Each Piper voice is its own
package; Kitten, Kokoro and ZipVoice offer their packages through a load-scope "model"
option. ZipVoice also needs the Vocos vocoder, downloaded once next to the packages, and
gets the reference recording (WAV) and its transcript per request. The model stays in
memory until Eject, and Stop cancels generation through sherpa-onnx's progress callback.

## Models

An engine whose model is a choice (a `select` option named `model`, or the manifest's
`model_option`) implements `engines.ModelStore`: which choices are downloaded, and
`InstallModel` to fetch one - an on-device package, or for a Python engine a first load
(recorded in `models.json`). The engine card asks which model to download on install and
offers more later; `manager.Values` resolves an untouched model setting to a downloaded
model, so the Studio and Transcribe pages start on one, while a model picked there
downloads on first use.

## Speech to text

An engine whose manifest has `task: "transcribe"` also implements
`engines.Transcriber`. The manager (`manager/transcribe.go`) first decodes the picked or
recorded file with `audio/decode` - WAV, MP3, FLAC, Ogg Vorbis and Opus in Go, anything
else through ffmpeg - resampling to 16 kHz mono block by block into a temporary WAV, so
every engine gets the same 16 kHz WAV. The transcript keeps an Opus copy of the audio
(24 kbps) in `<data>/transcripts`.

The microphone (`audio/record`) captures 48 kHz mono and a goroutine encodes it to an
Ogg Opus file as it arrives (`audio/opusfile`, pure Go via gopus), so a recording's length
costs disk, not memory. Cloning references recorded this way are converted to a
temporary WAV before an engine sees them.

On device, `onnx/asr.go` cuts the audio into utterances with Silero VAD (each widened a
little so first syllables survive, at most 20 s for Whisper) and recognizes them one by
one, which gives the segments their timestamps; Whisper's language and task are fixed
when the model loads, so changing them reloads it. Transcripts export as text, SRT or
WebVTT (`transcripts/export.go`).

## UI

- `ui/app` wires services and pages; `ui/shell` is the frame (title bar, sidebar).
- `ui/pages/transcribe` is the speech-to-text page; the Engines page shows voice and
  transcription engines on two tabs.
- `ui/pages/*` are the four screens, each split into one file per part.
- `ui/components/*` are reusable pieces, one per package: picker, waveform, player bar,
  option fields (which render any engine's settings from its manifest), segmented
  control, toggle, badges, chips, log view.
- `ui/tr` translates the UI (giode's `i18n` underneath). Messages are keyed by their English
  text: `tr.T("Install")`, `tr.T("Downloading {file}", "file", name)`, `tr.N("{n} words", n)`.
  Text drawn every frame calls `tr.T` there; elements built once take theirs through
  `tr.Bind`, which sets it again when the language changes; English kept in tables is
  marked `tr.Msg` and translated where it's shown. Engine descriptions, option labels and
  the manager's status messages are English data, translated where they're shown.
  Catalogs are `ui/tr/locales/<code>.json` (English needs only plural forms).
  `go run ./tools/i18nkeys -missing de` lists what a catalog lacks, and `ui/tr`'s test fails
  when a catalog misses a key, changes a placeholder or lacks a plural form.
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

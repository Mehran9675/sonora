# Sonora

**A desktop studio for text-to-speech models — no code, no terminal, no Python setup.**

Type some text, pick an engine and a voice, press **Generate**. Sonora installs models
like **Kitten TTS 2** and **Audio8 TTS** with one click, keeps every clip in a library,
and can drive almost any other TTS model: Python models through a small plugin, servers
through the OpenAI-compatible speech API, and command-line programs through a template.

Built in Go with [giode](https://github.com/Mehran9675/giode). Runs on Windows, macOS
and Linux.

![Studio](docs/screenshots/studio.png)

| Engines | Voices | Library (light theme) |
| --- | --- | --- |
| ![Engines](docs/screenshots/engines.png) | ![Voices](docs/screenshots/voices.png) | ![Library](docs/screenshots/library-light.png) |

## Features

- **One-click engines.** Sonora downloads its own Python (via [uv](https://github.com/astral-sh/uv)),
  creates a private environment per engine, picks the right PyTorch for your GPU, and
  shows progress and logs while it installs. Nothing to configure.
- **Studio.** A large script editor, searchable voice picker, speed control and every
  engine setting rendered as proper controls. Voice cloning from a recording for engines
  that support it.
- **Player.** A waveform you can click to seek, play/pause, save as WAV, show in folder.
- **Library.** Every clip is kept with its text, engine and voice. Search, replay, export,
  or load a clip's settings back into the Studio.
- **Any model.** Built-in engines, Python plugins, API servers and command-line tools all
  appear in the same list and work the same way.
- **Modern design.** Dark and light themes, a frameless window with a custom title bar,
  the Inter typeface, a violet-to-cyan accent.

## Engines

| Engine | What it's good at | Cloning | Runs on |
| --- | --- | --- | --- |
| **Kitten TTS 2** (KittenML) | 47 voices, 20 languages, emotion tags | ✓ | CPU · GPU · Apple silicon |
| **Audio8 TTS** (Audio8) | 0.6B multilingual, 11 languages | ✓ | CPU · GPU |
| **Kitten TTS Lite** | 15–80M models, tiny and fast | – | CPU |
| **Kokoro** | Natural voices in 8 languages, 82M | – | CPU · GPU |
| **Chatterbox** (Resemble AI) | Strong cloning, emotion control | ✓ | GPU recommended |
| **Piper** | Real-time voices in 40+ languages | – | CPU |
| **Coqui TTS** | XTTS-v2 or any of 100+ Coqui models by name | ✓ | CPU · GPU |
| **Hugging Face models** | Any Transformers TTS model by ID (MMS, Bark, VITS…) | – | CPU · GPU |
| **System voices** | Your OS's built-in voices, offline, instant | – | – |
| **API server** *(add your own)* | OpenAI, Kokoro-FastAPI, LocalAI, Audio8's server… | – | remote |
| **Command line** *(add your own)* | espeak-ng, piper CLI, `say`, anything that writes WAV | – | local |

Each model has its own license (some, like XTTS-v2, are non-commercial). The engine card
shows it; check it before commercial use.

## Getting started

### Download

Grab the build for your OS from the CI artifacts / releases and run it. That's it: the
first time you install a Python engine, Sonora downloads its toolchain (~30 MB) by itself.

- **Linux:** System voices need `espeak-ng` (`sudo apt install espeak-ng`); file dialogs
  use `zenity` or `kdialog` when available.
- **macOS / Windows:** nothing extra.
- **Optional:** `ffmpeg` lets Sonora accept MP3/OGG from API servers and command-line tools.

### Build from source

Requires Go 1.26+.

```sh
# Linux only: Gio and audio dependencies
sudo apt install gcc pkg-config libwayland-dev libx11-dev libx11-xcb-dev \
  libxkbcommon-x11-dev libgles2-mesa-dev libegl1-mesa-dev libffi-dev \
  libxcursor-dev libvulkan-dev libasound2-dev

go build -o sonora ./cmd/sonora                              # Linux / macOS
go build -ldflags "-H windowsgui" -o sonora.exe ./cmd/sonora  # Windows
```

## Using it

1. **Engines** → press **Install** on an engine (or use *System Voices* right away).
2. **Studio** → write or paste text, choose the engine and voice, press **Generate speech**.
3. Click the waveform to seek; use the download button to save a WAV.
4. **Library** keeps every clip; ↺ loads one back into the Studio.

Settings lets you switch theme, force CPU or GPU, and add a Hugging Face token for gated
models.

## Adding other models

- **A server you run** (or a cloud API): *Engines → Add engine → API server*, enter its
  URL. Voices are read from the server when it lists them.
- **A program on your computer**: *Add engine → Command line*, e.g.
  `espeak-ng -w {output} -v {voice} -f {text_file}`.
- **A Python model**: write an `engine.json` and a ~20-line `adapter.py`, drop them in the
  plugins folder. See [docs/plugins.md](docs/plugins.md).

## Project layout

```
cmd/sonora/            entry point
ui/
  app/                 wiring: services, pages, window
  shell/               title bar, sidebar, frame
  pages/               studio/ engines/ library/ settings/ (one file per part)
  components/          one package per component (picker, waveform, playerbar, …)
  core/ theme/ icons/  shared context, palette & fonts, icon paths
internal/
  engines/             Engine interface + python/ openai/ command/ system/ backends
  manager/             engine state for the UI, jobs, friendly errors
  catalog/             finds built-in engines and plugins
  pyenv/               uv toolchain, virtualenvs, downloads
  audio/               clip ops, wav/, playback/, convert/
  library/ settings/ paths/ sysdlg/
python/
  runtime/             worker.py + the `sonora` helper package
  engines/<id>/        engine.json + adapter.py for each built-in engine
docs/                  architecture, plugin guide, screenshots
```

See [docs/architecture.md](docs/architecture.md) for how the pieces fit together.

## Development

```sh
go test ./...          # unit + integration tests (the worker test uses python3 if present)
GIODE_MCP=1 ./sonora   # exposes the UI over MCP (giode) for automated UI testing
SONORA_HOME=/tmp/s ./sonora   # use a throwaway data folder
```

## Credits

[giode](https://github.com/Mehran9675/giode) and [Gio](https://gioui.org) for the UI,
[oto](https://github.com/ebitengine/oto) for audio, [uv](https://github.com/astral-sh/uv)
for Python, [Inter](https://rsms.me/inter/) (SIL OFL) for type, and the authors of every
model Sonora runs.

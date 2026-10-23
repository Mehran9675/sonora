"""Audio helpers that only need the standard library and numpy."""

import wave


def _to_numpy(audio):
    import numpy as np

    if hasattr(audio, "detach"):  # torch tensor
        audio = audio.detach().cpu().float().numpy()
    arr = np.asarray(audio, dtype=np.float32)
    arr = np.squeeze(arr)
    if arr.ndim > 1:  # (channels, n) or (n, channels): mix down
        arr = arr.mean(axis=0 if arr.shape[0] < arr.shape[-1] else 1)
    return arr


def save_wav(path, audio, sample_rate):
    """Write float audio (numpy, torch or list) as 16-bit mono PCM."""
    import numpy as np

    arr = _to_numpy(audio)
    if arr.size == 0:
        raise RuntimeError("the model returned no audio")
    peak = float(np.max(np.abs(arr)))
    if peak > 1.0:
        arr = arr / peak
    pcm = (np.clip(arr, -1.0, 1.0) * 32767).astype("<i2")
    with wave.open(path, "wb") as w:
        w.setnchannels(1)
        w.setsampwidth(2)
        w.setframerate(int(sample_rate))
        w.writeframes(pcm.tobytes())
    return {"sample_rate": int(sample_rate), "samples": int(arr.size)}


def read_wav(path):
    """Read a WAV as (float32 mono numpy array, sample_rate). Uses
    soundfile when installed (any encoding), else 8/16/32-bit PCM."""
    import numpy as np

    try:
        import soundfile

        data, rate = soundfile.read(path, dtype="float32", always_2d=True)
        return data.mean(axis=1), rate
    except ImportError:
        pass
    with wave.open(path, "rb") as w:
        rate, channels, width = w.getframerate(), w.getnchannels(), w.getsampwidth()
        raw = w.readframes(w.getnframes())
    dtype = {1: np.uint8, 2: "<i2", 4: "<i4"}[width]
    arr = np.frombuffer(raw, dtype=dtype).astype(np.float32)
    if width == 1:
        arr = (arr - 128) / 128
    else:
        arr /= float(2 ** (8 * width - 1))
    if channels > 1:
        arr = arr.reshape(-1, channels).mean(axis=1)
    return arr, rate


def concat(chunks, sample_rate, gap=0.15):
    """Join audio chunks with a short pause between them."""
    import numpy as np

    silence = np.zeros(int(gap * sample_rate), dtype=np.float32)
    parts = []
    for i, c in enumerate(chunks):
        if i:
            parts.append(silence)
        parts.append(_to_numpy(c))
    return np.concatenate(parts) if parts else np.zeros(0, dtype=np.float32)

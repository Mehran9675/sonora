"""Text helpers for engines that work best on short inputs."""

import re

_SENTENCE = re.compile(r"(?<=[.!?。！？;:])\s+|\n+")


def split_text(text, max_chars=300):
    """Split text into chunks of whole sentences, each at most max_chars
    where possible (a single over-long sentence is split on spaces)."""
    chunks, current = [], ""
    for sentence in (s.strip() for s in _SENTENCE.split(text)):
        if not sentence:
            continue
        while len(sentence) > max_chars:
            cut = sentence.rfind(" ", 0, max_chars)
            cut = cut if cut > max_chars // 2 else max_chars
            head, sentence = sentence[:cut].strip(), sentence[cut:].strip()
            if current:
                chunks.append(current)
                current = ""
            chunks.append(head)
        if current and len(current) + 1 + len(sentence) > max_chars:
            chunks.append(current)
            current = sentence
        else:
            current = f"{current} {sentence}".strip()
    if current:
        chunks.append(current)
    return chunks

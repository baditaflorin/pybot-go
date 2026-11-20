# Linux retrieval sample

On 2026-11-20, PyBot was run against 100 of the highest-voted questions tagged `linux` on [Unix & Linux Stack Exchange](https://unix.stackexchange.com/). The question list came from the [Stack Exchange API](https://api.stackexchange.com/2.3/questions?order=desc&sort=votes&tagged=linux&site=unix.stackexchange.com&pagesize=100). Each title was sent to a fresh `pybot.New(0).Reply` instance.

After tightening topic routing and adding focused Linux articles, all 100 questions returned a topic-matched, source-linked answer; none fell through to the unknown route. The sample caught and helped fix unrelated matches for empty-directory deletion, Wayland/X11 session checks, cache questions, `top` memory units, and other command topics.

This is a retrieval coverage and relevance check, not a 100-question independent correctness certification. The questions are a vote-ranked snapshot, may span multiple distributions and software versions, and do not represent every Linux task. Answers are curated and linked to manuals so users can inspect the underlying guidance. Unsupported or ambiguous Linux topics should still abstain instead of borrowing an answer from another topic.

To inspect the current response for one question locally:

```sh
go run ./cmd/pybot -p 'How do I make ls show file sizes in megabytes?' --verbose
```

The Stack Exchange API result can change over time, so this historical sample is not guaranteed to reproduce the same 100 titles later.

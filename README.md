# PyBot Go

PyBot Go is a small, CPU-only assistant inspired by [Hugging Face's PyBot](https://github.com/huggingface/pybot). It combines typed conversation memory, deterministic solvers, source-backed retrieval, and an optional pretrained transformer embedding router in the browser. It does not use a generative LLM, GPU, or third-party Go package.

PyBot can answer open-topic factual questions through cited Wikipedia retrieval, along with deterministic calculations, conversational memory, and source-backed Linux guidance. The 100 example conversations help test phrasing and follow-ups; they do not limit which subjects can be searched.

## Try the live browser demo

Open [pybot.0sample.com](https://pybot.0sample.com). The Go assistant and quantized MiniLM text embedder run in your browser on CPU. The model is downloaded from Hugging Face on first use (about 23 MB) and cached by the browser. It compares questions with inspectable intent examples and uses CPU embeddings to rerank Wikipedia article introductions and sentences. The 100 two-turn examples cover geography, science, history, culture, and everyday facts; they are demonstrations, not a topic whitelist, canned answers, or training for a text generator. Factual answers come from cited Wikipedia extracts. The latest 24 user/assistant turns are saved in this tab's `sessionStorage` and restored on reload. Starting a new chat clears the saved conversation. Nothing is sent to a PyBot server; open-topic questions use Wikipedia and show a link to the source.

## What it can do

- **Compare ages:** remember names and age statements for the current conversation, compare two or more people, and report the age difference.
- **Follow up on ages:** add another person and place them in the remembered age order, including how many years older or younger they are than nearby people.
- **Calculate arithmetic:** evaluate numeric and supported spoken-number expressions with parentheses and standard operator precedence. Arithmetic uses exact rational values; square roots remain exact when possible and otherwise return a decimal approximation.
- **Convert units:** convert common length, mass, volume, and temperature units, and reject incompatible or unsupported units.
- **Work with dates:** calculate weekday/date offsets, date differences, and relative dates from today. Explicit dates use `YYYY-MM-DD`.
- **Help with Linux administration:** retrieve source-linked guidance for common service failures, logs, storage, processes, CPU and memory, networking, permissions, packages, mounts, SSH, firewalls, Docker storage, and desktop troubleshooting. Linux retrieval is restricted to matching topics and abstains when it cannot find a relevant article. Start with read-only diagnostics, explain distro differences, and ask for the distribution, version, and error output when needed. PyBot does not execute commands.
- **Answer sourced questions:** search an embedded, curated library about Linux administration and basic Linux concepts, onions, one named electric kettle model, and egg-dish food safety. For general facts and geography, the browser routes natural questions and follow-ups to a short Wikipedia introduction with a source link. The kettle instructions are explicitly model-specific; egg cooking time is not inferred from a food-safety temperature source.
- **Handle multi-turn chat:** follow up on the last answer, repeat or shorten it, explain commands, show its sources, resolve factual follow-ups using the prior question and cited topic, summarize recent topics, and return to an earlier answer by topic words.
- **Personalize replies:** ask for brief or detailed answers, beginner explanations, bullets, or numbered lists. Preferences are scoped to the current chat and can be reviewed or cleared.
- **Work through Linux steps:** enable one-step-at-a-time guidance, request the next or current step, ask what a numbered step means, show remaining steps, and share command output when something fails.
- **Abstain:** ask for missing information or return a clear fallback when no supported answer is available.

The Go router includes a small multinomial Naive Bayes classifier trained from inspectable examples. The browser can also use cosine similarity between CPU transformer embeddings and example prompts to route semantic paraphrases and follow-ups. The two-turn examples are in [`site/factual-conversations.json`](site/factual-conversations.json); they are authored test-like examples, not collected user transcripts. Neither router generates answers or establishes facts. Explicit Go solvers perform calculations, and factual answers come from the bundled corpus or the cited Wikipedia extract. Similarity scores are routing signals, not calibrated probabilities.

## Try it

Requires Go 1.22 or newer.

```sh
go run ./cmd/pybot
```

Ask one question and exit:

```sh
go run ./cmd/pybot -p "Alice is 12 and Bob is 50. Who is older?"
```

Or keep facts across piped messages in one process:

```sh
printf 'Alice is 12.\nBob is 50.\nWho is older?\n' | go run ./cmd/pybot
```

Other examples:

```sh
go run ./cmd/pybot -p "What is (18 + 6) / 3?"
go run ./cmd/pybot -p "How many inches are in 2 meters?"
go run ./cmd/pybot -p "What date is 30 days after 2026-11-18?"
go run ./cmd/pybot -p "Why do onions make my eyes water?"
```

Build or install the command:

```sh
go build -o pybot ./cmd/pybot
go install github.com/baditaflorin/pybot-go/cmd/pybot@latest
```

Build the static GitHub Pages bundle locally:

```sh
./scripts/build-pages.sh
python3 -m http.server --directory docs 8080
```

Then open `http://localhost:8080`. The build script creates the published `docs/` files, including the Go runtime bundle and the matching `wasm_exec.js` from the installed Go toolchain.

## CLI

| Option | Behavior |
|---|---|
| `-p TEXT`, `--prompt TEXT` | Answer one message and exit |
| `-i`, `--interactive` | With `-p`, continue into an interactive conversation |
| `-s N`, `--seed N` | Make wording choices repeatable |
| `-v`, `--verbose` | Print the selected module and response time |
| `--version` | Print the Go implementation version |

Interactive commands: `/reset` clears conversation memory, `/context` summarizes remembered facts, preferences, and recent topics, `/forget name|age|distro|preferences` removes a remembered detail, `/help` shows commands, and `/exit` quits. Piped input gets one response per line while keeping the same conversation state for that process. Retrieval answers print their source links below the answer.

## Architecture and boundaries

`Bot.Reply` routes each message through a small planner. Exact solvers handle arithmetic, square roots, unit conversion, dates, age comparisons, and age ranking. The Go CLI works without pretrained models or a network. The browser adds a pinned, quantized MiniLM model running through Transformers.js and WebAssembly on CPU; its similarity scores help route intent and rerank retrieved Wikipedia pages and sentences, but never generate answers. Facts are typed, attributed to the message that supplied them, and scoped to one `Bot` instance. `Reset` clears them. The browser does not send chat to a PyBot server. Open-topic factual questions use Wikipedia's public API and show the matching source.

The embedded search uses a BM25-style ranker over topic-scoped curated articles. It returns document answers and source URLs, not an open-ended generated summary. The corpus is in [`knowledge/docs.json`](knowledge/docs.json); each entry needs a source title and valid HTTP(S) source URL. The 100-question Linux retrieval sample and its limits are documented in [`docs/linux-benchmark.md`](docs/linux-benchmark.md). The first 50 conversation improvements are listed in [`docs/conversation-improvements.md`](docs/conversation-improvements.md), and the next 50 in [`docs/conversation-improvements-v0.8.md`](docs/conversation-improvements-v0.8.md). The 100 factual conversation examples route through the browser's CPU embedder and use live, cited Wikipedia extracts. The included egg source supports a safe temperature for egg dishes, not universal boiling times. The kettle answer applies only to the linked KRUPS Control Line model.

This design trades generated prose for CPU inference, reproducible calculations, inspectable routing, and explicit evidence. Embeddings help match a question to retrieved sources; they do not establish facts or remove ambiguity. General factual answers depend on Wikipedia availability, while Linux guidance still depends on its curated library. It can retrieve across topics, but it will not handle arbitrary conversation like a generative language model, and its fixed regression examples are not a broad accuracy benchmark.

## Develop and evaluate

```sh
gofmt -w $(find . -type f -name '*.go')
go test ./...
go test -race ./...
go vet ./...
go build ./...
```

Tests cover held-out intent paraphrases and end-to-end behavior for memory, calculations, conversions, dates, retrieval sources, abstention, and the CLI. The held-out set is small and hand-authored; passing it is a capability regression check, not evidence of general conversational accuracy.

## Attribution and license

This is an independent Go implementation inspired by [Hugging Face's PyBot](https://github.com/huggingface/pybot), which is licensed under Apache-2.0. This repository is also licensed under Apache-2.0; see [LICENSE](LICENSE) and [NOTICE](NOTICE).

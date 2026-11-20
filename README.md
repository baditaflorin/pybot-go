# PyBot Go

PyBot Go is a small, CPU-only assistant inspired by [Hugging Face's PyBot](https://github.com/huggingface/pybot). It combines typed conversation memory, deterministic solvers, source-backed retrieval, and an optional pretrained transformer embedding router in the browser. It does not use a generative LLM, GPU, or third-party Go package.

This is a deliberately bounded assistant, not an attempt to reproduce the upstream Python experiment or its large generated knowledge base. It is designed to answer a useful set of questions reliably and to say when it does not know.

## Try the live browser demo

Open [pybot.0sample.com](https://pybot.0sample.com). The Go assistant and quantized MiniLM text embedder run in your browser on CPU. The model is downloaded from Hugging Face on first use (about 23 MB) and cached by the browser. It compares question embeddings with a small set of intent examples; a match selects a Go solver or retrieval route and never generates an answer. Chat and conversation facts stay in the tab. If the local corpus cannot answer a factual question, the page sends only that current question to Wikipedia and shows a link to the source. Starting a new chat or reloading clears conversation memory.

## What it can do

- **Compare ages:** remember names and age statements for the current conversation, compare two or more people, and report the age difference.
- **Follow up on ages:** add another person and place them in the remembered age order, including how many years older or younger they are than nearby people.
- **Calculate arithmetic:** evaluate numeric and supported spoken-number expressions with parentheses and standard operator precedence. Arithmetic uses exact rational values; square roots remain exact when possible and otherwise return a decimal approximation.
- **Convert units:** convert common length, mass, volume, and temperature units, and reject incompatible or unsupported units.
- **Work with dates:** calculate weekday/date offsets, date differences, and relative dates from today. Explicit dates use `YYYY-MM-DD`.
- **Help with Linux administration:** retrieve sourced guidance for service failures, system and application logs, disk space, large files, CPU and memory pressure, listening ports, networking and DNS, permissions, packages, mounts, SSH, firewalls, and Docker storage. Start with read-only diagnostics, explain distro differences, and ask for the distribution, version, and error output when needed. PyBot does not execute commands.
- **Answer sourced questions:** search an embedded local corpus about Linux administration, onions, one named electric kettle model, and egg-dish food safety. For other factual questions, the browser can retrieve a short Wikipedia introduction and link its source. The kettle instructions are explicitly model-specific; egg cooking time is not inferred from a food-safety temperature source.
- **Handle conversation basics:** recognize a small set of greetings, thanks, identity questions, name questions, and short follow-ups.
- **Abstain:** ask for missing information or return a clear fallback when no supported answer is available.

The Go router includes a small multinomial Naive Bayes classifier trained from inspectable examples. The browser can also use cosine similarity between CPU transformer embeddings and example prompts to route semantic paraphrases and follow-ups. Neither router generates answers or establishes facts. Explicit Go solvers perform calculations, and retrieval answers come from the bundled corpus or the cited Wikipedia extract. Similarity scores are routing signals, not calibrated probabilities.

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

Interactive commands: `/reset` clears conversation memory, `/help` shows commands, and `/exit` quits. Piped input gets one response per line while keeping the same conversation state for that process. Retrieval answers print their source links below the answer.

## Architecture and boundaries

`Bot.Reply` routes each message through a small planner. Exact solvers handle arithmetic, square roots, unit conversion, dates, age comparisons, and age ranking. The Go CLI works without pretrained models or a network. The browser adds a pinned, quantized MiniLM model running through Transformers.js and WebAssembly on CPU; its output is only a similarity-based intent hint. Facts are typed, attributed to the message that supplied them, and scoped to one `Bot` instance. `Reset` clears them. The browser does not send chat to a PyBot server. For an unmatched factual question, it sends the current question to Wikipedia's public API and shows the matching source.

The embedded search uses a BM25-style ranker over a few curated documents. It returns document answers and source URLs, not an open-ended generated summary. The corpus is in [`knowledge/docs.json`](knowledge/docs.json); each entry needs a source title and valid HTTP(S) source URL. The included egg source supports a safe temperature for egg dishes, not universal boiling times. The kettle answer applies only to the linked KRUPS Control Line model.

This design trades broad language coverage for CPU inference, reproducible calculations, inspectable routing, and explicit evidence. Embeddings improve wording and follow-up matching, but do not remove ambiguity or supply world knowledge by themselves. It still needs examples and curated material for new intents and topics. Wikipedia lookup improves general factual coverage but depends on network availability and sends the current question to that service. It will not handle arbitrary conversation like a generative language model, and its fixed paraphrase checks are regression tests rather than a broad benchmark.

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

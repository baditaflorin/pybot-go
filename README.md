# PyBot Go

PyBot Go is a small, CPU-only conversational assistant inspired by [Hugging Face's PyBot](https://github.com/huggingface/pybot). It combines a statistical intent router, typed conversation memory, exact task solvers, and source-backed local search. It does not use an LLM, GPU, network service, or third-party Go package.

This is a deliberately bounded assistant, not an attempt to reproduce the upstream Python experiment or its large generated knowledge base. It is designed to answer a useful set of questions reliably and to say when it does not know.

## What it can do

- **Compare ages:** remember names and age statements for the current conversation, compare two or more people, and report the age difference.
- **Calculate arithmetic:** evaluate numeric and supported spoken-number expressions with parentheses and standard operator precedence. Arithmetic uses exact rational values.
- **Convert units:** convert common length, mass, volume, and temperature units, and reject incompatible or unsupported units.
- **Work with dates:** calculate weekday/date offsets, date differences, and relative dates from today. Explicit dates use `YYYY-MM-DD`.
- **Answer a few sourced questions:** search an embedded local corpus about onions, one named electric kettle model, and egg-dish food safety. Answers include their source links. The kettle instructions are explicitly model-specific; egg cooking time is not inferred from a food-safety temperature source.
- **Handle conversation basics:** recognize a small set of greetings, thanks, identity questions, name questions, and short follow-ups.
- **Abstain:** ask for missing information or return a clear fallback when no supported answer is available.

The intent router is a small multinomial Naive Bayes classifier trained from inspectable, hand-written examples. It helps route paraphrases; it does not generate answers or establish facts. Explicit solvers perform calculations, and retrieval answers come from the bundled corpus. The classifier score is a routing signal, not a calibrated confidence value.

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

`Bot.Reply` routes each message through a small planner. Exact solvers handle arithmetic, unit conversion, dates, and age comparisons. The intent model is intentionally small and inspectable; it uses text features and hand-written examples rather than pretrained embeddings. Facts are typed, attributed to the message that supplied them, and scoped to one `Bot` instance. `Reset` clears them. Nothing is persisted or sent over the network.

The embedded search uses a BM25-style ranker over a few curated documents. It returns document answers and source URLs, not an open-ended generated summary. The corpus is in [`knowledge/docs.json`](knowledge/docs.json); each entry needs a source title and valid HTTP(S) source URL. The included egg source supports a safe temperature for egg dishes, not universal boiling times. The kettle answer applies only to the linked KRUPS Control Line model.

This design trades broad language coverage for low resource use, reproducible calculations, inspectable routing, and explicit evidence. It still needs examples and curated material for new intents and topics. It will not handle arbitrary conversation or general knowledge like a large language model, and its fixed paraphrase checks are regression tests rather than a broad benchmark.

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

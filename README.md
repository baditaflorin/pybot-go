# PyBot Go

A small, standalone Go reimplementation inspired by [Hugging Face's PyBot](https://github.com/huggingface/pybot). It uses explicit rules and conversation state; it does not use a language model, network service, or third-party Go package.

## Why this exists

The original PyBot is an experiment in how far a large swarm of coding agents can push a rule-based chatbot. Its README reports a 26% win rate on SmolTalkBench, a benchmark where an LLM judge compares replies with reference answers. That is an interesting result for a code-generation experiment, but it is not evidence that PyBot is a dependable general-purpose assistant. A fixed collection of rules can sound fluent in cases it recognizes and still fail on basic reasoning it was never taught.

This Go project starts with a smaller, explicit goal: make a lightweight chatbot that handles a few useful tasks predictably, including remembering names and ages during a conversation and comparing ages with a clear explanation. It is a compact Go-native implementation, not a line-by-line translation of the upstream project's roughly 640,000 lines of Python or a port of its full knowledge base.

## Use it

Requires Go 1.22 or newer.

```sh
go run ./cmd/pybot
```

Ask one question and exit:

```sh
go run ./cmd/pybot -p "Alice is 12 and Bob is 50. Who is older?"
```

The answer identifies Bob and gives the 38-year difference. Facts also persist across piped lines:

```sh
printf 'Alice is 12.\nBob is 50.\nWho is older?\n' | go run ./cmd/pybot
```

Build a binary or install the command:

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
| `-v`, `--verbose` | Print the selected rule module and response time |
| `--version` | Print the Go implementation version |

Interactive commands: `/reset` clears conversation memory, `/help` shows commands, and `/exit` quits. With input piped to stdin, PyBot prints one reply per line and keeps the same conversation memory for the duration of the process.

## Current behavior

- Remembers a name stated as `I'm Sam` or `my name is Sam`.
- Remembers age statements such as `Alice is 12 years old` and answers `Who is older?` or `Who is younger?`, including the numeric difference.
- Includes a small set of direct rules for cooking questions, greetings, and follow-ups.
- Uses a stable score-based arbiter so more specific rules beat general conversation rules.
- Keeps each conversation's state and random source inside one `Bot`; `Reply` and `Reset` are safe for concurrent calls.

This is intentionally a small first release. It will not understand arbitrary phrasing or answer the broad range of questions supported by a large language model. Unknown questions receive an explicit fallback instead of a fabricated factual answer. The upstream Python project remains the source for the larger swarm-generated experiment and its benchmark results.

## Attribution and license

This project is an independent Go implementation inspired by Hugging Face's PyBot. The original project is available at [huggingface/pybot](https://github.com/huggingface/pybot) and is licensed under Apache-2.0. This repository is also licensed under Apache-2.0; see [LICENSE](LICENSE) and [NOTICE](NOTICE).

## A path beyond fixed rules

A much stronger non-LLM assistant should not be built by adding thousands of unrelated phrase checks. The next architecture can combine:

1. **Language understanding:** normalize text, identify intent and entities, and use a small CPU-friendly classifier trained on examples to recognize paraphrases. It routes a message; it does not invent the answer.
2. **Structured memory:** store typed facts such as `person: Alice`, `age: 12 years`, and where each fact came from. Keep facts scoped to a conversation unless the user explicitly saves them.
3. **A query planner and solvers:** turn a question into operations over facts—compare ages, calculate totals, convert units, or reason over dates—and produce a checkable result.
4. **Local retrieval:** search a curated, source-backed knowledge base for questions that need facts. Retrieval can return the supporting passage, while deterministic code handles calculations and comparisons.
5. **Uncertainty handling:** ask one targeted question when a name, value, or reference is ambiguous; say when the local knowledge base has no answer.
6. **Capability evaluation:** measure exact answer accuracy and appropriate clarification on a fixed set of realistic prompts, including paraphrases and follow-ups. A broad benchmark score alone can hide failures on simple tasks.

This approach can stay CPU-only and much lighter than a language model. It improves coverage and composition while keeping answers inspectable. The trade-off is that the system still needs examples, a maintained knowledge base, and explicit support for each kind of reasoning; it will remain weaker than an LLM on open-ended questions. A useful target is high accuracy in a chosen set of tasks, rather than claiming universal conversation ability.

# Changelog

## 0.5.2 - 2026-11-20

- Answer capability questions directly with a concrete list of supported tasks and limits instead of sending them to factual lookup.

## 0.5.1 - 2026-11-20

- Version browser asset URLs so repeat visitors do not keep using cached pre-embedding JavaScript or WebAssembly after an update.

## 0.5.0 - 2026-11-20

- Add an optional, quantized MiniLM embedding router that runs on CPU in the browser and routes follow-ups to existing Go solvers without generating answers.
- Rank remembered people by age, including the exact George-between-Alice-and-Bob follow-up and age differences.
- Solve square roots exactly for perfect squares and with a clear decimal approximation for irrational results.
- Add cited Wikipedia lookup for factual questions outside the small embedded corpus; only the current question is sent for that lookup.
- Document model download, browser processing, and external factual lookup behavior.

## 0.4.0 - 2026-11-20

- Add a browser chat interface that runs the Go bot locally as WebAssembly.
- Add age, arithmetic, unit-conversion, and sourced-question examples; show route details and source links in the UI.
- Add a reproducible GitHub Pages build for `https://pybot.0sample.com` with no chat API or external runtime dependency.

## 0.3.0 - 2026-11-18

- Replace generic rule scoring with a small, inspectable Naive Bayes intent router and a conservative query planner.
- Add conversation-scoped typed facts with age comparison, exact arithmetic, unit conversion, and date solvers.
- Add BM25-style local retrieval with source links for a small curated corpus, plus explicit clarification and abstention.
- Print evidence links in the CLI and add regression coverage for core capabilities and paraphrase routing.
- Expand the Woodpecker pipeline to format, test, vet, and build the full Go project.

## 0.2.0 - 2026-11-18

- Add a CPU-only Naive Bayes intent router trained from a small built-in phrase set.
- Use statistical routing alongside explicit handlers for names, age comparisons, cooking questions, and conversation.
- Document the hybrid direction and its current evaluation limits.

## 0.1.0 - 2026-11-18

- Add a standalone Go CLI with interactive, one-shot, and piped-input modes.
- Add per-conversation name and age memory, with explicit age comparison answers.
- Add score-based rules, reproducible wording seeds, and module/timing output.
- Document the project's relationship to upstream PyBot and its current scope.

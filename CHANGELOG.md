# Changelog

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

# Changelog

## 0.9.2 - 2026-11-21

- Strip common polite overview and “tell me about” phrasing before Wikipedia search so the entity is ranked ahead of the request wrapper.

## 0.9.1 - 2026-11-21

- Route natural factual questions with varied phrasing, including polite “can you give me…” forms, to cited lookup instead of the unknown fallback.

## 0.9.0 - 2026-11-21

- Add 100 natural two-turn factual conversations spanning geography, science, history, culture, and everyday knowledge.
- Use their first questions and contextual follow-ups as CPU transformer routing examples; factual answers still come from cited Wikipedia extracts.
- Resolve entity follow-ups by retaining the prior question and cited topic in the next Wikipedia search.
- Bundle the examples with the browser site and show their count when the semantic router is ready.

## 0.8.1 - 2026-11-21

- Add a 50-scenario conversation evaluation covering memory, follow-ups, calculations, Linux help, answer preferences, sources, and topic navigation.
- Acknowledge standalone remembered facts before comparison routing, so age statements are stored instead of triggering a premature comparison question.
- Understand explicit Linux distributions in package questions, natural requests for more detail, and negative requests to stop guided steps.
- Clarify incomplete calculation requests and correctly explain that square roots of negative values are not real.
- Use clear preference summaries such as “bullet points,” and refresh the browser assets for this release.

## 0.8.0 - 2026-11-21

- Remember answer preferences for brief or detailed replies, beginner explanations, bullet or numbered formatting, and one-step Linux guidance.
- Keep a bounded recent-topic list so users can recap it or return to an earlier answer by topic words.
- Add stepwise Linux help with repeat, explain, skip, next, show remaining steps, and output-based recovery turns.
- Answer location, population, area, river, and history follow-ups from the latest cited article when it contains the requested fact.
- Expand CPU embedding examples for topic navigation and guided follow-up language; rebuild the browser bundle.

## 0.7.1 - 2026-11-20

- Normalize smart apostrophes and recognize natural statements like “I’m using Ubuntu 24.04” as environment memory.
- Tailor package update instructions and citations to the remembered Linux distribution.

## 0.7.0 - 2026-11-20

- Add a bounded turn history that tracks prior questions, answers, routes, sources, and external topics.
- Restore the last 24 turns from per-tab session storage; the Go runtime replays the saved turns so its typed memory and follow-up context come back after reload.
- Resolve contextual follow-ups for repeating, shortening, explaining, checking safety, citing sources, changing Linux distributions, and asking more about a sourced topic.
- Remember a user's stated Linux distribution for later help, answer distro-recall questions, and support selective forgetting of a name, age, or distribution.
- Add contextual follow-up prototypes to the CPU embedding router without allowing the model to generate response text.
- Add suggested follow-up actions, a live thinking state, answer copying with citations, Markdown transcript export, and a message length counter.
- Add `/context` and `/forget` to interactive CLI chat; document 50 observable conversational improvements.
- Version browser assets for the new conversation runtime.

## 0.6.1 - 2026-11-20

- Expand the CPU-only Linux reference library across common shell, storage, process, desktop, networking, and service questions.
- Restrict Linux retrieval to matching topics and abstain on unsupported Linux questions instead of returning an unrelated article.
- Add regression coverage for previously mismatched Linux topics and publish the 100-question retrieval sample results.
- Keep the browser asset URLs aligned with the new version.

## 0.6.0 - 2026-11-20

- Add a CPU-only Linux administration route with cited guidance for common service, storage, networking, permissions, package, SSH, and container tasks.
- Add distro-aware diagnostic commands, read-only first steps, safe troubleshooting follow-ups, and sourced definitions for Linux, systemd, and SSH.
- Expand browser examples and documentation for Linux help.

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

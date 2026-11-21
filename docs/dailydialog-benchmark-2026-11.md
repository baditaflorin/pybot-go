# DailyDialog replay benchmark (2026-11-21)

## Scope and method

This is a small, seeded conversation-routing audit, not a factual-answer accuracy benchmark. DailyDialog is a human-written corpus of everyday multi-turn conversations with dialogue-act and emotion labels. Its held-out test split contains 1,000 dialogues and is licensed CC BY-NC-SA 4.0 ([dataset card](https://huggingface.co/datasets/ConvLab/dailydialog/blob/main/README.md); [original paper](https://aclanthology.org/I17-1099/)). No dialogue text or derived examples are included in this repository.

A fixed seed (`20261121`) selected 100 sorted dialogue indices using Python's `random.Random(20261121).sample(range(1000), 100)`. The sample contained 773 utterances. The tested source was PyBot Go v0.10.4 at commit `f8d576f8b3c06eb0361f7b26a119215bd7316e0f`. The audit sent the even-indexed speaker's 372 turns, in order, to a fresh Go `Bot` for each dialogue. It recorded native Go routes and estimated whether the browser would fall back to Wikipedia using the current `site/app.js` predicate.

The run did **not** execute the browser's MiniLM/WASM classifier or fetch Wikipedia pages. It therefore measures local routing and the browser's lookup handoff, not end-to-end answer quality. The other speaker's scripted utterance was not fed back as the bot's reply, so this is not a full interactive simulation. Dialogue-act labels are descriptive; a question label does not mean the utterance asks for an encyclopedia fact.

## Results

| First-speaker turn outcome | Count | Share |
| --- | ---: | ---: |
| Native Go route produced a non-unknown response | 56 | 15.1% |
| Browser would hand an unknown response to Wikipedia | 140 | 37.6% |
| Remained `conversation/unknown` without lookup | 176 | 47.3% |
| **Total** | **372** | **100%** |

The 56 native responses were mostly narrow routes: name memory (27), greeting (13), small talk (4), thanks (4), explanation (3), identity (2), and one each for arithmetic, follow-up, and age comparison. The 140 Wikipedia cases are *pending lookups*, not 140 verified answers.

The largest gaps by dialogue-act label were:

| User-turn act | Turns | Unknown | Wikipedia handoff |
| --- | ---: | ---: | ---: |
| Inform (act 1) | 141 | 121 (85.8%) | 2 |
| Question (act 2) | 145 | 4 | 114 (78.6%) |
| Directive (act 3) | 64 | 32 (50.0%) | 24 (37.5%) |
| Commissive (act 4) | 22 | 19 (86.4%) | 0 |

Of the 173 evaluated turns containing a question mark, 138 (79.8%) would be sent to Wikipedia. That set includes ordinary interpersonal and coordination questions. The fallback in `site/app.js` treats any string ending in `?` as factual, so the handoff is too broad. The converse problem is that ordinary statements, offers, and commitments usually get the same abstention fallback instead of a context-aware conversational response.

## v0.10.5 dialogue-policy replay

The same 100 conversations and 372 first-speaker turns were replayed again against the candidate Go router. This pass added the browser's MiniLM intent hint using the pinned `all-MiniLM-L6-v2` revision and q8 weights, evaluated on CPU with Transformers.js. The Go response policy and browser factual-lookup predicate were both applied. The test used the same Go source as this candidate; its MiniLM inference ran in Node's CPU backend, not in a real browser's WebAssembly runtime. It still did not feed the second scripted speaker turn back as a bot reply, so this remains a routing replay rather than a realistic live chat or gold-answer evaluation.

| First-speaker turn outcome | v0.10.4 baseline | v0.10.5 candidate |
| --- | ---: | ---: |
| Native conversational, solver, or cited-fact route | 56 (15.1%) | 372 (100%) |
| Eligible for external factual lookup | 140 (37.6%) | 0 (0%) |
| Dead-end `conversation/unknown` without lookup | 176 (47.3%) | 0 (0%) |
| Question-mark turns handed to lookup | 138 / 173 (79.8%) | 0 / 173 (0%) |
| Local clarification responses | Not measured | 60 / 372 (16.1%) |

| User-turn act | Turns | Native | Lookup candidates | Unknown |
| --- | ---: | ---: | ---: | ---: |
| Inform | 141 | 141 | 0 | 0 |
| Question | 145 | 145 | 0 | 0 |
| Directive | 64 | 64 | 0 | 0 |
| Commissive | 22 | 22 | 0 | 0 |

For the candidate, the browser handoff was simulated using the same JavaScript question gate and accepted MiniLM hint that the UI uses. The 372 local routes include 60 clarification responses (48 general question clarifications, eight unresolved-reference prompts, and four specific context prompts). These are not counted as verified answers; some are expected, while others still need better intent recognition or query reformulation. The 85 `conversation/social-question` routes now include common personal questions, offers, and invitations that the previous pass mishandled.

An initial retrieval audit selected unrelated pages for three broad questions about an inter-city time offset, a solar-system relation, and university adviser duties. The relevance gate rejected those results. This led to three small, source-backed answers with appropriate caveats: seasonal time-zone differences link to both city time-zone pages, the planet relation links to NASA, and the adviser-role answer notes that duties vary by institution and links to a university job description. A separate smoke set accepted relevant sources for Romania, Bucharest, Marie Curie, photosynthesis, and the Danube, including a yes/no capital question. These results demonstrate better coverage for the tested prompts, not general factual accuracy across arbitrary questions.

The candidate adds conversation-act routes for informative statements, positive and negative emotional disclosures, acknowledgments, offers, requests, and social questions. It also separates polite requests for explanations from preference questions, gives practical apartment-search guidance, asks for a location or role when a fact depends on missing context, and rejects unrelated retrieval pages. These are deterministic response policies; the embedding model only suggests a route and ranks sources. The model's route labels are not a validated classifier for DailyDialog's four gold dialogue-act labels, and local routing coverage must not be read as answer accuracy.

## What the bot needs next

1. **Close the retrieval gaps surfaced by the replay.** Improve query reformulation for common relation questions, time-zone differences, and role descriptions; add reliable sources or exact solvers where appropriate, and retain abstention when evidence is weak.
2. **Improve context grounding.** The small sample shows that a route can be conversational yet still miss what a user means. Feed the other speaker's reply back into a multi-turn simulation, then evaluate references, topic changes, corrections, and user-provided facts.
3. **Measure answer quality separately from routing coverage.** Use answer-bearing QA sets to score factual correctness and citations. Use dialogue corpora to measure act recognition, response type, context carry-over, and abstention. DailyDialog does not provide a gold assistant answer for this bot.
4. **Run the actual browser path in CI or a browser smoke job.** Verify MiniLM WebAssembly loading, intent/score, source relevance, source coverage, failures, latency, CPU use, and follow-up memory after reload. Node CPU model results do not establish browser runtime behavior.
5. **Have people rate a blinded sample.** A deterministic template can be on-topic but still feel stiff. Human ratings should measure relevance, clarity, naturalness, and whether a clarification question was necessary.

## Interpretation

The current design is a compact CPU-only router, deterministic solvers, short conversation memory, and retrieval. It is not an open-ended language generator. Embeddings can make routing and retrieval more tolerant of wording, but the main product gap is dialogue policy plus grounded response composition—not simply adding more intent examples. A useful next milestone is a held-out evaluation that shows fewer irrelevant web lookups, fewer dead-end turns, and sourced answers that match the question, while preserving explicit abstention when the evidence is weak.

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

## What the bot needs next

1. **Separate question form from question purpose.** Route fact questions to retrieval, but recognize social questions, preferences, offers, invitations, and requests as dialogue. Use the CPU embedding classifier as one signal; keep conservative lexical fallbacks when the model is unavailable. Add explicit cases for yes/no factual questions so the filter does not simply suppress them.
2. **Add dialogue-act handling and lightweight state.** Distinguish disclosures, opinions, acknowledgments, apologies, offers, and actionable requests. Respond with a short, relevant turn and ask for clarification only when needed. Track the current entity and task across turns so “what about the other one?” has a real referent.
3. **Make retrieval end-to-end and evidence-aware.** For factual questions, retrieve from a local searchable corpus and/or an external source, check that the selected page answers the question, and return a concise extract with a citation. A handoff to Wikipedia is not proof of an answer.
4. **Compose answers without a generative model.** Use retrieved sentences, typed facts, exact solvers, and small response templates. MiniLM embeddings can match paraphrases, rank evidence, and resolve a follow-up; they do not provide facts or generate a coherent answer by themselves.
5. **Evaluate the right things separately.** Use answer-bearing QA sets to score factual correctness and citations. Use dialogue corpora to measure act recognition, appropriate response type, context carry-over, and abstention. Have people rate a blinded sample of responses; DailyDialog does not provide a gold assistant answer for this bot.
6. **Run the browser path in the next evaluation.** Measure MiniLM load success, intent/score, Wikipedia result relevance, source coverage, lookup failures, latency, and CPU use. The Go-only replay above cannot establish those outcomes.

## Interpretation

The current design is a compact CPU-only router, deterministic solvers, short conversation memory, and retrieval. It is not an open-ended language generator. Embeddings can make routing and retrieval more tolerant of wording, but the main product gap is dialogue policy plus grounded response composition—not simply adding more intent examples. A useful next milestone is a held-out evaluation that shows fewer irrelevant web lookups, fewer dead-end turns, and sourced answers that match the question, while preserving explicit abstention when the evidence is weak.

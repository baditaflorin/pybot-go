# 50 conversation improvements

PyBot Go v0.7.0 adds a deterministic, CPU-only dialogue layer. These are concrete behaviors supported by the browser and Go conversation engine. They improve context and turn-taking; they do not make PyBot a generative LLM.

## Turn memory and restoration

1. Keep the user's latest message as part of the active conversation state.
2. Keep the last assistant answer for direct follow-up questions.
3. Keep the route that produced the answer so follow-up handling can stay topic-aware.
4. Keep the answer's source links so a user can ask for them again.
5. Keep the subject of the latest cited Wikipedia answer for entity follow-ups.
6. Bound retained history to the latest 24 turns.
7. Store browser history in per-tab `sessionStorage`.
8. Rebuild the visible conversation after reloading the same tab.
9. Replay restored user turns through the Go runtime to rebuild typed facts and topic state.
10. Restore the saved embedding intent hint with each turn.
11. Retain age facts and names across the restored turns.
12. Allow a newer fact for the same person to replace the older fact.
13. Reattach the final browser-side answer after restoring a turn.
14. Clear the saved history and Go memory with **New chat**.
15. Keep separate browser tabs' conversations separate.
16. Keep conversation text off a PyBot server.
17. Send only the current or context-resolved factual query to Wikipedia.
18. Continue to use local Go solvers for arithmetic, dates, conversions, and age reasoning.
19. Include the latest assistant answer in the transformer context for short follow-ups.
20. Use that transformer context only for short, explicitly contextual messages.

## Natural follow-ups

21. Recognize a broader CPU-embedding intent for conversational follow-ups.
22. Repeat the previous answer on request.
23. Show the previous answer's citations again.
24. Produce a concise version using the first complete sentence.
25. Explain why read-only Linux checks were suggested.
26. Explain common commands from the preceding answer.
27. Flag data-changing commands when a user asks whether they are safe.
28. Identify when the existing answer is already the full local guidance and ask which detail needs unpacking.
29. Continue a Linux answer with the documented next-step guidance.
30. Ask a targeted clarification when the requested explanation is underspecified.

## Linux context

31. Remember a stated Linux distribution for the current conversation.
32. Retain an explicitly stated distribution version such as Ubuntu 24.04.
33. Accept a short distro-only answer when PyBot just asked for the environment.
34. Answer “what distro am I using?” from the current conversation.
35. Use the remembered distro when searching for later Linux help.
36. Re-run the active Linux topic when asked how it applies to another distribution.
37. Keep the original task active across a short distro clarification.
38. State when Windows or macOS need different instructions from the Linux answer.
39. Carry source links into “why?” and safety follow-ups.
40. Keep Linux help read-only first and do not execute suggested commands.

## Memory controls and cited facts

41. Summarize explicitly remembered facts when asked what PyBot remembers.
42. Identify remembered names in that summary.
43. Identify remembered ages in that summary.
44. Identify the remembered Linux distribution in that summary.
45. Forget a stored name on request.
46. Forget a stored age on request.
47. Forget a stored distribution on request.
48. Resolve “where is it?” against the latest cited Wikipedia topic.
49. Resolve follow-ups about size, population, history, or additional detail against that topic.
50. Show a retry action when a factual lookup cannot reach Wikipedia.

The app also exposes the conversational loop directly: it shows when it is matching a question or looking up a citation, offers next-turn suggestions, supports copying answers with citations, counts message characters, and exports a Markdown transcript. The 50-item list focuses on context behaviors rather than counting those interface controls as separate language capabilities.

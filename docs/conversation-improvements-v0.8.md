# 50 more conversation improvements

PyBot Go v0.8.0 extends the deterministic, CPU-only dialogue layer with answer preferences, recent-topic navigation, guided Linux steps, and more precise source-backed follow-ups. These are bounded conversation behaviors, not generated text.

## Answer preferences

1. Accept “keep it brief,” “keep it short,” and “be concise” as a chat-level preference.
2. Keep that brief-answer preference for later Linux replies in the same conversation.
3. Accept requests for detailed answers and add a known command explanation when available.
4. Accept beginner requests and explain common commands in plain language.
5. Let an advanced-user preference turn off the beginner explanation.
6. Format multi-part answers as bullets when requested.
7. Format multi-part answers as numbered lines when requested.
8. Enable guided mode with “one step at a time” or “step by step.”
9. Acknowledge the preference so the user knows the setting took effect.
10. Summarize the current answer preferences on request.
11. Clear answer preferences without clearing age, name, or distro facts.
12. Restore preferences by replaying the saved turns after a page reload.
13. Keep preference state local to the current browser tab or CLI conversation.
14. Apply formatting preferences to Linux guidance rather than changing factual lookup text.
15. Keep command safety notes and source links when changing the response format.

## Recent topics and returning to an earlier answer

16. Record topics from retrieved answers and deterministic reasoning replies.
17. Keep only the eight most recent topic entries.
18. Show a numbered recap of recent topics.
19. Include the original user question in the topic recap.
20. Include a concise topic label derived from the retrieved article or answer route.
21. Return to the previous topic with “go back to the previous topic.”
22. Return to a named topic with “go back to disk space” or similar words.
23. Accept “return to,” “resume,” “pick up,” and “back to” as topic-navigation phrasing.
24. Match a named topic using title, prior question, and route words.
25. Prefer the most recent topic when several topic labels match equally well.
26. Replay the earlier answer rather than asking the user to reconstruct it.
27. Bring the earlier answer's citations back with it.
28. Preserve the selected Linux task when returning to that topic.
29. Keep topic history within the browser's existing 24-turn session restore.
30. Ask for clarification when the requested topic does not match recent history.

## Guided Linux troubleshooting

31. Split multi-part Linux guidance into readable steps when guided mode is enabled.
32. Show the first step and wait before revealing the rest.
33. Accept “next,” “continue,” “what now,” and “what should I try?” to proceed.
34. Repeat the current step on request.
35. Explain a step by its number.
36. Accept numeric step numbers and common ordinal names such as first, second, and third.
37. Show all remaining steps when asked.
38. Let the user skip the current step or skip ahead.
39. Accept “that worked” and “done with that step” as progress updates.
40. Mark when the current answer has no more steps.
41. Ask for the exact command and complete output when a step fails.
42. Ask the user to redact passwords, tokens, private keys, or sensitive hostnames from pasted output.
43. Preserve source links during step repeat, explanation, and progress turns.
44. Clear a stale guided plan when the user starts a different substantive topic.
45. Stop guided mode and return to complete Linux answers on request.

## More precise follow-ups on sourced facts

46. Answer “where is it?” from a location sentence already present in the cited article.
47. Answer population questions from a cited population or residents sentence.
48. Answer “which river?” from a cited river sentence when the article includes one.
49. Answer size, area, and history questions from matching sentences in the cited article.
50. Keep the previous entity and citation in scope, and request a fresh lookup when the cited text does not contain the requested fact.

The browser uses CPU MiniLM embeddings to route paraphrases. Go code selects every response and does not use a generative LLM or GPU.

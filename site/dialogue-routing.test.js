const test = require("node:test");
const assert = require("node:assert/strict");
const { hasRelevantSource, isSocialQuestion, normalizeFactualQuestion, shouldLookupFactualQuestion } = require("./dialogue-routing.js");

test("factual open questions still request a sourced lookup", () => {
  for (const question of [
    "What is the capital of Romania?",
    "Who was Marie Curie?",
    "Where are the Carpathian Mountains?",
    "Can you explain photosynthesis?",
    "Tell me about the Danube River.",
  ]) {
    assert.equal(shouldLookupFactualQuestion(question), true, question);
  }
});

test("yes-no questions require a factual semantic signal", () => {
  assert.equal(shouldLookupFactualQuestion("Is Bucharest the capital of Romania?", "knowledge_question"), true);
  assert.equal(shouldLookupFactualQuestion("Is Bucharest the capital of Romania?"), false);
});

test("polite and embedded factual questions normalize to the actual subject", () => {
  assert.equal(normalizeFactualQuestion("Can you help me understand what Romania is?"), "Romania");
  assert.equal(normalizeFactualQuestion("Could you give me a brief overview of Romania?"), "Romania");
  assert.equal(normalizeFactualQuestion("What is the capital of Romania?"), "capital of Romania");
  assert.equal(normalizeFactualQuestion("Okay. What is the closest planet to our star?"), "closest planet to our star");
  assert.equal(normalizeFactualQuestion("Excuse me, what is the clock offset between Tokyo and London?"), "clock offset between Tokyo and London");
  assert.equal(normalizeFactualQuestion("Would you mind explaining what photosynthesis is?"), "photosynthesis");
});

test("personal questions and offers never go to Wikipedia by punctuation alone", () => {
  for (const question of [
    "Could I be useful to you?",
    "Would you travel to a conference for work?",
    "Would you pick a coat or a jacket?",
    "What personal ability would you work on improving?",
    "Do you think my explanation is understandable?",
    "Could you bring me a glass of water?",
    "Would Friday afternoon suit you?",
    "Can you tell whether I seem excited?",
    "How much longer do you intend to stay in Chicago?",
    "Which courses would you prefer to study?",
    "Would it be a long drive to the museum?",
    "Why are you acting unusually today?",
    "What is the reason you sent a gift to your colleague?",
    "Why won't you try a different meal?",
    "How do you usually structure your weekdays, Alex?",
    "What does your parent do for work?",
    "What name should I use for this person?",
    "Do you mind if I change the music?",
    "What place did you grow up in?",
    "Is there anything I can help you with?",
    "What details would you like from me?",
    "Which teammates do you collaborate with most easily?",
    "Have you ever attended a live show?",
    "Do you keep a regular exercise schedule?",
    "Should we visit the science museum this weekend?",
    "Could you send my teammate a note asking them to call?",
  ]) {
    assert.equal(isSocialQuestion(question), true, question);
    assert.equal(shouldLookupFactualQuestion(question, ""), false, question);
    assert.equal(shouldLookupFactualQuestion(question, "knowledge_question"), false, question);
  }
});

test("polite requests for an explanation remain factual tasks", () => {
  const question = "Would you mind describing how photosynthesis works?";
  assert.equal(isSocialQuestion(question), false);
  assert.equal(shouldLookupFactualQuestion(question), true);
});

test("advice questions do not get sent to an encyclopedia", () => {
  for (const question of [
    "How can I find a reasonably priced apartment?",
    "How should I compare two job offers?",
  ]) {
    assert.equal(shouldLookupFactualQuestion(question), false, question);
  }
});

test("knowledge questions wrapped in do-you-know wording remain eligible", () => {
  const question = "Do you know where Romania is?";
  assert.equal(isSocialQuestion(question), false);
  assert.equal(shouldLookupFactualQuestion(question, "knowledge_question"), true);
});

test("accepted dialogue-act labels suppress encyclopedia fallback for open questions", () => {
  assert.equal(shouldLookupFactualQuestion("Which courses would you prefer to study?", "social_question"), false);
  assert.equal(shouldLookupFactualQuestion("Which courses would you prefer to study?", "dialogue_request"), false);
  assert.equal(shouldLookupFactualQuestion("Can you list the major cities?", "dialogue_request"), true);
  assert.equal(shouldLookupFactualQuestion("What is the time difference between Tokyo and London?", "date_math"), true);
  assert.equal(shouldLookupFactualQuestion("Can you help me understand what Romania is?", "dialogue_request"), true);
  assert.equal(shouldLookupFactualQuestion("Can you help me with the capital of Romania?", "dialogue_request"), true);
  assert.equal(shouldLookupFactualQuestion("What about this red dress?", ""), false);
  assert.equal(shouldLookupFactualQuestion("What about Abraham Lincoln?", "context_followup"), false);
});

test("question marks alone do not trigger external lookup", () => {
  for (const text of ["Maybe tomorrow?", "I wonder if that is right?", "Same here?"]) {
    assert.equal(shouldLookupFactualQuestion(text), false, text);
  }
});

test("retrieval abstains when the selected page has no lexical or semantic support", () => {
  assert.equal(hasRelevantSource(0, null), false);
  assert.equal(hasRelevantSource(0.64, 0.61), false);
  assert.equal(hasRelevantSource(0.65, 0.1), true);
  assert.equal(hasRelevantSource(0, 0.62), true);
});

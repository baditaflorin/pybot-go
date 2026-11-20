package pybot

import (
	"fmt"
	"strings"
)

const statisticalRouteThreshold = 0.55
const semanticRouteThreshold = 0.34

// Reply parses a message, updates conversation memory, then routes it through
// exact solvers, source-backed local retrieval, and conservative fallbacks.
func (b *Bot) Reply(input string) Response {
	return b.replyWithSemanticHint(input, "", 0)
}

// ReplyWithSemanticHint adds a CPU embedding similarity hint to the regular Go
// router. The hint can choose a deterministic route, but cannot supply or
// generate an answer. Confidence is cosine similarity, not a probability.
func (b *Bot) ReplyWithSemanticHint(input, hintedIntent string, confidence float64) Response {
	return b.replyWithSemanticHint(input, hintedIntent, confidence)
}

func (b *Bot) replyWithSemanticHint(input, hintedIntent string, confidence float64) Response {
	b.mu.Lock()
	defer b.mu.Unlock()

	text := strings.TrimSpace(input)
	if text == "" {
		return Response{Text: "Send me a question or a fact to work with.", Module: "conversation/empty"}
	}
	updates := b.memory.remember(text)
	lower := normalizeText(text)
	if pronoun := unresolvedAgePronoun(lower); pronoun != "" {
		return Response{
			Text:   fmt.Sprintf("Who do you mean by %q? Tell me the person's name so I don't attach the age to the wrong person.", pronoun),
			Module: "memory/age", NeedsClarification: true,
		}
	}
	messageIntent, intentScore := defaultIntentClassifier.classify(lower)
	semanticAccepted := false
	if hint := parseSemanticIntent(hintedIntent); hint != "" && confidence >= semanticRouteThreshold &&
		(hint != intentLinuxExpert || isLinuxQuestion(lower)) {
		// Similarity alone should not invent missing age context. Keep ranking
		// routes tied to the words in the question and the facts in memory.
		if hint != intentAgeRanking || isAgeRankingQuestion(lower, len(b.memory.ages())) ||
			(len(b.memory.ages()) >= 2 && looksLikeComparativeAgeFollowup(lower)) {
			messageIntent, intentScore = hint, confidence
			semanticAccepted = true
		}
	}
	if isLinuxQuestion(lower) {
		messageIntent, intentScore = intentLinuxExpert, 1
	}
	routeTrusted := intentScore >= statisticalRouteThreshold || semanticAccepted

	if isAgeRankingQuestion(lower, len(b.memory.ages())) || (messageIntent == intentAgeRanking && routeTrusted) {
		return b.answerAgeRanking(lower)
	}
	if isAgeComparison(lower) || (messageIntent == intentAgeComparison && routeTrusted) {
		return b.answerAgeComparison(lower)
	}
	if response, handled := solveUnitConversion(lower); handled {
		return response
	}
	if response, handled := solveDateQuestion(lower, b.clock()); handled {
		return response
	}
	if response, handled := solveArithmetic(lower); handled {
		return response
	}
	if isNameQuestion(lower) || (messageIntent == intentNameQuestion && routeTrusted) {
		if name := b.memory.name(); name != "" {
			return Response{Text: "Your name is " + name + ".", Module: "memory/name"}
		}
		return Response{Text: "I don't know your name yet. What should I call you?", Module: "memory/name", NeedsClarification: true}
	}
	if isAgeQuestion(lower) {
		return b.answerAgeFact(lower)
	}
	if isCapabilityQuestion(lower) {
		return Response{
			Text:   "I can remember facts you tell me in this chat and answer follow-ups; compare or rank ages; calculate arithmetic and square roots; convert common units; and work out dates. I can answer a few built-in topics and look up other factual questions with a source link. For example, ask who is older, the square root of 2, or the capital of Romania. The CPU transformer helps match wording to these routes; it doesn't write answers. I don't use a generative LLM or GPU.",
			Module: "conversation/capabilities",
		}
	}
	if updates > 0 && isStandaloneStatement(lower) {
		return b.acknowledgeFacts()
	}
	if response, handled := b.answerFollowup(lower); handled {
		return response
	}
	if isGreeting(lower) || (messageIntent == intentGreeting && routeTrusted) {
		if name := b.memory.name(); name != "" {
			return Response{Text: "Hello, " + name + "! What can I help with?", Module: "conversation/greeting"}
		}
		return Response{Text: "Hello! What can I help with?", Module: "conversation/greeting"}
	}
	if isThanks(lower) || (messageIntent == intentThanks && routeTrusted) {
		return Response{Text: "You're welcome!", Module: "conversation/thanks"}
	}
	if strings.Contains(lower, "who are you") || strings.Contains(lower, "what are you") ||
		(messageIntent == intentIdentity && routeTrusted) {
		return Response{Text: "I'm PyBot Go: a CPU-only assistant that routes with a small statistical model, stores typed facts for this conversation, solves exact problems, and retrieves from a local knowledge base. I don't use an LLM.", Module: "conversation/identity"}
	}
	if strings.Contains(lower, "how are you") || (messageIntent == intentSmallTalk && routeTrusted) {
		return Response{Text: "I'm running fine. What would you like to work out?", Module: "conversation/small-talk"}
	}
	if result, found := b.knowledge.Search(lower, messageIntent, routeTrusted); found {
		if len(result.Ambiguous) > 0 {
			return Response{
				Text:   fmt.Sprintf("I found two likely topics: %s and %s. Which one do you mean?", result.Ambiguous[0].Title, result.Ambiguous[1].Title),
				Module: "retrieval/clarify", NeedsClarification: true,
			}
		}
		b.memory.lastKnowledgeID = result.Document.ID
		b.memory.lastKnowledgeTurn = b.memory.turn
		return Response{
			Text:    result.Document.Answer,
			Module:  "retrieval/" + result.Document.ID,
			Sources: allSources(result.Document),
		}
	}
	return Response{
		Text:   randomFallback(b.rng),
		Module: "conversation/unknown",
	}
}

func parseSemanticIntent(value string) intent {
	switch intent(value) {
	case intentAgeComparison, intentAgeRanking, intentArithmetic, intentConversion, intentDate, intentKnowledge, intentLinuxExpert:
		return intent(value)
	default:
		return ""
	}
}

func normalizeText(text string) string {
	return strings.Join(strings.Fields(normalizeIntentAliases(text)), " ")
}

func isNameQuestion(text string) bool {
	return strings.Contains(text, "what's my name") || strings.Contains(text, "what is my name") ||
		strings.Contains(text, "do you remember my name") || strings.Contains(text, "who am i")
}

func isCapabilityQuestion(text string) bool {
	for _, phrase := range []string{
		"what can you do", "what you can do", "what can i ask", "what can i ask you", "what are you able to do",
		"what are your capabilities", "what questions can you answer", "how can you help me",
		"what kind of questions can you answer", "what do you do", "list your capabilities",
	} {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}

func isAgeComparison(text string) bool {
	comparative := strings.Contains(text, "older") || strings.Contains(text, "younger") ||
		strings.Contains(text, "oldest") || strings.Contains(text, "youngest") ||
		strings.Contains(text, "higher age") || strings.Contains(text, "greater age") ||
		strings.Contains(text, "lower age") || strings.Contains(text, "compare their ages") ||
		strings.Contains(text, "compare the ages") || strings.Contains(text, "how much older") ||
		strings.Contains(text, "how much younger")
	return comparative && (strings.Contains(text, "age") || strings.Contains(text, "year old") ||
		strings.Contains(text, "years old") || strings.Contains(text, "who ") ||
		strings.HasPrefix(text, "who's ") || strings.HasPrefix(text, "which ") ||
		strings.HasPrefix(text, "how much older") || strings.HasPrefix(text, "how much younger"))
}

func isAgeRankingQuestion(text string, knownAgeCount int) bool {
	positionWord := strings.Contains(text, "rank") || strings.Contains(text, "position") ||
		strings.Contains(text, "age order") || strings.Contains(text, "age ranking") ||
		strings.Contains(text, "sit") || strings.Contains(text, "fit") || strings.Contains(text, "stand") ||
		strings.Contains(text, "fall between") || strings.Contains(text, "come in")
	if !positionWord {
		return false
	}
	if strings.Contains(text, "age") || strings.Contains(text, "older") || strings.Contains(text, "younger") ||
		strings.Contains(text, "oldest") || strings.Contains(text, "youngest") || strings.Contains(text, "years old") {
		return true
	}
	return knownAgeCount >= 2 && (strings.Contains(text, "where") || strings.Contains(text, "between"))
}

func looksLikeComparativeAgeFollowup(text string) bool {
	return strings.Contains(text, "compare") || strings.Contains(text, "relative") ||
		strings.Contains(text, "among") || strings.Contains(text, "where") ||
		strings.Contains(text, "sit") || strings.Contains(text, "fit") ||
		strings.Contains(text, "position") || strings.Contains(text, "rank")
}

func isAgeQuestion(text string) bool {
	return strings.Contains(text, "how old is") || strings.Contains(text, "how old are") ||
		strings.Contains(text, "how old am i") || strings.Contains(text, "what is the age of") || strings.Contains(text, "what's the age of")
}

func isStandaloneStatement(text string) bool {
	if strings.Contains(text, "?") {
		return false
	}
	for _, prefix := range []string{"who ", "what ", "when ", "where ", "why ", "how ", "which ", "is ", "are ", "can ", "could "} {
		if strings.HasPrefix(text, prefix) {
			return false
		}
	}
	return true
}

func (b *Bot) acknowledgeFacts() Response {
	facts := b.memory.factsInOrder()
	latest := make([]Fact, 0)
	for _, fact := range facts {
		if fact.Turn == b.memory.turn {
			latest = append(latest, fact)
		}
	}
	if len(latest) == 0 {
		return Response{Text: "Got it.", Module: "memory/facts"}
	}
	fact := latest[len(latest)-1]
	if fact.Attribute == "name" {
		return Response{Text: "I'll remember your name as " + fact.Value + " for this conversation.", Module: "memory/name"}
	}
	if fact.Subject == "user" {
		return Response{Text: fmt.Sprintf("I'll remember that you are %s years old for this conversation.", fact.Value), Module: "memory/age"}
	}
	return Response{Text: fmt.Sprintf("I'll remember that %s is %s years old for this conversation.", displayPerson(fact.Subject), fact.Value), Module: "memory/age"}
}

func isGreeting(text string) bool {
	for _, word := range []string{"hi", "hello", "hey", "good morning", "good afternoon", "good evening"} {
		if text == word || strings.HasPrefix(text, word+" ") || strings.HasPrefix(text, word+",") || strings.HasPrefix(text, word+"!") {
			return true
		}
	}
	return false
}

func isThanks(text string) bool {
	return text == "thanks" || text == "thank you" || strings.HasPrefix(text, "thanks ") || strings.HasPrefix(text, "thank you ")
}

func (b *Bot) answerFollowup(text string) (Response, bool) {
	if b.memory.lastKnowledgeID == "" || b.memory.turn-b.memory.lastKnowledgeTurn > 2 {
		return Response{}, false
	}
	if strings.Contains(text, "how long") || strings.Contains(text, "how much time") {
		if answer, ok := b.knowledge.Followup(b.memory.lastKnowledgeID, "duration"); ok {
			return answer, true
		}
		return Response{Text: "Which part do you mean? I don't have a more specific timing note for the previous answer.", Module: "memory/follow-up", NeedsClarification: true}, true
	}
	if isNextStepFollowup(text) && strings.HasPrefix(b.memory.lastKnowledgeID, "linux-") {
		if answer, ok := b.knowledge.Followup(b.memory.lastKnowledgeID, "next_steps"); ok {
			return answer, true
		}
		return Response{
			Text:   "Tell me the Linux distribution and version, the exact command you ran, its complete error or output, and what changed recently. Redact passwords, tokens, private keys, and sensitive host details.",
			Module: "linux/follow-up", NeedsClarification: true,
		}, true
	}
	return Response{}, false
}

func isNextStepFollowup(text string) bool {
	for _, phrase := range []string{
		"what next", "what should i do next", "what should i try", "what do i try next",
		"what if that doesn't work", "what if that does not work", "then what", "how do i troubleshoot that",
	} {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	return false
}

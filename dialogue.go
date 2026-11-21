package pybot

import (
	"fmt"
	"regexp"
	"strings"
)

type conversationPreferences struct {
	verbosity string
	audience  string
	format    string
	guided    bool
}

type preferenceChange struct {
	field           string
	value           string
	acknowledgement string
}

type conversationTopic struct {
	key      string
	title    string
	query    string
	response Response
	turn     int
}

const conversationTopicLimit = 8

func preferenceStatement(text string) (preferenceChange, bool) {
	switch {
	case containsAny(text, "stop step by step", "stop one step at a time", "no more step by step", "turn off guided steps", "stop guiding me"):
		return preferenceChange{"guided", "false", "Okay. I'll give you the complete answer again, with the usual safety notes."}, true
	case containsAny(text, "one step at a time", "step by step", "guide me through", "walk me through one step", "guided steps"):
		return preferenceChange{"guided", "true", "Understood. I'll give Linux guidance one step at a time and wait for you before continuing."}, true
	case containsAny(text, "keep it brief", "keep answers short", "short answers", "be concise", "answer concisely", "in one sentence", "keep it short", "less detail"):
		return preferenceChange{"verbosity", "brief", "Understood. I'll keep answers brief for this conversation."}, true
	case containsAny(text, "give more detail by default", "detailed answers", "be more detailed", "explain in detail by default", "full explanations"):
		return preferenceChange{"verbosity", "detailed", "Understood. I'll include more explanation when the available guidance supports it."}, true
	case containsAny(text, "i'm a beginner", "i am a beginner", "i'm new to linux", "i am new to linux", "explain like i'm new", "beginner friendly", "i'm learning linux", "i am learning linux"):
		return preferenceChange{"audience", "beginner", "Got it. I'll explain Linux terms and commands in beginner-friendly language."}, true
	case containsAny(text, "i'm an expert", "i am an expert", "assume i know linux", "skip the basics", "skip basic explanations", "advanced mode"):
		return preferenceChange{"audience", "advanced", "Understood. I'll skip basic explanations where I can."}, true
	case containsAny(text, "use bullet points", "use bullets", "format as bullets", "bullet list"):
		return preferenceChange{"format", "bullets", "Okay. I'll use bullet points when an answer has several parts."}, true
	case containsAny(text, "use numbered steps", "number the steps", "numbered list", "format as a list"):
		return preferenceChange{"format", "numbered", "Okay. I'll number multi-part answers."}, true
	default:
		return preferenceChange{}, false
	}
}

func (p *conversationPreferences) set(change preferenceChange) {
	switch change.field {
	case "verbosity":
		p.verbosity = change.value
	case "audience":
		p.audience = change.value
	case "format":
		p.format = change.value
	case "guided":
		p.guided = change.value == "true"
	}
}

func (p conversationPreferences) summary() string {
	parts := make([]string, 0, 4)
	if p.verbosity != "" {
		parts = append(parts, p.verbosity+" answers")
	}
	if p.audience != "" {
		parts = append(parts, p.audience+"-friendly explanations")
	}
	if p.format != "" {
		parts = append(parts, p.format+" formatting")
	}
	if p.guided {
		parts = append(parts, "Linux guidance one step at a time")
	}
	return strings.Join(parts, "; ")
}

func isPreferenceQuestion(text string) bool {
	return oneOf(text, "what preferences do you remember", "what answer style are you using", "how did i ask you to answer", "what style did i ask for", "what are my answer preferences")
}

func isPreferenceReset(text string) bool {
	return oneOf(text, "forget my preferences", "forget my answer preferences", "reset my answer style", "clear my response preferences", "use your default answer style")
}

func (b *Bot) answerPreferenceControl(text string) (Response, bool) {
	if isPreferenceReset(text) {
		b.memory.preferences = conversationPreferences{}
		b.memory.preferenceAck = "I've cleared your answer preferences for this conversation."
		b.memory.guidedSteps = nil
		b.memory.guidedIndex = 0
		b.memory.guidedSources = nil
		return Response{Text: b.memory.preferenceAck, Module: "memory/preferences"}, true
	}
	if isPreferenceQuestion(text) {
		if summary := b.memory.preferences.summary(); summary != "" {
			return Response{Text: "For this conversation, I'll use " + summary + ". You can ask me to change or forget these preferences.", Module: "memory/preferences"}, true
		}
		return Response{Text: "I haven't saved any answer preferences for this conversation. You can ask for brief answers, beginner explanations, bullet points, or one-step-at-a-time Linux guidance.", Module: "memory/preferences"}, true
	}
	return Response{}, false
}

func isTopicRecapQuestion(text string) bool {
	return oneOf(text, "what were we discussing", "what were we talking about", "what topics have we covered", "what have we discussed", "show recent topics", "recap our topics", "list the topics from this chat")
}

func parseTopicNavigation(text string) (target string, previous bool, ok bool) {
	if oneOf(text, "previous topic", "go back", "go back to the previous topic", "return to the previous topic", "back to the previous topic", "resume the previous topic") {
		return "", true, true
	}
	for _, prefix := range []string{"go back to ", "return to ", "back to ", "resume ", "pick up "} {
		if strings.HasPrefix(text, prefix) {
			target = strings.TrimSpace(strings.Trim(text[len(prefix):], " ?.!,"))
			if strings.HasPrefix(target, "the ") {
				target = strings.TrimPrefix(target, "the ")
			}
			if target != "" && target != "previous topic" {
				return target, false, true
			}
			return "", true, true
		}
	}
	return "", false, false
}

func (b *Bot) answerTopicControl(text string) (Response, bool) {
	if isTopicRecapQuestion(text) {
		if len(b.memory.topics) == 0 {
			return Response{Text: "We haven't completed a topic in this chat yet. Ask a question and I'll keep track of it here.", Module: "conversation/topics"}, true
		}
		lines := make([]string, 0, len(b.memory.topics))
		for i := len(b.memory.topics) - 1; i >= 0; i-- {
			topic := b.memory.topics[i]
			lines = append(lines, fmt.Sprintf("%d. %s — %s", len(b.memory.topics)-i, topic.title, topic.query))
		}
		return Response{Text: "Recent topics in this chat:\n" + strings.Join(lines, "\n"), Module: "conversation/topics"}, true
	}
	target, previous, ok := parseTopicNavigation(text)
	if !ok {
		return Response{}, false
	}
	var selected *conversationTopic
	if previous {
		if len(b.memory.topics) >= 2 {
			selected = &b.memory.topics[len(b.memory.topics)-2]
		}
	} else {
		selected = b.findConversationTopic(target)
	}
	if selected == nil {
		return Response{
			Text:   "I couldn't match that to a recent topic. Ask “what topics have we covered?” to see the conversation list, or name a topic from it.",
			Module: "conversation/topics", NeedsClarification: true,
		}, true
	}
	response := selected.response
	response.Text = fmt.Sprintf("Picking up %s. Earlier you asked: %s\n\n%s", selected.title, selected.query, selected.response.Text)
	response.Sources = append([]Source(nil), selected.response.Sources...)
	response.Suggestions = nil
	if strings.HasPrefix(response.Module, "retrieval/linux-") || strings.HasPrefix(response.Module, "linux/") {
		b.memory.pendingTopicQuery = selected.query
	}
	return response, true
}

func (b *Bot) findConversationTopic(target string) *conversationTopic {
	wanted := topicWords(target)
	if len(wanted) == 0 {
		return nil
	}
	bestScore := 0
	var best *conversationTopic
	for i := range b.memory.topics {
		topic := &b.memory.topics[i]
		corpus := strings.ToLower(topic.title + " " + topic.query + " " + topic.key)
		score := 0
		for word := range wanted {
			if strings.Contains(corpus, word) {
				score++
			}
		}
		if strings.Contains(corpus, strings.ToLower(strings.TrimSpace(target))) {
			score += 3
		}
		if score >= bestScore && score > 0 {
			bestScore = score
			best = topic
		}
	}
	if bestScore == 0 {
		return nil
	}
	return best
}

func topicWords(text string) map[string]struct{} {
	stop := map[string]struct{}{"the": {}, "a": {}, "an": {}, "to": {}, "on": {}, "about": {}, "how": {}, "what": {}, "where": {}, "linux": {}}
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return r < 'a' || r > 'z'
	})
	result := make(map[string]struct{})
	for _, word := range words {
		if len(word) < 3 {
			continue
		}
		if _, found := stop[word]; !found {
			result[word] = struct{}{}
		}
	}
	return result
}

func isTopicResponse(response Response) bool {
	return strings.HasPrefix(response.Module, "retrieval/") || strings.HasPrefix(response.Module, "reasoning/") || strings.HasPrefix(response.Module, "linux/guided-step")
}

func (m *conversationMemory) recordTopic(query, title string, response Response) {
	_, _, navigation := parseTopicNavigation(normalizeText(query))
	if !isTopicResponse(response) || strings.HasPrefix(response.Module, "retrieval/follow-up/") || navigation {
		return
	}
	key := response.Module + "|" + normalizeText(query)
	for i := range m.topics {
		if m.topics[i].key == key {
			m.topics = append(m.topics[:i], m.topics[i+1:]...)
			break
		}
	}
	response.Sources = append([]Source(nil), response.Sources...)
	response.Suggestions = append([]string(nil), response.Suggestions...)
	m.topics = append(m.topics, conversationTopic{key: key, title: title, query: query, response: response, turn: m.turn})
	if len(m.topics) > conversationTopicLimit {
		m.topics = append([]conversationTopic(nil), m.topics[len(m.topics)-conversationTopicLimit:]...)
	}
}

func (m *conversationMemory) replaceTopicResponse(query string, response Response) {
	for i := range m.topics {
		if normalizeText(m.topics[i].query) != normalizeText(query) {
			continue
		}
		response.Sources = append([]Source(nil), response.Sources...)
		response.Suggestions = append([]string(nil), response.Suggestions...)
		m.topics[i].response = response
		if topic := externalTopic(response); topic != "" {
			m.topics[i].title = "Wikipedia: " + topic
		}
	}
}

func (b *Bot) conversationTopicTitle(response Response) string {
	if topic := externalTopic(response); topic != "" {
		return "Wikipedia: " + topic
	}
	if strings.HasPrefix(response.Module, "retrieval/") {
		id := strings.TrimPrefix(response.Module, "retrieval/")
		if result, found := b.knowledge.documentByID(id); found {
			return result.Document.Title
		}
	}
	switch response.Module {
	case "reasoning/age-comparison":
		return "Age comparison"
	case "reasoning/age-ranking":
		return "Age ranking"
	case "reasoning/arithmetic":
		return "Calculation"
	case "reasoning/date":
		return "Date calculation"
	case "reasoning/unit-conversion":
		return "Unit conversion"
	default:
		return "Linux troubleshooting"
	}
}

func (b *Bot) applyConversationPreferences(user string, response Response) Response {
	if !strings.HasPrefix(response.Module, "retrieval/linux-") && !strings.HasPrefix(response.Module, "linux/") {
		return response
	}
	preferences := b.memory.preferences
	if preferences.guided && strings.HasPrefix(response.Module, "retrieval/linux-") {
		b.memory.guidedSteps = nil
		b.memory.guidedIndex = 0
		b.memory.guidedSources = nil
		steps := splitAnswerSteps(response.Text)
		if len(steps) > 1 {
			b.memory.guidedSteps = steps
			b.memory.guidedIndex = 1
			b.memory.guidedSources = append([]Source(nil), response.Sources...)
			b.memory.lastLinuxQuestion = user
			response.Text = guidedStepText(1, len(steps), steps[0]) + "\n\nRun only this step, then share its output before making changes."
			response.Module = "linux/guided-step"
		}
	}
	if preferences.audience == "beginner" {
		if explanation := explainPreviousCommand(response.Text); explanation != "" && !strings.Contains(response.Text, "Plain English:") {
			response.Text += "\n\nPlain English: " + explanation
		}
	}
	if preferences.verbosity == "detailed" {
		if explanation := explainPreviousCommand(response.Text); explanation != "" && !strings.Contains(response.Text, "Command detail:") {
			response.Text += "\n\nCommand detail: " + explanation
		}
	}
	if preferences.verbosity == "brief" {
		response.Text = briefAnswer(response.Text)
	}
	if preferences.format != "" {
		parts := splitAnswerSteps(response.Text)
		if len(parts) > 1 {
			for i := range parts {
				if preferences.format == "numbered" {
					parts[i] = fmt.Sprintf("%d. %s", i+1, parts[i])
				} else {
					parts[i] = "- " + parts[i]
				}
			}
			response.Text = strings.Join(parts, "\n")
		}
	}
	return response
}

func briefAnswer(text string) string {
	parts := splitAnswerSteps(text)
	if len(parts) <= 1 {
		return strings.TrimSpace(text)
	}
	result := []string{parts[0]}
	for _, part := range parts[1:] {
		lower := strings.ToLower(part)
		if containsAny(lower, "production", "before confirming", "before making changes", "share its output", "share the result", "backup", "delete", "remove", "destructive", "side effect", "verify", "confirm", "risky", "do not", "don't", "must not") {
			result = append(result, part)
			break
		}
	}
	return strings.Join(result, " ")
}

var answerStepBoundaryRE = regexp.MustCompile(`([.!?])\s+`)

func splitAnswerSteps(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	text = answerStepBoundaryRE.ReplaceAllString(text, "$1\n")
	parts := make([]string, 0)
	for _, line := range strings.Split(text, "\n") {
		for _, fragment := range strings.Split(line, "; ") {
			fragment = strings.TrimSpace(fragment)
			if fragment != "" {
				parts = append(parts, fragment)
			}
		}
	}
	return parts
}

func guidedStepText(number, total int, step string) string {
	return fmt.Sprintf("Step %d of %d: %s", number, total, strings.TrimSpace(step))
}

func (b *Bot) answerGuidedFollowup(text string) (Response, bool) {
	if len(b.memory.guidedSteps) == 0 || b.memory.guidedIndex == 0 {
		return Response{}, false
	}
	current := b.memory.guidedIndex - 1
	if oneOf(text, "repeat step", "repeat this step", "repeat the current step", "what was that step") {
		return Response{Text: guidedStepText(current+1, len(b.memory.guidedSteps), b.memory.guidedSteps[current]), Module: "linux/guided-step", Sources: append([]Source(nil), b.memory.guidedSources...)}, true
	}
	if containsAny(text, "that didn't work", "that did not work", "it failed", "the command failed", "i got an error", "i got this error") {
		return Response{Text: "Paste the exact command and complete output, including the exit status if you have it. Redact passwords, tokens, private keys, and hostnames you consider sensitive.", Module: "linux/guided-output", NeedsClarification: true, Sources: append([]Source(nil), b.memory.guidedSources...)}, true
	}
	if containsAny(text, "what does step", "explain step", "what is step") {
		if number := requestedStepNumber(text); number > 0 && number <= len(b.memory.guidedSteps) {
			step := b.memory.guidedSteps[number-1]
			if explanation := explainPreviousCommand(step); explanation != "" {
				return Response{Text: fmt.Sprintf("Step %d means: %s", number, explanation), Module: "linux/guided-explanation", Sources: append([]Source(nil), b.memory.guidedSources...)}, true
			}
			return Response{Text: fmt.Sprintf("Step %d is: %s", number, step), Module: "linux/guided-explanation", Sources: append([]Source(nil), b.memory.guidedSources...)}, true
		}
		return Response{Text: "Which step number do you mean?", Module: "linux/guided-explanation", NeedsClarification: true}, true
	}
	if containsAny(text, "show all steps", "show the rest", "give me all the steps") {
		lines := make([]string, 0, len(b.memory.guidedSteps)-current)
		for i := current; i < len(b.memory.guidedSteps); i++ {
			lines = append(lines, fmt.Sprintf("%d. %s", i+1, b.memory.guidedSteps[i]))
		}
		return Response{Text: strings.Join(lines, "\n"), Module: "linux/guided-plan", Sources: append([]Source(nil), b.memory.guidedSources...)}, true
	}
	if containsAny(text, "skip this step", "skip the step", "skip ahead") || containsAny(text, "that worked", "it worked", "i finished that", "done with that step", "next", "continue", "what now") || isNextStepFollowup(text) {
		if containsAny(text, "skip ahead") && b.memory.guidedIndex < len(b.memory.guidedSteps) {
			b.memory.guidedIndex++
		}
		if b.memory.guidedIndex >= len(b.memory.guidedSteps) {
			return Response{Text: "That's the last step in this answer. Tell me whether it worked, or paste the exact output if it didn't.", Module: "linux/guided-complete", Sources: append([]Source(nil), b.memory.guidedSources...)}, true
		}
		step := b.memory.guidedSteps[b.memory.guidedIndex]
		b.memory.guidedIndex++
		return Response{Text: guidedStepText(b.memory.guidedIndex, len(b.memory.guidedSteps), step) + "\n\nShare the result before proceeding if anything looks unexpected.", Module: "linux/guided-step", Sources: append([]Source(nil), b.memory.guidedSources...)}, true
	}
	return Response{}, false
}

func requestedStepNumber(text string) int {
	match := regexp.MustCompile(`\b(?:step\s+|the\s+)(\d+)\b`).FindStringSubmatch(text)
	if len(match) != 2 {
		for word, number := range map[string]int{"first": 1, "second": 2, "third": 3, "fourth": 4, "fifth": 5} {
			if strings.Contains(text, "step "+word) || strings.Contains(text, "the "+word+" step") {
				return number
			}
		}
		return 0
	}
	var number int
	_, _ = fmt.Sscanf(match[1], "%d", &number)
	return number
}

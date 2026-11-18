// Package pybot provides a small CPU-only conversational bot with statistical intent routing and explicit answer rules.
//
// A Bot owns its conversation state. Calls to Reply and Reset are safe to make
// from multiple goroutines, though one Bot still represents one conversation.
package pybot

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

// Version is the version of this Go implementation.
const Version = "0.2.0"

// Response is a reply and the rule module that selected it.
type Response struct {
	Text   string
	Module string
}

// Bot keeps conversational facts and selects the strongest matching rule.
type Bot struct {
	mu       sync.Mutex
	seed     int64
	rng      *rand.Rand
	name     string
	ages     map[string]int
	ageOrder []string
	topic    string
}

type proposal struct {
	text   string
	module string
	score  int
}

var (
	nameStatementRE = regexp.MustCompile(`(?i)(?:my name is|i am|i'm|call me)\s+([\p{L}][\p{L}'-]*)`)
	ageStatementRE  = regexp.MustCompile(`(?i)\b([\p{L}][\p{L}'-]*(?:\s+[\p{L}][\p{L}'-]*)?)\s+(?:is\s+(?:(?:the\s+)?age(?:\s+of)?\s+|aged\s+)?|has\s+(?:the\s+)?age\s+of\s+)(\d{1,3})(?:\s*(?:years?|yrs?)\s*old)?\b`)
)

// New creates a bot whose wording choices are repeatable for seed.
func New(seed int64) *Bot {
	return &Bot{
		seed: seed,
		rng:  rand.New(rand.NewSource(seed)),
		ages: make(map[string]int),
	}
}

// Reset starts a fresh conversation and restarts seeded wording choices.
func (b *Bot) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.name = ""
	b.ages = make(map[string]int)
	b.ageOrder = nil
	b.topic = ""
	b.rng = rand.New(rand.NewSource(b.seed))
}

// Reply answers one message and remembers simple facts for later turns.
func (b *Bot) Reply(input string) Response {
	b.mu.Lock()
	defer b.mu.Unlock()

	text := strings.TrimSpace(input)
	if text == "" {
		return Response{Text: "Hello!", Module: "conversation/fallback"}
	}
	lower := strings.ToLower(text)
	messageIntent, intentScore := defaultIntentClassifier.classify(lower)
	statisticalRoute := intentScore >= intentRouteThreshold
	ageUpdates := b.rememberAges(text)

	proposals := make([]proposal, 0, 8)
	if isAgeComparison(lower) || b.hasKnownAgeComparison(lower) ||
		(messageIntent == intentAgeComparison && statisticalRoute && strings.Contains(lower, "age")) {
		if len(b.ageOrder) >= 2 {
			proposals = append(proposals, proposal{b.compareAges(lower), "reasoning/age-comparison", 120})
		} else {
			proposals = append(proposals, proposal{
				"Tell me each person's age and I can compare them, for example: Alice is 12 and Bob is 50. Who is older?",
				"reasoning/age-comparison", 120,
			})
		}
	}

	if isNameQuestion(lower) || (messageIntent == intentNameQuestion && statisticalRoute) {
		if b.name != "" {
			proposals = append(proposals, proposal{"Your name is " + b.name + ".", "memory/name", 110})
		} else {
			proposals = append(proposals, proposal{"I don't know your name yet. Tell me, for example, \"I'm Sam.\"", "memory/name", 110})
		}
	}
	if match := nameStatementRE.FindStringSubmatch(text); len(match) == 2 {
		b.name = capitalizeName(match[1])
		proposals = append(proposals, proposal{"Nice to meet you, " + b.name + ". What can I help with?", "memory/name", 100})
	}
	if ageUpdates > 0 && len(proposals) == 0 {
		proposals = append(proposals, proposal{"Got it. I'll remember those ages for this conversation.", "memory/ages", 65})
	}

	if isGreeting(lower) || (messageIntent == intentGreeting && statisticalRoute) {
		greeting := "Hello! What can I help with?"
		if b.name != "" {
			greeting = "Hello, " + b.name + "! What can I help with?"
		}
		proposals = append(proposals, proposal{greeting, "conversation/greeting", 60})
	}
	if strings.Contains(lower, "who are you") || strings.Contains(lower, "what are you") ||
		(messageIntent == intentIdentity && statisticalRoute) {
		proposals = append(proposals, proposal{"I'm PyBot Go, a small CPU-only chatbot. I use a lightweight statistical router and explicit answer rules, not a language model.", "conversation/identity", 75})
	}
	if strings.Contains(lower, "how are you") || (messageIntent == intentSmallTalk && statisticalRoute) {
		proposals = append(proposals, proposal{"I'm running fine. What would you like to talk about?", "conversation/small-talk", 55})
	}
	if isThanks(lower) || (messageIntent == intentThanks && statisticalRoute) {
		proposals = append(proposals, proposal{"You're welcome!", "conversation/thanks", 55})
	}

	if (strings.Contains(lower, "onion") && (strings.Contains(lower, "cry") || strings.Contains(lower, "eyes"))) ||
		(messageIntent == intentOnion && statisticalRoute) {
		proposals = append(proposals, proposal{"Cutting an onion releases volatile sulfur compounds. They react with moisture in your eyes and form a mild irritant, which makes your eyes water. A sharp knife and chilling the onion first can reduce the effect.", "knowledge/cooking", 80})
		b.topic = "onion"
	}
	if (strings.Contains(lower, "kettle") && (strings.Contains(lower, "lime") || strings.Contains(lower, "scale") || strings.Contains(lower, "clean") || strings.Contains(lower, "descal"))) ||
		(messageIntent == intentKettle && statisticalRoute) {
		proposals = append(proposals, proposal{"Half fill the kettle with equal parts water and white vinegar, or use citric acid. Boil it, leave it for about 30 minutes, empty it, then boil and discard fresh water twice before using the kettle again.", "knowledge/cooking", 85})
		b.topic = "kettle-descaling"
	}
	if isDurationFollowUp(lower) && b.topic == "kettle-descaling" {
		proposals = append(proposals, proposal{"Leave the solution in the kettle for about 30 minutes.", "memory/follow-up", 90})
	}
	if (strings.Contains(lower, "boil") && strings.Contains(lower, "egg")) ||
		(messageIntent == intentEggs && statisticalRoute) {
		proposals = append(proposals, proposal{"For a large egg placed into boiling water, cook about 6 minutes for a soft yolk, 8–9 minutes for a jammy yolk, or 10–12 minutes for a firm yolk. Cool it in cold water to stop the cooking.", "knowledge/cooking", 80})
		b.topic = "boiling-eggs"
	}
	if len(proposals) == 0 {
		proposals = append(proposals, proposal{b.fallback(), "conversation/fallback", 0})
	}
	winner := arbitrate(proposals)
	return Response{Text: winner.text, Module: winner.module}
}

func arbitrate(proposals []proposal) proposal {
	// Stable ordering makes equal-score rules predictable and easy to inspect.
	winner := proposals[0]
	for _, candidate := range proposals[1:] {
		if candidate.score > winner.score {
			winner = candidate
		}
	}
	return winner
}

func (b *Bot) fallback() string {
	choices := []string{
		"I don't have a rule for that yet. Try asking a concrete question, or tell me a fact I can use.",
		"I can't work that out from the rules I have. Can you give me a little more detail?",
		"I don't know. I'm a small rule-based bot, so I may need a more specific question.",
	}
	return choices[b.rng.Intn(len(choices))]
}

func (b *Bot) rememberAges(text string) int {
	matches := ageStatementRE.FindAllStringSubmatch(text, -1)
	genericIndex := 0
	updated := 0
	for _, match := range matches {
		if len(match) != 3 {
			continue
		}
		age, err := strconv.Atoi(match[2])
		if err != nil || age > 150 {
			continue
		}
		person := canonicalPerson(match[1])
		if isGenericPerson(person) {
			genericIndex++
			person = fmt.Sprintf("person %d", genericIndex)
		}
		if _, exists := b.ages[person]; !exists {
			b.ageOrder = append(b.ageOrder, person)
		}
		b.ages[person] = age
		updated++
	}
	return updated
}

func canonicalPerson(person string) string {
	person = strings.ToLower(strings.TrimSpace(person))
	person = strings.Join(strings.Fields(person), " ")
	return strings.Trim(person, " ,.;:!?'")
}

func isGenericPerson(person string) bool {
	switch person {
	case "a person", "one person", "another person", "the other person", "person", "someone", "somebody", "someone else", "somebody else", "first person", "second person", "person one", "person two", "person a", "person b":
		return true
	default:
		return false
	}
}

func displayPerson(person string) string {
	if strings.HasPrefix(person, "person ") {
		return "Person " + strings.TrimPrefix(person, "person ")
	}
	fields := strings.Fields(person)
	for i, field := range fields {
		if field == "" {
			continue
		}
		runes := []rune(field)
		runes[0] = unicode.ToUpper(runes[0])
		fields[i] = string(runes)
	}
	return strings.Join(fields, " ")
}

func isAgeComparison(text string) bool {
	comparative := strings.Contains(text, "older") || strings.Contains(text, "younger") ||
		strings.Contains(text, "oldest") || strings.Contains(text, "youngest") ||
		strings.Contains(text, "higher age") || strings.Contains(text, "greater age") || strings.Contains(text, "lower age")
	if !comparative {
		return false
	}
	if strings.Contains(text, "age") || strings.Contains(text, "year old") || strings.Contains(text, "years old") {
		return true
	}
	// A direct people-comparison question may rely on ages stored in earlier turns.
	return strings.HasPrefix(text, "who ") || strings.HasPrefix(text, "who's ") ||
		strings.HasPrefix(text, "which ")
}

func (b *Bot) hasKnownAgeComparison(text string) bool {
	comparative := strings.Contains(text, "older") || strings.Contains(text, "younger") ||
		strings.Contains(text, "oldest") || strings.Contains(text, "youngest")
	if !comparative {
		return false
	}
	matched := 0
	for person := range b.ages {
		if !strings.HasPrefix(person, "person ") && strings.Contains(text, person) {
			matched++
		}
	}
	return matched >= 2
}

func (b *Bot) compareAges(question string) string {
	if len(b.ageOrder) < 2 {
		return "I need two or more ages to compare."
	}
	olderQuestion := strings.Contains(question, "older") || strings.Contains(question, "oldest") ||
		strings.Contains(question, "higher age") || strings.Contains(question, "greater age")
	ordered := make([]string, 0, len(b.ageOrder))
	for _, person := range b.ageOrder {
		if !strings.HasPrefix(person, "person ") && strings.Contains(question, person) {
			ordered = append(ordered, person)
		}
	}
	if len(ordered) < 2 {
		ordered = append(ordered, b.ageOrder...)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if olderQuestion {
			return b.ages[ordered[i]] > b.ages[ordered[j]]
		}
		return b.ages[ordered[i]] < b.ages[ordered[j]]
	})
	first, second := ordered[0], ordered[1]
	firstAge, secondAge := b.ages[first], b.ages[second]
	if firstAge == secondAge {
		return fmt.Sprintf("%s and %s are the same age: %d.", displayPerson(first), displayPerson(second), firstAge)
	}
	if olderQuestion {
		return fmt.Sprintf("%s is older: %d versus %d years old, a difference of %d years.", displayPerson(first), firstAge, secondAge, firstAge-secondAge)
	}
	return fmt.Sprintf("%s is younger: %d versus %d years old, a difference of %d years.", displayPerson(first), firstAge, secondAge, secondAge-firstAge)
}

func isNameQuestion(text string) bool {
	return strings.Contains(text, "what's my name") || strings.Contains(text, "what is my name") ||
		strings.Contains(text, "do you remember my name") || strings.Contains(text, "who am i")
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

func isDurationFollowUp(text string) bool {
	return strings.Contains(text, "how long") && (strings.Contains(text, "that take") || strings.Contains(text, "leave it"))
}

func capitalizeName(name string) string {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}
	return string(runes)
}

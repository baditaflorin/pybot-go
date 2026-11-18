package pybot

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type factKey struct {
	subject   string
	attribute string
}

type conversationMemory struct {
	turn              int
	facts             map[factKey]Fact
	factOrder         []factKey
	nextGenericPerson int
	lastPersonSubject string
	lastKnowledgeID   string
	lastKnowledgeTurn int
}

var (
	nameStatementRE = regexp.MustCompile(`(?i)(?:my name is|i am|i'm|call me)\s+([\p{L}][\p{L}'-]*)`)
	ageStatementRE  = regexp.MustCompile(`(?i)\b([\p{L}][\p{L}'-]*(?:\s+[\p{L}][\p{L}'-]*)?)\s+(?:is\s+(?:(?:the\s+)?age(?:\s+of)?\s+|aged\s+)?|has\s+(?:the\s+)?age\s+of\s+)(\d{1,3})(?:\s*(?:years?|yrs?)\s*old)?\b`)
	selfAgeRE       = regexp.MustCompile(`(?i)\b(?:i am|i'm)\s+(\d{1,3})\s*(?:years?|yrs?)\s*old\b`)
)

func newConversationMemory() conversationMemory {
	return conversationMemory{facts: make(map[factKey]Fact)}
}

func (m *conversationMemory) remember(text string) int {
	m.turn++
	updates := 0
	if match := nameStatementRE.FindStringSubmatch(text); len(match) == 2 {
		m.setFact(Fact{
			Subject: "user", Attribute: "name", Value: capitalizeName(match[1]),
			Source: text, Turn: m.turn,
		})
		updates++
	}
	if match := selfAgeRE.FindStringSubmatch(text); len(match) == 2 {
		if age, err := strconv.Atoi(match[1]); err == nil && age <= 150 {
			m.setFact(Fact{Subject: "user", Attribute: "age", Value: strconv.Itoa(age), Unit: "years", Source: text, Turn: m.turn})
			updates++
		}
	}
	for _, match := range ageStatementRE.FindAllStringSubmatch(text, -1) {
		if len(match) != 3 {
			continue
		}
		age, err := strconv.Atoi(match[2])
		if err != nil || age > 150 {
			continue
		}
		subject := canonicalPerson(match[1])
		if isAgePronoun(subject) && m.lastPersonSubject != "" {
			subject = m.lastPersonSubject
		}
		if isUnresolvedAgePronoun(subject) {
			continue
		}
		if isGenericPerson(subject) {
			m.nextGenericPerson++
			subject = fmt.Sprintf("person %d", m.nextGenericPerson)
		}
		m.setFact(Fact{
			Subject: subject, Attribute: "age", Value: strconv.Itoa(age), Unit: "years",
			Source: text, Turn: m.turn,
		})
		if !strings.HasPrefix(subject, "person ") {
			m.lastPersonSubject = subject
		}
		updates++
	}
	return updates
}

func unresolvedAgePronoun(text string) string {
	for _, match := range ageStatementRE.FindAllStringSubmatch(text, -1) {
		if len(match) != 3 {
			continue
		}
		subject := canonicalPerson(match[1])
		if isUnresolvedAgePronoun(subject) {
			return subject
		}
	}
	return ""
}

func (m *conversationMemory) setFact(fact Fact) {
	key := factKey{subject: fact.Subject, attribute: fact.Attribute}
	if _, exists := m.facts[key]; !exists {
		m.factOrder = append(m.factOrder, key)
	}
	m.facts[key] = fact
}

func (m *conversationMemory) factsInOrder() []Fact {
	facts := make([]Fact, 0, len(m.factOrder))
	for _, key := range m.factOrder {
		if fact, exists := m.facts[key]; exists {
			facts = append(facts, fact)
		}
	}
	return facts
}

func (m *conversationMemory) name() string {
	fact, ok := m.facts[factKey{subject: "user", attribute: "name"}]
	if !ok {
		return ""
	}
	return fact.Value
}

func (m *conversationMemory) ages() []Fact {
	ages := make([]Fact, 0)
	for _, fact := range m.factsInOrder() {
		if fact.Attribute == "age" {
			ages = append(ages, fact)
		}
	}
	return ages
}

func canonicalPerson(person string) string {
	person = strings.ToLower(strings.TrimSpace(person))
	person = strings.Join(strings.Fields(person), " ")
	person = strings.Trim(person, " ,.;:!?'\"()")
	person = strings.TrimPrefix(person, "and ")
	person = strings.TrimPrefix(person, "then ")
	return person
}

func isGenericPerson(person string) bool {
	switch person {
	case "a person", "one person", "another person", "the other person", "that person", "this person", "person", "someone", "somebody", "someone else", "somebody else", "first person", "second person", "person one", "person two", "person a", "person b":
		return true
	default:
		return false
	}
}

func isAgePronoun(subject string) bool {
	switch subject {
	case "this person", "that person":
		return true
	default:
		return false
	}
}

func isUnresolvedAgePronoun(subject string) bool {
	switch subject {
	case "he", "she", "they", "him", "her", "them":
		return true
	default:
		return false
	}
}

func displayPerson(person string) string {
	if person == "user" {
		return "You"
	}
	if strings.HasPrefix(person, "person ") {
		return "Person " + strings.TrimPrefix(person, "person ")
	}
	fields := strings.Fields(person)
	for i, field := range fields {
		runes := []rune(field)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
		}
		fields[i] = string(runes)
	}
	return strings.Join(fields, " ")
}

func capitalizeName(name string) string {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) > 0 {
		runes[0] = unicode.ToUpper(runes[0])
	}
	return string(runes)
}

package pybot

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func (b *Bot) answerAgeComparison(question string) Response {
	ages := b.memory.ages()
	if len(ages) == 0 {
		return Response{
			Text:   "I don't have any ages to compare yet. Tell me something like \"Alice is 12 and Bob is 50.\"",
			Module: "reasoning/age-comparison", NeedsClarification: true,
		}
	}

	selected := selectNamedAgeFacts(question, ages)
	if len(selected) < 2 {
		if len(ages) == 1 {
			return Response{
				Text:   fmt.Sprintf("I have %s at %s years old. Who is the other person, and how old are they?", displayPerson(ages[0].Subject), ages[0].Value),
				Module: "reasoning/age-comparison", NeedsClarification: true,
			}
		}
		selected = ages
	}

	olderQuestion := strings.Contains(question, "older") || strings.Contains(question, "oldest") ||
		strings.Contains(question, "higher age") || strings.Contains(question, "greater age")
	sort.SliceStable(selected, func(i, j int) bool {
		left, _ := parseAge(selected[i].Value)
		right, _ := parseAge(selected[j].Value)
		if olderQuestion {
			return left > right
		}
		return left < right
	})
	first, _ := parseAge(selected[0].Value)
	second, _ := parseAge(selected[1].Value)
	if first == second {
		return Response{
			Text:   fmt.Sprintf("%s and %s are the same age: %d.", displayPerson(selected[0].Subject), displayPerson(selected[1].Subject), first),
			Module: "reasoning/age-comparison",
		}
	}
	if olderQuestion {
		if len(selected) > 2 {
			return Response{Text: fmt.Sprintf("%s is oldest at %d. %s is next at %d, a difference of %d years.",
				displayPerson(selected[0].Subject), first, displayPerson(selected[1].Subject), second, first-second), Module: "reasoning/age-comparison"}
		}
		return Response{Text: fmt.Sprintf("%s is older: %d versus %d years old, a difference of %d years.",
			displayPerson(selected[0].Subject), first, second, first-second), Module: "reasoning/age-comparison"}
	}
	if len(selected) > 2 {
		return Response{Text: fmt.Sprintf("%s is youngest at %d. %s is next at %d, a difference of %d years.",
			displayPerson(selected[0].Subject), first, displayPerson(selected[1].Subject), second, second-first), Module: "reasoning/age-comparison"}
	}
	return Response{Text: fmt.Sprintf("%s is younger: %d versus %d years old, a difference of %d years.",
		displayPerson(selected[0].Subject), first, second, second-first), Module: "reasoning/age-comparison"}
}

func selectNamedAgeFacts(question string, ages []Fact) []Fact {
	selected := make([]Fact, 0, len(ages))
	for _, fact := range ages {
		if strings.HasPrefix(fact.Subject, "person ") {
			continue
		}
		if phraseMentioned(question, fact.Subject) {
			selected = append(selected, fact)
		}
	}
	if len(selected) >= 2 {
		return selected
	}
	return ages
}

func phraseMentioned(text, phrase string) bool {
	paddedText := " " + strings.NewReplacer("?", " ", ",", " ", ".", " ", "!", " ", "'", " ").Replace(text) + " "
	paddedPhrase := " " + strings.ReplaceAll(phrase, "'", " ") + " "
	return strings.Contains(paddedText, paddedPhrase)
}

func parseAge(value string) (int, bool) {
	var age int
	_, err := fmt.Sscanf(value, "%d", &age)
	return age, err == nil
}

var agePersonQuestionRE = regexp.MustCompile(`(?i)(?:how old is|how old are|what is the age of|what's the age of)\s+([\p{L}][\p{L}'-]*(?:\s+[\p{L}][\p{L}'-]*)?)`)

func (b *Bot) answerAgeFact(question string) Response {
	subject := ""
	if strings.Contains(question, "how old am i") {
		subject = "user"
	} else if match := agePersonQuestionRE.FindStringSubmatch(question); len(match) == 2 {
		subject = canonicalPerson(match[1])
		if isAgePronoun(subject) {
			if b.memory.lastPersonSubject == "" {
				return Response{Text: "Which person do you mean by that? Please tell me their name.", Module: "memory/age", NeedsClarification: true}
			}
			subject = b.memory.lastPersonSubject
		} else if isUnresolvedAgePronoun(subject) {
			return Response{Text: fmt.Sprintf("Who do you mean by %q? Please tell me their name.", subject), Module: "memory/age", NeedsClarification: true}
		}
	}
	if subject == "" {
		return Response{Text: "Whose age do you mean?", Module: "memory/age", NeedsClarification: true}
	}
	for _, fact := range b.memory.ages() {
		if fact.Subject == subject {
			if fact.Subject == "user" {
				return Response{Text: fmt.Sprintf("You are %s years old.", fact.Value), Module: "memory/age"}
			}
			return Response{Text: fmt.Sprintf("%s is %s years old.", displayPerson(fact.Subject), fact.Value), Module: "memory/age"}
		}
	}
	return Response{Text: fmt.Sprintf("I don't have %s's age in this conversation. Tell me their age and I'll remember it here.", displayPerson(subject)), Module: "memory/age", NeedsClarification: true}
}

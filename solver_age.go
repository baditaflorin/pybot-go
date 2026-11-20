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
var ageRankingSubjectRE = regexp.MustCompile(`(?i)\b(?:where\s+(?:does|do|would)|how\s+does)\s+([\p{L}][\p{L}'-]*(?:\s+[\p{L}][\p{L}'-]*)?)\s+(?:sit|rank|fit|stand|fall|place)\b`)

type rankedAgeFact struct {
	fact Fact
	age  int
}

func (b *Bot) answerAgeRanking(question string) Response {
	ages := b.memory.ages()
	if len(ages) < 2 {
		return Response{
			Text:   "I need at least two remembered ages to place someone in the ranking. Tell me each person's age first.",
			Module: "reasoning/age-ranking", NeedsClarification: true,
		}
	}

	subject := ""
	if match := ageRankingSubjectRE.FindStringSubmatch(question); len(match) == 2 {
		subject = canonicalPerson(match[1])
		if isAgeRankingPronoun(subject) {
			subject = b.memory.lastPersonSubject
		}
	}
	if subject == "" {
		subject = b.memory.lastPersonSubject
	}
	if subject == "" {
		return Response{
			Text:   "Whose position should I place in the age ranking? Name the person and I'll compare their age with the others.",
			Module: "reasoning/age-ranking", NeedsClarification: true,
		}
	}

	ranked := make([]rankedAgeFact, 0, len(ages))
	targetFound := false
	for _, fact := range ages {
		age, ok := parseAge(fact.Value)
		if !ok {
			continue
		}
		ranked = append(ranked, rankedAgeFact{fact: fact, age: age})
		if fact.Subject == subject {
			targetFound = true
		}
	}
	if !targetFound {
		return Response{
			Text:   fmt.Sprintf("I don't have %s's age in this conversation. Tell me their age and I'll place them.", displayPerson(subject)),
			Module: "reasoning/age-ranking", NeedsClarification: true,
		}
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].age < ranked[j].age })

	target := -1
	for i, item := range ranked {
		if item.fact.Subject == subject {
			target = i
			break
		}
	}
	if target < 0 || len(ranked) < 2 {
		return Response{
			Text:   "I need at least two remembered ages to place someone in the ranking.",
			Module: "reasoning/age-ranking", NeedsClarification: true,
		}
	}

	person := displayPerson(ranked[target].fact.Subject)
	targetAge := ranked[target].age
	firstEqual, lastEqual := target, target
	for firstEqual > 0 && ranked[firstEqual-1].age == targetAge {
		firstEqual--
	}
	for lastEqual+1 < len(ranked) && ranked[lastEqual+1].age == targetAge {
		lastEqual++
	}
	if firstEqual != lastEqual {
		others := make([]string, 0, lastEqual-firstEqual)
		for i := firstEqual; i <= lastEqual; i++ {
			if i != target {
				others = append(others, displayPerson(ranked[i].fact.Subject))
			}
		}
		return Response{
			Text: fmt.Sprintf("%s is tied with %s at %d years old (%s of %d by age).", person,
				strings.Join(others, " and "), targetAge, ordinal(firstEqual+1), len(ranked)),
			Module: "reasoning/age-ranking",
		}
	}

	if target == 0 {
		next := ranked[1]
		return Response{
			Text: fmt.Sprintf("%s is youngest of %d at %d years old. %s is next at %d, so %s is %d years younger.",
				person, len(ranked), targetAge, displayPerson(next.fact.Subject), next.age, person, next.age-targetAge),
			Module: "reasoning/age-ranking",
		}
	}
	if target == len(ranked)-1 {
		previous := ranked[target-1]
		return Response{
			Text: fmt.Sprintf("%s is oldest of %d at %d years old. %s is next at %d, so %s is %d years older.",
				person, len(ranked), targetAge, displayPerson(previous.fact.Subject), previous.age, person, targetAge-previous.age),
			Module: "reasoning/age-ranking",
		}
	}

	younger := ranked[target-1]
	older := ranked[target+1]
	placement := fmt.Sprintf("That puts %s %s youngest of %d.", person, ordinal(target+1), len(ranked))
	if len(ranked) == 3 && target == 1 {
		placement = fmt.Sprintf("That puts %s in the middle of the three.", person)
	}
	return Response{
		Text: fmt.Sprintf("%s sits between %s (%d) and %s (%d) at %d years old: %d years older than %s and %d years younger than %s. %s",
			person, displayPerson(younger.fact.Subject), younger.age, displayPerson(older.fact.Subject), older.age,
			targetAge, targetAge-younger.age, displayPerson(younger.fact.Subject), older.age-targetAge,
			displayPerson(older.fact.Subject), placement),
		Module: "reasoning/age-ranking",
	}
}

func isAgeRankingPronoun(subject string) bool {
	switch subject {
	case "he", "she", "they", "him", "her", "them", "it", "this", "that", "this person", "that person":
		return true
	default:
		return false
	}
}

func ordinal(value int) string {
	suffix := "th"
	if value%100 < 11 || value%100 > 13 {
		switch value % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return fmt.Sprintf("%d%s", value, suffix)
}

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

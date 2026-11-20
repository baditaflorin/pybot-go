package pybot

import (
	"math"
	"regexp"
	"strings"
	"unicode"
)

type intent string

const (
	intentGreeting        intent = "greeting"
	intentNameQuestion    intent = "name_question"
	intentAgeComparison   intent = "age_comparison"
	intentAgeRanking      intent = "age_ranking"
	intentArithmetic      intent = "arithmetic"
	intentConversion      intent = "unit_conversion"
	intentDate            intent = "date_math"
	intentIdentity        intent = "identity"
	intentOnion           intent = "onion"
	intentKettle          intent = "kettle"
	intentEggs            intent = "eggs"
	intentThanks          intent = "thanks"
	intentSmallTalk       intent = "small_talk"
	intentKnowledge       intent = "knowledge_question"
	intentLinuxExpert     intent = "linux_expert"
	intentContextFollowup intent = "context_followup"
)

type intentExample struct {
	label intent
	text  string
}

type intentClassifier struct {
	labels      []intent
	documentN   map[intent]int
	tokenCounts map[intent]map[string]int
	totalTokens map[intent]int
	vocabulary  map[string]struct{}
	totalDocs   int
}

type phraseAlias struct {
	pattern   *regexp.Regexp
	canonical string
}

var intentAliases = []phraseAlias{
	{regexp.MustCompile(`\bhow have you been\b`), "how are you"},
	{regexp.MustCompile(`\bhow have things been\b`), "how are you"},
	{regexp.MustCompile(`\bwork out\b`), "calculate"},
	{regexp.MustCompile(`\bfigure out\b`), "calculate"},
	{regexp.MustCompile(`\bmultiplied by\b`), "times"},
	{regexp.MustCompile(`\bexpress\b`), "convert"},
	{regexp.MustCompile(`\btranslate\b`), "convert"},
	{regexp.MustCompile(`\bweekday\b`), "date day"},
	{regexp.MustCompile(`\beldest\b`), "older"},
	{regexp.MustCompile(`\belders?\b`), "older"},
	{regexp.MustCompile(`\bmost senior\b`), "older"},
	{regexp.MustCompile(`\bmost junior\b`), "younger"},
	{regexp.MustCompile(`\bhard[- ]cook\b`), "boil"},
	{regexp.MustCompile(`\bmany thanks\b`), "thank you"},
	{regexp.MustCompile(`\bthank you so much\b`), "thank you"},
	{regexp.MustCompile(`\bwatery eyes\b`), "eyes water"},
}

func normalizeIntentAliases(text string) string {
	text = strings.ToLower(text)
	text = strings.NewReplacer("’", "'", "‘", "'", "“", `"`, "”", `"`).Replace(text)
	for _, alias := range intentAliases {
		text = alias.pattern.ReplaceAllString(text, alias.canonical)
	}
	return text
}

// The examples are intentionally inspectable. Add paraphrases from the
// evaluation corpus; no remote model or generated training data is used.
var intentExamples = []intentExample{
	{intentGreeting, "hello"}, {intentGreeting, "hi there"}, {intentGreeting, "hey"},
	{intentGreeting, "good morning"}, {intentGreeting, "good evening"}, {intentGreeting, "howdy"},
	{intentGreeting, "yo there"}, {intentGreeting, "greetings"},

	{intentNameQuestion, "what is my name"}, {intentNameQuestion, "what's my name"},
	{intentNameQuestion, "do you remember my name"}, {intentNameQuestion, "who am i"},
	{intentNameQuestion, "what do you call me"}, {intentNameQuestion, "what name did i tell you"},
	{intentNameQuestion, "what was the name i gave you"},

	{intentContextFollowup, "repeat your last answer"},
	{intentContextFollowup, "can you say that again"},
	{intentContextFollowup, "show me the sources for that"},
	{intentContextFollowup, "what references support that answer"},
	{intentContextFollowup, "make your last response shorter"},
	{intentContextFollowup, "explain that step to me"},
	{intentContextFollowup, "why did you recommend those commands"},
	{intentContextFollowup, "is that command safe to run"},
	{intentContextFollowup, "walk me through the previous answer"},
	{intentContextFollowup, "continue with the next steps"},

	{intentAgeComparison, "who is older"}, {intentAgeComparison, "who's younger"},
	{intentAgeComparison, "which person is older"}, {intentAgeComparison, "compare their ages"},
	{intentAgeComparison, "who has the higher age"}, {intentAgeComparison, "who is the youngest"},
	{intentAgeComparison, "alice is 12 and bob is 50 who is older"},
	{intentAgeComparison, "one person has the age of 12 somebody is the age of 50 who is older"},
	{intentAgeComparison, "is sam older than pat"}, {intentAgeComparison, "who is oldest in the family"},
	{intentAgeRanking, "where does george rank by age"},
	{intentAgeRanking, "where does he sit among these people"},
	{intentAgeRanking, "put alice bob and george in age order"},
	{intentAgeRanking, "who is between alice and bob in age"},

	{intentArithmetic, "what is 12 plus 4"}, {intentArithmetic, "calculate 8 times 7"},
	{intentArithmetic, "compute 100 divided by 4"}, {intentArithmetic, "what is 18 minus 6"},
	{intentArithmetic, "solve (3 + 4) * 2"}, {intentArithmetic, "add twelve and fifty"},
	{intentArithmetic, "how much is 2 plus 2"}, {intentArithmetic, "what is the sum of 8 and 9"},

	{intentConversion, "convert 100 centimeters to meters"},
	{intentConversion, "how many inches are in 2 meters"},
	{intentConversion, "what is 5 feet in centimeters"},
	{intentConversion, "change 20 celsius to fahrenheit"},
	{intentConversion, "how many pounds is 3 kilograms"},
	{intentConversion, "convert 10 miles into kilometers"},

	{intentDate, "what date is 30 days after 2026-11-18"},
	{intentDate, "how many days between 2026-01-01 and 2026-01-10"},
	{intentDate, "what day will it be 2 weeks from today"},
	{intentDate, "what was the date 3 months before 2025-05-01"},

	{intentIdentity, "who are you"}, {intentIdentity, "what are you"},
	{intentIdentity, "are you an ai"}, {intentIdentity, "are you a language model"},
	{intentIdentity, "are you a chatbot"},

	{intentKnowledge, "what is romania"}, {intentKnowledge, "tell me about the country romania"},
	{intentKnowledge, "who was marie curie"}, {intentKnowledge, "what does photosynthesis mean"},
	{intentKnowledge, "where is the carpathian mountains"},

	{intentLinuxExpert, "how do i troubleshoot a linux server"},
	{intentLinuxExpert, "how do i check disk space on ubuntu"},
	{intentLinuxExpert, "why is my systemd service failing"},
	{intentLinuxExpert, "how do i see which process uses port 8080"},
	{intentLinuxExpert, "how do i inspect linux logs and network configuration"},
	{intentLinuxExpert, "explain chmod permissions on debian"},
	{intentLinuxExpert, "how do i update packages on fedora"},

	{intentOnion, "why do onions make you cry"},
	{intentOnion, "why do onions make my eyes water"},
	{intentOnion, "what makes onions sting your eyes"},
	{intentOnion, "how do onions affect your eyes"},
	{intentOnion, "why does chopping an onion make my eyes tear"},
	{intentOnion, "how can i cut onions without crying"},

	{intentKettle, "how do i get rid of limescale in my kettle"},
	{intentKettle, "how can i descale a kettle"},
	{intentKettle, "how to remove mineral deposits from a kettle"},
	{intentKettle, "how do i clean scale buildup from a tea kettle"},
	{intentKettle, "my kettle has hard water deposits how can i clean it"},

	{intentEggs, "how long should i boil an egg"},
	{intentEggs, "how to cook a soft boiled egg"},
	{intentEggs, "how many minutes for hard boiled eggs"},
	{intentEggs, "what is the boiling time for eggs"},
	{intentEggs, "how long do i cook an egg in boiling water"},

	{intentThanks, "thanks"}, {intentThanks, "thank you"}, {intentThanks, "thanks a lot"},
	{intentThanks, "i appreciate it"}, {intentThanks, "much appreciated"},
	{intentThanks, "thank you for the help"},

	{intentSmallTalk, "how are you"}, {intentSmallTalk, "how are you doing"},
	{intentSmallTalk, "are you doing okay"}, {intentSmallTalk, "how is it going"},
	{intentSmallTalk, "how do you feel"},
}

var defaultIntentClassifier = trainIntentClassifier(intentExamples)

func trainIntentClassifier(examples []intentExample) *intentClassifier {
	model := &intentClassifier{
		documentN:   make(map[intent]int),
		tokenCounts: make(map[intent]map[string]int),
		totalTokens: make(map[intent]int),
		vocabulary:  make(map[string]struct{}),
	}
	seen := make(map[intent]bool)
	for _, example := range examples {
		if !seen[example.label] {
			seen[example.label] = true
			model.labels = append(model.labels, example.label)
			model.tokenCounts[example.label] = make(map[string]int)
		}
		model.documentN[example.label]++
		model.totalDocs++
		for _, feature := range textFeatures(example.text) {
			model.tokenCounts[example.label][feature]++
			model.totalTokens[example.label]++
			model.vocabulary[feature] = struct{}{}
		}
	}
	return model
}

// classify returns the most likely intent and its normalized Naive Bayes score.
// The score is a routing signal, not a calibrated probability of correctness.
func (model *intentClassifier) classify(text string) (intent, float64) {
	features := textFeatures(text)
	if len(features) == 0 || model.totalDocs == 0 {
		return "", 0
	}
	known := make([]string, 0, len(features))
	for _, feature := range features {
		for _, label := range model.labels {
			if model.tokenCounts[label][feature] > 0 {
				known = append(known, feature)
				break
			}
		}
	}
	if len(known) < 2 {
		return "", 0
	}
	features = known

	best := intent("")
	bestScore := math.Inf(-1)
	scores := make([]float64, 0, len(model.labels))
	for _, label := range model.labels {
		score := math.Log(float64(model.documentN[label]) / float64(model.totalDocs))
		denominator := float64(model.totalTokens[label] + len(model.vocabulary))
		for _, feature := range features {
			score += math.Log(float64(model.tokenCounts[label][feature]+1) / denominator)
		}
		scores = append(scores, score)
		if score > bestScore {
			best = label
			bestScore = score
		}
	}

	probabilityTotal := 0.0
	for _, score := range scores {
		probabilityTotal += math.Exp(score - bestScore)
	}
	return best, 1 / probabilityTotal
}

func textFeatures(text string) []string {
	words := make([]string, 0, 12)
	var current strings.Builder
	flush := func() {
		if current.Len() == 0 {
			return
		}
		word := stemWord(current.String())
		if word != "" {
			words = append(words, word)
		}
		current.Reset()
	}
	for _, r := range normalizeIntentAliases(text) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			current.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()

	features := make([]string, 0, len(words)*2)
	for _, word := range words {
		if !isIntentStopWord(word) {
			features = append(features, word)
		}
	}
	for i := 0; i+1 < len(words); i++ {
		features = append(features, words[i]+"_"+words[i+1])
	}
	return features
}

func stemWord(word string) string {
	if strings.HasSuffix(word, "ies") && len(word) > 4 {
		return strings.TrimSuffix(word, "ies") + "y"
	}
	for _, suffix := range []string{"ing", "ed", "es", "s"} {
		if strings.HasSuffix(word, suffix) && len(word) > len(suffix)+3 {
			stem := strings.TrimSuffix(word, suffix)
			if (suffix == "ing" || suffix == "ed") && len(stem) > 2 && stem[len(stem)-1] == stem[len(stem)-2] {
				stem = stem[:len(stem)-1]
			}
			return stem
		}
	}
	return word
}

func isIntentStopWord(word string) bool {
	switch word {
	case "a", "an", "the", "i", "you", "me", "my", "your", "is", "are", "am", "do", "does", "did", "to", "in", "of", "for", "it", "and", "can", "could", "would", "should", "please":
		return true
	default:
		return false
	}
}

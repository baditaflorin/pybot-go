package pybot

import (
	"math"
	"strings"
	"unicode"
)

type intent string

const (
	intentGreeting       intent = "greeting"
	intentNameQuestion   intent = "name_question"
	intentAgeComparison  intent = "age_comparison"
	intentIdentity       intent = "identity"
	intentOnion          intent = "onion"
	intentKettle         intent = "kettle"
	intentEggs           intent = "eggs"
	intentThanks         intent = "thanks"
	intentSmallTalk      intent = "small_talk"
	intentRouteThreshold        = 0.55
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

// This small corpus demonstrates classic statistical routing without a model
// download, GPU, or generated answers. It should grow from evaluated examples.
var intentExamples = []intentExample{
	{intentGreeting, "hello"},
	{intentGreeting, "hi there"},
	{intentGreeting, "hey"},
	{intentGreeting, "good morning"},
	{intentGreeting, "good evening"},
	{intentGreeting, "howdy"},
	{intentGreeting, "yo there"},

	{intentNameQuestion, "what is my name"},
	{intentNameQuestion, "what's my name"},
	{intentNameQuestion, "do you remember my name"},
	{intentNameQuestion, "who am i"},
	{intentNameQuestion, "what do you call me"},
	{intentNameQuestion, "what name did i tell you"},

	{intentAgeComparison, "who is older"},
	{intentAgeComparison, "who's younger"},
	{intentAgeComparison, "which person is older"},
	{intentAgeComparison, "compare their ages"},
	{intentAgeComparison, "who has the higher age"},
	{intentAgeComparison, "who is the youngest"},
	{intentAgeComparison, "alice is 12 and bob is 50 who is older"},
	{intentAgeComparison, "one person has the age of 12 somebody is the age of 50 who is older"},

	{intentIdentity, "who are you"},
	{intentIdentity, "what are you"},
	{intentIdentity, "are you an ai"},
	{intentIdentity, "are you a language model"},
	{intentIdentity, "are you a chatbot"},

	{intentOnion, "why do onions make you cry"},
	{intentOnion, "why do onions make my eyes water"},
	{intentOnion, "what makes onions sting your eyes"},
	{intentOnion, "how do onions affect your eyes"},
	{intentOnion, "why does chopping an onion make my eyes tear"},

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

	{intentThanks, "thanks"},
	{intentThanks, "thank you"},
	{intentThanks, "thanks a lot"},
	{intentThanks, "i appreciate it"},
	{intentThanks, "much appreciated"},

	{intentSmallTalk, "how are you"},
	{intentSmallTalk, "how are you doing"},
	{intentSmallTalk, "are you doing okay"},
	{intentSmallTalk, "how is it going"},
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
	knownFeatures := 0
	for _, feature := range features {
		for _, label := range model.labels {
			if model.tokenCounts[label][feature] > 0 {
				knownFeatures++
				break
			}
		}
	}
	if knownFeatures < 2 {
		return "", 0
	}

	best := intent("")
	bestScore := math.Inf(-1)
	scores := make([]float64, 0, len(model.labels))
	for _, label := range model.labels {
		score := math.Log(float64(model.documentN[label]) / float64(model.totalDocs))
		denominator := float64(model.totalTokens[label] + len(model.vocabulary))
		for _, feature := range features {
			count := model.tokenCounts[label][feature]
			score += math.Log(float64(count+1) / denominator)
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
		words = append(words, current.String())
		current.Reset()
	}
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			current.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()

	features := append([]string(nil), words...)
	for i := 0; i+1 < len(words); i++ {
		features = append(features, words[i]+"_"+words[i+1])
	}
	return features
}

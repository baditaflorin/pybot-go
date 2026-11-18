package pybot

import "testing"

func TestIntentClassifierRecognizesHeldOutParaphrases(t *testing.T) {
	cases := []struct {
		text string
		want intent
	}{
		{"hey there, good to see you", intentGreeting},
		{"can you recall what name I gave you", intentNameQuestion},
		{"which of these two people is the eldest", intentAgeComparison},
		{"please work out 9 multiplied by 6", intentArithmetic},
		{"express 3 kilograms in pounds", intentConversion},
		{"what weekday is two weeks from today", intentDate},
		{"what makes chopped onions sting", intentOnion},
		{"remove hard water chalk from my tea kettle", intentKettle},
		{"how do I hard cook eggs safely", intentEggs},
		{"many thanks for the help", intentThanks},
		{"how have you been doing", intentSmallTalk},
	}
	correct := 0
	for _, test := range cases {
		got, score := defaultIntentClassifier.classify(test.text)
		if got == test.want && score >= statisticalRouteThreshold {
			correct++
		} else {
			t.Logf("miss: %q => %s (%.3f), want %s", test.text, got, score, test.want)
		}
	}
	if correct < 10 {
		t.Fatalf("classifier recognized %d/%d held-out paraphrases; want at least 10", correct, len(cases))
	}
}

package pybot

import (
	"strings"
	"testing"
)

func TestCuratedGeneralFactsAreSourcedAndScoped(t *testing.T) {
	cases := []struct {
		question string
		phrase   string
		sources  int
	}{
		{"Which planet is closest to our Sun?", "Mercury", 1},
		{"How many hours separate Tokyo and London?", "9 hours", 2},
		{"How does an international student adviser usually help?", "campus services", 1},
	}
	for _, test := range cases {
		t.Run(test.question, func(t *testing.T) {
			response := New(0).Reply(test.question)
			if response.Module != "retrieval/general-fact" || !strings.Contains(response.Text, test.phrase) {
				t.Fatalf("response = %#v, want sourced answer containing %q", response, test.phrase)
			}
			if len(response.Sources) != test.sources {
				t.Fatalf("got %d sources, want %d: %#v", len(response.Sources), test.sources, response.Sources)
			}
		})
	}
	if response, handled := answerCuratedGeneralFact("What is the nearest planet to Earth?"); handled {
		t.Fatalf("Sun-specific fact rule matched another planet question: %#v", response)
	}
}

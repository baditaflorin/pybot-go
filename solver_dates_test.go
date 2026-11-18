package pybot

import (
	"testing"
	"time"
)

func fixedDateBot() *Bot {
	clock := func() time.Time {
		return time.Date(2026, time.November, 18, 12, 0, 0, 0, time.FixedZone("EET", 2*60*60))
	}
	return newBot(0, clock, defaultKnowledgeBase)
}

func TestDateArithmetic(t *testing.T) {
	cases := []struct {
		question string
		want     string
	}{
		{"What date is 30 days after 2026-11-18?", "2026-12-18 is Friday."},
		{"How many days between 2026-01-01 and 2026-01-10?", "There are 9 days between those dates."},
		{"What day will it be 2 weeks from today?", "2026-12-02 is Wednesday."},
		{"What date is 1 month after 2025-01-31?", "2025-02-28 is Friday."},
		{"What day is it tomorrow?", "2026-11-19 is Thursday."},
		{"How many days until 2026-11-20?", "That date is 2 days from today (2026-11-18)."},
	}
	for _, test := range cases {
		t.Run(test.question, func(t *testing.T) {
			response := fixedDateBot().Reply(test.question)
			if response.Module != "reasoning/date" || response.Text != test.want {
				t.Fatalf("response = %#v, want %q", response, test.want)
			}
		})
	}
}

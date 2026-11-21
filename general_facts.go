package pybot

import "strings"

// answerCuratedGeneralFact handles a few stable relations that are both common
// and easy to verify against a named source. It complements broad web retrieval;
// it is deliberately small rather than pretending to be a complete knowledge
// base.
func answerCuratedGeneralFact(text string) (Response, bool) {
	text = strings.ToLower(strings.TrimSpace(text))
	switch {
	case strings.Contains(text, "planet") && strings.Contains(text, "sun") && containsAny(text, "nearest", "closest"):
		return Response{
			Text:        "Mercury is the planet closest to the Sun.",
			Module:      "retrieval/general-fact",
			Sources:     []Source{{Title: "NASA Science · About the Planets", URL: "https://science.nasa.gov/solar-system/planets/"}},
			Suggestions: []string{"Tell me more about Mercury", "Ask another question"},
		}, true
	case strings.Contains(text, "tokyo") && strings.Contains(text, "london") &&
		(strings.Contains(text, "time difference") || containsAny(text, "how many hours", "hours separate", "hours between")):
		return Response{
			Text:   "Tokyo is 9 hours ahead of London during UK winter and 8 hours ahead during British Summer Time. Japan stays on UTC+9; London shifts between GMT (UTC+0) and BST (UTC+1). Give me a date if you need the exact offset around a clock change.",
			Module: "retrieval/general-fact",
			Sources: []Source{
				{Title: "Timeanddate · Tokyo time zone", URL: "https://www.timeanddate.com/worldclock/japan/tokyo"},
				{Title: "Timeanddate · London time zone", URL: "https://www.timeanddate.com/worldclock/uk/london"},
			},
			Suggestions: []string{"Give me a date", "Ask another question"},
		}, true
	case strings.Contains(text, "foreign student adviser") || strings.Contains(text, "foreign student advisor") || strings.Contains(text, "international student adviser") || strings.Contains(text, "international student advisor"):
		return Response{
			Text:        "At a university, an international student adviser commonly helps students navigate enrollment and campus services, connect with academic or personal support, and understand student-status requirements. Exact duties depend on the country and institution; for a specific case, check that university's international office.",
			Module:      "retrieval/general-fact",
			Sources:     []Source{{Title: "University of Michigan · International Student and Scholar Advisor", URL: "https://careers.umich.edu/job_detail/284648/international-student-and-scholar-advisor"}},
			Suggestions: []string{"Ask about a specific university", "Ask another question"},
		}, true
	}
	return Response{}, false
}

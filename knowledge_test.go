package pybot

import (
	"strings"
	"testing"
)

func TestLocalRetrievalReturnsSources(t *testing.T) {
	response := New(0).Reply("Why do onions make my eyes water when I cut them?")
	if !strings.HasPrefix(response.Module, "retrieval/") {
		t.Fatalf("module = %q, want retrieval", response.Module)
	}
	if len(response.Sources) == 0 || !strings.Contains(response.Sources[0].URL, "cornell.edu") {
		t.Fatalf("sources = %#v, want Cornell source", response.Sources)
	}
}

func TestKettleAnswerIsModelSpecificAndFollowupUsesSource(t *testing.T) {
	bot := New(0)
	response := bot.Reply("How do I descale a kettle?")
	if !strings.Contains(response.Text, "model") || len(response.Sources) == 0 {
		t.Fatalf("unexpected kettle answer: %#v", response)
	}
	followup := bot.Reply("How long does that take?")
	if !strings.Contains(followup.Text, "15 minutes") || len(followup.Sources) == 0 {
		t.Fatalf("unexpected follow-up: %#v", followup)
	}
}

func TestEggAnswerUsesFoodSafetyEvidence(t *testing.T) {
	response := New(0).Reply("How should I cook egg dishes safely?")
	if !strings.Contains(response.Text, "160°F") || len(response.Sources) == 0 {
		t.Fatalf("unexpected egg answer: %#v", response)
	}
}

func TestUnknownDoesNotRetrieveUnrelatedDocument(t *testing.T) {
	response := New(0).Reply("What is the population of Mars?")
	if response.Module != "conversation/unknown" || len(response.Sources) > 0 {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestKnowledgeCorpusRejectsMissingAndInvalidSources(t *testing.T) {
	valid := `[{"id":"x","title":"title","answer":"answer","text":"searchable","source_title":"source","source_url":"https://example.com"}]`
	if _, err := loadKnowledge([]byte(valid)); err != nil {
		t.Fatalf("valid corpus rejected: %v", err)
	}
	for name, corpus := range map[string]string{
		"empty":            `[]`,
		"missing source":   strings.Replace(valid, `"source_title":"source","source_url":"https://example.com"`, `"source_title":"","source_url":"file:///tmp/doc"`, 1),
		"bad extra source": strings.Replace(valid, `"source_url":"https://example.com"`, `"source_url":"https://example.com","extra_sources":[{"title":"bad","url":"javascript:alert(1)"}]`, 1),
		"duplicate id":     strings.TrimSuffix(valid, "]") + `,{"id":"x","title":"other","answer":"answer","source_title":"source","source_url":"https://example.org"}]`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := loadKnowledge([]byte(corpus)); err == nil {
				t.Fatal("expected corpus validation error")
			}
		})
	}
}

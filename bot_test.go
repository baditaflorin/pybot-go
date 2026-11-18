package pybot

import (
	"strings"
	"sync"
	"testing"
)

func TestAgeComparisonFromOneMessage(t *testing.T) {
	bot := New(4)
	response := bot.Reply("One person has the age of 12 and somebody is the age of 50. Who is older?")
	if response.Module != "reasoning/age-comparison" {
		t.Fatalf("module = %q, want age comparison", response.Module)
	}
	if !strings.Contains(response.Text, "Person 2") || !strings.Contains(response.Text, "38 years") {
		t.Fatalf("unexpected comparison: %q", response.Text)
	}
}

func TestAgeFactsAcrossTurnsAndReset(t *testing.T) {
	bot := New(2)
	bot.Reply("Alice is 12 years old.")
	bot.Reply("Bob is 50 years old.")
	response := bot.Reply("Who is older?")
	if !strings.Contains(response.Text, "Bob") || !strings.Contains(response.Text, "38 years") {
		t.Fatalf("unexpected comparison: %q", response.Text)
	}
	if got := len(bot.Facts()); got != 2 {
		t.Fatalf("remembered facts = %d, want 2", got)
	}
	bot.Reset()
	if got := len(bot.Facts()); got != 0 {
		t.Fatalf("facts after reset = %d, want 0", got)
	}
}

func TestAgeComparisonSelectsPeopleNamedInQuestion(t *testing.T) {
	bot := New(2)
	bot.Reply("Alice is 12 years old.")
	bot.Reply("Bob is 50 years old.")
	bot.Reply("Charlie is 40 years old.")
	response := bot.Reply("Who is older, Alice or Bob?")
	if !strings.Contains(response.Text, "Bob is older") || !strings.Contains(response.Text, "38 years") {
		t.Fatalf("unexpected named comparison: %q", response.Text)
	}
}

func TestRememberAndAnswerUserAge(t *testing.T) {
	bot := New(1)
	bot.Reply("I am 37 years old.")
	response := bot.Reply("How old am I?")
	if !strings.Contains(response.Text, "You are 37 years old") {
		t.Fatalf("self-age response = %q", response.Text)
	}
	facts := bot.Facts()
	if len(facts) != 1 || facts[0].Subject != "user" || facts[0].Attribute != "age" {
		t.Fatalf("self-age facts = %#v", facts)
	}
}

func TestAgePronounsClarifyInsteadOfGuessing(t *testing.T) {
	bot := New(1)
	bot.Reply("Alice is 12 years old.")
	response := bot.Reply("She is 50 years old.")
	if !response.NeedsClarification || !strings.Contains(response.Text, "Who do you mean") {
		t.Fatalf("pronoun response = %#v, want clarification", response)
	}
	if facts := bot.Facts(); len(facts) != 1 || facts[0].Subject != "alice" {
		t.Fatalf("ambiguous pronoun was stored as a fact: %#v", facts)
	}
	response = bot.Reply("How old is she?")
	if !response.NeedsClarification {
		t.Fatalf("pronoun age question should clarify: %#v", response)
	}
}

func TestNameMemoryAndQuestion(t *testing.T) {
	bot := New(3)
	bot.Reply("I'm Sam.")
	response := bot.Reply("What do you call me?")
	if !strings.Contains(response.Text, "Sam") {
		t.Fatalf("name answer = %q, want Sam", response.Text)
	}
}

func TestUnknownQuestionAbstains(t *testing.T) {
	bot := New(1)
	response := bot.Reply("What is the population of Mars?")
	if response.Module != "conversation/unknown" {
		t.Fatalf("module = %q, want unknown", response.Module)
	}
	if len(response.Sources) != 0 {
		t.Fatalf("unknown answer unexpectedly cites sources: %#v", response.Sources)
	}
}

func TestFallbackWordingIsRepeatableForSeed(t *testing.T) {
	first := New(12)
	second := New(12)
	question := "What is the population of Mars?"
	initial := first.Reply(question).Text
	if got, want := second.Reply(question).Text, initial; got != want {
		t.Fatalf("same seed produced different fallbacks: %q != %q", got, want)
	}
	first.Reset()
	if got, want := first.Reply(question).Text, initial; got != want {
		t.Fatalf("reset did not restore seeded fallback sequence: %q != %q", got, want)
	}
}

func TestBotMethodsAreSafeForConcurrentUse(t *testing.T) {
	bot := New(5)
	var group sync.WaitGroup
	for i := 0; i < 32; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			if i%3 == 0 {
				bot.Reset()
				return
			}
			bot.Reply("Alice is 12 years old.")
			_ = bot.Facts()
		}(i)
	}
	group.Wait()
}

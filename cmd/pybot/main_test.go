package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestOneShotPrintsEvidenceLinks(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"-s", "1", "-p", "Why do onions make my eyes water?"}, strings.NewReader(""), &stdout, &stderr, false)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "source: Cornell Chronicle") || !strings.Contains(stdout.String(), "https://news.cornell.edu/") {
		t.Fatalf("one-shot output omitted evidence link: %q", stdout.String())
	}
}

func TestPipedMessagesShareConversationMemory(t *testing.T) {
	var stdout, stderr bytes.Buffer
	input := "Alice is 12 years old.\nBob is 50 years old.\nWho is older?\n"
	code := run(nil, strings.NewReader(input), &stdout, &stderr, false)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Bob is older") || !strings.Contains(stdout.String(), "38 years") {
		t.Fatalf("piped conversation lost age facts: %q", stdout.String())
	}
}

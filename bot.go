// Package pybot provides a CPU-only conversational assistant built from a
// statistical intent router, structured conversation memory, deterministic
// solvers, and source-backed local retrieval. It does not generate text with an LLM.
package pybot

import (
	"math/rand"
	"strings"
	"sync"
	"time"
)

// Version is the version of this Go implementation.
const Version = "0.9.0"

// Source identifies a local knowledge document's original source.
type Source struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Response contains an answer, its route, and optional evidence links.
type Response struct {
	Text               string   `json:"text"`
	Module             string   `json:"module"`
	Sources            []Source `json:"sources,omitempty"`
	NeedsClarification bool     `json:"needs_clarification,omitempty"`
	LookupQuery        string   `json:"lookup_query,omitempty"`
	Suggestions        []string `json:"suggestions,omitempty"`
	RetryQuery         string   `json:"retry_query,omitempty"`
}

// ConversationTurn is the serializable boundary used by the browser to restore
// a short, tab-local conversation without sending it to a server.
type ConversationTurn struct {
	User       string   `json:"user"`
	Assistant  Response `json:"assistant"`
	Intent     string   `json:"intent,omitempty"`
	Confidence float64  `json:"confidence,omitempty"`
}

// Fact is a typed, conversation-scoped value extracted from a user message.
type Fact struct {
	Subject   string
	Attribute string
	Value     string
	Unit      string
	Source    string
	Turn      int
}

// Bot owns one conversation's memory and deterministic response choices.
// Reply, Reset, and Facts are safe to call concurrently.
type Bot struct {
	mu        sync.Mutex
	seed      int64
	rng       *rand.Rand
	memory    conversationMemory
	clock     func() time.Time
	knowledge *knowledgeBase
}

// New creates a bot with a repeatable wording seed. Date questions that mention
// "today" use the machine's local clock.
func New(seed int64) *Bot {
	return newBot(seed, time.Now, defaultKnowledgeBase)
}

func newBot(seed int64, clock func() time.Time, knowledge *knowledgeBase) *Bot {
	if clock == nil {
		clock = time.Now
	}
	if knowledge == nil {
		knowledge = defaultKnowledgeBase
	}
	return &Bot{
		seed:      seed,
		rng:       rand.New(rand.NewSource(seed)),
		memory:    newConversationMemory(),
		clock:     clock,
		knowledge: knowledge,
	}
}

// Reset starts a fresh conversation and restarts seeded wording choices.
func (b *Bot) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.memory = newConversationMemory()
	b.rng = rand.New(rand.NewSource(b.seed))
}

// Facts returns a copy of facts remembered in the current conversation.
func (b *Bot) Facts() []Fact {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.memory.factsInOrder()
}

// RestoreConversation reconstructs local reasoning state from browser history.
// The saved assistant responses are retained for context, while the user's
// messages are replayed through the current deterministic engine.
func (b *Bot) RestoreConversation(turns []ConversationTurn) {
	b.Reset()
	start := max(0, len(turns)-conversationHistoryLimit)
	for _, turn := range turns[start:] {
		if strings.TrimSpace(turn.User) == "" {
			continue
		}
		b.ReplyWithSemanticHint(turn.User, turn.Intent, turn.Confidence)
		if turn.Assistant.Text != "" {
			b.RememberResponse(turn.User, turn.Assistant)
		}
	}
}

// RememberResponse records the final response after browser-side retrieval,
// such as a Wikipedia lookup, completes for the current user message.
func (b *Bot) RememberResponse(question string, response Response) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.memory.rememberResponse(question, response)
}

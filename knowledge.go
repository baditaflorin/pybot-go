package pybot

import (
	"embed"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"sort"
	"strings"
)

//go:embed knowledge/docs.json
var embeddedKnowledge embed.FS

type knowledgeDocument struct {
	ID           string            `json:"id"`
	Title        string            `json:"title"`
	Answer       string            `json:"answer"`
	Text         string            `json:"text"`
	SourceTitle  string            `json:"source_title"`
	SourceURL    string            `json:"source_url"`
	ExtraSources []Source          `json:"extra_sources"`
	Followups    map[string]string `json:"followups"`
}

type indexedDocument struct {
	document knowledgeDocument
	termFreq map[string]int
	length   int
}

type knowledgeBase struct {
	documents           []indexedDocument
	documentFrequencies map[string]int
	averageLength       float64
}

type retrievalResult struct {
	Document  knowledgeDocument
	Score     float64
	Ambiguous []knowledgeDocument
}

var defaultKnowledgeBase = mustLoadKnowledge()

func mustLoadKnowledge() *knowledgeBase {
	data, err := embeddedKnowledge.ReadFile("knowledge/docs.json")
	if err != nil {
		panic(err)
	}
	kb, err := loadKnowledge(data)
	if err != nil {
		panic(err)
	}
	return kb
}

func loadKnowledge(data []byte) (*knowledgeBase, error) {
	var documents []knowledgeDocument
	if err := json.Unmarshal(data, &documents); err != nil {
		return nil, err
	}
	kb := &knowledgeBase{documentFrequencies: make(map[string]int)}
	if len(documents) == 0 {
		return nil, fmt.Errorf("knowledge base has no documents")
	}
	seenIDs := make(map[string]struct{}, len(documents))
	totalLength := 0
	for _, document := range documents {
		if document.ID == "" || document.Title == "" || document.Answer == "" || document.SourceTitle == "" {
			return nil, fmt.Errorf("knowledge document is missing required metadata")
		}
		if _, exists := seenIDs[document.ID]; exists {
			return nil, fmt.Errorf("duplicate knowledge document id %q", document.ID)
		}
		seenIDs[document.ID] = struct{}{}
		if !validSourceURL(document.SourceURL) {
			return nil, fmt.Errorf("knowledge document %q has an invalid source URL", document.ID)
		}
		for _, source := range document.ExtraSources {
			if source.Title == "" || !validSourceURL(source.URL) {
				return nil, fmt.Errorf("knowledge document %q has an invalid extra source", document.ID)
			}
		}
		terms := retrievalTokens(document.Title + " " + document.Text + " " + document.Answer)
		frequency := make(map[string]int)
		for _, term := range terms {
			frequency[term]++
		}
		for term := range frequency {
			kb.documentFrequencies[term]++
		}
		indexed := indexedDocument{document: document, termFreq: frequency, length: len(terms)}
		kb.documents = append(kb.documents, indexed)
		totalLength += indexed.length
	}
	if len(kb.documents) > 0 {
		kb.averageLength = float64(totalLength) / float64(len(kb.documents))
	}
	return kb, nil
}

func validSourceURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func (kb *knowledgeBase) Search(query string, route intent, routeTrusted bool) (retrievalResult, bool) {
	queryTerms := uniqueTerms(retrievalTokens(query))
	if len(queryTerms) == 0 {
		return retrievalResult{}, false
	}
	if route == intentLinuxExpert || route == intentKnowledge {
		if documentID := linuxConceptDocumentID(query); documentID != "" {
			for _, indexed := range kb.documents {
				if indexed.document.ID == documentID {
					return retrievalResult{Document: indexed.document, Score: 1}, true
				}
			}
		}
	}
	if route == intentLinuxExpert && isGenericLinuxHelpQuery(query) {
		for _, indexed := range kb.documents {
			if indexed.document.ID == "linux-help" {
				return retrievalResult{Document: indexed.document, Score: 1}, true
			}
		}
	}
	searchTerms := append([]string(nil), queryTerms...)
	if routeTrusted {
		searchTerms = append(searchTerms, intentTerms(route)...)
	}
	if route == intentLinuxExpert {
		searchTerms = append(searchTerms, linuxSearchTerms(query)...)
	}
	searchTerms = uniqueTerms(searchTerms)

	type scored struct {
		doc        knowledgeDocument
		score      float64
		rawMatches int
	}
	ranked := make([]scored, 0, len(kb.documents))
	for _, indexed := range kb.documents {
		score := 0.0
		rawMatches := 0
		for _, term := range searchTerms {
			tf := indexed.termFreq[term]
			if tf == 0 {
				continue
			}
			if containsTerm(queryTerms, term) {
				rawMatches++
			}
			idf := math.Log(1 + (float64(len(kb.documents)-kb.documentFrequencies[term])+0.5)/(float64(kb.documentFrequencies[term])+0.5))
			const k1, b = 1.2, 0.75
			lengthNorm := 1 - b + b*float64(indexed.length)/maxFloat(kb.averageLength, 1)
			score += idf * (float64(tf) * (k1 + 1)) / (float64(tf) + k1*lengthNorm)
		}
		if score > 0 {
			ranked = append(ranked, scored{doc: indexed.document, score: score, rawMatches: rawMatches})
		}
	}
	if len(ranked) == 0 {
		return retrievalResult{}, false
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })
	best := ranked[0]
	minimumMatches := 2
	if routeTrusted {
		minimumMatches = 1
	}
	if best.rawMatches < minimumMatches || best.score < 0.75 {
		return retrievalResult{}, false
	}
	if len(ranked) > 1 && ranked[1].rawMatches >= 2 &&
		ranked[1].score >= best.score*0.92 && !routeTrusted {
		return retrievalResult{Ambiguous: []knowledgeDocument{best.doc, ranked[1].doc}}, true
	}
	return retrievalResult{Document: best.doc, Score: best.score}, true
}

func (kb *knowledgeBase) Followup(documentID, kind string) (Response, bool) {
	for _, indexed := range kb.documents {
		if indexed.document.ID != documentID {
			continue
		}
		answer := indexed.document.Followups[kind]
		if answer == "" {
			return Response{}, false
		}
		return Response{
			Text: answer, Module: "retrieval/follow-up/" + documentID,
			Sources: allSources(indexed.document),
		}, true
	}
	return Response{}, false
}

func allSources(document knowledgeDocument) []Source {
	sources := []Source{{Title: document.SourceTitle, URL: document.SourceURL}}
	return append(sources, document.ExtraSources...)
}

func intentTerms(route intent) []string {
	switch route {
	case intentOnion:
		return []string{"onion", "eye", "tear", "cut"}
	case intentKettle:
		return []string{"kettle", "scale", "descale", "clean"}
	case intentEggs:
		return []string{"egg", "cook", "safety", "temperature"}
	default:
		return nil
	}
}

func linuxSearchTerms(query string) []string {
	text := strings.ToLower(query)
	terms := make([]string, 0, 12)
	addWhen := func(markers []string, additions ...string) {
		for _, marker := range markers {
			if strings.Contains(text, marker) {
				terms = append(terms, additions...)
				return
			}
		}
	}
	addWhen([]string{"systemd", "systemctl", "service", "unit", "failed", "failure"}, "systemctl", "journalctl", "unit", "status", "failure")
	addWhen([]string{"journal", "log", "boot", "kernel message"}, "journalctl", "kernel", "warning", "error")
	addWhen([]string{"disk", "space", "inode", "storage", "full"}, "filesystem", "df", "usage", "inode")
	addWhen([]string{"large file", "big file", "find file", "files taking"}, "find", "size", "file", "large")
	addWhen([]string{"process", "cpu", "memory", "ram", "load", "slow"}, "ps", "free", "uptime", "process")
	addWhen([]string{"port", "listen", "socket", "tcp", "udp"}, "ss", "socket", "tcp", "udp")
	addWhen([]string{"dns", "resolve", "hostname", "resolv"}, "getent", "resolvectl", "resolver")
	addWhen([]string{"route", "gateway", "internet", "interface", "network"}, "ip", "route", "address", "curl")
	addWhen([]string{"chmod", "permission", "owner", "chown", "access denied"}, "chmod", "owner", "mode", "namei")
	addWhen([]string{"apt", "dnf", "pacman", "update", "upgrade", "package"}, "package", "apt", "dnf", "pacman")
	addWhen([]string{"ssh", "key", "login", "authentication"}, "ssh", "sshd", "authentication", "verbose")
	addWhen([]string{"firewall", "ufw", "nft", "blocked"}, "firewall", "nft", "ufw", "rules")
	addWhen([]string{"mount", "partition", "filesystem", "fstab", "lsblk"}, "findmnt", "lsblk", "device", "partition")
	addWhen([]string{"docker", "volume", "container"}, "docker", "volume", "image", "prune")
	addWhen([]string{"man ", "manual", "--help"}, "man", "manual", "help", "option")
	return uniqueTerms(terms)
}

func retrievalTokens(text string) []string {
	terms := make([]string, 0)
	var word strings.Builder
	flush := func() {
		if word.Len() == 0 {
			return
		}
		token := stemWord(word.String())
		if len(token) > 1 && !isRetrievalStopWord(token) {
			terms = append(terms, token)
		}
		word.Reset()
	}
	for _, r := range strings.ToLower(text) {
		if r == '°' || r == '-' {
			flush()
			continue
		}
		if isLetterOrNumber(r) {
			word.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return terms
}

func isLetterOrNumber(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') ||
		(r >= 'à' && r <= 'ÿ')
}

func isRetrievalStopWord(word string) bool {
	switch word {
	case "a", "an", "the", "i", "you", "me", "my", "your", "is", "are", "am", "do", "does", "did", "to", "in", "of", "for", "it", "and", "can", "could", "would", "should", "how", "what", "why", "when", "where", "who", "which", "with", "on", "from", "after", "before", "this", "that", "be", "will", "was", "were", "their", "them", "there":
		return true
	default:
		return false
	}
}

func uniqueTerms(terms []string) []string {
	seen := make(map[string]struct{}, len(terms))
	unique := make([]string, 0, len(terms))
	for _, term := range terms {
		if _, ok := seen[term]; ok {
			continue
		}
		seen[term] = struct{}{}
		unique = append(unique, term)
	}
	return unique
}

func containsTerm(terms []string, term string) bool {
	for _, candidate := range terms {
		if candidate == term {
			return true
		}
	}
	return false
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// randomFallback is used only when the user's seed requests varied wording.
func randomFallback(rng *rand.Rand) string {
	choices := []string{
		"I don't have a reliable answer for that in my local knowledge base.",
		"I can't answer that from the facts and sources I have here.",
		"I don't know that one yet. I can calculate, convert units, compare remembered ages, or search my local sources.",
	}
	return choices[rng.Intn(len(choices))]
}

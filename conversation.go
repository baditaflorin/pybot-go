package pybot

import (
	"fmt"
	"regexp"
	"strings"
)

var linuxDistroRE = regexp.MustCompile(`(?i)\b(ubuntu|debian|fedora|rhel|red\s+hat|rocky(?:\s+linux)?|alma(?:linux)?|arch(?:\s+linux)?|opensuse|suse|alpine|linux\s+mint|mint|pop!?_?os|manjaro|gentoo|nixos)(?:\s+(\d+(?:\.\d+)*))?\b`)
var commandInAnswerRE = regexp.MustCompile("`([^`]+)`")
var trailingDistroPrepositionRE = regexp.MustCompile(`(?i)\b(on|for|using)\s*$`)
var externalEntityPronounRE = regexp.MustCompile(`(?i)\b(?:it|that|there|they|he|she|its|their|those|these|them)\b`)
var externalSingularPronounRE = regexp.MustCompile(`(?i)\b(?:it|that|there|he|she|its)\b`)

func suggestedFollowups(response Response) []string {
	switch {
	case response.NeedsClarification:
		return []string{"What detail do you need?", "Show an example"}
	case strings.HasPrefix(response.Module, "retrieval/linux-") || strings.HasPrefix(response.Module, "linux/"):
		return []string{"What next?", "Why these steps?", "Show source"}
	case response.Module == "retrieval/wikipedia" || response.Module == "retrieval/wikipedia-follow-up":
		return []string{"Tell me more", "Show source", "Where is it?"}
	case strings.HasPrefix(response.Module, "reasoning/age-"):
		return []string{"Add another person", "Rank them", "Show the age difference"}
	case response.Module == "conversation/unknown":
		return []string{"What can you do?", "Ask a Linux question", "Try a calculation"}
	default:
		return []string{"Tell me more", "What can you do?"}
	}
}

func externalTopic(response Response) string {
	for _, source := range response.Sources {
		if strings.HasPrefix(source.Title, "Wikipedia · ") {
			return strings.TrimSpace(strings.TrimPrefix(source.Title, "Wikipedia · "))
		}
	}
	return ""
}

func (b *Bot) answerContextControl(text string) (Response, bool) {
	if isMemoryQuestion(text) {
		return b.answerConversationSummary(), true
	}
	if isLinuxDistributionQuestion(text) {
		if distro := b.memory.linuxDistribution(); distro != "" {
			return Response{Text: "You told me you use " + distro + " in this conversation.", Module: "memory/environment"}, true
		}
		return Response{Text: "I don't know yet. Tell me the Linux distribution and version, such as Ubuntu 24.04 or Debian 13, and I'll keep it in mind for this chat.", Module: "memory/environment", NeedsClarification: true}, true
	}
	if attribute := requestedForgetAttribute(text); attribute != "" {
		b.memory.forget("user", attribute)
		return Response{Text: "Okay. I removed that detail from my conversation memory.", Module: "memory/forget"}, true
	}
	if b.memory.lastResponse.Text == "" {
		return Response{}, false
	}
	if isRepeatRequest(text) {
		response := b.memory.lastResponse
		response.Module = "conversation/repeat"
		return response, true
	}
	if isSourceRequest(text) {
		if len(b.memory.lastResponse.Sources) == 0 {
			return Response{Text: "My previous answer did not include an external source link.", Module: "conversation/source"}, true
		}
		return Response{Text: "These are the sources for my previous answer.", Module: "conversation/source", Sources: append([]Source(nil), b.memory.lastResponse.Sources...)}, true
	}
	if isShorterRequest(text) {
		short := firstSentence(b.memory.lastResponse.Text)
		if short == "" {
			short = b.memory.lastResponse.Text
		}
		return Response{Text: short, Module: "conversation/shorten", Sources: append([]Source(nil), b.memory.lastResponse.Sources...)}, true
	}
	if isWhyFollowup(text) {
		if strings.HasPrefix(b.memory.lastResponse.Module, "retrieval/linux-") || strings.HasPrefix(b.memory.lastResponse.Module, "linux/") {
			return Response{
				Text:   "I started with read-only checks because they show the current state without changing your system. Their output helps narrow the cause before trying a fix. If you share the exact error and your distribution/version, I can make the next step specific.",
				Module: "linux/follow-up/why", Sources: append([]Source(nil), b.memory.lastResponse.Sources...),
			}, true
		}
		if strings.HasPrefix(b.memory.lastResponse.Module, "reasoning/age-") {
			return Response{Text: "I compared the ages stated in this conversation and used the numeric difference between them.", Module: "conversation/explanation"}, true
		}
		return Response{Text: "I used the details in your previous question and the source shown with my answer. Ask about a specific sentence if you want me to unpack it.", Module: "conversation/explanation", Sources: append([]Source(nil), b.memory.lastResponse.Sources...)}, true
	}
	if isExplainFollowup(text) {
		if explanation := explainPreviousCommand(b.memory.lastResponse.Text); explanation != "" {
			return Response{Text: explanation, Module: "conversation/explanation", Sources: append([]Source(nil), b.memory.lastResponse.Sources...)}, true
		}
		return Response{
			Text:   "Which part of my previous answer should I explain? You can paste a command or name the step.",
			Module: "conversation/explanation", NeedsClarification: true,
		}, true
	}
	if isDetailRequest(text) {
		return Response{
			Text:   "The previous reply contains the full guidance I have for that topic. Point to a command, step, or sentence and I can explain it directly.",
			Module: "conversation/detail", Sources: append([]Source(nil), b.memory.lastResponse.Sources...),
		}, true
	}
	if isSafetyFollowup(text) {
		return answerSafetyFollowup(b.memory.lastResponse)
	}
	return Response{}, false
}

func (b *Bot) answerSemanticContextFollowup(text string) (Response, bool) {
	if b.memory.lastResponse.Text == "" {
		return Response{}, false
	}
	switch {
	case containsAny(text, "source", "cite", "citation", "reference"):
		return b.answerContextControl("show sources")
	case containsAny(text, "short", "brief", "concise", "summary"):
		return b.answerContextControl("shorter")
	case containsAny(text, "repeat", "again", "restate", "reiterate"):
		return b.answerContextControl("repeat that")
	case containsAny(text, "safe", "danger", "risk", "break"):
		return b.answerContextControl("is that safe")
	case containsAny(text, "why", "reason", "justify", "recommend"):
		return b.answerContextControl("why")
	case containsAny(text, "explain", "command", "step", "mean", "walk me through"):
		return b.answerContextControl("explain that")
	case containsAny(text, "next", "continue"):
		if response, handled := b.answerFollowup("what next"); handled {
			return response, true
		}
		return b.answerContextControl("more detail")
	case containsAny(text, "detail", "elaborate", "expand"):
		return b.answerContextControl("more detail")
	default:
		return Response{Text: "Which part of my previous answer should I follow up on: the source, a command, or the next step?", Module: "conversation/follow-up", NeedsClarification: true}, true
	}
}

func isLinuxDistributionQuestion(text string) bool {
	return oneOf(text, "what distro am i using", "which distro am i using", "what linux distro am i on", "which linux distro am i on", "what is my linux distribution", "which distribution do i use", "what os am i using", "what os am i on", "what did i say my distro was")
}

func isMemoryQuestion(text string) bool {
	return oneOf(text, "what do you remember", "what do you remember about me", "what did i tell you", "what have i told you", "what is the context", "summarize our conversation", "recap this chat", "what preferences do you remember", "what answer style are you using", "what style did i ask for")
}

func requestedForgetAttribute(text string) string {
	switch {
	case oneOf(text, "forget my name", "forget my name please"):
		return "name"
	case oneOf(text, "forget my age", "forget how old i am"):
		return "age"
	case oneOf(text, "forget my linux distribution", "forget my distro", "forget which linux i use"):
		return "linux_distribution"
	case oneOf(text, "forget my preferences", "forget my answer preferences", "reset my answer style", "clear my response preferences", "use your default answer style"):
		return "preferences"
	default:
		return ""
	}
}

func (b *Bot) answerConversationSummary() Response {
	parts := make([]string, 0)
	for _, fact := range b.memory.factsInOrder() {
		switch fact.Attribute {
		case "name":
			parts = append(parts, "you told me your name is "+fact.Value)
		case "age":
			if fact.Subject == "user" {
				parts = append(parts, "you told me you are "+fact.Value+" years old")
			} else {
				parts = append(parts, displayPerson(fact.Subject)+" is "+fact.Value+" years old")
			}
		case "linux_distribution":
			parts = append(parts, "you use "+fact.Value)
		}
	}
	if preferences := b.memory.preferences.summary(); preferences != "" {
		parts = append(parts, "your answer preferences are "+preferences)
	}
	if len(b.memory.topics) > 0 {
		recent := make([]string, 0, 3)
		for i := len(b.memory.topics) - 1; i >= 0 && len(recent) < 3; i-- {
			recent = append(recent, b.memory.topics[i].title)
		}
		parts = append(parts, "recent topics include "+strings.Join(recent, ", "))
	}
	if len(parts) == 0 {
		if b.memory.lastExternalTopic != "" {
			return Response{Text: "We just looked up " + b.memory.lastExternalTopic + ". I don't have any personal facts saved from this chat.", Module: "conversation/summary"}
		}
		if b.memory.lastKnowledgeID != "" {
			if result, found := b.knowledge.documentByID(b.memory.lastKnowledgeID); found {
				return Response{Text: "We have been discussing " + result.Document.Title + ". I don't have any personal facts saved from this chat.", Module: "conversation/summary"}
			}
		}
		return Response{Text: "I don't have any personal facts saved from this chat yet.", Module: "conversation/summary"}
	}
	if b.memory.lastExternalTopic != "" {
		parts = append(parts, "we just looked up "+b.memory.lastExternalTopic)
	} else if b.memory.lastKnowledgeID != "" {
		if result, found := b.knowledge.documentByID(b.memory.lastKnowledgeID); found {
			parts = append(parts, "we have been discussing "+result.Document.Title)
		}
	}
	return Response{Text: "From this chat, I remember " + strings.Join(parts, "; ") + ". You can say “forget my name,” “forget my age,” “forget my distro,” or “forget my preferences” to remove a saved detail.", Module: "conversation/summary"}
}

func (b *Bot) answerLinuxEnvironmentFollowup(text string) (Response, bool) {
	if b.memory.lastLinuxQuestion == "" {
		return Response{}, false
	}
	if containsAny(text, "windows", "macos", "mac os", "freebsd") {
		return Response{
			Text:   "The previous instructions were for Linux. Windows and macOS use different tools; tell me which operating system and what task you want to do there.",
			Module: "linux/follow-up/platform", NeedsClarification: true,
		}, true
	}
	distro := linuxDistroRE.FindString(text)
	if distro == "" {
		return Response{}, false
	}
	query := replaceLinuxDistro(b.memory.lastLinuxQuestion, distro)
	result, found := b.knowledge.Search(query, intentLinuxExpert, true)
	if !found || result.Document.ID == "" {
		return Response{
			Text:   fmt.Sprintf("I don't have a source-backed version of that guidance for %s yet. Share the command and exact error and I can check what applies.", distro),
			Module: "linux/follow-up/distro", NeedsClarification: true,
		}, true
	}
	b.memory.lastKnowledgeID = result.Document.ID
	b.memory.lastKnowledgeTurn = b.memory.turn
	return Response{
		Text:    fmt.Sprintf("For %s: %s", distro, result.Document.Answer),
		Module:  "retrieval/" + result.Document.ID,
		Sources: allSources(result.Document),
	}, true
}

func (b *Bot) answerExternalFollowup(text string) (Response, bool) {
	topic := b.memory.lastExternalTopic
	if topic == "" || !isEntityFollowup(text) {
		return Response{}, false
	}
	if response, handled := answerExternalFactFollowup(text, topic, b.memory.lastResponse); handled {
		return response, true
	}
	query := externalFollowupQuery(text, topic, b.memory.lastExternalQuestion)
	if query == "" {
		return Response{}, false
	}
	return Response{
		Text:        "I'll look that up about " + topic + ".",
		Module:      "retrieval/wikipedia-follow-up",
		LookupQuery: query,
	}, true
}

func answerExternalFactFollowup(text, topic string, previous Response) (Response, bool) {
	wanted := ""
	switch {
	case containsAny(text, "population", "how many people", "people live", "residents", "inhabitants"):
		wanted = "population"
	case containsAny(text, "river", "which waterway"):
		wanted = "river"
	case containsAny(text, "how big", "how large", "what area", "size of it", "how much area"):
		wanted = "area"
	case containsAny(text, "where is", "where are", "where was", "where can"):
		wanted = "location"
	case containsAny(text, "when was", "when did", "when founded", "when born"):
		wanted = "history"
	default:
		return Response{}, false
	}
	markers := map[string][]string{
		"population": {"population", "residents", "inhabitants"},
		"river":      {"river", "waterway"},
		"area":       {"area", "square kilomet", "km2", "sq mi"},
		"location":   {"located", "lies in", "stands on", "situated", "in south", "in northern", "in eastern", "in western", "in central"},
		"history":    {"founded", "established", "born", "created in", "first recorded"},
	}[wanted]
	for _, sentence := range splitAnswerSteps(previous.Text) {
		if containsAny(strings.ToLower(sentence), markers...) {
			return Response{
				Text:    fmt.Sprintf("From the cited article about %s: %s", topic, strings.TrimSpace(sentence)),
				Module:  "retrieval/follow-up/" + wanted,
				Sources: append([]Source(nil), previous.Sources...),
			}, true
		}
	}
	return Response{}, false
}

func isRepeatRequest(text string) bool {
	return oneOf(text, "repeat", "repeat that", "repeat it", "say that again", "can you repeat that", "what did you say")
}

func isSourceRequest(text string) bool {
	return oneOf(text, "source", "sources", "show source", "show sources", "where did you get that", "what is your source", "cite that")
}

func isShorterRequest(text string) bool {
	return oneOf(text, "shorter", "make it shorter", "be concise", "briefly", "in short", "one sentence", "give me the short version")
}

func isWhyFollowup(text string) bool {
	return oneOf(text, "why", "why?", "why is that", "why is that?", "why did you suggest that", "why those steps", "why these steps", "why did you say that")
}

func isExplainFollowup(text string) bool {
	return oneOf(text, "explain that", "explain it", "what does that mean", "what does it mean", "what does that command do", "explain the command", "what is that command", "unpack that")
}

func isDetailRequest(text string) bool {
	return oneOf(text, "more detail", "more details", "tell me more", "go deeper", "elaborate", "can you elaborate", "expand on that")
}

func isSafetyFollowup(text string) bool {
	return oneOf(text, "is that safe", "is it safe", "is this safe", "could that break anything", "will this delete anything", "is that dangerous")
}

func oneOf(text string, values ...string) bool {
	text = strings.Trim(strings.TrimSpace(text), " ?.!\t\n")
	for _, value := range values {
		if text == strings.Trim(strings.TrimSpace(value), " ?.!\t\n") {
			return true
		}
	}
	return false
}

func firstSentence(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	for _, marker := range []string{". ", "? ", "! "} {
		if index := strings.Index(text, marker); index >= 0 {
			return strings.TrimSpace(text[:index+1])
		}
	}
	return text
}

func explainPreviousCommand(answer string) string {
	for _, match := range commandInAnswerRE.FindAllStringSubmatch(answer, -1) {
		if len(match) < 2 {
			continue
		}
		command := strings.TrimSpace(match[1])
		if description := explainLinuxCommand(command); description != "" {
			return fmt.Sprintf("`%s` %s", command, description)
		}
	}
	return ""
}

func explainLinuxCommand(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return ""
	}
	switch fields[0] {
	case "systemctl":
		return "asks systemd for the state and properties of a service or unit; `status` is observational and does not restart it."
	case "journalctl":
		return "reads log entries from the systemd journal; the `-u` option filters them to one service."
	case "df":
		return "reports filesystem space; `-h` formats sizes in human-readable units."
	case "du":
		return "estimates how much disk space files and directories use."
	case "find":
		return "walks a directory tree and applies the tests or actions supplied in the command. Review any destructive action before running it."
	case "free":
		return "reports physical memory and swap usage."
	case "ps":
		return "shows a snapshot of processes currently running on the system."
	case "ss":
		return "shows socket information, including listening network ports."
	case "ip":
		return "inspects or manages network interfaces, addresses, and routes; the `route` subcommand only displays routes."
	case "lsblk":
		return "lists block devices and their partitions and mount points."
	case "getent":
		return "queries the system name-service configuration, which can help check DNS resolution."
	case "namei":
		return "shows the permissions along each component of a path."
	case "chmod":
		return "changes file or directory permission bits. Check the target and requested mode before running it."
	case "chown":
		return "changes file ownership. Check the target path and owner before running it."
	case "apt", "dnf", "pacman", "zypper":
		return "manages software packages on a Linux distribution. Installation and removal commands change the system."
	case "ssh":
		return "starts a secure remote shell connection using the host and options that follow."
	case "curl":
		return "makes an HTTP request to the URL that follows; inspect the destination before sending private data."
	default:
		return "is a command mentioned in the previous guidance. Tell me its exact output if you want help interpreting it."
	}
}

func answerSafetyFollowup(previous Response) (Response, bool) {
	text := strings.ToLower(previous.Text)
	if containsAny(text, "-delete", "rm ", "prune", "will remove", "delete") {
		return Response{
			Text:   "That step can change or remove data. Preview the exact targets first, verify the path carefully, and make sure you have a backup before running the destructive command. The read-only preview command is the safer first step.",
			Module: "linux/follow-up/safety", NeedsClarification: true, Sources: append([]Source(nil), previous.Sources...),
		}, true
	}
	return Response{
		Text:   "The diagnostic commands I suggested only inspect the current state. Any command that installs, restarts, changes permissions, edits files, or deletes data can have side effects; check the exact target and impact before running it.",
		Module: "linux/follow-up/safety", Sources: append([]Source(nil), previous.Sources...),
	}, true
}

func linuxDistro(text string) string {
	match := linuxDistroRE.FindStringSubmatch(text)
	if len(match) == 0 {
		return ""
	}
	distro := strings.Join(strings.Fields(match[1]), " ")
	if len(match) > 2 && match[2] != "" {
		distro += " " + match[2]
	}
	switch strings.ToLower(distro) {
	case "mint":
		return "Linux Mint"
	case "red hat":
		return "RHEL"
	case "suse":
		return "openSUSE"
	case "popos", "pop os":
		return "Pop!_OS"
	default:
		return distro
	}
}

func replaceLinuxDistro(question, distro string) string {
	question = linuxDistroRE.ReplaceAllString(question, "")
	question = trailingDistroPrepositionRE.ReplaceAllString(question, "")
	question = strings.Join(strings.Fields(question), " ")
	question = strings.TrimRight(question, " ?!.;,")
	if strings.TrimSpace(question) == "" {
		question = "Linux troubleshooting"
	}
	return question + " on " + distro
}

func isEntityFollowup(text string) bool {
	text = strings.ToLower(text)
	if len(strings.Fields(text)) > 14 {
		return false
	}
	if externalEntityPronounRE.MatchString(text) || containsAny(text, "tell me more", "what about", "more detail", "elaborate") {
		return true
	}
	return oneOf(text, "population", "how many people", "people live", "river", "area", "how big", "how large", "residents", "inhabitants", "history")
}

func containsWholeWord(text, word string) bool {
	return regexp.MustCompile(`\b` + regexp.QuoteMeta(word) + `\b`).MatchString(text)
}

func externalFollowupQuery(text, topic, previousQuestion string) string {
	if containsAny(text, "what is it", "what is that", "who is it", "who is that") {
		return "What is " + topic + "?"
	}
	if containsAny(text, "its capital", "their capital", "capital there") {
		return "What is the capital of " + topic + "?"
	}
	if containsAny(text, "how does it work", "how does that work", "how does it function") {
		return "How does " + topic + " work?"
	}
	if containsAny(text, "why is it", "why are they", "why does it") {
		pronouns := regexp.MustCompile(`\b(it|that|there|they|he|she|its|their)\b`)
		return pronouns.ReplaceAllString(strings.TrimSpace(text), topic)
	}
	if containsAny(text, "how big", "how large", "population", "how many people", "people live", "residents", "inhabitants") {
		return "What is the size or population of " + topic + "?"
	}
	if containsAny(text, "when founded", "when was it founded", "when was it created", "when did it start") {
		return "When was " + topic + " founded or created?"
	}
	if containsAny(text, "history", "what about its history") {
		return "What is the history of " + topic + "?"
	}
	if containsAny(text, "tell me more", "more about", "more detail", "elaborate", "what about") {
		return "Tell me more about " + topic
	}
	if expanded := externalSingularPronounRE.ReplaceAllString(strings.TrimSpace(text), topic); expanded != strings.TrimSpace(text) {
		return expanded
	}
	// Preserve the user's actual follow-up and attach both known context clues.
	// This lets the factual search resolve questions such as "Which of those
	// four is largest?" without turning every question into a generic "Who is
	// <topic>?" or "Where is <topic>?" lookup.
	query := strings.TrimSpace(text + " " + topic)
	if previousQuestion != "" && containsAny(text, "those", "these", "them", "they") {
		query += " " + previousQuestion
	}
	return query
}

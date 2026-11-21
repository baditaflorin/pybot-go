package pybot

import (
	"regexp"
	"strings"
)

var socialQuestionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^(?:can|could|would) you (?:please )?(?:help|assist)(?: me)?(?: with (?:this|that|it))?$`),
	regexp.MustCompile(`^(?:can|could|would) i help you\b`),
	regexp.MustCompile(`^(?:can|could|would) i (?:be useful|be of help)(?: to you)?\b`),
	regexp.MustCompile(`^(?:is there|can|could|would) (?:anything|something) i (?:can|could) help you with\b`),
	regexp.MustCompile(`^would you (?:accept|choose|pick|take|wear|travel|rather|be willing to)\b`),
	regexp.MustCompile(`^(?:what|which) personal (?:ability|skill) would you\b`),
	regexp.MustCompile(`^(?:what|which)\b.*\bwould you (?:like|want|prefer|choose|hope|plan)\b`),
	regexp.MustCompile(`^(?:do|does|did) you (?:like|love|prefer|want|think|feel|believe|remember|have|eat|drink|exercise|travel|watch|listen|read|enjoy)\b`),
	regexp.MustCompile(`^have you (?:ever )?(?:seen|been|tried|done|visited|read|watched|heard|met|experienced)\b`),
	regexp.MustCompile(`^have you (?:ever )?(?:attended|gone to|been to)\b`),
	regexp.MustCompile(`^did you (?:go|attend|watch|see|visit|try|read|exercise|travel|eat|drink)\b`),
	regexp.MustCompile(`^how often do you\b`),
	regexp.MustCompile(`^are you (?:more of|a |an |available|doing anything|free|ready|interested|okay|all right|alright|happy|sad|sure|excited|married|single)\b`),
	regexp.MustCompile(`^(?:what|who|how|where|when|why|which) (?:do|does|did|are|is|were|was) you (?:think|feel|like|prefer|want|believe|mean|eat|drink|enjoy|call|remember|watch|read)\b`),
	regexp.MustCompile(`^(?:what|which) (?:do|would) you (?:like|want|prefer|choose)\b`),
	regexp.MustCompile(`^what would you like to improve about yourself\b`),
	regexp.MustCompile(`^(?:what|who) (?:is|are|was|were) your (?:personal |favorite |favourite )?(?:weakness|strength|opinion|preference|favorite|favourite|name|family|hobby|job|experience)`),
	regexp.MustCompile(`^(?:what|which) (?:type|kind) of people do you work with\b`),
	regexp.MustCompile(`^(?:which|what) (?:colleagues|coworkers|teammates) do you (?:work best with|collaborate with)\b`),
	regexp.MustCompile(`^what (?:things|qualities) impress you\b`),
	regexp.MustCompile(`^what foods do you eat\b`),
	regexp.MustCompile(`^where are you from\b`),
	regexp.MustCompile(`^what place did you grow up in\b`),
	regexp.MustCompile(`^(?:what's|what is) your hometown\b`),
	regexp.MustCompile(`^how (?:is|are) your family\b`),
	regexp.MustCompile(`^what can i do for you\b`),
	regexp.MustCompile(`^what do you need(?: to know)?\b`),
	regexp.MustCompile(`^what information do you need from me\b`),
	regexp.MustCompile(`^what details would you like from me\b`),
	regexp.MustCompile(`^how can i assist you\b`),
	regexp.MustCompile(`^do you need (?:a guide|any help|help)\b`),
	regexp.MustCompile(`^shall we (?:visit|go|meet|have|drop in|stop)\b`),
	regexp.MustCompile(`^should we (?:visit|go|meet|tour|have|drop in|stop)\b`),
	regexp.MustCompile(`^do you have (?:a )?(?:regular )?(?:workout|exercise) (?:routine|schedule)\b`),
	regexp.MustCompile(`^do you (?:keep|follow|maintain) (?:a )?(?:regular )?(?:workout|exercise) schedule\b`),
	regexp.MustCompile(`^can you tell (?:that|if|whether)\b`),
	regexp.MustCompile(`^would (?:next )?(?:monday|tuesday|wednesday|thursday|friday|saturday|sunday)\b.*\bsuit you\b`),
	regexp.MustCompile(`^(?:is|are|was|were) (?:this|that|it)\b.*\b(?:okay|all right|alright|fine)\b`),
	regexp.MustCompile(`^(?:is|would) (?:next )?(?:monday|tuesday|wednesday|thursday|friday|saturday|sunday)\b.*\b(?:okay|fine|good|work|suit)\b`),
	regexp.MustCompile(`^would (?:next )?(?:monday|tuesday|wednesday|thursday|friday|saturday|sunday)\b.*\bsuit you\b`),
	regexp.MustCompile(`^(?:do|does|did) you have (?:a free moment|anything to do|time to|any plans)\b`),
	regexp.MustCompile(`^(?:do|would) you mind\b`),
	regexp.MustCompile(`^(?:would|could) you (?:pour|bring|carry|hand|pick up|drive|visit|meet|lend|spare)\b`),
	regexp.MustCompile(`^(?:would|could) you ask (?:him|her|them|someone)\b`),
	regexp.MustCompile(`^(?:can|could|would) you (?:send|pass|deliver) (?:a )?message to\b`),
	regexp.MustCompile(`^(?:can|could|would) you (?:send|pass|deliver) .+ (?:a )?(?:note|message)\b`),
	regexp.MustCompile(`^can you stop somewhere\b`),
	regexp.MustCompile(`^(?:what|who|how|where|when|which) (?:kind|type|sort) of\b.*\b(?:would|do|does|did) you (?:like|want|prefer|plan|hope)\b`),
	regexp.MustCompile(`^how long (?:will|would|do|does|did|are) you\b`),
	regexp.MustCompile(`^how (?:much longer|long) do you (?:plan|intend|expect|want) to\b`),
	regexp.MustCompile(`^(?:is|would|will) it (?:be )?(?:a )?(?:long|short) (?:ride|walk|drive|trip|journey)\b`),
	regexp.MustCompile(`^(?:are|were) you (?:planning|going|coming|acting|being|looking|seeming)\b`),
	regexp.MustCompile(`^why(?:['’]d| did| do| does| are| were) you\b`),
	regexp.MustCompile(`^why (?:aren't|are not|won't|will not|can't|cannot) you\b`),
	regexp.MustCompile(`^how do you (?:usually )?(?:spend|structure) (?:your|the) (?:day|days|weekday|weekdays|time)\b`),
	regexp.MustCompile(`^(?:what|where|how) (?:does|do|did|is|are) (?:your|my|his|her|their) (?:mother|father|parent|friend|brother|sister|manager|husband|wife|child|daughter|son|boss|neighbor)\b`),
	regexp.MustCompile(`^what shall i call (?:him|her|them)\b`),
	regexp.MustCompile(`^what name should i use for (?:him|her|them|this person|that person)\b`),
	regexp.MustCompile(`^(?:what's|what is) the reason (?:you|she|he|they)\b`),
	regexp.MustCompile(`^(?:what's up|what is up|how's it going|how is it going)$`),
	regexp.MustCompile(`^how about (?:this|that|the|a|an|my)\b`),
	regexp.MustCompile(`^why not\b`),
	regexp.MustCompile(`^do you know (?:what|which|where|when|who|how) .+\b(?:she|he|they|my|your)\b`),
}

var knowledgeQuestionWrapper = regexp.MustCompile(`^(?:do|does|did) you know (?:what|who|where|when|why|which|how)\b`)
var personalPronounPattern = regexp.MustCompile(`\b(?:she|he|they|my|your)\b`)
var unresolvedReferencePattern = regexp.MustCompile(`(?i)\b(it|this|that|they|them|those|these|he|she|him|her|its|their|our)\b`)
var unresolvedQuestionStartPattern = regexp.MustCompile(`(?i)^(?:what|who|where|when|why|which|how|does|do|is|are|what's|who's|where's|how's|and)\b`)
var properNamePattern = regexp.MustCompile(`\b[A-Z][a-z]+(?:\s+[A-Z][a-z]+)+\b`)
var personalExperienceRequestPattern = regexp.MustCompile(`^(?:describe (?:a time|an experience)|tell me about (?:a time|an experience)|give (?:me )?an example of (?:a time|an experience))\b`)
var politeMindInformationRequestPattern = regexp.MustCompile(`^(?:do|would) you mind (?:please )?(?:telling me about|explaining|showing|describing|defining|summarizing|listing|naming|calculating|comparing|converting)\b`)
var informationRequestPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^(?:can|could|would) you (?:please )?(?:tell me|give me|name|list|explain|show|describe|define|summarize|calculate|compare|convert|find|look up)\b`),
	regexp.MustCompile(`^(?:please )?(?:tell me about|explain|show|describe|define|summarize)\b`),
	regexp.MustCompile(`^(?:can|could|would) you (?:please )?help me (?:understand|learn about|find out about|look up)\b`),
}
var informationQuestionPattern = regexp.MustCompile(`^(?:what|who|where|when|why|which|how|is|are|was|were|do|does|did|can|could|would|will|should|have|has|had)\b`)
var apartmentSearchAdvicePattern = regexp.MustCompile(`^(?:(?:can|could|would) you (?:please )?(?:tell me|explain)\s+)?(?:(?:what is|what's)\s+)?(?:the\s+)?(?:best way|how (?:should|can|do) i)\b.*\b(?:look for|find|search for|apartment|flat)\b`)
var generalAdvicePattern = regexp.MustCompile(`^(?:(?:can|could|would) you (?:please )?(?:tell me|explain)\s+)?(?:(?:what is|what's)\s+)?(?:the\s+)?(?:best|easiest|safest|quickest|cheapest)\s+way\s+to\b|^how (?:should|can|do) i\b.*\b(?:find|look for|choose|compare|plan|learn|start)\b`)

func isSocialQuestion(text string) bool {
	question := strings.TrimSpace(strings.TrimRight(normalizeText(text), "?.!"))
	if question == "" || isLinuxQuestion(question) || isNameQuestion(question) {
		return false
	}
	if politeMindInformationRequestPattern.MatchString(question) {
		return false
	}
	if knowledgeQuestionWrapper.MatchString(question) && !personalPronounPattern.MatchString(question) {
		return false
	}
	for _, pattern := range socialQuestionPatterns {
		if pattern.MatchString(question) {
			return true
		}
	}
	return false
}

func answerSocialQuestion(text string) Response {
	switch {
	case containsAny(text, "your name"):
		return Response{Text: "I'm PyBot Go, a CPU-only assistant that routes to deterministic solvers and sourced retrieval.", Module: "conversation/identity"}
	case strings.HasPrefix(text, "what can i do for you") || strings.HasPrefix(text, "what do you need") || strings.HasPrefix(text, "what information do you need") || strings.HasPrefix(text, "what details would you like from me") || strings.HasPrefix(text, "how can i assist you") || strings.HasPrefix(text, "is there anything i can help you") || strings.HasPrefix(text, "is there something i can help you"):
		return Response{
			Text:        "Nothing for myself, but I can help you. Tell me what you're looking for or what you're trying to do.",
			Module:      "conversation/offer",
			Suggestions: []string{"Ask a question", "Describe a task"},
		}
	case strings.HasPrefix(text, "where are you from") || strings.HasPrefix(text, "what place did you grow up") || strings.HasPrefix(text, "what's your hometown") || strings.HasPrefix(text, "what is your hometown"):
		return Response{
			Text:        "I don't have a hometown; I'm software. If you name a place, I can help with information about it.",
			Module:      "conversation/identity",
			Suggestions: []string{"Ask about a place", "Ask something else"},
		}
	case strings.HasPrefix(text, "shall we visit") || strings.HasPrefix(text, "shall we go") || strings.HasPrefix(text, "should we visit") || strings.HasPrefix(text, "should we go") || strings.HasPrefix(text, "should we tour"):
		return Response{
			Text:        "I can't travel in person, but I can help plan a visit. Which place are you considering, and when would you go?",
			Module:      "conversation/advice",
			Suggestions: []string{"Plan a visit", "Ask something else"},
		}
	case strings.HasPrefix(text, "is next ") || strings.HasPrefix(text, "would next "):
		return Response{
			Text:        "I don't keep a personal schedule, but I can help you plan around that day. What time and activity do you have in mind?",
			Module:      "conversation/clarify",
			Suggestions: []string{"Share the time and activity", "Ask something else"},
		}
	case isPersonalExperienceQuestion(text):
		return Response{
			Text:        "I don't have personal experiences, a job, or a family. I can still discuss the topic or help you think through your own situation.",
			Module:      "conversation/social-question",
			Suggestions: []string{"Discuss the topic", "Share your situation"},
		}
	case strings.HasPrefix(text, "could you ask ") || strings.HasPrefix(text, "would you ask ") || strings.HasPrefix(text, "can you send a message") || strings.HasPrefix(text, "could you send a message") || strings.HasPrefix(text, "would you send a message") || strings.HasPrefix(text, "can you pass a message") || strings.HasPrefix(text, "could you pass a message") || strings.HasPrefix(text, "would you pass a message") || strings.HasPrefix(text, "could you send ") || strings.HasPrefix(text, "would you send ") || strings.HasPrefix(text, "can you send "):
		return Response{
			Text:        "I can't contact someone for you, but I can draft a short message asking them to call back.",
			Module:      "conversation/request",
			Suggestions: []string{"Draft the message", "Ask something else"},
		}
	case strings.HasPrefix(text, "can you stop somewhere"):
		return Response{
			Text:        "I can't control a vehicle or make a physical stop. If you share your route or area, I can help you look for a suitable stop.",
			Module:      "conversation/request",
			Suggestions: []string{"Share the route", "Ask something else"},
		}
	case containsAny(text, "do you mind if", "would you mind if"):
		return Response{
			Text:        "No, go ahead. I don't have personal preferences; if you want to discuss something afterward, tell me what it is.",
			Module:      "conversation/social-question",
			Suggestions: []string{"Go ahead", "Ask something else"},
		}
	case containsAny(text, "do you mind", "would you mind"):
		return Response{
			Text:        "I don't mind. Tell me what you'd like me to do or explain, and I'll help if I can.",
			Module:      "conversation/social-question",
			Suggestions: []string{"Explain a topic", "Ask a factual question"},
		}
	case containsAny(text, "would you pour", "could you pour", "would you bring", "could you bring", "would you carry", "could you carry", "would you hand", "could you hand", "would you pick up", "could you pick up", "would you drive", "could you drive", "would you visit", "could you visit", "would you meet", "could you meet", "would you lend", "could you lend", "would you spare", "could you spare"):
		return Response{
			Text:        "I can't do physical errands or meet in person. I can help you plan it or draft a message if that would help.",
			Module:      "conversation/request",
			Suggestions: []string{"Help me plan it", "Draft a message"},
		}
	case containsAny(text, "can i help you", "could i be useful", "could i be of help", "can i be useful", "can you help me", "could you help me", "do you need a guide", "do you need help"):
		return Response{
			Text:        "Yes. Tell me what you're trying to do, what you've tried, and what result you want. I'll say clearly if it's outside what I can help with.",
			Module:      "conversation/social-question",
			Suggestions: []string{"Ask a factual question", "Ask for Linux help", "Try a calculation"},
		}
	case containsAny(text, "weakness", "strength", "leader or a follower", "personal opinion", "your opinion", "improve about yourself"):
		return Response{
			Text:        "I don't have a personal life or experiences. My limits are that I use deterministic routes and retrieved sources rather than open-ended text generation, so I can miss ambiguous wording. What are you testing me on?",
			Module:      "conversation/social-question",
			Suggestions: []string{"Ask a factual question", "Try a follow-up"},
		}
	case containsAny(text, "your mother", "your father", "your family", "your brother", "your sister", "your sibling", "what shall i call", "what name should i use for"):
		return Response{
			Text:               "I don't have family or personal details about people unless you share them here. Tell me who you mean and what you know, and I'll work from that context.",
			Module:             "conversation/social-question",
			NeedsClarification: true,
			Suggestions:        []string{"Share the context", "Ask something else"},
		}
	case containsAny(text, "tell that", "tell if", "tell whether", "are you feeling", "do you feel"):
		return Response{
			Text:        "I don't have feelings, and I can't reliably infer yours from a short message. If you tell me what you're feeling, I can help think through it.",
			Module:      "conversation/social-question",
			Suggestions: []string{"Tell me more", "Ask for a next step"},
		}
	case knowledgeQuestionWrapper.MatchString(text) && personalPronounPattern.MatchString(text):
		return Response{
			Text:               "I can't know private details about someone unless you share them. Tell me what you know, and I can help you reason from it.",
			Module:             "conversation/social-question",
			NeedsClarification: true,
			Suggestions:        []string{"Share what you know", "Ask something else"},
		}
	case containsAny(text, "what's up", "what is up", "how's it going", "how is it going"):
		return Response{Text: "I'm here and ready to help. What's on your mind?", Module: "conversation/small-talk"}
	case containsAny(text, "why not", "why won't you", "why will not you", "why can't you", "how about this", "how about that", "how about the"):
		return Response{
			Text:        "I can't make personal plans or judge an outfit from text alone, but I can help compare options or plan the outing if you tell me what matters to you.",
			Module:      "conversation/social-question",
			Suggestions: []string{"Compare options", "Plan the outing"},
		}
	case containsAny(text, "why'd you", "why did you", "why are you acting", "why are you being", "what's the reason you", "what is the reason you"):
		return Response{
			Text:               "I don't know another person's reason from that question alone. If you share what happened, I can help think through plausible explanations.",
			Module:             "conversation/social-question",
			NeedsClarification: true,
			Suggestions:        []string{"Share what happened", "Ask a different question"},
		}
	case containsAny(text, "long ride", "long walk", "long drive", "long trip", "long journey", "short ride", "short walk", "short drive", "short trip", "short journey"):
		return Response{
			Text:               "I need the starting point and how you're travelling to estimate that. Where are you starting from, and are you walking, driving, or taking transit?",
			Module:             "conversation/clarify",
			NeedsClarification: true,
			Suggestions:        []string{"Give me the starting point", "Ask something else"},
		}
	case containsAny(text, "like", "love", "prefer", "want", "choose", "accept", "improve", "favorite", "favourite", "eat", "drink", "exercise", "travel", "watch", "listen", "read", "enjoy", "seen", "been", "visited", "tried", "doing anything", "available", "free", "ready", "married", "single", "stay", "plan to stay", "intend to stay", "suit you", "collaborate", "planning", "going", "leader", "follower"):
		return Response{
			Text:        "I don't have personal preferences, experiences, or a schedule. I can still compare options or help with information about the topic—what would be useful?",
			Module:      "conversation/social-question",
			Suggestions: []string{"Compare some options", "Ask for information"},
		}
	default:
		return Response{
			Text:        "I'm here to help with questions and text-based tasks. What would you like to work on?",
			Module:      "conversation/social-question",
			Suggestions: []string{"Ask a question", "Tell me what you're trying to do"},
		}
	}
}

func isPersonalExperienceQuestion(text string) bool {
	for _, prefix := range []string{
		"did you ", "have you attended", "have you been to", "how often do you",
		"do you have a regular workout", "do you have a workout", "what foods do you eat",
		"what type of people do you work", "what kind of people do you work",
		"which colleagues do you work", "what colleagues do you work",
		"which coworkers do you work", "what coworkers do you work",
		"which teammates do you work", "what teammates do you work",
		"what things impress you", "what qualities impress you", "how do you usually spend", "how is your family",
	} {
		if strings.HasPrefix(text, prefix) {
			return true
		}
	}
	return false
}

func answerDialogueIntent(text string, dialogueIntent intent) (Response, bool) {
	switch dialogueIntent {
	case intentDialoguePositive:
		return Response{
			Text:        "That sounds like good news. What part are you most pleased about?",
			Module:      "conversation/positive-response",
			Suggestions: []string{"Tell me more", "Plan a next step"},
		}, true
	case intentDialogueNegative:
		return Response{
			Text:        "I'm sorry you're dealing with that. Would you rather talk it through or focus on a practical next step?",
			Module:      "conversation/support",
			Suggestions: []string{"Talk it through", "Find a next step"},
		}, true
	case intentDialogueAgree:
		return Response{
			Text:        "Understood. What should I take into account as we continue?",
			Module:      "conversation/acknowledgment",
			Suggestions: []string{"Continue", "Ask a follow-up"},
		}, true
	case intentDialogueOffer:
		return Response{
			Text:        "Thanks for offering. If there's information or an option you want me to work with, share it here and I'll use it in this chat.",
			Module:      "conversation/offer",
			Suggestions: []string{"Share the details", "Ask something else"},
		}, true
	case intentDialogueRequest:
		return Response{
			Text:        "Sure. What are you trying to do, and what would a useful result look like?",
			Module:      "conversation/request",
			Suggestions: []string{"Describe the outcome", "Share the relevant details"},
		}, true
	case intentDialogueStatement:
		if containsAny(text, "i apologize", "i'm sorry", "i am sorry") {
			return Response{Text: "No worries. What would you like to do next?", Module: "conversation/acknowledgment"}, true
		}
		return Response{
			Text:        "Thanks for sharing that. Would you like to explore it, compare options, or plan a next step?",
			Module:      "conversation/acknowledgment",
			Suggestions: []string{"Explore this", "Compare options", "Plan a next step"},
		}, true
	default:
		return Response{}, false
	}
}

func isConversationalStatement(text string) bool {
	if strings.Contains(text, "?") {
		return false
	}
	words := strings.Fields(text)
	if len(words) > 1 {
		if oneOf(words[0], "what", "who", "when", "where", "why", "which", "how") &&
			oneOf(words[1], "is", "are", "was", "were", "do", "does", "did", "can", "could", "would", "will", "should", "have", "has", "had") {
			return false
		}
		if oneOf(words[0], "is", "are", "was", "were", "do", "does", "did", "can", "could", "would", "will", "should", "have", "has", "had", "what's", "who's", "where's", "how's") {
			return false
		}
	}
	for _, prefix := range []string{
		"please ", "tell me ", "show me ",
		"give me ", "help me ", "find ", "list ", "explain ", "describe ", "calculate ", "compare ", "convert ",
		"install ", "run ", "check ", "set ", "turn ", "go ", "let's ",
	} {
		if strings.HasPrefix(text, prefix) {
			return false
		}
	}
	return true
}

func unresolvedContextQuestion(text string) (string, bool) {
	if !strings.Contains(text, "?") || !unresolvedReferencePattern.MatchString(text) || properNamePattern.MatchString(text) {
		return "", false
	}
	if !unresolvedQuestionStartPattern.MatchString(text) {
		return "", false
	}
	match := unresolvedReferencePattern.FindStringSubmatch(text)
	if len(match) < 2 {
		return "", false
	}
	return match[1], true
}

func answerPersonalExperienceRequest(text string) (Response, bool) {
	if !personalExperienceRequestPattern.MatchString(strings.TrimSpace(text)) && !strings.HasPrefix(text, "tell me about a time when") {
		return Response{}, false
	}
	return Response{
		Text:        "I don't have personal experiences to draw from. I can help you draft a strong answer from a real example in your life—tell me the situation, what you did, and the result.",
		Module:      "conversation/request",
		Suggestions: []string{"Share a real example", "Ask for a response outline"},
	}, true
}

func answerPracticalAdvice(text string) (Response, bool) {
	text = strings.ToLower(strings.TrimSpace(text))
	if isLinuxQuestion(text) {
		return Response{}, false
	}
	if apartmentSearchAdvicePattern.MatchString(text) && containsAny(text, "apartment", "flat", "place to rent") {
		return Response{
			Text:               "Start with a total monthly budget, including utilities and commuting. Pick a few neighborhoods, compare current listings across more than one reputable site, and verify the landlord and lease before paying a deposit. View the place in person or by live video when possible. Which city, budget, and move-in date should I use to narrow this down?",
			Module:             "conversation/advice",
			NeedsClarification: true,
			Suggestions:        []string{"Share the city and budget", "Ask about rental scams"},
		}, true
	}
	if generalAdvicePattern.MatchString(text) {
		return Response{
			Text:               "Start by deciding what matters most and what constraints are fixed. Compare the options against the same short list of criteria, check the evidence for each, and note the trade-offs before choosing a next step. What options are you weighing, and which factors matter most to you?",
			Module:             "conversation/advice",
			NeedsClarification: true,
			Suggestions:        []string{"Share the options", "List what matters most"},
		}, true
	}
	return Response{}, false
}

func answerContextDependentQuestion(text string) (Response, bool) {
	switch {
	case !isLinuxQuestion(text) && containsAny(text, "package", "gift box", "bundle") && containsAny(text, "included", "includes", "contain", "inside"):
		return Response{
			Text:               "Which package or product do you mean? Share its name or description, and I'll help interpret what's included.",
			Module:             "conversation/clarify",
			NeedsClarification: true,
			Suggestions:        []string{"Name the package", "Share a link"},
		}, true
	case strings.Contains(text, "landlord") && containsAny(text, "rent", "charge", "cost", "price"),
		containsAny(text, "rent", "rental", "lease") && containsAny(text, "how much", "cost", "price", "pay"):
		return Response{
			Text:               "Rent depends on location, property size, and when you plan to move. Which city or area and what size place do you mean?",
			Module:             "conversation/clarify",
			NeedsClarification: true,
			Suggestions:        []string{"Give me the city and size", "Ask something else"},
		}, true
	case containsAny(text, "qualification", "qualifications", "credential", "credentials", "requirement", "requirements") && containsAny(text, "position", "job", "role", "opening"):
		return Response{
			Text:               "The required qualifications depend on the role and location. What job title and country or region is the position in?",
			Module:             "conversation/clarify",
			NeedsClarification: true,
			Suggestions:        []string{"Give me the job title and location", "Ask something else"},
		}, true
	case strings.Contains(text, "where should i go") && containsAny(text, "university", "campus"),
		strings.Contains(text, "office") && containsAny(text, "campus", "university") && containsAny(text, "visit", "find", "reach", "arrive"):
		return Response{
			Text:               "That depends on the campus and what you need to do. Which university are you at, and are you looking for admissions, student services, housing, or a specific office?",
			Module:             "conversation/clarify",
			NeedsClarification: true,
			Suggestions:        []string{"Name the university and task", "Ask something else"},
		}, true
	}
	return Response{}, false
}

func isInformationSeekingRequest(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	if politeMindInformationRequestPattern.MatchString(text) {
		return true
	}
	for _, pattern := range informationRequestPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func isInformationSeekingMessage(text string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	return isInformationSeekingRequest(text) || informationQuestionPattern.MatchString(text)
}

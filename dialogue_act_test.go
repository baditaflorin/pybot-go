package pybot

import (
	"strings"
	"testing"
)

func TestInformativeStatementsGetAConversationalTurn(t *testing.T) {
	response := New(0).Reply("We've reduced energy use at our factory by 15 percent.")
	if response.Module != "conversation/acknowledgment" {
		t.Fatalf("module = %q, want a conversational acknowledgment; response: %#v", response.Module, response)
	}
	if !strings.Contains(response.Text, "Would you like") || len(response.Sources) != 0 {
		t.Fatalf("statement response should acknowledge without inventing a source: %#v", response)
	}
}

func TestPersonalAndSocialQuestionsAvoidFactualSearch(t *testing.T) {
	cases := []string{
		"Can I help you?",
		"Would you accept a work trip?",
		"Would you choose boots or sneakers?",
		"Which personal skill would you most like to improve?",
		"Do you think listeners understand my pronunciation?",
		"Would Friday suit you?",
		"Can you tell whether I'm excited?",
		"How much longer do you plan to stay in Toronto?",
		"Which courses would you prefer to take?",
		"Why are you behaving unusually today?",
		"Why did you send her a gift?",
		"Why won't you eat more?",
		"How do you usually spend your weekdays, Alex?",
		"What name should I use for them?",
		"Would you mind if I put the movie on?",
		"Do you know which snack my teammate enjoys?",
	}
	for _, input := range cases {
		t.Run(input, func(t *testing.T) {
			response := New(0).Reply(input)
			if response.Module != "conversation/social-question" {
				t.Fatalf("module = %q, want social-question; response: %#v", response.Module, response)
			}
			if len(response.Sources) != 0 {
				t.Fatalf("social question unexpectedly retrieved a source: %#v", response.Sources)
			}
		})
	}
}

func TestPersonalExperiencePromptOffersAUsefulDraftingPath(t *testing.T) {
	for _, input := range []string{
		"Tell me about a time you made a difficult decision.",
		"Give an example of a time when you led a challenging task.",
		"Describe an experience where you had to adapt quickly.",
	} {
		response := New(0).Reply(input)
		if response.Module != "conversation/request" || !strings.Contains(response.Text, "I don't have personal experiences") || !strings.Contains(response.Text, "real example") {
			t.Errorf("%q: response = %#v, want an honest prompt to draft from the user's real experience", input, response)
		}
	}
}

func TestPolitePermissionAndPhysicalRequestsGetAppropriateReplies(t *testing.T) {
	permission := New(0).Reply("Do you mind if I ask something?")
	if permission.Module != "conversation/social-question" || !strings.Contains(permission.Text, "go ahead") {
		t.Fatalf("permission response = %#v, want an invitation to continue", permission)
	}

	request := New(0).Reply("Could you explain photosynthesis for me?")
	if request.Module == "conversation/request" || !isInformationSeekingRequest("Could you explain photosynthesis for me?") {
		t.Fatalf("information request was treated as a generic dialogue act: %#v", request)
	}

	errands := New(0).Reply("Could you carry this box for me?")
	if errands.Module != "conversation/request" || !strings.Contains(errands.Text, "physical errands") {
		t.Fatalf("physical request response = %#v, want a clear capability boundary and useful alternative", errands)
	}
}

func TestAdviceGetsUsefulGuidanceAndDependentFactsAskForContext(t *testing.T) {
	if isLinuxQuestion("How can I find a reasonably priced apartment?") {
		t.Fatal("ordinary use of find was misclassified as a Linux command question")
	}
	if !isLinuxQuestion("How do I find large files on Ubuntu?") {
		t.Fatal("an explicit Linux file-search question lost its domain classification")
	}

	advice := New(0).Reply("How can I find a reasonably priced apartment?")
	if advice.Module != "conversation/advice" || !strings.Contains(advice.Text, "monthly budget") || !advice.NeedsClarification {
		t.Fatalf("apartment advice = %#v, want practical steps and a location/budget follow-up", advice)
	}
	decisionAdvice := New(0).Reply("How should I compare two job offers?")
	if decisionAdvice.Module != "conversation/advice" || !strings.Contains(decisionAdvice.Text, "same short list of criteria") {
		t.Fatalf("decision advice = %#v, want a reusable comparison approach", decisionAdvice)
	}
	if answer, handled := answerPracticalAdvice("I found an apartment near the university."); handled {
		t.Fatalf("a declarative statement was mistaken for an advice request: %#v", answer)
	}

	cases := []struct {
		input string
		part  string
	}{
		{"How much does a rental usually cost?", "Which city or area"},
		{"What credentials does this job opening require?", "job title and country"},
		{"Which office should I find when I reach campus?", "Which university"},
	}
	for _, test := range cases {
		response := New(0).Reply(test.input)
		if response.Module != "conversation/clarify" || !response.NeedsClarification || !strings.Contains(response.Text, test.part) {
			t.Errorf("%q: response = %#v, want clarification containing %q", test.input, response, test.part)
		}
	}
}

func TestInformationalQuestionsCannotBeSwallowedByDialogueActHints(t *testing.T) {
	cases := []struct {
		input  string
		intent string
	}{
		{"What is the capital of Romania?", "dialogue_statement"},
		{"Can you explain photosynthesis?", "dialogue_request"},
		{"Is Bucharest the capital of Romania?", "dialogue_positive_emotion"},
	}
	for _, test := range cases {
		response := New(0).ReplyWithSemanticHint(test.input, test.intent, 0.9)
		if response.Module == "conversation/acknowledgment" || response.Module == "conversation/request" || response.Module == "conversation/support" {
			t.Errorf("%q with %s hint was swallowed by a dialogue reply: %#v", test.input, test.intent, response)
		}
	}
}

func TestTravelQuestionRequestsMissingRouteContext(t *testing.T) {
	response := New(0).Reply("Would it be a long drive to the art gallery from here?")
	if response.Module != "conversation/clarify" || !response.NeedsClarification {
		t.Fatalf("response = %#v, want a request for origin and transport", response)
	}
	if !strings.Contains(response.Text, "starting point") {
		t.Fatalf("response does not ask for the missing origin: %q", response.Text)
	}
}

func TestAmbiguousWhyAndPackageQuestionsStayInConversation(t *testing.T) {
	why := New(0).Reply("Why'd you call me earlier?")
	if why.Module != "conversation/social-question" || !why.NeedsClarification {
		t.Fatalf("why-question = %#v, want a social question asking for context", why)
	}

	packageQuestion := New(0).Reply("What does the seasonal bundle contain?")
	if packageQuestion.Module != "conversation/clarify" || !packageQuestion.NeedsClarification || !strings.Contains(packageQuestion.Text, "Which package or product") {
		t.Fatalf("package-question = %#v, want a clarification for the missing product", packageQuestion)
	}
}

func TestUnresolvedReferencesAskForTheMissingEntity(t *testing.T) {
	for _, input := range []string{"What does it say?", "How far is it from our hotel?", "How much is it?", "Who are they for?", "What's he doing?"} {
		response := New(0).Reply(input)
		if response.Module != "conversation/context-clarify" || !response.NeedsClarification {
			t.Errorf("%q: response = %#v, want a contextual clarification", input, response)
			continue
		}
	}
	if _, unresolved := unresolvedContextQuestion("What about Abraham Lincoln? Did you get questions about him?"); unresolved {
		t.Fatal("a named entity in the same turn should not be treated as an unresolved reference")
	}
}

func TestPersonalKnowledgeQuestionsAskForContextInsteadOfLookingUpAProfile(t *testing.T) {
	response := New(0).Reply("Do you know which gift my teammate enjoys?")
	if response.Module != "conversation/social-question" || !response.NeedsClarification {
		t.Fatalf("response = %#v, want a clarification about private context", response)
	}
	if !strings.Contains(response.Text, "private details") {
		t.Fatalf("response should explain the missing context: %q", response.Text)
	}
}

func TestPersonalAndServiceTurnsDoNotFallIntoFactualLookup(t *testing.T) {
	cases := []struct {
		input  string
		module string
		phrase string
	}{
		{"What's your hometown?", "conversation/identity", "hometown"},
		{"How can I assist you?", "conversation/offer", "Nothing for myself"},
		{"What information do you need from me?", "conversation/offer", "Nothing for myself"},
		{"Which coworkers do you work best with?", "conversation/social-question", "personal experiences"},
		{"Have you attended a live music show?", "conversation/social-question", "personal experiences"},
		{"Do you have a regular workout routine?", "conversation/social-question", "personal experiences"},
		{"Should we tour the art museum tomorrow?", "conversation/advice", "can't travel"},
		{"Could you pass a message to my teammate asking them to call?", "conversation/request", "can't contact someone"},
	}
	for _, test := range cases {
		t.Run(test.input, func(t *testing.T) {
			response := New(0).Reply(test.input)
			if !isSocialQuestion(test.input) {
				t.Errorf("question was not recognized as a social/service turn")
			}
			if response.Module != test.module || !strings.Contains(response.Text, test.phrase) {
				t.Errorf("response = %#v, want module %q containing %q", response, test.module, test.phrase)
			}
			if len(response.Sources) > 0 {
				t.Errorf("social/service turn unexpectedly returned sources: %#v", response.Sources)
			}
		})
	}
}

func TestConversationActHintsChooseGroundedResponseStyles(t *testing.T) {
	cases := []struct {
		input  string
		intent string
		module string
		phrase string
	}{
		{"I got a new job and feel excited.", "dialogue_positive_emotion", "conversation/positive-response", "good news"},
		{"I've been feeling very down lately.", "dialogue_negative_emotion", "conversation/support", "sorry"},
		{"I see, that makes sense.", "dialogue_acknowledgment", "conversation/acknowledgment", "Understood"},
		{"I could refer you for the open position if you want.", "dialogue_offer", "conversation/offer", "Thanks for offering"},
		{"Book me a ride to the station, please.", "dialogue_request", "conversation/request", "useful result"},
	}
	for _, test := range cases {
		t.Run(test.intent, func(t *testing.T) {
			response := New(0).ReplyWithSemanticHint(test.input, test.intent, 0.8)
			if response.Module != test.module || !strings.Contains(response.Text, test.phrase) {
				t.Fatalf("response = %#v, want module %q and phrase %q", response, test.module, test.phrase)
			}
		})
	}
}

func TestSocialQuestionRulesDoNotInterceptExpertQuestions(t *testing.T) {
	if isSocialQuestion("Can you help me with Linux?") {
		t.Fatal("Linux help request was classified as social small talk")
	}
	if isSocialQuestion("Do you know where Romania is?") {
		t.Fatal("factual question wrapper was classified as social small talk")
	}
	response := New(0).Reply("Can you help me with Linux?")
	if response.Module != "retrieval/linux-help" {
		t.Fatalf("module = %q, want Linux help; response: %s", response.Module, response.Text)
	}
}

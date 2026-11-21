package pybot

import (
	"strings"
	"testing"
)

type evaluationTurn struct {
	input              string
	module             string
	contains           []string
	needsClarification bool
	checkClarification bool
	wantSource         bool
}

type evaluationScenario struct {
	name  string
	steps []evaluationTurn
}

func evalStep(input, module string, contains ...string) evaluationTurn {
	return evaluationTurn{input: input, module: module, contains: contains}
}

func evalClarification(input, module string, contains ...string) evaluationTurn {
	return evaluationTurn{input: input, module: module, contains: contains, needsClarification: true, checkClarification: true}
}

func evalSourced(input, module string, contains ...string) evaluationTurn {
	return evaluationTurn{input: input, module: module, contains: contains, wantSource: true}
}

// TestConversationEvaluation50Scenarios keeps a varied, multi-turn regression
// set for the product's supported conversation routes. These are behavior
// checks, not a claim that the bot can answer arbitrary questions.
func TestConversationEvaluation50Scenarios(t *testing.T) {
	scenarios := []evaluationScenario{
		{name: "compare two ages in one question", steps: []evaluationTurn{evalStep("One person is 12 and someone else is 50. Who is older?", "reasoning/age-comparison", "Person 2", "38 years")}},
		{name: "compare ages remembered across turns", steps: []evaluationTurn{evalStep("Alice is 12 years old.", "memory/age", "Alice"), evalStep("Bob is 50 years old.", "memory/age", "Bob"), evalStep("Who is older?", "reasoning/age-comparison", "Bob", "38 years")}},
		{name: "insert a third person into the age ranking", steps: []evaluationTurn{evalStep("Alice is 12 years old.", "memory/age", "Alice"), evalStep("Bob is 50 years old.", "memory/age", "Bob"), evalStep("And if George is 30 where does he sit?", "reasoning/age-ranking", "George", "18 years older than Alice", "20 years younger than Bob")}},
		{name: "compare two named people among three facts", steps: []evaluationTurn{evalStep("Alice is 12 years old.", "memory/age", "Alice"), evalStep("Bob is 50 years old.", "memory/age", "Bob"), evalStep("Charlie is 40 years old.", "memory/age", "Charlie"), evalStep("Who is older, Alice or Bob?", "reasoning/age-comparison", "Bob is older", "38 years")}},
		{name: "find oldest of three remembered people", steps: []evaluationTurn{evalStep("Mira is 24 years old.", "memory/age", "Mira"), evalStep("Noah is 31 years old.", "memory/age", "Noah"), evalStep("Omar is 19 years old.", "memory/age", "Omar"), evalStep("Who is the oldest?", "reasoning/age-comparison", "Noah", "31")}},
		{name: "recall the user's age", steps: []evaluationTurn{evalStep("I am 37 years old.", "memory/age", "37"), evalStep("How old am I?", "memory/age", "You are 37 years old")}},
		{name: "ask for clarification on an ambiguous age pronoun", steps: []evaluationTurn{evalStep("Alice is 12 years old.", "memory/age", "Alice"), evalClarification("She is 50 years old.", "memory/age", "Who do you mean")}},
		{name: "update a remembered age", steps: []evaluationTurn{evalStep("Alice is 12 years old.", "memory/age", "Alice"), evalStep("Alice is 13 years old now.", "memory/age", "Alice"), evalStep("How old is Alice?", "memory/age", "Alice is 13 years old")}},
		{name: "ask about an age not supplied", steps: []evaluationTurn{evalClarification("How old is Marta?", "memory/age", "don't have", "Marta")}},
		{name: "rank a new person with a natural follow-up", steps: []evaluationTurn{evalStep("Alice is 12 years old.", "memory/age", "Alice"), evalStep("Bob is 50 years old.", "memory/age", "Bob"), evalStep("And George is 30. Where does that put him?", "reasoning/age-ranking", "George", "middle")}},
		{name: "calculate square root of two", steps: []evaluationTurn{evalStep("What is the square root of 2?", "reasoning/arithmetic", "1.414213562373")}},
		{name: "multiply in natural language", steps: []evaluationTurn{evalStep("Calculate 8 times 7", "reasoning/arithmetic", "56")}},
		{name: "evaluate parenthesized arithmetic", steps: []evaluationTurn{evalStep("Compute (3 + 4) * 2", "reasoning/arithmetic", "14")}},
		{name: "handle division by zero without inventing a result", steps: []evaluationTurn{evalClarification("What is 9 divided by 0?", "reasoning/arithmetic", "undefined")}},
		{name: "reject a negative real square root", steps: []evaluationTurn{evalClarification("What is the square root of -1?", "reasoning/arithmetic", "not a real number")}},
		{name: "ask for a missing arithmetic expression", steps: []evaluationTurn{evalClarification("Can you calculate this for me?", "reasoning/arithmetic", "arithmetic expression")}},
		{name: "convert centimeters to meters", steps: []evaluationTurn{evalStep("Convert 100 centimeters to meters", "reasoning/unit-conversion", "100 cm = 1 m")}},
		{name: "convert Celsius to Fahrenheit", steps: []evaluationTurn{evalStep("How much is 20 Celsius in Fahrenheit?", "reasoning/unit-conversion", "20 °C = 68 °F")}},
		{name: "reject incompatible unit dimensions", steps: []evaluationTurn{evalClarification("Convert 5 kilograms to meters", "reasoning/unit-conversion", "different things")}},
		{name: "explain an unsupported unit", steps: []evaluationTurn{evalClarification("Convert 5 furlongs to meters", "reasoning/unit-conversion", "don't recognize", "units")}},
		{name: "add days to a date", steps: []evaluationTurn{evalStep("What date is 30 days after 2026-11-18?", "reasoning/date", "2026-12-18", "Friday")}},
		{name: "calculate days between dates", steps: []evaluationTurn{evalStep("How many days between 2026-01-01 and 2026-01-10?", "reasoning/date", "9 days")}},
		{name: "remember a Linux distribution", steps: []evaluationTurn{evalStep("I use Fedora 41.", "memory/environment", "Fedora 41")}},
		{name: "apply remembered distribution to package advice", steps: []evaluationTurn{evalStep("I use Fedora 41.", "memory/environment", "Fedora 41"), evalSourced("How do I update packages?", "retrieval/linux-package-updates", "dnf upgrade")}},
		{name: "recall the remembered distribution", steps: []evaluationTurn{evalStep("I use Debian 13.", "memory/environment", "Debian 13"), evalStep("Which distro am I using?", "memory/environment", "Debian 13")}},
		{name: "remove a remembered distribution", steps: []evaluationTurn{evalStep("I use Ubuntu 24.04.", "memory/environment", "Ubuntu"), evalStep("Forget my distro.", "memory/forget", "removed"), evalStep("Which distro am I using?", "memory/environment", "don't know yet")}},
		{name: "diagnose low disk space with a source", steps: []evaluationTurn{evalSourced("How do I check disk space on Ubuntu?", "retrieval/linux-filesystem-space", "df -hT")}},
		{name: "investigate a failed systemd service", steps: []evaluationTurn{evalSourced("My systemd service failed. How do I see why?", "retrieval/linux-systemd-service-failure", "journalctl -u SERVICE")}},
		{name: "ask for the next service diagnostic", steps: []evaluationTurn{evalSourced("My systemd service failed. How do I see why?", "retrieval/linux-systemd-service-failure", "journalctl"), evalSourced("What next?", "retrieval/follow-up/linux-systemd-service-failure", "systemctl status SERVICE")}},
		{name: "identify process listening on a port", steps: []evaluationTurn{evalSourced("Which Linux process is listening on port 8080?", "retrieval/linux-listening-ports", "ss -lntup")}},
		{name: "explain chmod permissions", steps: []evaluationTurn{evalSourced("What does chmod 755 mean in Linux?", "retrieval/linux-chmod-755", "owner read, write, and execute")}},
		{name: "troubleshoot DNS on Debian", steps: []evaluationTurn{evalSourced("How do I troubleshoot DNS on Debian?", "retrieval/linux-dns", "getent ahosts")}},
		{name: "inspect Docker storage", steps: []evaluationTurn{evalSourced("How do I inspect Docker disk use?", "retrieval/linux-docker-storage", "docker system df -v")}},
		{name: "ask for guided Linux steps", steps: []evaluationTurn{evalStep("One step at a time, please.", "memory/preferences", "one step at a time"), evalSourced("My systemd service failed. What should I check?", "linux/guided-step", "Step 1 of", "Run only this step")}},
		{name: "advance guided Linux diagnosis", steps: []evaluationTurn{evalStep("One step at a time.", "memory/preferences", "one step at a time"), evalSourced("My systemd service failed. What should I check?", "linux/guided-step", "Step 1 of"), evalSourced("Next step", "linux/guided-step", "Step 2 of")}},
		{name: "repeat the current guided step", steps: []evaluationTurn{evalStep("One step at a time.", "memory/preferences", "one step at a time"), evalSourced("My systemd service failed. What should I check?", "linux/guided-step", "Step 1 of"), evalSourced("Repeat the current step", "linux/guided-step", "Step 1 of")}},
		{name: "explain a numbered guided step", steps: []evaluationTurn{evalStep("One step at a time.", "memory/preferences", "one step at a time"), evalSourced("My systemd service failed. What should I check?", "linux/guided-step", "Step 1 of"), evalSourced("What does step 2 mean?", "linux/guided-explanation", "Step 2 means")}},
		{name: "show the remaining guided plan", steps: []evaluationTurn{evalStep("One step at a time.", "memory/preferences", "one step at a time"), evalSourced("My systemd service failed. What should I check?", "linux/guided-step", "Step 1 of"), evalSourced("Show the rest", "linux/guided-plan", "2.")}},
		{name: "recover from a failed guided command safely", steps: []evaluationTurn{evalStep("One step at a time.", "memory/preferences", "one step at a time"), evalSourced("My systemd service failed. What should I check?", "linux/guided-step", "Step 1 of"), evalClarification("I got an error.", "linux/guided-output", "exact command", "Redact passwords")}},
		{name: "turn off guided Linux mode from a negative request", steps: []evaluationTurn{evalStep("One step at a time.", "memory/preferences", "one step at a time"), evalStep("I don't need one step at a time.", "memory/preferences", "complete answer"), evalSourced("My systemd service failed. What should I check?", "retrieval/linux-systemd-service-failure", "journalctl")}},
		{name: "repeat, shorten, and cite a sourced response", steps: []evaluationTurn{evalSourced("How do I check disk space on Ubuntu?", "retrieval/linux-filesystem-space", "df -hT"), evalSourced("Repeat that", "conversation/repeat", "df -hT"), evalSourced("Make it shorter", "conversation/shorten", "df -hT"), {input: "Show sources", module: "conversation/source", wantSource: true}}},
		{name: "keep a safety warning in a brief Linux answer", steps: []evaluationTurn{evalStep("Keep it brief.", "memory/preferences", "keep answers brief"), evalSourced("How do I update packages on Ubuntu?", "retrieval/linux-package-updates", "sudo apt update", "production machine")}},
		{name: "set detailed answer preference", steps: []evaluationTurn{evalStep("Give me more detail.", "memory/preferences", "more explanation")}},
		{name: "set beginner explanation preference", steps: []evaluationTurn{evalStep("I'm a beginner.", "memory/preferences", "beginner-friendly")}},
		{name: "set bullet formatting preference", steps: []evaluationTurn{evalStep("Use bullet points.", "memory/preferences", "bullet points")}},
		{name: "summarize saved preferences", steps: []evaluationTurn{evalStep("Keep it brief.", "memory/preferences", "keep answers brief"), evalStep("Use bullet points.", "memory/preferences", "bullet points"), evalStep("What preferences do you remember?", "memory/preferences", "brief", "bullet points")}},
		{name: "forget saved answer preferences", steps: []evaluationTurn{evalStep("Keep it brief.", "memory/preferences", "keep answers brief"), evalStep("Forget my preferences.", "memory/preferences", "cleared"), evalStep("What answer style are you using?", "memory/preferences", "haven't saved")}},
		{name: "recap recent conversation topics", steps: []evaluationTurn{evalSourced("How do I check disk space on Ubuntu?", "retrieval/linux-filesystem-space", "df -hT"), evalStep("What is 12 plus 50?", "reasoning/arithmetic", "62"), evalStep("What topics have we covered?", "conversation/topics", "disk", "calculation")}},
		{name: "return to an earlier Linux topic", steps: []evaluationTurn{evalSourced("How do I check disk space on Ubuntu?", "retrieval/linux-filesystem-space", "df -hT"), evalSourced("How do I troubleshoot DNS on Debian?", "retrieval/linux-dns", "getent"), evalSourced("Go back to disk space", "retrieval/linux-filesystem-space", "Picking up", "df -hT")}},
		{name: "answer a capability question with concrete boundaries", steps: []evaluationTurn{evalStep("What can you do?", "conversation/capabilities", "troubleshoot Linux", "follow-ups", "capital of Romania", "I don't use a generative LLM or GPU")}},
	}

	if len(scenarios) != 50 {
		t.Fatalf("evaluation scenarios = %d, want 50", len(scenarios))
	}
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			bot := New(0)
			for index, step := range scenario.steps {
				response := bot.Reply(step.input)
				if step.module != "" && response.Module != step.module {
					t.Fatalf("turn %d %q: module = %q, want %q; answer: %s", index+1, step.input, response.Module, step.module, response.Text)
				}
				for _, expected := range step.contains {
					if !strings.Contains(strings.ToLower(response.Text), strings.ToLower(expected)) {
						t.Errorf("turn %d %q: answer %q does not contain %q", index+1, step.input, response.Text, expected)
					}
				}
				if step.checkClarification && response.NeedsClarification != step.needsClarification {
					t.Errorf("turn %d %q: clarification = %t, want %t", index+1, step.input, response.NeedsClarification, step.needsClarification)
				}
				if step.wantSource && len(response.Sources) == 0 {
					t.Errorf("turn %d %q: expected a source, got none", index+1, step.input)
				}
			}
		})
	}
}

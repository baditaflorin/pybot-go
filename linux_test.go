package pybot

import (
	"strings"
	"testing"
)

func TestLinuxQuestionsReturnRelevantSourcedDiagnostics(t *testing.T) {
	cases := []struct {
		question string
		document string
		answer   string
	}{
		{"How do I check disk space on Ubuntu?", "linux-filesystem-space", "df -hT"},
		{"My systemd service failed. How do I see why?", "linux-systemd-service-failure", "journalctl -u SERVICE"},
		{"Which Linux process is listening on port 8080?", "linux-listening-ports", "ss -lntup"},
		{"What does chmod 755 mean in Linux?", "linux-chmod-755", "owner read, write, and execute"},
		{"How do I troubleshoot DNS on Debian?", "linux-dns", "getent ahosts"},
		{"How do I upgrade packages on Fedora?", "linux-package-updates", "dnf upgrade"},
		{"How can I find large files on Linux?", "linux-find-large-files", "find /var"},
		{"How do I inspect Docker disk use?", "linux-docker-storage", "docker system df -v"},
	}
	for _, test := range cases {
		t.Run(test.question, func(t *testing.T) {
			response := New(0).Reply(test.question)
			if response.Module != "retrieval/"+test.document {
				t.Fatalf("module = %q, want retrieval/%s; answer: %s", response.Module, test.document, response.Text)
			}
			if !strings.Contains(response.Text, test.answer) {
				t.Fatalf("answer %q does not contain %q", response.Text, test.answer)
			}
			if len(response.Sources) == 0 || !strings.HasPrefix(response.Sources[0].URL, "https://") {
				t.Fatalf("response has no HTTPS source: %#v", response)
			}
		})
	}
}

func TestLinuxTroubleshootingFollowupRequestsUsefulEvidence(t *testing.T) {
	bot := New(0)
	first := bot.Reply("My systemd service failed. How do I see why?")
	if first.Module != "retrieval/linux-systemd-service-failure" {
		t.Fatalf("initial response = %#v", first)
	}
	followup := bot.Reply("What next?")
	if followup.Module != "retrieval/follow-up/linux-systemd-service-failure" ||
		!strings.Contains(followup.Text, "systemctl status SERVICE") || len(followup.Sources) == 0 {
		t.Fatalf("follow-up = %#v", followup)
	}
}

func TestLinuxHelpQuestionSetsDistroAndRedactionExpectations(t *testing.T) {
	response := New(0).Reply("Can you help me with Linux?")
	if response.Module != "retrieval/linux-help" {
		t.Fatalf("module = %q, want retrieval/linux-help; answer: %s", response.Module, response.Text)
	}
	for _, want := range []string{"distribution and version", "exact error", "read-only checks", "Redact passwords"} {
		if !strings.Contains(response.Text, want) {
			t.Errorf("answer %q is missing %q", response.Text, want)
		}
	}
}

func TestLinuxIntentRecognizesAdminCommandsAndRespectsExplicitOtherOS(t *testing.T) {
	for _, question := range []string{
		"What does chmod 755 mean?", "How do I check disk space on Ubuntu?",
		"How do I inspect systemd logs?", "Which process uses port 8080?",
	} {
		if !isLinuxQuestion(question) {
			t.Errorf("isLinuxQuestion(%q) = false, want true", question)
		}
	}
	if isLinuxQuestion("How do I check disk space on Windows with PowerShell?") {
		t.Fatal("explicit Windows question was classified as Linux administration")
	}
}

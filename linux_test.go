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
		{"How do I make ls show sizes in megabytes?", "linux-ls", "--block-size=MB"},
		{"Execute vs read bit: how do directory permissions work?", "linux-permissions", "execute allows traversal"},
		{"How do I know whether Wayland or X11 is being used?", "linux-display-session", "XDG_SESSION_TYPE"},
		{"How do you empty the buffers and cache on a Linux system?", "linux-memory-cache", "available"},
		{"How do I recursively delete empty directories in my home directory?", "linux-empty-directories", `find "$HOME" -type d -empty -print`},
		{"How do I set a default memory limit for a single Linux process?", "linux-process-memory-limit", "MemoryMax="},
		{"Why does /etc/resolv.conf point at 127.0.0.53?", "linux-systemd-resolved-stub", "local loopback DNS stub"},
		{"How do I list all mounted filesystems?", "linux-mounts", "findmnt --list"},
		{"How do I identify Linux virtualization technology?", "linux-virtualization", "systemd-detect-virt"},
		{"How do I kill all a user's processes using their UID?", "linux-process-user-stop", "loginctl terminate-user"},
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

func TestLinuxRetrievalDoesNotSubstituteUnrelatedTopics(t *testing.T) {
	cases := []struct {
		question string
		module   string
		avoid    string
	}{
		{"How do I recursively delete empty directories in my home directory?", "retrieval/linux-empty-directories", "kettle"},
		{"How do I check whether Wayland or X11 is active?", "retrieval/linux-display-session", "memory"},
		{"How do I clear Linux buffers and cache?", "retrieval/linux-memory-cache", "docker"},
		{"How do I display meminfo in megabytes in top?", "retrieval/linux-top-memory", "--sort=-%cpu"},
	}
	for _, test := range cases {
		t.Run(test.question, func(t *testing.T) {
			response := New(0).Reply(test.question)
			if response.Module != test.module {
				t.Fatalf("module = %q, want %q; answer: %s", response.Module, test.module, response.Text)
			}
			if strings.Contains(strings.ToLower(response.Text), strings.ToLower(test.avoid)) {
				t.Fatalf("answer contains unrelated topic %q: %s", test.avoid, response.Text)
			}
			if len(response.Sources) == 0 {
				t.Fatalf("response has no source: %#v", response)
			}
		})
	}
}

func TestCapabilityQuestionListsUsefulLinuxAndReasoningTasks(t *testing.T) {
	response := New(0).Reply("What can you do?")
	if response.Module != "conversation/capabilities" {
		t.Fatalf("module = %q, want conversation/capabilities; answer: %s", response.Module, response.Text)
	}
	for _, want := range []string{"troubleshoot Linux", "square roots", "capital of Romania", "I don't use a generative LLM or GPU"} {
		if !strings.Contains(response.Text, want) {
			t.Errorf("capability answer %q is missing %q", response.Text, want)
		}
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

func TestLinuxConceptQuestionsReturnSourcedDefinitions(t *testing.T) {
	cases := []struct {
		question string
		document string
		answer   string
	}{
		{"What is Linux?", "linux-overview", "Linux is Unix-like"},
		{"What is systemd?", "linux-systemd-overview", "system and service manager"},
		{"What is SSH?", "linux-ssh-overview", "secure remote login"},
	}
	for _, test := range cases {
		t.Run(test.question, func(t *testing.T) {
			response := New(0).Reply(test.question)
			if response.Module != "retrieval/"+test.document {
				t.Fatalf("module = %q, want retrieval/%s; answer: %s", response.Module, test.document, response.Text)
			}
			if !strings.Contains(strings.ToLower(response.Text), strings.ToLower(test.answer)) || len(response.Sources) == 0 {
				t.Fatalf("response is not a sourced answer containing %q: %#v", test.answer, response)
			}
		})
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
	if isLinuxQuestion("What is Linux?") {
		t.Fatal("general Linux definition was classified as administration help")
	}
}

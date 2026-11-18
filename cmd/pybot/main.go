package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/baditaflorin/pybot-go"
)

const (
	logo = `    ____        ____        __
   / __ \__  __/ __ )____  / /_
  / /_/ / / / / __  / __ \/ __/
 / ____/ /_/ / /_/ / /_/ / /_
/_/    \__, /_____/\____/\__/
      /____/`
	tagline  = "A small chatbot written in Go. No weights, no training, just rules."
	helpText = `Commands:
  /reset   clear the conversation memory
  /help    show this help
  /exit    quit (Ctrl+D works too)`
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, isTerminal(os.Stdin)))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer, stdinIsTerminal bool) int {
	fs := flag.NewFlagSet("pybot", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var prompt string
	var interactive, verbose, showVersion bool
	defaultSeed := time.Now().UnixNano()
	var seed int64
	fs.StringVar(&prompt, "p", "", "reply to one message and exit")
	fs.StringVar(&prompt, "prompt", "", "reply to one message and exit")
	fs.BoolVar(&interactive, "i", false, "with -p, continue chatting after the first reply")
	fs.BoolVar(&interactive, "interactive", false, "with -p, continue chatting after the first reply")
	fs.Int64Var(&seed, "s", defaultSeed, "seed for repeatable wording choices")
	fs.Int64Var(&seed, "seed", defaultSeed, "seed for repeatable wording choices")
	fs.BoolVar(&verbose, "v", false, "show the rule module and response time")
	fs.BoolVar(&verbose, "verbose", false, "show the rule module and response time")
	fs.BoolVar(&showVersion, "version", false, "print version and exit")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: pybot [options]")
		fmt.Fprintln(stderr, "Chat with a small rule-based bot. Piped input gets one reply per line.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintf(stderr, "pybot: unexpected argument %q\n", fs.Arg(0))
		fs.Usage()
		return 2
	}
	if showVersion {
		fmt.Fprintf(stdout, "pybot-go %s\n", pybot.Version)
		return 0
	}
	promptProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "p" || f.Name == "prompt" {
			promptProvided = true
		}
	})

	bot := pybot.New(seed)
	shouldChat := (!promptProvided && stdinIsTerminal) || (promptProvided && interactive)
	if shouldChat {
		if promptProvided {
			return chat(bot, prompt, stdin, stdout, stderr, verbose, stdinIsTerminal)
		}
		return chat(bot, "", stdin, stdout, stderr, verbose, stdinIsTerminal)
	}
	if promptProvided {
		return oneShot(bot, prompt, stdout, stderr, verbose)
	}
	return pipe(bot, stdin, stdout, stderr, verbose)
}

func oneShot(bot *pybot.Bot, prompt string, stdout, stderr io.Writer, verbose bool) int {
	start := time.Now()
	response := bot.Reply(prompt)
	if _, err := fmt.Fprintln(stdout, response.Text); err != nil {
		fmt.Fprintf(stderr, "pybot: write reply: %v\n", err)
		return 1
	}
	if verbose {
		fmt.Fprintf(stderr, "[module: %s | %d ms]\n", response.Module, time.Since(start).Milliseconds())
	}
	return 0
}

func pipe(bot *pybot.Bot, stdin io.Reader, stdout, stderr io.Writer, verbose bool) int {
	scanner := newScanner(stdin)
	for scanner.Scan() {
		start := time.Now()
		response := bot.Reply(strings.TrimSuffix(strings.TrimSuffix(scanner.Text(), "\n"), "\r"))
		if _, err := fmt.Fprintln(stdout, response.Text); err != nil {
			fmt.Fprintf(stderr, "pybot: write reply: %v\n", err)
			return 1
		}
		if verbose {
			fmt.Fprintf(stderr, "[module: %s | %d ms]\n", response.Module, time.Since(start).Milliseconds())
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(stderr, "pybot: read input: %v\n", err)
		return 1
	}
	return 0
}

func chat(bot *pybot.Bot, first string, stdin io.Reader, stdout, stderr io.Writer, verbose, clearable bool) int {
	if _, err := fmt.Fprintf(stdout, "%s\n\n%s\nType a message and press Enter. /reset starts a new conversation; /exit quits.\n\n", logo, tagline); err != nil {
		fmt.Fprintf(stderr, "pybot: write banner: %v\n", err)
		return 1
	}
	scanner := newScanner(stdin)
	firstPending := first != ""
	for {
		var line string
		if firstPending {
			line = first
			firstPending = false
			if _, err := fmt.Fprintf(stdout, "user: %s\n", line); err != nil {
				fmt.Fprintf(stderr, "pybot: write prompt: %v\n", err)
				return 1
			}
		} else {
			if _, err := io.WriteString(stdout, "user: "); err != nil {
				fmt.Fprintf(stderr, "pybot: write prompt: %v\n", err)
				return 1
			}
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					fmt.Fprintf(stderr, "pybot: read input: %v\n", err)
					return 1
				}
				_, _ = fmt.Fprintln(stdout)
				return 0
			}
			line = strings.TrimSuffix(strings.TrimSuffix(scanner.Text(), "\n"), "\r")
		}

		command := strings.ToLower(strings.TrimSpace(line))
		switch command {
		case "/exit", "/quit", "/bye":
			return 0
		case "/reset":
			bot.Reset()
			if clearable && runtime.GOOS != "windows" && os.Getenv("TERM") != "" {
				_, _ = io.WriteString(stdout, "\033[H\033[2J\033[3J")
				_, _ = fmt.Fprintf(stdout, "%s\n\n%s\nType a message and press Enter. /reset starts a new conversation; /exit quits.\n\n", logo, tagline)
			} else {
				_, _ = fmt.Fprintln(stdout, "New conversation.")
			}
			continue
		case "/help":
			_, _ = fmt.Fprint(stdout, helpText+"\n\n")
			continue
		case "":
			continue
		}
		start := time.Now()
		response := bot.Reply(line)
		if _, err := fmt.Fprintf(stdout, "pybot: %s\n", response.Text); err != nil {
			fmt.Fprintf(stderr, "pybot: write reply: %v\n", err)
			return 1
		}
		if verbose {
			fmt.Fprintf(stdout, "[module: %s | %d ms]\n", response.Module, time.Since(start).Milliseconds())
		}
		if _, err := fmt.Fprintln(stdout); err != nil {
			fmt.Fprintf(stderr, "pybot: write reply: %v\n", err)
			return 1
		}
	}
}

func newScanner(input io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4096), 4*1024*1024)
	return scanner
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

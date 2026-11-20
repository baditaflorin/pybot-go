//go:build js && wasm

package main

import (
	"encoding/json"
	"unicode/utf8"

	pybot "github.com/baditaflorin/pybot-go"
	"syscall/js"
)

var callbacks []js.Func

func main() {
	bot := pybot.New(0)
	global := js.Global()

	reply := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) == 0 || args[0].Type() != js.TypeString {
			return marshal(pybot.Response{Text: "Enter a message to get started."})
		}
		prompt := args[0].String()
		if utf8.RuneCountInString(prompt) > 2000 {
			return marshal(pybot.Response{
				Text:               "Please keep each message under 2,000 characters.",
				Module:             "conversation/input-limit",
				NeedsClarification: true,
			})
		}
		hint := ""
		if len(args) > 1 && args[1].Type() == js.TypeString {
			hint = args[1].String()
		}
		confidence := 0.0
		if len(args) > 2 && args[2].Type() == js.TypeNumber {
			confidence = args[2].Float()
		}
		return marshal(bot.ReplyWithSemanticHint(prompt, hint, confidence))
	})
	reset := js.FuncOf(func(_ js.Value, _ []js.Value) any {
		bot.Reset()
		return nil
	})
	restore := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) == 0 || args[0].Type() != js.TypeString {
			return false
		}
		var turns []pybot.ConversationTurn
		if err := json.Unmarshal([]byte(args[0].String()), &turns); err != nil {
			return false
		}
		bot.RestoreConversation(turns)
		return true
	})
	rememberResponse := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) < 2 || args[0].Type() != js.TypeString || args[1].Type() != js.TypeString {
			return false
		}
		var response pybot.Response
		if err := json.Unmarshal([]byte(args[1].String()), &response); err != nil {
			return false
		}
		bot.RememberResponse(args[0].String(), response)
		return true
	})
	callbacks = append(callbacks, reply, reset, restore, rememberResponse)
	global.Set("pybotReply", reply)
	global.Set("pybotReset", reset)
	global.Set("pybotRestoreConversation", restore)
	global.Set("pybotRememberResponse", rememberResponse)
	global.Set("pybotReady", true)
	select {}
}

func marshal(response pybot.Response) string {
	data, err := json.Marshal(response)
	if err != nil {
		return `{"text":"I couldn't format that response.","module":"conversation/error"}`
	}
	return string(data)
}

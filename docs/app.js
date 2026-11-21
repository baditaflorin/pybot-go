(() => {
  const form = document.getElementById("composer");
  const prompt = document.getElementById("prompt");
  const messages = document.getElementById("messages");
  const status = document.getElementById("runtime-status");
  const sendButton = document.getElementById("send-button");
  const resetButton = document.getElementById("reset-button");
  const exportButton = document.getElementById("export-button");
  const thinking = document.getElementById("thinking-indicator");
  const messageCounter = document.getElementById("message-counter");
  const promptCards = Array.from(document.querySelectorAll("[data-prompt]"));
  const recentTurns = [];
  const conversation = [];
  const storageKey = "pybot-go-conversation-v2";
  let runtimeReady = false;
  let sending = false;
  let semanticMessage = "CPU language model loading…";
  let semanticState = "loading";

  function renderStatus(message, state = "ready") {
    const dot = status.querySelector(".status-dot");
    dot.classList.toggle("loading", state === "loading");
    dot.classList.toggle("error", state === "error");
    status.lastElementChild.textContent = message;
  }

  function setReady(ready, message) {
    runtimeReady = ready;
    renderStatus(ready ? semanticMessage : message, ready ? semanticState : "error");
    prompt.disabled = !ready;
    sendButton.disabled = !ready || sending;
    resetButton.disabled = !ready || sending;
    exportButton.disabled = !ready || conversation.length === 0;
    promptCards.forEach((card) => { card.disabled = !ready; });
  }

  window.addEventListener("pybot-semantic-status", (event) => {
    semanticMessage = event.detail?.message || "CPU language model status unavailable";
    semanticState = event.detail?.state || "ready";
    if (runtimeReady) renderStatus(semanticMessage, semanticState);
  });

  function scrollToLatest() {
    messages.scrollTop = messages.scrollHeight;
  }

  function addUserMessage(text) {
    const article = document.createElement("article");
    article.className = "message user-message";

    const content = document.createElement("div");
    content.className = "message-content";
    const meta = document.createElement("div");
    meta.className = "message-meta";
    const label = document.createElement("span");
    label.textContent = "You";
    meta.append(label);

    const bubble = document.createElement("div");
    bubble.className = "bubble";
    bubble.textContent = text;
    content.append(meta, bubble);
    article.append(content);
    messages.append(article);
    scrollToLatest();
  }

  function addSourceLinks(parent, sources) {
    if (!Array.isArray(sources) || sources.length === 0) return;
    const list = document.createElement("div");
    list.className = "source-list";
    sources.forEach((source) => {
      if (!source || typeof source.title !== "string" || typeof source.url !== "string") return;
      let url;
      try { url = new URL(source.url); } catch { return; }
      if (url.protocol !== "https:" && url.protocol !== "http:") return;

      const link = document.createElement("a");
      link.className = "source-link-chip";
      link.href = url.href;
      link.target = "_blank";
      link.rel = "noopener noreferrer";
      const title = document.createElement("span");
      title.textContent = source.title;
      const arrow = document.createElementNS("http://www.w3.org/2000/svg", "svg");
      arrow.setAttribute("viewBox", "0 0 12 12");
      arrow.setAttribute("aria-hidden", "true");
      const path = document.createElementNS("http://www.w3.org/2000/svg", "path");
      path.setAttribute("d", "M3 9 9 3M4 3h5v5");
      arrow.append(path);
      link.append(title, arrow);
      list.append(link);
    });
    if (list.childElementCount) parent.append(list);
  }

  function addAssistantMessage(response) {
    const article = document.createElement("article");
    article.className = "message assistant-message";

    const avatar = document.createElement("div");
    avatar.className = "message-avatar";
    avatar.setAttribute("aria-hidden", "true");
    avatar.textContent = "p.";

    const content = document.createElement("div");
    content.className = "message-content";
    const meta = document.createElement("div");
    meta.className = "message-meta";
    const label = document.createElement("span");
    label.textContent = "PyBot";
    const separator = document.createElement("span");
    separator.className = "meta-separator";
    separator.textContent = "·";
    const time = document.createElement("span");
    time.textContent = new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    meta.append(label, separator, time);

    const bubble = document.createElement("div");
    bubble.className = "bubble";
    bubble.textContent = response.text || "I couldn't produce a response.";
    content.append(meta, bubble);

    const details = document.createElement("div");
    details.className = "answer-detail";
    if (response.module) {
      const module = document.createElement("span");
      module.className = "module-label";
      module.textContent = response.module;
      details.append(module);
    }
    if (response.needs_clarification) {
      const clarification = document.createElement("span");
      clarification.className = "clarification-label";
      clarification.textContent = "needs one detail";
      details.append(clarification);
    }
    if (details.childElementCount) content.append(details);
    addSourceLinks(content, response.sources);

    const tools = document.createElement("div");
    tools.className = "message-tools";
    const copy = document.createElement("button");
    copy.type = "button";
    copy.className = "copy-button";
    copy.textContent = "Copy";
    copy.setAttribute("aria-label", "Copy this answer");
    copy.addEventListener("click", async () => {
      try {
        const sources = Array.isArray(response.sources) ? response.sources.map((source) => `${source.title}: ${source.url}`).join("\n") : "";
        await navigator.clipboard.writeText([response.text || "", sources].filter(Boolean).join("\n\n"));
        copy.textContent = "Copied";
        window.setTimeout(() => { copy.textContent = "Copy"; }, 1400);
      } catch {
        copy.textContent = "Copy unavailable";
        window.setTimeout(() => { copy.textContent = "Copy"; }, 1800);
      }
    });
    tools.append(copy);
    if (response.retry_query) {
      const retry = document.createElement("button");
      retry.type = "button";
      retry.className = "copy-button";
      retry.textContent = "Retry lookup";
      retry.addEventListener("click", () => sendMessage(response.retry_query));
      tools.append(retry);
    }
    content.append(tools);

    if (Array.isArray(response.suggestions) && response.suggestions.length) {
      const suggestions = document.createElement("div");
      suggestions.className = "suggestion-list";
      suggestions.setAttribute("aria-label", "Suggested follow-ups");
      response.suggestions.slice(0, 4).forEach((suggestion) => {
        if (typeof suggestion !== "string" || !suggestion.trim()) return;
        const button = document.createElement("button");
        button.type = "button";
        button.className = "suggestion-button";
        button.textContent = suggestion;
        button.addEventListener("click", () => sendMessage(suggestion));
        suggestions.append(button);
      });
      if (suggestions.childElementCount) content.append(suggestions);
    }

    article.append(avatar, content);
    messages.append(article);
    scrollToLatest();
  }

  function setThinking(active, message = "Thinking…") {
    if (!messages.contains(thinking)) messages.append(thinking);
    thinking.hidden = !active;
    thinking.lastElementChild.textContent = message;
    scrollToLatest();
  }

  function persistConversation() {
    try {
      sessionStorage.setItem(storageKey, JSON.stringify(conversation.slice(-24)));
    } catch {
      renderStatus("Conversation storage is full · this tab only", "error");
    }
    exportButton.disabled = !runtimeReady || conversation.length === 0;
  }

  function restoreConversation() {
    let stored;
    try {
      stored = JSON.parse(sessionStorage.getItem(storageKey) || "[]");
    } catch {
      try { sessionStorage.removeItem(storageKey); } catch { /* Storage can be disabled by the browser. */ }
      return;
    }
    if (!Array.isArray(stored)) return;
    const turns = stored.slice(-24).filter((turn) => turn && typeof turn.user === "string" && turn.assistant && typeof turn.assistant.text === "string");
    if (!turns.length) return;
    if (typeof window.pybotRestoreConversation !== "function" || !window.pybotRestoreConversation(JSON.stringify(turns))) return;
    conversation.push(...turns);
    recentTurns.push(...turns.map((turn) => turn.user));
    messages.replaceChildren();
    turns.forEach((turn) => {
      addUserMessage(turn.user);
      addAssistantMessage(turn.assistant);
    });
    exportButton.disabled = false;
  }

  function exportConversation() {
    if (!conversation.length) return;
    const lines = ["# PyBot Go conversation", ""];
    conversation.forEach((turn) => {
      lines.push("## You", "", turn.user, "", "## PyBot", "", turn.assistant.text || "");
      if (Array.isArray(turn.assistant.sources) && turn.assistant.sources.length) {
        lines.push("", "Sources:", ...turn.assistant.sources.map((source) => `- [${source.title}](${source.url})`));
      }
      lines.push("");
    });
    const url = URL.createObjectURL(new Blob([lines.join("\n")], { type: "text/markdown;charset=utf-8" }));
    const link = document.createElement("a");
    link.href = url;
    link.download = `pybot-conversation-${new Date().toISOString().slice(0, 10)}.md`;
    link.click();
    window.setTimeout(() => URL.revokeObjectURL(url), 1000);
  }

  async function lookupWikipedia(question) {
    const params = new URLSearchParams({
      action: "query",
      generator: "search",
      gsrsearch: question,
      gsrnamespace: "0",
      gsrlimit: "3",
      prop: "extracts",
      exintro: "1",
      explaintext: "1",
      exsentences: "4",
      format: "json",
      formatversion: "2",
      origin: "*",
    });
    const response = await fetch(`https://en.wikipedia.org/w/api.php?${params}`, {
      headers: { Accept: "application/json" },
      signal: AbortSignal.timeout(7000),
    });
    if (!response.ok) throw new Error("Wikipedia lookup failed");
    const data = await response.json();
    const pages = (data.query?.pages || []).filter((page) => page.extract && page.title)
      .sort((a, b) => a.index - b.index);
    const page = pages[0];
    if (!page) throw new Error("No sourced answer found");
    const extract = String(page.extract).replace(/\s+/g, " ").trim().slice(0, 1000);
    return {
      text: extract || `I found a Wikipedia page for ${page.title}, but it has no short introduction.`,
      module: "retrieval/wikipedia",
      suggestions: ["Tell me more", "Show source", "Where is it?"],
      sources: [{
        title: `Wikipedia · ${page.title}`,
        url: `https://en.wikipedia.org/wiki/${encodeURIComponent(page.title.replace(/ /g, "_"))}`,
      }],
    };
  }

  function isFactualQuestion(text) {
    return /^(what|who|where|when|why|which|how|define|explain|tell me about)\b/i.test(text.trim());
  }

  async function sendMessage(text) {
    const value = text.trim();
    if (!value || typeof window.pybotReply !== "function" || sending) return;
    addUserMessage(value);
    prompt.value = "";
    resizePrompt();
    sending = true;
    sendButton.disabled = true;
    resetButton.disabled = true;
    setThinking(true, "Matching your question to the conversation…");
    const context = recentTurns.slice(-2);
    const previousTurn = conversation[conversation.length - 1];
    if (previousTurn?.assistant?.text) context.push(previousTurn.assistant.text.slice(0, 400));
    recentTurns.push(value);
    if (recentTurns.length > 8) recentTurns.shift();
    try {
      const semantic = typeof window.pybotSemanticClassify === "function"
        ? await window.pybotSemanticClassify(value, context)
        : { intent: "", score: 0 };
      let response = JSON.parse(window.pybotReply(value, semantic.intent || "", semantic.score || 0));
      const lookupQuery = response.lookup_query || (response.module === "conversation/unknown" && isFactualQuestion(value) ? value : "");
      if (lookupQuery) {
        setThinking(true, "Looking up a cited answer…");
        try {
          response = await lookupWikipedia(lookupQuery);
        } catch {
          response = {
            text: "I couldn't reach Wikipedia just now, so I don't have a cited answer for that question. You can retry the lookup.",
            module: "retrieval/wikipedia-error",
            retry_query: lookupQuery,
          };
        }
      }
      if (typeof window.pybotRememberResponse === "function") {
        window.pybotRememberResponse(value, JSON.stringify(response));
      }
      conversation.push({user: value, assistant: response, intent: semantic.intent || "", confidence: semantic.score || 0});
      if (conversation.length > 24) conversation.shift();
      persistConversation();
      addAssistantMessage(response);
    } catch {
      const response = {
        text: "The local bot runtime returned an unreadable response. Start a new chat and try again.",
        module: "browser/runtime-error",
      };
      conversation.push({user: value, assistant: response});
      if (conversation.length > 24) conversation.shift();
      persistConversation();
      addAssistantMessage(response);
    }
    sending = false;
    setThinking(false);
    sendButton.disabled = !runtimeReady;
    resetButton.disabled = !runtimeReady;
    prompt.focus();
  }

  function resizePrompt() {
    prompt.style.height = "auto";
    prompt.style.height = `${Math.min(prompt.scrollHeight, 115)}px`;
  }

  form.addEventListener("submit", (event) => {
    event.preventDefault();
    sendMessage(prompt.value);
  });

  prompt.addEventListener("input", resizePrompt);
  prompt.addEventListener("input", () => {
    const count = prompt.value.length;
    messageCounter.textContent = `${count.toLocaleString()} / 2,000`;
    messageCounter.classList.toggle("near-limit", count >= 1800 && count < 2000);
    messageCounter.classList.toggle("at-limit", count >= 2000);
  });
  prompt.addEventListener("keydown", (event) => {
    if (event.key === "Enter" && !event.shiftKey && !event.isComposing) {
      event.preventDefault();
      form.requestSubmit();
    }
  });

  promptCards.forEach((card) => {
    card.addEventListener("click", () => sendMessage(card.dataset.prompt || ""));
  });

  resetButton.addEventListener("click", () => {
    if (typeof window.pybotReset === "function") window.pybotReset();
    recentTurns.length = 0;
    conversation.length = 0;
    try { sessionStorage.removeItem(storageKey); } catch { /* Memory still resets for this page load. */ }
    exportButton.disabled = true;
    messages.replaceChildren();
    addAssistantMessage({
      text: "New conversation. Tell me a fact or try one of the example prompts.",
      module: "conversation/reset",
    });
    prompt.focus();
  });

  exportButton.addEventListener("click", exportConversation);

  async function startRuntime() {
    if (typeof Go !== "function") {
      setReady(false, "Runtime failed to load.");
      return;
    }
    try {
      const go = new Go();
      const binary = await fetch("./pybot.wasm?v=0.8.1");
      if (!binary.ok) throw new Error("WebAssembly bundle is unavailable.");
      const result = await WebAssembly.instantiate(await binary.arrayBuffer(), go.importObject);
      void go.run(result.instance).catch(() => setReady(false, "Runtime failed to load."));
      const deadline = Date.now() + 10000;
      const waitForGo = () => {
        if (window.pybotReady) {
          restoreConversation();
          setReady(true, semanticMessage);
          prompt.focus();
          return;
        }
        if (Date.now() >= deadline) {
          setReady(false, "Runtime failed to load.");
          return;
        }
        window.setTimeout(waitForGo, 25);
      };
      waitForGo();
    } catch {
      setReady(false, "Runtime failed to load.");
    }
  }

  setReady(false, "Loading browser runtime…");
  startRuntime();
})();

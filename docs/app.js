(() => {
  const form = document.getElementById("composer");
  const prompt = document.getElementById("prompt");
  const messages = document.getElementById("messages");
  const status = document.getElementById("runtime-status");
  const sendButton = document.getElementById("send-button");
  const resetButton = document.getElementById("reset-button");
  const promptCards = Array.from(document.querySelectorAll("[data-prompt]"));

  function setReady(ready, message) {
    const dot = status.querySelector(".status-dot");
    dot.classList.toggle("loading", !ready);
    dot.classList.toggle("error", message === "Runtime failed to load.");
    status.lastElementChild.textContent = message;
    prompt.disabled = !ready;
    sendButton.disabled = !ready;
    resetButton.disabled = !ready;
    promptCards.forEach((card) => { card.disabled = !ready; });
  }

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

    article.append(avatar, content);
    messages.append(article);
    scrollToLatest();
  }

  function sendMessage(text) {
    const value = text.trim();
    if (!value || typeof window.pybotReply !== "function") return;
    addUserMessage(value);
    prompt.value = "";
    resizePrompt();
    try {
      const response = JSON.parse(window.pybotReply(value));
      addAssistantMessage(response);
    } catch {
      addAssistantMessage({
        text: "The local bot runtime returned an unreadable response. Start a new chat and try again.",
        module: "browser/runtime-error",
      });
    }
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
    messages.replaceChildren();
    addAssistantMessage({
      text: "New conversation. Tell me a fact or try one of the example prompts.",
      module: "conversation/reset",
    });
    prompt.focus();
  });

  async function startRuntime() {
    if (typeof Go !== "function") {
      setReady(false, "Runtime failed to load.");
      return;
    }
    try {
      const go = new Go();
      const binary = await fetch("./pybot.wasm");
      if (!binary.ok) throw new Error("WebAssembly bundle is unavailable.");
      const result = await WebAssembly.instantiate(await binary.arrayBuffer(), go.importObject);
      void go.run(result.instance).catch(() => setReady(false, "Runtime failed to load."));
      const deadline = Date.now() + 10000;
      const waitForGo = () => {
        if (window.pybotReady) {
          setReady(true, "Ready · runs locally");
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

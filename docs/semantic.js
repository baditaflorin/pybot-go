// Small semantic router for the browser demo. MiniLM embeds text; it does not
// generate answers. All labels map to deterministic Go solvers or retrieval.
(() => {
  const MODEL = "Xenova/all-MiniLM-L6-v2";
  const REVISION = "cb3d680149bf9a3209564e1b27ab3bb355b65707";
  const TRANSFORMERS = "https://cdn.jsdelivr.net/npm/@huggingface/transformers@4.3.2";
  const intentExamples = {
    age_comparison: [
      "Who is older?", "Which person is the oldest?", "Compare their ages.",
      "Who is younger between these people?", "How many years older is one person than another?",
    ],
    age_ranking: [
      "Where does George rank by age among these people?", "Put these people in age order.",
      "Where does he sit among them by age?", "Who is in the middle by age?",
      "How does this person's age compare with everyone else?",
    ],
    knowledge_question: [
      "What is Romania?", "Tell me about Romania.", "Who was Marie Curie?",
      "Where are the Carpathian Mountains?", "Explain what photosynthesis means.",
      "Give me a factual answer about a place, person, or concept.",
    ],
    linux_expert: [
      "How do I troubleshoot a Linux server?", "How can I check disk space on Ubuntu?",
      "My systemd service failed. Which logs should I inspect?",
      "How do I find the process listening on a port?", "Help me diagnose Linux DNS and routing.",
      "Explain chmod 755 and file permissions.", "How do I safely update packages on Debian?",
      "What should I check when SSH cannot connect?",
    ],
    arithmetic: [
      "What is the square root of 2?", "Calculate 12 plus 4.",
      "What is 18 minus 6?", "Compute 8 times 7.", "Solve this math problem.",
    ],
    unit_conversion: [
      "Convert 100 centimeters to meters.", "How many inches are in 2 meters?",
      "Change 20 Celsius to Fahrenheit.", "Convert 3 kilograms to pounds.",
    ],
    date_math: [
      "What date is 30 days after January 1?", "How many days are between two dates?",
      "What day will it be two weeks from today?",
    ],
    context_followup: [
      "Repeat your last answer.", "Can you say that again?", "Show me the sources for that.",
      "What references support that answer?", "Make your last response shorter.",
      "Explain that step to me.", "Why did you recommend those commands?",
      "Is that command safe to run?", "Walk me through the previous answer.",
      "Continue with the next steps.",
    ],
  };

  let extractorPromise;
  let prototypePromise;
  let routerReady = false;

  function setStatus(message, state = "loading") {
    window.dispatchEvent(new CustomEvent("pybot-semantic-status", { detail: { message, state } }));
  }

  async function getExtractor() {
    if (!extractorPromise) {
      extractorPromise = (async () => {
        setStatus("Loading CPU language model · first use downloads about 23 MB…");
        const transformers = await import(TRANSFORMERS);
        transformers.env.allowRemoteModels = true;
        transformers.env.allowLocalModels = false;
        transformers.env.backends.onnx.wasm.numThreads = 1;
        return transformers.pipeline("feature-extraction", MODEL, {
          device: "wasm",
          dtype: "q8",
          revision: REVISION,
        });
      })().catch((error) => {
        extractorPromise = undefined;
        throw error;
      });
    }
    return extractorPromise;
  }

  async function embed(extractor, texts) {
    const result = await extractor(texts, { pooling: "mean", normalize: true });
    const rows = result.tolist();
    return rows.map((row) => row.map(Number));
  }

  async function getPrototypes(extractor) {
    if (!prototypePromise) {
      const labels = Object.keys(intentExamples);
      const texts = labels.flatMap((label) => intentExamples[label]);
      prototypePromise = embed(extractor, texts).then((vectors) => {
        const result = {};
        let index = 0;
        labels.forEach((label) => {
          const examples = intentExamples[label].map(() => vectors[index++]);
          result[label] = examples;
        });
        return result;
      }).catch((error) => {
        prototypePromise = undefined;
        throw error;
      });
    }
    return prototypePromise;
  }

  function isQuestion(text) {
    return /^(what|who|where|when|why|which|how|define|explain|tell me about)\b/i.test(text.trim());
  }

  function shouldUseContext(text) {
    const words = text.trim().split(/\s+/).length;
    return words <= 10 && /\b(it|they|them|he|she|him|her|that|those|these|there|same|again|previous|more|why|source|command|step|safe|next|what about|and if|where does)\b/i.test(text);
  }

  function cosine(vector, prototype) {
    let dot = 0;
    let a = 0;
    let b = 0;
    for (let i = 0; i < vector.length; i++) {
      dot += vector[i] * prototype[i];
      a += vector[i] * vector[i];
      b += prototype[i] * prototype[i];
    }
    return dot / (Math.sqrt(a * b) || 1);
  }

  async function classify(text, recentTurns = []) {
    if (!routerReady) return { intent: "", score: 0 };
    try {
      const extractor = await getExtractor();
      const prototypes = await getPrototypes(extractor);
      const context = shouldUseContext(text) ? recentTurns.slice(-2).join(" ") : "";
      const [vector] = await embed(extractor, [context ? `${context} ${text}` : text]);
      const scores = Object.entries(prototypes).map(([intent, examples]) => ({
        intent,
        score: Math.max(...examples.map((example) => cosine(vector, example))),
      })).sort((a, b) => b.score - a.score);
      const best = scores[0];
      const next = scores[1];
      const threshold = best.intent === "knowledge_question" ? 0.38 : 0.34;
      const margin = best.score - (next?.score ?? 0);
      const factualQuestion = isQuestion(text);
      const accepted = best.score >= threshold && (margin >= 0.015 || (best.intent === "knowledge_question" && factualQuestion));
      setStatus("CPU transformer ready · answers stay rule-based", "ready");
      return accepted ? { intent: best.intent, score: best.score } : { intent: "", score: 0 };
    } catch (error) {
      console.warn("PyBot semantic model unavailable; using local rules only.", error);
      setStatus("Rules ready · semantic model unavailable", "error");
      return { intent: "", score: 0 };
    }
  }

  window.pybotSemanticClassify = classify;
  // Warm the model without blocking the rule-based chat from becoming usable.
  getExtractor().then(async (extractor) => {
    await getPrototypes(extractor);
    routerReady = true;
    setStatus("CPU transformer ready · answers stay rule-based", "ready");
  })
    .catch(() => setStatus("Rules ready · semantic model unavailable", "error"));
})();

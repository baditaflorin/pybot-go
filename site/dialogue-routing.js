// Keep conversational questions out of web lookup unless there is a clear
// factual shape or the CPU semantic classifier identifies a knowledge query.
(() => {
  const personalPatterns = [
    /^(?:can|could|would) you (?:please )?(?:help|assist)(?: me)?(?: with (?:this|that|it))?$/i,
    /^(?:can|could|would) i help you\b/i,
    /^(?:can|could|would) i (?:be useful|be of help)(?: to you)?\b/i,
    /^(?:is there|can|could|would) (?:anything|something) i (?:can|could) help you with\b/i,
    /^would you (?:accept|choose|pick|take|wear|travel|rather|be willing to)\b/i,
    /^(?:what|which) personal (?:ability|skill) would you\b/i,
    /^(?:what|which)\b.*\bwould you (?:like|want|prefer|choose|hope|plan)\b/i,
    /^(?:do|does|did) you (?:like|love|prefer|want|think|feel|believe|remember|have|eat|drink|exercise|travel|watch|listen|read|enjoy)\b/i,
    /^have you (?:ever )?(?:seen|been|tried|done|visited|read|watched|heard|met|experienced)\b/i,
    /^have you (?:ever )?(?:attended|gone to|been to)\b/i,
    /^did you (?:go|attend|watch|see|visit|try|read|exercise|travel|eat|drink)\b/i,
    /^how often do you\b/i,
    /^are you (?:more of|a |an |available|doing anything|free|ready|interested|okay|all right|alright|happy|sad|sure|excited|married|single)\b/i,
    /^(?:what|who|how|where|when|why|which) (?:do|does|did|are|is|were|was) you (?:think|feel|like|prefer|want|believe|mean|eat|drink|enjoy|call|remember|watch|read)\b/i,
    /^(?:what|who) (?:is|are|was|were) your (?:personal |favorite |favourite )?(?:weakness|strength|opinion|preference|favorite|favourite|name|family|hobby|job|experience)/i,
    /^(?:what|which) (?:type|kind) of people do you work with\b/i,
    /^(?:which|what) (?:colleagues|coworkers|teammates) do you work best with\b/i,
    /^(?:which|what) (?:colleagues|coworkers|teammates) do you collaborate with\b/i,
    /^what (?:things|qualities) impress you\b/i,
    /^what foods do you eat\b/i,
    /^where are you from\b/i,
    /^what place did you grow up in\b/i,
    /^(?:what's|what is) your hometown\b/i,
    /^how (?:is|are) your family\b/i,
    /^what can i do for you\b/i,
    /^what do you need(?: to know)?\b/i,
    /^what information do you need from me\b/i,
    /^what details would you like from me\b/i,
    /^how can i assist you\b/i,
    /^(?:can|could|would) you (?:send|pass|deliver) .+ (?:a )?(?:note|message)\b/i,
    /^do you need (?:a guide|any help|help)\b/i,
    /^shall we (?:visit|go|meet|have|drop in|stop)\b/i,
    /^should we (?:visit|go|meet|tour|have|drop in|stop)\b/i,
    /^do you have (?:a )?(?:regular )?(?:workout|exercise) routine\b/i,
    /^do you have (?:a )?(?:regular )?(?:workout|exercise) schedule\b/i,
    /^do you (?:keep|follow|maintain) (?:a )?(?:regular )?(?:workout|exercise) schedule\b/i,
    /^can you tell (?:that|if|whether)\b/i,
    /^(?:is|are|was|were) (?:this|that|it)\b.*\b(?:okay|all right|alright|fine)\b/i,
    /^would (?:next )?(?:monday|tuesday|wednesday|thursday|friday|saturday|sunday)\b.*\bsuit you\b/i,
    /^(?:do|does|did) you have (?:a free moment|anything to do|time to|any plans)\b/i,
    /^(?:do|would) you mind\b/i,
    /^(?:would|could) you (?:pour|bring|carry|pick up|drive|visit|meet|lend|spare)\b/i,
    /^(?:would|could) you ask (?:him|her|them|someone)\b/i,
    /^(?:can|could|would) you (?:send|pass|deliver) (?:a )?message to\b/i,
    /^can you stop somewhere\b/i,
    /^(?:what|who|how|where|when|which) (?:kind|type|sort) of\b.*\b(?:would|do|does|did) you (?:like|want|prefer|plan|hope)\b/i,
    /^how long (?:will|would|do|does|did|are) you\b/i,
    /^how (?:much longer|long) do you (?:plan|intend|expect|want) to\b/i,
    /^(?:is|would|will) it (?:be )?(?:a )?(?:long|short) (?:ride|walk|drive|trip|journey)\b/i,
    /^(?:are|were) you (?:planning|going|coming|acting|being|looking|seeming)\b/i,
    /^why(?:['’]d| did| do| does| are| were) you\b/i,
    /^why aren't you\b/i,
    /^why (?:are not|won't|will not|can't|cannot) you\b/i,
    /^how do you (?:usually )?(?:spend|structure) (?:your|the) (?:day|days|weekday|weekdays|time)\b/i,
    /^how (?:much longer|long) do you (?:plan|intend|expect|want) to\b/i,
    /^(?:what|where|how) (?:does|do|did|is|are) (?:your|my|his|her|their) (?:mother|father|friend|brother|sister|manager|husband|wife|child|daughter|son|boss|neighbor)\b/i,
    /^(?:what|where|how) (?:does|do|did|is|are) (?:your|my|his|her|their) parent\b/i,
    /^what shall i call (?:him|her|them)\b/i,
    /^what name should i use for (?:him|her|them|this person|that person)\b/i,
    /^(?:what's|what is) the reason (?:you|she|he|they)\b/i,
    /^(?:what's up|what is up|how's it going|how is it going)$/i,
    /^do you know (?:what|which|where|when|who|how) .+\b(?:she|he|they|my|your)\b/i,
  ];
  const knowledgeWrapped = /^(?:do|does|did) you know (?:what|who|where|when|why|which|how)\b/i;
  const personalReference = /\b(?:she|he|they|my|your)\b/i;
  const factualOpeners = /^(?:what|who|where|when|why|which|how|define|explain|tell me about|at what)\b/i;
  const factualRequests = /^(?:can|could|would) you (?:please )?(?:tell|give|name|list|explain|show|describe|define|summarize)\b/i;
  const politeMindInformationRequest = /^(?:do|would) you mind (?:please )?(?:telling me about|explaining|showing|describing|defining|summarizing|listing|naming|calculating|comparing|converting)\b/i;
  const embeddedKnowledgeRequest = /^(?:can|could|would) you (?:please )?help me (?:understand|learn about|find out about|look up)\b/i;
  const subjectKnowledgeRequest = /^(?:can|could|would) you (?:please )?help me with (?:the )?(?:capital|population|area|currency|meaning|definition|history|location|distance|square root|weather|president|prime minister)\b/i;
  const adviceRequest = /^(?:(?:can|could|would) you (?:please )?(?:tell me|explain)\s+)?(?:(?:what is|what's)\s+)?(?:the\s+)?(?:best|easiest|safest|quickest|cheapest)\s+way\s+to\b|^how (?:should|can|do) i\b.*\b(?:find|look for|choose|compare|plan|learn|start)\b/i;
  const factualWrappers = [
    /^(?:can|could|would) you (?:please )?help me (?:understand|learn about|find out about|look up)\s+/i,
    /^(?:can|could|would) you (?:please )?(?:give me|tell me|show me)\s+(?:(?:a|an|the)\s+)?(?:(?:quick|brief|short)\s+)?(?:overview|summary|description|introduction)\s+(?:of|about)\s+/i,
    /^(?:can|could|would) you (?:please )?tell me about\s+/i,
    /^(?:please\s+)?tell me about\s+/i,
    /^(?:can|could|would) you (?:please )?(?:explain|describe|define|summarize)\s+/i,
    /^(?:please\s+)?(?:explain|describe|define|summarize)\s+/i,
    /^(?:what(?: is|'s|’s)|who(?: is|'s|’s)|where(?: is|'s|’s))\s+(?:(?:the|a|an)\s+)?/i,
  ];

  function isSocialQuestion(text) {
    const question = String(text || "").trim().replace(/[?.!]+\s*$/, "");
    if (politeMindInformationRequest.test(question)) return false;
    if (!question || (knowledgeWrapped.test(question) && !personalReference.test(question))) return false;
    return personalPatterns.some((pattern) => pattern.test(question));
  }

  function shouldLookupFactualQuestion(text, semanticIntent = "") {
    const question = String(text || "").trim();
    if (!question || isSocialQuestion(question)) return false;
    if (adviceRequest.test(question)) return false;
    if (/^(?:what|how) about\b/i.test(question) && semanticIntent !== "knowledge_question") return false;
    if (semanticIntent === "knowledge_question") return true;
    if (factualRequests.test(question) || politeMindInformationRequest.test(question) || embeddedKnowledgeRequest.test(question) || subjectKnowledgeRequest.test(question)) return true;
    if (semanticIntent === "social_question" || semanticIntent.startsWith("dialogue_")) return false;
    if (semanticIntent === "context_followup") return false;
    return factualOpeners.test(question);
  }

  function hasRelevantSource(lexicalScore, semanticScore) {
    return lexicalScore >= 0.65 || (Number.isFinite(semanticScore) && semanticScore >= 0.62);
  }

  function normalizeFactualQuestion(text) {
    let question = String(text || "").trim();
    const filler = /^(?:all right|alright|okay|ok|well|excuse me|by the way)[\s,.!?-]+/i;
    while (filler.test(question)) question = question.replace(filler, "").trim();
    const invertedQuestion = question.match(/^(?:can|could|would) you (?:please )?help me understand\s+(?:what|who|where)\s+(.+?)\s+(?:is|are|was|were)[?.!]*$/i);
    if (invertedQuestion) return invertedQuestion[1].trim();
    const mindRequest = question.match(politeMindInformationRequest);
    if (mindRequest) question = question.replace(politeMindInformationRequest, "").trim();
    const copularQuestion = question.match(/^(?:what|who|where)\s+(.+?)\s+(?:is|are|was|were)[?.!]*$/i);
    if (copularQuestion) return copularQuestion[1].trim();
    let previous;
    do {
      previous = question;
      for (const wrapper of factualWrappers) {
        if (wrapper.test(question)) {
          question = question.replace(wrapper, "").replace(/[?.!]+\s*$/, "").trim();
          break;
        }
      }
    } while (question !== previous);
    return question;
  }

  const api = { isSocialQuestion, shouldLookupFactualQuestion, hasRelevantSource, normalizeFactualQuestion };
  if (typeof window !== "undefined") window.pybotDialogueRouting = api;
  if (typeof module !== "undefined" && module.exports) module.exports = api;
})();

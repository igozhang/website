/* 德州扑克页面：交互、难度、Worker AI */
(function () {
  "use strict";

  var R = window.FsTexasRules;
  var BoardView = window.FsTexasBoard;
  if (!R || !BoardView) return;

  var LS_LEVEL = "fs_texas_level";
  var LS_OPP = "fs_texas_opponents";

  var canvas = document.getElementById("board");
  var summaryEl = document.getElementById("handSummary");
  var statusEl = document.getElementById("status");
  var btnFold = document.getElementById("btnFold");
  var btnCheckCall = document.getElementById("btnCheckCall");
  var btnRaise = document.getElementById("btnRaise");
  var btnAllIn = document.getElementById("btnAllIn");
  var btnRestart = document.getElementById("btnRestart");
  var btnNextHand = document.getElementById("btnNextHand");
  var btnEasy = document.getElementById("btnLevelEasy");
  var btnHard = document.getElementById("btnLevelHard");
  var btnOpp3 = document.getElementById("btnOpp3");
  var btnOpp4 = document.getElementById("btnOpp4");
  var btnOpp5 = document.getElementById("btnOpp5");
  var raiseMultEl = document.getElementById("raiseMult");

  var ops = (typeof window !== "undefined" && window.__fsTexasOps) || {};
  var level = localStorage.getItem(LS_LEVEL) || "easy";
  if (level !== "easy" && level !== "hard") level = "easy";

  var oppRaw = localStorage.getItem(LS_OPP);
  var oppPick = null;
  if (oppRaw === "3" || oppRaw === "4" || oppRaw === "5") oppPick = parseInt(oppRaw, 10);

  var opponents = 3;
  var view = new BoardView(canvas, { summaryEl: summaryEl });
  var game = null;
  var busy = false;
  var worker = null;
  var aiReqId = 0;
  var aiTimer = null;
  var uiAnimId = null;
  var aiPending = null;
  var aiThinkStarted = 0;
  var aiThinkLimitMs = 0;
  var aiGapTimer = null;
  var aiGen = 0;
  var fastFinishUntil = 0;
  var lastBoardLen = 0;
  var lastHandNo = 0;
  var texasAudioCtx = null;

  function getTexasAudioCtx() {
    if (texasAudioCtx) return texasAudioCtx;
    var Ctx = window.AudioContext || window.webkitAudioContext;
    if (!Ctx) return null;
    try { texasAudioCtx = new Ctx(); } catch (e) { return null; }
    return texasAudioCtx;
  }

  function unlockTexasAudio() {
    var ac = getTexasAudioCtx();
    if (ac && ac.state === "suspended") ac.resume().catch(function () {});
  }

  function playDealSound() {
    var ac = getTexasAudioCtx();
    if (!ac) return;
    if (ac.state === "suspended") {
      ac.resume().then(function () { playDealSound(); }).catch(function () {});
      return;
    }
    var t0 = ac.currentTime;
    [520, 680].forEach(function (freq, i) {
      var osc = ac.createOscillator();
      osc.type = "triangle";
      osc.frequency.setValueAtTime(freq, t0 + i * 0.04);
      var g = ac.createGain();
      g.gain.setValueAtTime(0.0001, t0 + i * 0.04);
      g.gain.exponentialRampToValueAtTime(0.07, t0 + i * 0.04 + 0.008);
      g.gain.exponentialRampToValueAtTime(0.0001, t0 + i * 0.04 + 0.07);
      osc.connect(g);
      g.connect(ac.destination);
      osc.start(t0 + i * 0.04);
      osc.stop(t0 + i * 0.04 + 0.08);
    });
  }

  function playBetSound() {
    var ac = getTexasAudioCtx();
    if (!ac) return;
    if (ac.state === "suspended") {
      ac.resume().then(function () { playBetSound(); }).catch(function () {});
      return;
    }
    var t0 = ac.currentTime;
    var osc = ac.createOscillator();
    osc.type = "sine";
    osc.frequency.setValueAtTime(920, t0);
    osc.frequency.exponentialRampToValueAtTime(520, t0 + 0.05);
    var g = ac.createGain();
    g.gain.setValueAtTime(0.12, t0);
    g.gain.exponentialRampToValueAtTime(0.0008, t0 + 0.06);
    osc.connect(g);
    g.connect(ac.destination);
    osc.start(t0);
    osc.stop(t0 + 0.07);
  }

  function texasVibrate() {
    try {
      if (typeof navigator.vibrate !== "function") return;
      if (window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
      navigator.vibrate(16);
    } catch (e) {}
  }

  function triggerFastFinish() {
    fastFinishUntil = Date.now() + 600;
  }

  function shouldFastFinish() {
    return Date.now() < fastFinishUntil;
  }

  function humanHandOver() {
    if (!game || !game.players[0]) return false;
    var h = game.players[0];
    return !!(h.folded || h.chips <= 0);
  }

  function maybeFastFinishAfterAction(idx, action) {
    if (idx !== 0 || !game) return;
    if (action === "fold" || game.players[0].chips <= 0) triggerFastFinish();
  }

  function playActionFeedback(action) {
    if (humanHandOver()) return;
    if (action === "call" || action === "raise" || action === "allin") playBetSound();
  }

  function trackDealSounds() {
    if (!game || game.street === "idle") {
      lastBoardLen = 0;
      return;
    }
    if (game.handNo !== lastHandNo) {
      lastHandNo = game.handNo;
      lastBoardLen = (game.board || []).length;
      playDealSound();
    } else if ((game.board || []).length > lastBoardLen) {
      lastBoardLen = game.board.length;
      playDealSound();
    }
  }

  function pubRoot() {
    if (typeof window.__fsAbsRoot === "function") return window.__fsAbsRoot();
    var foot = document.querySelector(".site-foot[data-base]");
    return foot ? foot.getAttribute("data-base") || "" : "";
  }

  function levelLabel() {
    return level === "easy" ? "简单" : "困难";
  }

  function oppLabel() {
    if (!game) return oppPick ? "1v" + oppPick : "随机 1v3～5";
    return "1v" + opponents;
  }

  function resolveOpponents() {
    if (oppPick >= 3 && oppPick <= 5) return oppPick;
    return 3 + Math.floor(Math.random() * 3);
  }

  function raiseMult() {
    if (!raiseMultEl) return 1;
    var v = parseInt(raiseMultEl.value, 10);
    return v >= 1 && v <= 9 ? v : 1;
  }

  function bigBlind() {
    return (game && game.opts && game.opts.bigBlind) || ops.bigBlind || 20;
  }

  function idlePublicState() {
    return {
      playerCount: 0,
      street: "idle",
      pot: 0,
      board: [],
      currentBet: 0,
      actor: -1,
      dealer: -1,
      bb: -1,
      message: "",
      showHole: false,
      revealAiHole: false,
      handSummaryLines: [],
      handSummaryFullLines: [],
      handSummaryTruncated: false,
      winners: [],
      players: []
    };
  }

  function setStatus(t) {
    if (statusEl) statusEl.textContent = t;
  }

  function idleStatusText() {
    return levelLabel() + " · " + oppLabel() + " · 点击「开始」发牌";
  }

  function updateLevelUI() {
    if (btnEasy) btnEasy.classList.toggle("primary", level === "easy");
    if (btnHard) btnHard.classList.toggle("primary", level === "hard");
  }

  function updateOppUI() {
    if (btnOpp3) btnOpp3.classList.toggle("primary", oppPick === 3);
    if (btnOpp4) btnOpp4.classList.toggle("primary", oppPick === 4);
    if (btnOpp5) btnOpp5.classList.toggle("primary", oppPick === 5);
  }

  function gameOpts(aiCount) {
    var ai = aiCount || opponents;
    return {
      playerCount: ai + 1,
      startingChips: ops.startingChips || 2000,
      smallBlind: ops.smallBlind || 10,
      bigBlind: ops.bigBlind || 20,
      maxRaisesPerRound: ops.maxRaisesPerRound || 2
    };
  }

  function publicState() {
    return R.serializePublic(game);
  }

  function needsUiAnim() {
    if (!game) return false;
    if (game.street === "hand_end" || game.street === "done") return true;
    return game.street !== "idle" && game.actor >= 0;
  }

  function stopUiAnim() {
    if (uiAnimId) {
      cancelAnimationFrame(uiAnimId);
      uiAnimId = null;
    }
    view.setAnimTime(0);
  }

  function startUiAnim() {
    if (uiAnimId) return;
    function tick(ts) {
      view.setAnimTime(ts);
      view.draw();
      uiAnimId = requestAnimationFrame(tick);
    }
    uiAnimId = requestAnimationFrame(tick);
  }

  function syncView() {
    trackDealSounds();
    view.setState(game ? publicState() : idlePublicState());
    if (needsUiAnim()) startUiAnim();
    else {
      stopUiAnim();
      view.draw();
    }
    updateButtons();
  }

  function statusText() {
    if (!game) return idleStatusText();
    var street = R.streetLabel(game.street);
    if (game.street === "hand_end") {
      return (game.message ? game.message + " · " : "") + "点击「下一手」继续";
    }
    if (game.street === "done") return game.message || "对局结束";
    if (game.message) return game.message;
    var actor = game.players[game.actor];
    if (actor && actor.isHuman) return levelLabel() + " · " + oppLabel() + " · " + street + " · 轮到你";
    if (actor) return levelLabel() + " · " + oppLabel() + " · " + street + " · " + actor.name + " 思考中…";
    return levelLabel() + " · " + oppLabel() + " · " + street;
  }

  function updateButtons() {
    var humanIdx = 0;
    var humanTurn = false;
    var legal = [];
    var handEnd = false;
    var done = false;
    var playing = !!game && game.street !== "idle";

    if (game) {
      legal = R.legalActions(game, humanIdx);
      humanTurn = game.actor === humanIdx && legal.length > 0;
      handEnd = game.street === "hand_end";
      done = game.street === "done";
    }

    if (btnFold) btnFold.disabled = !playing || !humanTurn || busy;
    if (btnAllIn) btnAllIn.disabled = !playing || !humanTurn || busy || legal.indexOf("allin") < 0;

    if (btnRaise) {
      var mult = raiseMult();
      btnRaise.textContent = "加注 " + mult * bigBlind();
      btnRaise.disabled = !playing || !humanTurn || busy || legal.indexOf("raise") < 0;
    }

    if (btnCheckCall) {
      var canCheck = legal.indexOf("check") >= 0;
      var canCall = legal.indexOf("call") >= 0;
      btnCheckCall.disabled = !playing || !humanTurn || busy || (!canCheck && !canCall);
      if (canCall) {
        btnCheckCall.textContent = "跟注 " + R.toCall(game, humanIdx);
      } else {
        btnCheckCall.textContent = "过牌";
      }
    }

    if (btnNextHand) {
      btnNextHand.hidden = !handEnd;
      btnNextHand.disabled = busy || done;
    }
    if (raiseMultEl) raiseMultEl.disabled = !playing || !humanTurn || busy || legal.indexOf("raise") < 0;

    var canChangeOpp = !game || game.street === "idle" || game.street === "hand_end" || game.street === "done";
    if (btnOpp3) btnOpp3.disabled = busy || !canChangeOpp;
    if (btnOpp4) btnOpp4.disabled = busy || !canChangeOpp;
    if (btnOpp5) btnOpp5.disabled = busy || !canChangeOpp;
  }

  function clearAiTimer() {
    if (aiTimer) { clearTimeout(aiTimer); aiTimer = null; }
  }

  function clearAiSchedulers() {
    clearAiTimer();
    if (aiGapTimer) { clearTimeout(aiGapTimer); aiGapTimer = null; }
  }

  function bumpAiGen() {
    aiGen += 1;
    return aiGen;
  }

  function scheduleAiTurn(delay) {
    if (aiGapTimer) { clearTimeout(aiGapTimer); aiGapTimer = null; }
    var gen = aiGen;
    aiGapTimer = setTimeout(function () {
      aiGapTimer = null;
      if (gen !== aiGen) return;
      runAiTurn();
    }, delay);
  }

  function clearAiPending() {
    aiPending = null;
    aiThinkStarted = 0;
    aiThinkLimitMs = 0;
  }

  function randomAiThinkMs() {
    var minMs = ops.easyThinkMs || 500;
    var maxMs = ops.hardThinkMs || 2000;
    if (maxMs < minMs) maxMs = minMs;
    return minMs + Math.floor(Math.random() * (maxMs - minMs + 1));
  }

  function beginAiThink(idx) {
    aiThinkStarted = Date.now();
    if (humanHandOver() && game.street !== "hand_end" && game.street !== "done") {
      var left = shouldFastFinish() ? fastFinishUntil - Date.now() : 600;
      aiThinkLimitMs = Math.max(15, Math.min(40, left));
    } else {
      aiThinkLimitMs = randomAiThinkMs();
    }
    aiPending = { idx: idx, action: null };
    clearAiTimer();
    aiTimer = setTimeout(onAiDeadline, aiThinkLimitMs);
  }

  function onAiDeadline() {
    if (!aiPending || !game) { busy = false; return; }
    if (game.players[aiPending.idx].isHuman) { clearAiPending(); busy = false; return; }
    var pending = aiPending;
    clearAiPending();
    applyAiAction(pending.idx, pending.action);
    busy = false;
    afterAction();
  }

  function ensureWorker() {
    if (worker) return worker;
    var rulesSrc = window.__fsTexasRulesSrc;
    var aiSrc = window.__fsTexasAiWorkerSrc;
    if (!rulesSrc || !aiSrc) return null;
    try {
      var blob = new Blob([rulesSrc + "\n" + aiSrc], { type: "application/javascript" });
      worker = new Worker(URL.createObjectURL(blob));
      worker.onmessage = onWorkerMessage;
      worker.onerror = function () {
        if (aiPending && aiPending.action == null) {
          aiPending.action = pickFallbackAction(aiPending.idx);
        }
      };
    } catch (e) { worker = null; }
    return worker;
  }

  function pickFallbackAction(idx) {
    var legal = R.legalActions(game, idx);
    if (!legal.length) return null;
    if (legal.indexOf("check") >= 0) return "check";
    if (legal.indexOf("call") >= 0) return "call";
    return legal[0];
  }

  function resolveAiAction(idx, preferred) {
    var legal = R.legalActions(game, idx);
    if (!legal.length) return null;
    if (preferred && legal.indexOf(preferred) >= 0) return preferred;
    return pickFallbackAction(idx);
  }

  function applyAiAction(idx, preferred) {
    var action = resolveAiAction(idx, preferred);
    if (!action) return false;
    if (applyActorAction(idx, action)) return true;
    var fallback = pickFallbackAction(idx);
    if (fallback && fallback !== action) return applyActorAction(idx, fallback);
    return false;
  }

  function applyActorAction(idx, action, extra) {
    if (!R.applyAction(game, idx, action, extra)) return false;
    maybeFastFinishAfterAction(idx, action);
    playActionFeedback(action);
    syncView();
    setStatus(statusText());
    return true;
  }

  function onWorkerMessage(e) {
    var d = e.data || {};
    if (d.type !== "decide" || d.id !== aiReqId) return;
    if (!aiPending || !game || game.actor !== aiPending.idx) return;
    aiPending.action = resolveAiAction(aiPending.idx, d.action);
  }

  function runAiTurn() {
    if (!game || game.street === "done" || game.street === "hand_end" || game.street === "showdown") {
      busy = false;
      updateButtons();
      return;
    }
    var idx = game.actor;
    if (idx < 0 || !game.players[idx] || game.players[idx].isHuman) {
      busy = false;
      updateButtons();
      return;
    }
    var legal = R.legalActions(game, idx);
    if (!legal.length) {
      busy = false;
      updateButtons();
      return;
    }
    busy = true;
    updateButtons();
    setStatus(statusText());
    beginAiThink(idx);
    var w = ensureWorker();
    var p = game.players[idx];

    // 计算位置信息（按存活座位相对庄家）
    var activeN = 0;
    var pi;
    for (pi = 0; pi < game.players.length; pi++) {
      if (!game.players[pi].eliminated) activeN += 1;
    }
    var posFromDealer = R.seatPositionFromDealer(game, idx);
    var isLatePosition = posFromDealer === 0 || posFromDealer === activeN - 1;
    var isEarlyPosition = (activeN >= 5 && posFromDealer >= 3 && posFromDealer <= activeN - 2)
      || (activeN === 4 && posFromDealer === 3);
    var isSmallBlind = game.sb === idx;
    var isBigBlind = game.bb === idx;
    
    var req = {
      type: "decide",
      id: ++aiReqId,
      level: level,
      hole: p.hole,
      board: game.board,
      pot: game.pot,
      toCall: R.toCall(game, idx),
      chips: p.chips,
      legal: legal,
      ops: gameOpts(),
      street: game.street,
      posFromDealer: posFromDealer,
      isEarlyPosition: isEarlyPosition,
      isLatePosition: isLatePosition,
      isSmallBlind: isSmallBlind,
      isBigBlind: isBigBlind,
      numPlayers: activeN,
      currentBet: game.currentBet,
      betThisRound: p.betThisRound
    };
    if (!w) {
      aiPending.action = pickFallbackAction(idx);
      return;
    }
    try { w.postMessage(req); }
    catch (e) {
      aiPending.action = pickFallbackAction(idx);
    }
  }

  function afterAction() {
    syncView();
    setStatus(statusText());
    if (game.street === "hand_end" || game.street === "showdown" || game.street === "done") {
      busy = false;
      updateButtons();
      return;
    }
    if (game.actor >= 0 && !game.players[game.actor].isHuman) {
      var aiGap = 100;
      if (humanHandOver()) {
        var left = shouldFastFinish() ? fastFinishUntil - Date.now() : 600;
        aiGap = Math.max(10, Math.min(25, left));
      }
      scheduleAiTurn(aiGap);
    } else {
      busy = false;
      updateButtons();
    }
  }

  function humanAction(action) {
    if (busy || !game || game.actor !== 0) return;
    unlockTexasAudio();
    texasVibrate();
    busy = true;
    updateButtons();
    var extra = action === "raise" ? { raiseMult: raiseMult() } : undefined;
    applyActorAction(0, action, extra);
    busy = false;
    afterAction();
  }

  function startGame() {
    clearAiSchedulers();
    clearAiPending();
    bumpAiGen();
    stopUiAnim();
    aiReqId++;
    fastFinishUntil = 0;
    lastBoardLen = 0;
    lastHandNo = 0;
    busy = false;
    unlockTexasAudio();
    opponents = resolveOpponents();
    game = R.createGame(gameOpts(opponents));
    game.street = "idle";
    R.startHand(game);
    syncView();
    setStatus(statusText());
    afterAction();
  }

  function nextHand() {
    if (!game || game.street !== "hand_end") return;
    clearAiSchedulers();
    clearAiPending();
    bumpAiGen();
    fastFinishUntil = 0;
    busy = false;
    unlockTexasAudio();
    R.startHand(game);
    syncView();
    setStatus(statusText());
    afterAction();
  }

  function setLevel(lv) {
    level = lv === "easy" ? "easy" : "hard";
    try { localStorage.setItem(LS_LEVEL, level); } catch (e) {}
    updateLevelUI();
    setStatus(game ? statusText() : idleStatusText());
  }

  function setOpponents(n) {
    if (busy && game && game.street !== "idle" && game.street !== "done" && game.street !== "hand_end") return;
    if (oppPick === n) {
      oppPick = null;
      try { localStorage.removeItem(LS_OPP); } catch (e) {}
    } else {
      oppPick = n;
      try { localStorage.setItem(LS_OPP, String(n)); } catch (e) {}
    }
    updateOppUI();
    if (!game) setStatus(idleStatusText());
  }

  if (btnFold) btnFold.addEventListener("click", function () { unlockTexasAudio(); humanAction("fold"); });
  if (btnCheckCall) btnCheckCall.addEventListener("click", function () {
    unlockTexasAudio();
    var legal = R.legalActions(game, 0);
    if (legal.indexOf("check") >= 0) humanAction("check");
    else humanAction("call");
  });
  if (btnRaise) btnRaise.addEventListener("click", function () { unlockTexasAudio(); humanAction("raise"); });
  if (btnAllIn) btnAllIn.addEventListener("click", function () { unlockTexasAudio(); humanAction("allin"); });
  if (btnRestart) btnRestart.addEventListener("click", function () { unlockTexasAudio(); startGame(); });
  if (btnNextHand) btnNextHand.addEventListener("click", function () { unlockTexasAudio(); nextHand(); });
  if (btnEasy) btnEasy.addEventListener("click", function () { setLevel("easy"); });
  if (btnHard) btnHard.addEventListener("click", function () { setLevel("hard"); });
  if (btnOpp3) btnOpp3.addEventListener("click", function () { setOpponents(3); });
  if (btnOpp4) btnOpp4.addEventListener("click", function () { setOpponents(4); });
  if (btnOpp5) btnOpp5.addEventListener("click", function () { setOpponents(5); });
  if (raiseMultEl) raiseMultEl.addEventListener("change", function () { updateButtons(); });

  updateLevelUI();
  updateOppUI();
  view.setPub(pubRoot());
  setStatus(idleStatusText());
  view.loadCardBack(pubRoot()).then(syncView);

  var ro = typeof ResizeObserver !== "undefined" ? new ResizeObserver(function () { view.draw(); }) : null;
  if (ro && canvas.parentElement) ro.observe(canvas.parentElement);
  window.addEventListener("resize", function () { view.draw(); });
})();

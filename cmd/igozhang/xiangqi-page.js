/* 象棋页面：交互、难度、Worker AI */
(function () {
  "use strict";

  var R = window.FsXiangqiRules;
  var BoardView = window.FsXiangqiBoard;
  if (!R || !BoardView) return;

  var LS_LEVEL = "fs_xiangqi_level";
  var HUMAN = "red";
  var AI = "black";

  var canvas = document.getElementById("board");
  var statusEl = document.getElementById("status");
  var btnUndo = document.getElementById("btnUndo");
  var btnRestart = document.getElementById("btnRestart");
  var btnEasy = document.getElementById("btnLevelEasy");
  var btnHard = document.getElementById("btnLevelHard");

  var ops = (typeof window !== "undefined" && window.__fsXiangqiOps) || {};
  var level = localStorage.getItem(LS_LEVEL) || "hard";
  if (level !== "easy" && level !== "hard") level = "hard";

  var view = new BoardView(canvas, {});
  var board = R.initialBoard();
  var history = [];
  var gameOver = false;
  var busy = false;
  var selected = null;
  var legal = [];
  var worker = null;
  var searchReqId = 0;
  var aiTimer = null;
  var lastTouchAt = 0;
  var touchStartX = 0, touchStartY = 0, touchStartT = 0, touchMoved = false;
  var TAP_MOVE_MAX = 16, TAP_TIME_MAX = 360;

  /** 由着法历史推导行棋方，避免 turn 与棋盘不同步 */
  function sideToMove() {
    return history.length % 2 === 0 ? HUMAN : AI;
  }

  function clearAiTimer() {
    if (aiTimer) {
      clearTimeout(aiTimer);
      aiTimer = null;
    }
  }

  function setStatus(t) {
    if (statusEl) statusEl.textContent = t;
  }

  function levelLabel() {
    return level === "easy" ? "简单" : "困难";
  }

  function updateLevelUI() {
    if (btnEasy) btnEasy.classList.toggle("primary", level === "easy");
    if (btnHard) btnHard.classList.toggle("primary", level === "hard");
  }

  function setLevel(lv) {
    level = lv === "easy" ? "easy" : "hard";
    try { localStorage.setItem(LS_LEVEL, level); } catch (e) {}
    updateLevelUI();
    if (!gameOver && !busy && sideToMove() === HUMAN) {
      setStatus(levelLabel() + " · 轮到红方");
    }
  }

  function syncView() {
    view.setBoard(board);
    view.setSelected(selected);
    view.setLegal(legal);
    var last = history.length ? history[history.length - 1] : null;
    view.setLastMove(last);
    view.draw();
  }

  function refreshLegal() {
    if (!selected) {
      legal = [];
      return;
    }
    var all = R.legalMoves(board, sideToMove());
    legal = [];
    for (var i = 0; i < all.length; i++) {
      if (all[i].from.r === selected.r && all[i].from.c === selected.c) legal.push(all[i]);
    }
  }

  function updateUndo() {
    if (btnUndo) btnUndo.disabled = history.length < 2 || busy || gameOver;
  }

  var stoneAudioCtx = null;
  function getStoneAudioCtx() {
    if (stoneAudioCtx) return stoneAudioCtx;
    var Ctx = window.AudioContext || window.webkitAudioContext;
    if (!Ctx) return null;
    try { stoneAudioCtx = new Ctx(); } catch (e) { return null; }
    return stoneAudioCtx;
  }
  function unlockStoneAudio() {
    var ac = getStoneAudioCtx();
    if (ac && ac.state === "suspended") ac.resume().catch(function () {});
  }
  function playStoneSound() {
    var ac = getStoneAudioCtx();
    if (!ac) return;
    if (ac.state === "suspended") {
      ac.resume().then(function () { playStoneSound(); }).catch(function () {});
      return;
    }
    var t0 = ac.currentTime;
    var dur = 0.055;
    var n = Math.max(1, Math.floor(ac.sampleRate * dur));
    var buf = ac.createBuffer(1, n, ac.sampleRate);
    var ch = buf.getChannelData(0);
    for (var i = 0; i < n; i++) {
      var env = 1 - i / n;
      ch[i] = (Math.random() * 2 - 1) * env * env;
    }
    var noise = ac.createBufferSource();
    noise.buffer = buf;
    var bp = ac.createBiquadFilter();
    bp.type = "bandpass";
    bp.frequency.value = 2800;
    var nGain = ac.createGain();
    nGain.gain.setValueAtTime(0.35, t0);
    nGain.gain.exponentialRampToValueAtTime(0.0008, t0 + dur);
    noise.connect(bp);
    bp.connect(nGain);
    nGain.connect(ac.destination);
    noise.start(t0);
    noise.stop(t0 + dur + 0.01);
  }

  function applyHistory() {
    board = R.initialBoard();
    for (var i = 0; i < history.length; i++) {
      board = R.applyMove(board, history[i]);
    }
  }

  function pushMove(m, skipValidate) {
    var who = sideToMove();
    if (!skipValidate && !R.isLegalMove(board, m, who)) return false;
    if (who === HUMAN) view.setAiHighlight(null);
    history.push(m);
    applyHistory();
    selected = null;
    legal = [];
    playStoneSound();
    syncView();
    updateUndo();
    return true;
  }

  function endGameIfNeeded(side) {
    if (!R.hasKing(board, side)) {
      gameOver = true;
      if (side === HUMAN) {
        setStatus(levelLabel() + " · 将帅被擒，红方负");
      } else {
        setStatus(levelLabel() + " · 将帅被擒，红方胜！");
      }
      return true;
    }
    if (R.isCheckmate(board, side)) {
      gameOver = true;
      if (side === HUMAN) {
        setStatus(levelLabel() + " · 将死，红方负");
      } else {
        setStatus(levelLabel() + " · 将死，红方胜！");
      }
      return true;
    }
    if (R.isStalemate(board, side)) {
      gameOver = true;
      setStatus(levelLabel() + " · 困毙，和棋");
      return true;
    }
    return false;
  }

  function releaseBusy() {
    busy = false;
    clearAiTimer();
    updateUndo();
  }

  function humanTurnStatus() {
    if (R.inCheck(board, HUMAN)) {
      setStatus(levelLabel() + " · 将军！轮到红方");
    } else {
      setStatus(levelLabel() + " · 轮到红方");
    }
  }

  function ensureWorker() {
    if (worker) return worker;
    var rulesSrc = window.__fsXiangqiRulesSrc;
    var aiSrc = window.__fsXiangqiAiWorkerSrc;
    if (!rulesSrc || !aiSrc) return null;
    try {
      var blob = new Blob([rulesSrc + "\n" + aiSrc], { type: "application/javascript" });
      worker = new Worker(URL.createObjectURL(blob));
      worker.onmessage = onWorkerMessage;
      worker.onerror = function () {
        onAiFailed("worker-error");
      };
    } catch (e) {
      worker = null;
    }
    return worker;
  }

  function cancelSearch() {
    clearAiTimer();
    if (worker) {
      searchReqId += 1;
      try {
        worker.postMessage({ type: "cancel", reqId: searchReqId });
      } catch (e) {}
    }
  }

  function aiThinkMs() {
    if (level === "easy") return ops.easyTimeMs || 700;
    return ops.hardTimeMs || 1600;
  }

  function scheduleAiTimeout(reqId) {
    clearAiTimer();
    aiTimer = setTimeout(function () {
      if (busy && searchReqId === reqId) {
        onAiFailed("timeout");
      }
    }, aiThinkMs() + 2500);
  }

  function pickFallbackMove() {
    var moves = R.legalMoves(board, AI);
    if (!moves.length) return null;
    moves.sort(function (a, b) {
      var ca = board[a.to.r][a.to.c];
      var cb = board[b.to.r][b.to.c];
      return (R.PIECE_VAL[cb] || 0) - (R.PIECE_VAL[ca] || 0);
    });
    if (level === "easy" && moves.length > 1 && Math.random() < 0.35) {
      return moves[1 + Math.floor(Math.random() * Math.min(3, moves.length - 1))];
    }
    return moves[0];
  }

  function commitAiMove(m) {
    if (!pushMove(m, true)) {
      releaseBusy();
      return;
    }
    view.setAiHighlight(m);
    releaseBusy();
    if (endGameIfNeeded(AI)) return;
    humanTurnStatus();
  }

  function applyAiMove(m, attempt) {
    attempt = attempt || 0;
    if (!m || !m.from || !m.to) {
      finishAiTurnNoMove();
      return;
    }
    if (!R.isLegalMove(board, m, AI)) {
      if (attempt >= 2) { finishAiTurnNoMove(); return; }
      var alt = pickFallbackMove();
      if (!alt) { finishAiTurnNoMove(); return; }
      applyAiMove(alt, attempt + 1);
      return;
    }
    var piece = board[m.from.r][m.from.c];
    if (!piece) { finishAiTurnNoMove(); return; }
    var move = { from: { r: m.from.r, c: m.from.c }, to: { r: m.to.r, c: m.to.c } };
    view.playMoveAnim(move, piece, function () {
      commitAiMove(move);
    });
  }

  function finishAiTurnNoMove() {
    releaseBusy();
    if (endGameIfNeeded(AI)) return;
    var m = pickFallbackMove();
    if (m) applyAiMove(m, 0);
    else endGameIfNeeded(AI);
  }

  function onAiFailed(reason) {
    if (!busy || sideToMove() !== AI) {
      releaseBusy();
      return;
    }
    if (searchReqId) { /* 作废进行中的 worker 响应 */ searchReqId += 1; }
    var m = pickFallbackMove();
    if (m) {
      setStatus(levelLabel() + " · 电脑应着（" + (reason === "timeout" ? "超时" : "备用") + "）");
      setTimeout(function () { applyAiMove(m); }, 120);
      return;
    }
    finishAiTurnNoMove();
  }

  function runAi() {
    if (gameOver || sideToMove() !== AI) {
      releaseBusy();
      return;
    }
    var moves = R.legalMoves(board, AI);
    if (!moves.length) {
      finishAiTurnNoMove();
      return;
    }
    var w = ensureWorker();
    if (!w) {
      setStatus(levelLabel() + " · 电脑思考中…");
      setTimeout(function () { applyAiMove(pickFallbackMove()); }, 280);
      return;
    }
    var reqId = ++searchReqId;
    setStatus(levelLabel() + " · 电脑思考中…");
    scheduleAiTimeout(reqId);
    try {
      w.postMessage({
        type: "search",
        reqId: reqId,
        board: R.cloneBoard(board),
        turn: AI,
        level: level,
        depth: level === "easy" ? (ops.easyDepth || 2) : (ops.hardDepth || 3),
        timeMs: aiThinkMs()
      });
    } catch (e) {
      onAiFailed("post");
    }
  }

  function onWorkerMessage(e) {
    var d = e.data || {};
    if (d.type === "error") {
      if (d.reqId === searchReqId) onAiFailed("error");
      return;
    }
    if (d.type !== "best") return;
    if (d.reqId !== searchReqId) return;
    if (!busy || sideToMove() !== AI) return;
    clearAiTimer();
    applyAiMove(d.from && d.to ? { from: d.from, to: d.to } : null);
  }

  function afterHumanMove() {
    if (endGameIfNeeded(HUMAN)) {
      releaseBusy();
      return;
    }
    if (sideToMove() !== AI) {
      releaseBusy();
      return;
    }
    busy = true;
    updateUndo();
    setTimeout(runAi, 60);
  }

  function trySelectOrMove(sq) {
    if (busy || gameOver || sideToMove() !== HUMAN) return;
    var piece = board[sq.r][sq.c];
    if (selected) {
      for (var i = 0; i < legal.length; i++) {
        var m = legal[i];
        if (m.to.r === sq.r && m.to.c === sq.c) {
          if (!pushMove(m)) return;
          busy = true;
          afterHumanMove();
          return;
        }
      }
    }
    if (piece && R.sideOf(piece) === HUMAN) {
      selected = sq;
      refreshLegal();
      syncView();
      return;
    }
    selected = null;
    legal = [];
    syncView();
  }

  function onPlay(ev) {
    unlockStoneAudio();
    var sq = view.posFromEvent(ev);
    if (!sq) return;
    trySelectOrMove(sq);
  }

  function undoPair() {
    if (history.length < 2 || busy) return;
    cancelSearch();
    view.stopAnim();
    view.setAiHighlight(null);
    history.pop();
    history.pop();
    applyHistory();
    gameOver = false;
    busy = false;
    selected = null;
    legal = [];
    syncView();
    setStatus(levelLabel() + " · 轮到红方");
    updateUndo();
  }

  function restart() {
    cancelSearch();
    view.stopAnim();
    view.setAiHighlight(null);
    board = R.initialBoard();
    history = [];
    gameOver = false;
    busy = false;
    selected = null;
    legal = [];
    syncView();
    setStatus(levelLabel() + " · 轮到红方（红先）");
    updateUndo();
    lastRelayoutW = 0;
    view.resize(true);
  }

  if (btnEasy) btnEasy.addEventListener("click", function () { setLevel("easy"); });
  if (btnHard) btnHard.addEventListener("click", function () { setLevel("hard"); });
  if (btnRestart) btnRestart.addEventListener("click", restart);
  if (btnUndo) btnUndo.addEventListener("click", undoPair);

  canvas.addEventListener("click", function (ev) {
    if (Date.now() - lastTouchAt < 450) return;
    onPlay(ev);
  });
  canvas.addEventListener("touchstart", function (ev) {
    unlockStoneAudio();
    touchMoved = false;
    if (ev.touches && ev.touches.length) {
      touchStartX = ev.touches[0].clientX;
      touchStartY = ev.touches[0].clientY;
      touchStartT = Date.now();
    }
  }, { passive: true });
  canvas.addEventListener("touchmove", function (ev) {
    if (ev.touches && ev.touches.length) {
      var dx = ev.touches[0].clientX - touchStartX;
      var dy = ev.touches[0].clientY - touchStartY;
      if (dx * dx + dy * dy > TAP_MOVE_MAX * TAP_MOVE_MAX) touchMoved = true;
    }
  }, { passive: true });
  canvas.addEventListener("touchend", function (ev) {
    var dt = Date.now() - touchStartT;
    if (!touchMoved && dt < TAP_TIME_MAX) {
      ev.preventDefault();
      lastTouchAt = Date.now();
      onPlay(ev);
    }
  }, { passive: false });

  var relayoutTimers = [];
  var relayoutDebounce = null;
  var lastRelayoutW = 0;

  function clearRelayoutTimers() {
    for (var i = 0; i < relayoutTimers.length; i++) clearTimeout(relayoutTimers[i]);
    relayoutTimers = [];
    if (relayoutDebounce) {
      clearTimeout(relayoutDebounce);
      relayoutDebounce = null;
    }
  }

  function onBoardViewportChange(force) {
    clearRelayoutTimers();
    var run = function () {
      var section = canvas.closest ? canvas.closest(".xiangqi-page") : null;
      var w = section && section.clientWidth > 0 ? section.clientWidth : window.innerWidth;
      if (!force && lastRelayoutW > 0 && Math.abs(w - lastRelayoutW) < 6) return;
      lastRelayoutW = w;
      view.resize(!!force);
    };
    if (force) {
      requestAnimationFrame(run);
      return;
    }
    if (relayoutDebounce) clearTimeout(relayoutDebounce);
    relayoutDebounce = setTimeout(function () {
      relayoutDebounce = null;
      requestAnimationFrame(run);
    }, 220);
  }

  window.addEventListener("resize", function () { onBoardViewportChange(false); });
  window.addEventListener("orientationchange", function () {
    lastRelayoutW = 0;
    onBoardViewportChange(true);
  });
  document.addEventListener("visibilitychange", function () {
    if (document.visibilityState === "visible") onBoardViewportChange(true);
  });
  window.addEventListener("pageshow", function (ev) {
    if (document.visibilityState === "visible") onBoardViewportChange(true);
    if (ev.persisted) onBoardViewportChange(true);
  });
  document.addEventListener("fs:pagevisible", function () { onBoardViewportChange(true); });
  if (window.visualViewport) {
    window.visualViewport.addEventListener("resize", function () { onBoardViewportChange(false); });
  }
  if (typeof ResizeObserver !== "undefined") {
    var roTarget = canvas.closest ? canvas.closest(".xiangqi-page") : null;
    if (!roTarget) roTarget = document.querySelector("main.shell");
    if (roTarget) {
      try {
        new ResizeObserver(function () { onBoardViewportChange(false); }).observe(roTarget);
      } catch (e) {}
    }
  }

  updateLevelUI();
  restart();
})();

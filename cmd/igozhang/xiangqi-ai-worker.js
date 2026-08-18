/* 象棋 AI Worker（须先加载 xiangqi-rules.js，挂到 FsXiangqiRules） */
(function () {
  "use strict";

  var R = self.FsXiangqiRules;
  if (!R) return;

  var tt = Object.create(null);
  var cancelReq = -1;

  var LEVEL_CFG = {
    easy: { depth: 2, timeMs: 700, blunderRate: 0.3, tacticalOnly: false, noise: 60 },
    hard: { depth: 3, timeMs: 1600, blunderRate: 0, tacticalOnly: false, noise: 0 }
  };

  function boardHash(board, turn) {
    var parts = [turn];
    for (var r = 0; r < R.RANKS; r++) parts.push(board[r].join(""));
    return parts.join("|");
  }

  function moveScore(board, m, side) {
    var cap = board[m.to.r][m.to.c];
    var s = 0;
    if (cap) s += (R.PIECE_VAL[cap] || 0) * 10;
    var nb = R.applyMove(board, m);
    if (R.inCheck(nb, R.opponent(side))) s += 80;
    return s;
  }

  function orderedMoves(board, side, moves) {
    moves.sort(function (a, b) {
      return moveScore(board, b, side) - moveScore(board, a, side);
    });
    return moves;
  }

  function searchRoot(board, turn, cfg, reqId, deadline) {
    var moves = cfg.tacticalOnly
      ? R.legalMoves(board, turn, { tacticalOnly: true })
      : R.legalMoves(board, turn);
    if (!moves.length) {
      moves = R.legalMoves(board, turn);
    }
    if (!moves.length) return null;
    orderedMoves(board, turn, moves);

    var bestScore = -1e15;
    var bestMoves = [];
    var alpha = -1e15;
    var beta = 1e15;
    var i, m, sc, nb;

    for (i = 0; i < moves.length; i++) {
      if (reqId === cancelReq || Date.now() >= deadline) break;
      m = moves[i];
      nb = R.applyMove(board, m);
      sc = -negamax(nb, R.opponent(turn), cfg.depth - 1, -beta, -alpha, cfg, reqId, deadline, 1);
      if (sc > bestScore) {
        bestScore = sc;
        bestMoves = [m];
      } else if (sc === bestScore) {
        bestMoves.push(m);
      }
      if (sc > alpha) alpha = sc;
    }

    if (!bestMoves.length) return moves[0];

    if (cfg.blunderRate > 0 && moves.length > 1 && Math.random() < cfg.blunderRate) {
      var alt = moves[Math.floor(Math.random() * moves.length)];
      if (alt && R.moveKey(alt) !== R.moveKey(bestMoves[0])) return alt;
    }
    return bestMoves[Math.floor(Math.random() * bestMoves.length)];
  }

  function negamax(board, turn, depth, alpha, beta, cfg, reqId, deadline, ply) {
    if (reqId === cancelReq || Date.now() >= deadline) return 0;

    var inChk = R.inCheck(board, turn);
    if (depth <= 0) {
      var ev = R.evaluate(board);
      return turn === "black" ? ev : -ev;
    }

    var key = boardHash(board, turn) + "#" + depth;
    if (cfg.tt && tt[key] !== undefined) return tt[key];

    var moves = R.legalMoves(board, turn);
    if (!moves.length) {
      if (inChk) return turn === "black" ? -99999 + ply : 99999 - ply;
      return 0;
    }

    orderedMoves(board, turn, moves);

    var best = -1e15;
    var i, m, nb, sc;

    for (i = 0; i < moves.length; i++) {
      if (reqId === cancelReq || Date.now() >= deadline) break;
      m = moves[i];
      nb = R.applyMove(board, m);
      sc = -negamax(nb, R.opponent(turn), depth - 1, -beta, -alpha, cfg, reqId, deadline, ply + 1);
      if (sc > best) best = sc;
      if (sc > alpha) alpha = sc;
      if (alpha >= beta) break;
    }

    if (cfg.tt) tt[key] = best;
    return best;
  }

  function runSearch(data) {
    var cfg = LEVEL_CFG[data.level] || LEVEL_CFG.hard;
    cfg = {
      depth: data.depth || cfg.depth,
      timeMs: data.timeMs || cfg.timeMs,
      blunderRate: data.level === "easy" ? cfg.blunderRate : 0,
      tacticalOnly: data.level === "easy" ? cfg.tacticalOnly : false,
      noise: data.level === "easy" ? cfg.noise : 0,
      tt: data.level === "hard"
    };
    tt = Object.create(null);
    var board = R.cloneBoard(data.board);
    var turn = data.turn || "black";
    var reqId = data.reqId || 0;
    var deadline = Date.now() + cfg.timeMs;
    var move = searchRoot(board, turn, cfg, reqId, deadline);
    return {
      type: "best",
      reqId: reqId,
      level: data.level,
      from: move ? move.from : null,
      to: move ? move.to : null
    };
  }

  self.onmessage = function (e) {
    var d = e.data || {};
    if (d.type === "cancel") {
      cancelReq = d.reqId;
      return;
    }
    if (d.type === "search") {
      cancelReq = -1;
      try {
        var result = runSearch(d);
        self.postMessage(result);
      } catch (err) {
        self.postMessage({ type: "error", reqId: d.reqId, message: String(err) });
      }
    }
  };
})();

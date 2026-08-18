/* 中国象棋规则（浏览器主线程与 Web Worker 共用） */
(function (root) {
  "use strict";

  var FILES = 9;
  var RANKS = 10;

  var PIECE_VAL = {
    K: 10000, k: 10000,
    R: 900, r: 900,
    C: 450, c: 450,
    N: 400, n: 400,
    B: 200, b: 200,
    A: 200, a: 200,
    P: 100, p: 100
  };

  var PIECE_LABEL = {
    K: "帅", A: "仕", B: "相", N: "马", R: "车", C: "炮", P: "兵",
    k: "将", a: "士", b: "象", n: "马", r: "车", c: "炮", p: "卒"
  };

  var KNIGHT_D = [
    [-2, -1], [-2, 1], [-1, -2], [-1, 2], [1, -2], [1, 2], [2, -1], [2, 1]
  ];
  var KNIGHT_LEG = [
    [-1, 0], [-1, 0], [0, -1], [0, 1], [0, -1], [0, 1], [1, 0], [1, 0]
  ];
  var BISHOP_D = [[-2, -2], [-2, 2], [2, -2], [2, 2]];
  var ADVISOR_D = [[-1, -1], [-1, 1], [1, -1], [1, 1]];
  var KING_D = [[0, 1], [0, -1], [1, 0], [-1, 0]];

  function isRed(p) {
    return p >= "A" && p <= "Z";
  }

  function isBlack(p) {
    return p >= "a" && p <= "z";
  }

  function sideOf(p) {
    if (isRed(p)) return "red";
    if (isBlack(p)) return "black";
    return "";
  }

  function opponent(side) {
    return side === "red" ? "black" : "red";
  }

  function inBounds(r, c) {
    return r >= 0 && r < RANKS && c >= 0 && c < FILES;
  }

  function inPalace(r, c, red) {
    if (c < 3 || c > 5) return false;
    return red ? r >= 7 && r <= 9 : r >= 0 && r <= 2;
  }

  function cloneBoard(b) {
    var out = new Array(RANKS);
    for (var r = 0; r < RANKS; r++) out[r] = b[r].slice();
    return out;
  }

  function initialBoard() {
    return [
      ["r", "n", "b", "a", "k", "a", "b", "n", "r"],
      ["", "", "", "", "", "", "", "", ""],
      ["", "c", "", "", "", "", "", "c", ""],
      ["p", "", "p", "", "p", "", "p", "", "p"],
      ["", "", "", "", "", "", "", "", ""],
      ["", "", "", "", "", "", "", "", ""],
      ["P", "", "P", "", "P", "", "P", "", "P"],
      ["", "C", "", "", "", "", "", "C", ""],
      ["", "", "", "", "", "", "", "", ""],
      ["R", "N", "B", "A", "K", "A", "B", "N", "R"]
    ];
  }

  function findKing(board, red) {
    var want = red ? "K" : "k";
    for (var r = 0; r < RANKS; r++) {
      for (var c = 0; c < FILES; c++) {
        if (board[r][c] === want) return { r: r, c: c };
      }
    }
    return null;
  }

  function kingsFace(board) {
    var rk = -1, rkc = -1, bk = -1;
    for (var r = 0; r < RANKS; r++) {
      for (var c = 0; c < FILES; c++) {
        if (board[r][c] === "K") { rk = r; rkc = c; }
        else if (board[r][c] === "k") { bk = r; }
      }
    }
    if (rk < 0 || bk < 0 || rkc < 0) return false;
    var bkPos = findKing(board, false);
    if (!bkPos || rkc !== bkPos.c) return false;
    var c = rkc;
    var lo = Math.min(rk, bk) + 1;
    var hi = Math.max(rk, bk) - 1;
    for (var rr = lo; rr <= hi; rr++) {
      if (board[rr][c] !== "") return false;
    }
    return true;
  }

  function addMove(list, fr, fc, tr, tc) {
    list.push({ from: { r: fr, c: fc }, to: { r: tr, c: tc } });
  }

  function canCapture(piece, target) {
    if (!target) return false;
    return isRed(piece) ? isBlack(target) : isRed(target);
  }

  function rookSlides(board, r, c, piece, list, captureOnly) {
    var red = isRed(piece);
    var dirs = [[0, 1], [0, -1], [1, 0], [-1, 0]];
    for (var d = 0; d < dirs.length; d++) {
      var dr = dirs[d][0], dc = dirs[d][1];
      var nr = r + dr, nc = c + dc;
      while (inBounds(nr, nc)) {
        var t = board[nr][nc];
        if (t === "") {
          if (!captureOnly) addMove(list, r, c, nr, nc);
        } else {
          if (canCapture(piece, t)) addMove(list, r, c, nr, nc);
          break;
        }
        nr += dr;
        nc += dc;
      }
    }
  }

  function cannonSlides(board, r, c, piece, list) {
    var dirs = [[0, 1], [0, -1], [1, 0], [-1, 0]];
    for (var d = 0; d < dirs.length; d++) {
      var dr = dirs[d][0], dc = dirs[d][1];
      var nr = r + dr, nc = c + dc;
      var jumped = false;
      while (inBounds(nr, nc)) {
        var t = board[nr][nc];
        if (!jumped) {
          if (t === "") addMove(list, r, c, nr, nc);
          else jumped = true;
        } else {
          if (t !== "") {
            if (canCapture(piece, t)) addMove(list, r, c, nr, nc);
            break;
          }
        }
        nr += dr;
        nc += dc;
      }
    }
  }

  function pieceMoves(board, r, c, piece, list) {
    var red = isRed(piece);
    var t, nr, nc, i, legR, legC, eyeR, eyeC;

    switch (piece.toUpperCase()) {
      case "R":
        rookSlides(board, r, c, piece, list, false);
        break;
      case "C":
        cannonSlides(board, r, c, piece, list);
        break;
      case "N":
        for (i = 0; i < KNIGHT_D.length; i++) {
          legR = r + KNIGHT_LEG[i][0];
          legC = c + KNIGHT_LEG[i][1];
          if (!inBounds(legR, legC) || board[legR][legC] !== "") continue;
          nr = r + KNIGHT_D[i][0];
          nc = c + KNIGHT_D[i][1];
          if (!inBounds(nr, nc)) continue;
          t = board[nr][nc];
          if (t !== "" && !canCapture(piece, t)) continue;
          addMove(list, r, c, nr, nc);
        }
        break;
      case "B":
        for (i = 0; i < BISHOP_D.length; i++) {
          eyeR = r + BISHOP_D[i][0] / 2;
          eyeC = c + BISHOP_D[i][1] / 2;
          nr = r + BISHOP_D[i][0];
          nc = c + BISHOP_D[i][1];
          if (!inBounds(nr, nc)) continue;
          if (red && nr < 5) continue;
          if (!red && nr > 4) continue;
          if (board[eyeR][eyeC] !== "") continue;
          t = board[nr][nc];
          if (t !== "" && !canCapture(piece, t)) continue;
          addMove(list, r, c, nr, nc);
        }
        break;
      case "A":
        for (i = 0; i < ADVISOR_D.length; i++) {
          nr = r + ADVISOR_D[i][0];
          nc = c + ADVISOR_D[i][1];
          if (!inPalace(nr, nc, red)) continue;
          t = board[nr][nc];
          if (t !== "" && !canCapture(piece, t)) continue;
          addMove(list, r, c, nr, nc);
        }
        break;
      case "K":
        for (i = 0; i < KING_D.length; i++) {
          nr = r + KING_D[i][0];
          nc = c + KING_D[i][1];
          if (!inPalace(nr, nc, red)) continue;
          t = board[nr][nc];
          if (t !== "" && !canCapture(piece, t)) continue;
          addMove(list, r, c, nr, nc);
        }
        break;
      case "P":
        if (red) {
          nr = r - 1;
          if (inBounds(nr, c)) {
            t = board[nr][c];
            if (t !== "" && !canCapture(piece, t)) { /* blocked */ } else addMove(list, r, c, nr, c);
          }
          if (r <= 4) {
            if (inBounds(r, c - 1)) {
              t = board[r][c - 1];
              if (t === "" || canCapture(piece, t)) addMove(list, r, c, r, c - 1);
            }
            if (inBounds(r, c + 1)) {
              t = board[r][c + 1];
              if (t === "" || canCapture(piece, t)) addMove(list, r, c, r, c + 1);
            }
          }
        } else {
          nr = r + 1;
          if (inBounds(nr, c)) {
            t = board[nr][c];
            if (t !== "" && !canCapture(piece, t)) { /* blocked */ } else addMove(list, r, c, nr, c);
          }
          if (r >= 5) {
            if (inBounds(r, c - 1)) {
              t = board[r][c - 1];
              if (t === "" || canCapture(piece, t)) addMove(list, r, c, r, c - 1);
            }
            if (inBounds(r, c + 1)) {
              t = board[r][c + 1];
              if (t === "" || canCapture(piece, t)) addMove(list, r, c, r, c + 1);
            }
          }
        }
        break;
    }
  }

  function applyMove(board, move) {
    var nb = cloneBoard(board);
    var p = nb[move.from.r][move.from.c];
    if (!p) return nb;
    nb[move.to.r][move.to.c] = p;
    nb[move.from.r][move.from.c] = "";
    return nb;
  }

  function isLegalMove(board, move, side) {
    if (!move || !move.from || !move.to) return false;
    var fr = move.from.r, fc = move.from.c, tr = move.to.r, tc = move.to.c;
    if (!inBounds(fr, fc) || !inBounds(tr, tc)) return false;
    var piece = board[fr][fc];
    if (!piece || sideOf(piece) !== side) return false;
    var all = legalMoves(board, side);
    for (var i = 0; i < all.length; i++) {
      var m = all[i];
      if (m.from.r === fr && m.from.c === fc && m.to.r === tr && m.to.c === tc) return true;
    }
    return false;
  }

  function squareAttacked(board, tr, tc, byRed) {
    var r, c, piece, pseudo = [];
    for (r = 0; r < RANKS; r++) {
      for (c = 0; c < FILES; c++) {
        piece = board[r][c];
        if (!piece) continue;
        if (isRed(piece) !== byRed) continue;
        pseudo.length = 0;
        pieceMoves(board, r, c, piece, pseudo);
        for (var i = 0; i < pseudo.length; i++) {
          if (pseudo[i].to.r === tr && pseudo[i].to.c === tc) return true;
        }
      }
    }
    return false;
  }

  function inCheck(board, side) {
    var king = findKing(board, side === "red");
    if (!king) return true;
    return squareAttacked(board, king.r, king.c, side !== "red");
  }

  function pseudoLegalForSide(board, side, opts) {
    opts = opts || {};
    var list = [];
    var r, c, piece, moves;
    for (r = 0; r < RANKS; r++) {
      for (c = 0; c < FILES; c++) {
        piece = board[r][c];
        if (!piece) continue;
        if (sideOf(piece) !== side) continue;
        moves = [];
        pieceMoves(board, r, c, piece, moves);
        for (var i = 0; i < moves.length; i++) list.push(moves[i]);
      }
    }
    if (opts.tacticalOnly) {
      var filtered = [];
      var wasInCheck = inCheck(board, side);
      for (var j = 0; j < list.length; j++) {
        var m = list[j];
        var cap = board[m.to.r][m.to.c] !== "";
        var nb = applyMove(board, m);
        var givesCheck = inCheck(nb, opponent(side));
        var escapes = wasInCheck && !inCheck(nb, side);
        if (cap || givesCheck || escapes) filtered.push(m);
      }
      return filtered;
    }
    return list;
  }

  function legalMoves(board, side, opts) {
    var pseudo = pseudoLegalForSide(board, side, opts);
    var legal = [];
    for (var i = 0; i < pseudo.length; i++) {
      var m = pseudo[i];
      var target = board[m.to.r][m.to.c];
      if (target === "K" || target === "k") continue;
      var nb = applyMove(board, m);
      if (kingsFace(nb)) continue;
      if (inCheck(nb, side)) continue;
      legal.push(m);
    }
    return legal;
  }

  function evaluate(board) {
    var score = 0;
    for (var r = 0; r < RANKS; r++) {
      for (var c = 0; c < FILES; c++) {
        var p = board[r][c];
        if (!p) continue;
        var v = PIECE_VAL[p] || 0;
        if (isBlack(p)) score += v;
        else score -= v;
      }
    }
    return score;
  }

  function moveKey(m) {
    return m.from.r + "," + m.from.c + ">" + m.to.r + "," + m.to.c;
  }

  function hasKing(board, side) {
    return findKing(board, side === "red") !== null;
  }

  function isCheckmate(board, side) {
    if (!hasKing(board, side)) return true;
    return inCheck(board, side) && legalMoves(board, side).length === 0;
  }

  function isStalemate(board, side) {
    if (!hasKing(board, side)) return false;
    return !inCheck(board, side) && legalMoves(board, side).length === 0;
  }

  var api = {
    FILES: FILES,
    RANKS: RANKS,
    PIECE_VAL: PIECE_VAL,
    PIECE_LABEL: PIECE_LABEL,
    isRed: isRed,
    isBlack: isBlack,
    sideOf: sideOf,
    opponent: opponent,
    inBounds: inBounds,
    inPalace: inPalace,
    cloneBoard: cloneBoard,
    initialBoard: initialBoard,
    findKing: findKing,
    kingsFace: kingsFace,
    applyMove: applyMove,
    isLegalMove: isLegalMove,
    legalMoves: legalMoves,
    inCheck: inCheck,
    evaluate: evaluate,
    moveKey: moveKey,
    isCheckmate: isCheckmate,
    isStalemate: isStalemate,
    hasKing: hasKing,
    pieceMoves: pieceMoves
  };

  root.FsXiangqiRules = api;
})(typeof self !== "undefined" ? self : this);

/* 德州扑克 AI Worker（须先加载 texas-rules.js，挂到 FsTexasRules） */
(function () {
  "use strict";

  var R = self.FsTexasRules;
  if (!R) return;

  // ==========================================
  // 工具函数
  // ==========================================

  function isPremiumPair(r) { return r >= 12; } // AA, KK, QQ
  function isStrongPair(r) { return r >= 10; }  // AA-KK-JJ
  function isMidPair(r) { return r >= 7 && r <= 9; } // 77-99
  function isSuited(c1, c2) { return c1.s === c2.s; }
  function isConnected(r1, r2) { var d = Math.abs(r1 - r2); return d === 1 || (d === 12 && Math.max(r1, r2) === 14); }
  function highCard(r) { return r >= 11; } // J+

  // 翻牌前起手牌评分（0-100）
  function preflopHandScore(hole, isEarlyPosition, level) {
    if (!hole || hole.length < 2) return 0;
    var r1 = hole[0].r, r2 = hole[1].r;
    var high = Math.max(r1, r2), low = Math.min(r1, r2);
    var pair = r1 === r2;
    var suited = isSuited(hole[0], hole[1]);
    var connected = isConnected(r1, r2);

    var score = 0;

    // 对子
    if (pair) {
      if (isPremiumPair(high)) score = 95;
      else if (isStrongPair(high)) score = 85;
      else if (isMidPair(high)) score = 70;
      else score = 50 + (high - 2) * 3;
    }
    // 非对子
    else {
      // 高牌
      if (high >= 14) score = 70; // Ax
      else if (high >= 13) score = 60; // Kx
      else if (high >= 12) score = 50; // Qx
      else if (high >= 11) score = 40; // Jx
      else score = 20 + (high - 2) * 2;

      // 同花加成
      if (suited) score += 12;
      // 连张加成
      if (connected) score += 10;
      // 同花连张加成
      if (suited && connected) score += 8;
    }

    // 位置惩罚/奖励
    if (isEarlyPosition) score -= 15;

    // 困难模式更激进一点
    if (level === "hard") score += 5;

    return Math.max(0, Math.min(100, score));
  }

  // 分析牌面 texture
  function analyzeBoard(board) {
    if (!board || board.length === 0) return { isDry: true, isDrawHeavy: false, hasPair: false, flushDraw: false, straightDraw: false };

    var ranks = board.map(function (c) { return c.r; });
    var suits = board.map(function (c) { return c.s; });

    // 成对牌面
    var rankCount = {};
    var hasPair = false;
    for (var i = 0; i < ranks.length; i++) {
      rankCount[ranks[i]] = (rankCount[ranks[i]] || 0) + 1;
      if (rankCount[ranks[i]] >= 2) hasPair = true;
    }

    // 同花听牌
    var suitCount = {};
    var flushDraw = false;
    for (var j = 0; j < suits.length; j++) {
      suitCount[suits[j]] = (suitCount[suits[j]] || 0) + 1;
      if (suitCount[suits[j]] >= 3) flushDraw = true;
    }

    // 顺子听牌（简化）
    var straightDraw = false;
    var sorted = ranks.slice().sort(function (a, b) { return b - a; });
    for (var k = 0; k < sorted.length - 1; k++) {
      if (sorted[k] - sorted[k + 1] <= 2) { straightDraw = true; break; }
    }

    var isDry = !flushDraw && !straightDraw && !hasPair;
    var isDrawHeavy = flushDraw || straightDraw;

    return { isDry: isDry, isDrawHeavy: isDrawHeavy, hasPair: hasPair, flushDraw: flushDraw, straightDraw: straightDraw };
  }

  // ==========================================
  // 简单模式策略（有经验的休闲玩家）
  // ==========================================

  function pickEasyAction(req, ctx) {
    var hole = req.hole, board = req.board, legal = req.legal;
    var strength = ctx.strength, pot = req.pot, toCall = req.toCall, chips = req.chips;
    var bb = (req.ops && req.ops.bigBlind) || 20;
    var street = ctx.street, isPreflop = street === "preflop";
    var isEarlyPosition = ctx.isEarlyPosition;

    // 翻牌前：用起手牌表
    if (isPreflop) {
      var pfScore = preflopHandScore(hole, isEarlyPosition, "easy");
      if (pfScore < 40) {
        if (toCall > bb) return "fold";
        if (legal.indexOf("check") >= 0) return "check";
        if (legal.indexOf("call") >= 0) return "call";
      }
      if (pfScore >= 80 && legal.indexOf("raise") >= 0 && Math.random() < 0.6) return "raise";
    }

    // 翻牌后
    if (legal.indexOf("fold") >= 0 && toCall > 0) {
      if (strength < 0.22 && toCall > bb) return "fold";
      if (strength < 0.35 && Math.random() < 0.4) return "fold";
    }

    // 强牌加注
    if (strength > 0.75 && legal.indexOf("raise") >= 0 && Math.random() < 0.5) return "raise";
    if (strength > 0.88 && legal.indexOf("allin") >= 0 && Math.random() < 0.15) return "allin";

    // 默认：过牌或跟注
    if (toCall === 0 && legal.indexOf("check") >= 0) return "check";
    if (legal.indexOf("call") >= 0) return "call";
    if (legal.indexOf("check") >= 0) return "check";
    if (legal.indexOf("fold") >= 0) return "fold";
    return legal[0];
  }

  // ==========================================
  // 困难模式策略（狡猾的常规玩家）
  // ==========================================

  function pickHardAction(req, ctx) {
    var hole = req.hole, board = req.board, legal = req.legal;
    var strength = ctx.strength, pot = req.pot, toCall = req.toCall, chips = req.chips;
    var bb = (req.ops && req.ops.bigBlind) || 20;
    var street = ctx.street, isPreflop = street === "preflop";
    var isEarlyPosition = ctx.isEarlyPosition, isLatePosition = ctx.isLatePosition;
    var boardTex = ctx.boardTex;
    var potOdds = toCall > 0 ? Math.min(toCall, chips) / (pot + Math.min(toCall, chips)) : 0;

    // 翻牌前
    if (isPreflop) {
      var pfScore = preflopHandScore(hole, isEarlyPosition, "hard");

      // 偷盲（后位）
      if (isLatePosition && toCall <= bb * 2 && legal.indexOf("raise") >= 0 && Math.random() < 0.25) {
        return "raise";
      }

      // 优质牌3bet
      if (pfScore >= 85 && legal.indexOf("raise") >= 0 && Math.random() < 0.7) return "raise";
      if (pfScore >= 70 && legal.indexOf("raise") >= 0 && Math.random() < 0.4) return "raise";

      // 太弱就弃
      if (pfScore < 35 && toCall > bb) return "fold";
    }

    // 翻牌后：基于牌面 texture
    if (!isPreflop) {
      // 诈唬（干燥牌面+没位置+弱牌）
      if (boardTex.isDry && strength < 0.35 && toCall === 0 && legal.indexOf("raise") >= 0 && Math.random() < 0.18) {
        return "raise";
      }

      // 半诈唬（听牌面+强听牌）
      if (boardTex.isDrawHeavy && strength > 0.35 && strength < 0.6 && legal.indexOf("raise") >= 0 && Math.random() < 0.25) {
        return "raise";
      }
    }

    // 弃牌判断（基于底池赔率）
    if (legal.indexOf("fold") >= 0 && toCall > 0) {
      if (strength < potOdds * 0.9 && toCall > bb / 2) return "fold";
      if (strength < 0.25 && toCall >= bb * 2) return "fold";
    }

    // 强牌动作
    if (strength > 0.8) {
      if (legal.indexOf("allin") >= 0 && Math.random() < 0.25) return "allin";
      if (legal.indexOf("raise") >= 0 && Math.random() < 0.75) return "raise";
    } else if (strength > 0.65 && legal.indexOf("raise") >= 0 && Math.random() < 0.5) {
      return "raise";
    } else if (strength > 0.45 && strength < 0.6 && legal.indexOf("raise") >= 0 && Math.random() < 0.15) {
      // 轻量级加注
      return "raise";
    }

    // 默认
    if (toCall === 0 && legal.indexOf("check") >= 0) return "check";
    if (legal.indexOf("call") >= 0) return "call";
    if (legal.indexOf("check") >= 0) return "check";
    if (legal.indexOf("fold") >= 0) return "fold";
    return legal[0];
  }

  // ==========================================
  // 主入口
  // ==========================================

  function pickAction(req) {
    var level = req.level === "easy" ? "easy" : "hard";
    var hole = req.hole || [];
    var board = req.board || [];
    var legal = req.legal || [];
    if (!legal.length) return "check";

    // 使用真实的上下文信息
    var street = req.street || (board.length === 0 ? "preflop" : (board.length === 3 ? "flop" : (board.length === 4 ? "turn" : "river")));
    var samples = level === "easy" ? 60 : 150;
    var strength = R.estimateStrength(hole, board, samples);
    var isEarlyPosition = !!req.isEarlyPosition;
    var isLatePosition = !!req.isLatePosition;
    var isSmallBlind = !!req.isSmallBlind;
    var isBigBlind = !!req.isBigBlind;
    var boardTex = analyzeBoard(board);

    var ctx = {
      strength: strength,
      street: street,
      isEarlyPosition: isEarlyPosition,
      isLatePosition: isLatePosition,
      isSmallBlind: isSmallBlind,
      isBigBlind: isBigBlind,
      boardTex: boardTex
    };

    if (level === "easy") {
      return pickEasyAction(req, ctx);
    } else {
      return pickHardAction(req, ctx);
    }
  }

  self.onmessage = function (ev) {
    var req = ev.data || {};
    if (req.type === "decide") {
      self.postMessage({ type: "decide", id: req.id, action: pickAction(req) });
    }
  };
})();

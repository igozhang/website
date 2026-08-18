/* 德州扑克简化规则（主线程与 Web Worker 共用） */
(function (root) {
  "use strict";

  var SUITS = ["♠", "♥", "♦", "♣"];
  var SUIT_COLORS = { 0: "#1a2a3a", 1: "#a03030", 2: "#a03030", 3: "#1a2a3a" };
  var RANK_CHARS = ["", "", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"];

  var HAND_NAMES = [
    "高牌", "一对", "两对", "三条", "顺子", "同花", "葫芦", "四条", "同花顺", "皇家同花顺"
  ];

  var STREETS = ["preflop", "flop", "turn", "river", "showdown"];

  function cloneCard(c) {
    return { r: c.r, s: c.s };
  }

  function makeDeck() {
    var d = [];
    for (var s = 0; s < 4; s++) {
      for (var r = 2; r <= 14; r++) d.push({ r: r, s: s });
    }
    return d;
  }

  function shuffle(arr) {
    var a = arr.slice();
    for (var i = a.length - 1; i > 0; i--) {
      var j = Math.floor(Math.random() * (i + 1));
      var t = a[i];
      a[i] = a[j];
      a[j] = t;
    }
    return a;
  }

  function cardLabel(c) {
    return RANK_CHARS[c.r] + SUITS[c.s];
  }

  function cardKey(c) {
    return c.r + "-" + c.s;
  }

  function combinations(cards, k) {
    var out = [];
    function pick(start, combo) {
      if (combo.length === k) {
        out.push(combo.slice());
        return;
      }
      for (var i = start; i <= cards.length - (k - combo.length); i++) {
        combo.push(cards[i]);
        pick(i + 1, combo);
        combo.pop();
      }
    }
    pick(0, []);
    return out;
  }

  function rankCounts(cards) {
    var m = {};
    for (var i = 0; i < cards.length; i++) {
      var r = cards[i].r;
      m[r] = (m[r] || 0) + 1;
    }
    return m;
  }

  function sortedRanksDesc(cards) {
    var rs = cards.map(function (c) { return c.r; }).sort(function (a, b) { return b - a; });
    return rs;
  }

  function isStraight(ranks) {
    var uniq = [];
    for (var i = 0; i < ranks.length; i++) {
      if (uniq.indexOf(ranks[i]) < 0) uniq.push(ranks[i]);
    }
    uniq.sort(function (a, b) { return b - a; });
    if (uniq.length < 5) return null;
    for (var start = 0; start <= uniq.length - 5; start++) {
      var ok = true;
      for (var j = 1; j < 5; j++) {
        if (uniq[start + j] !== uniq[start] - j) {
          ok = false;
          break;
        }
      }
      if (ok) return uniq[start];
    }
    if (uniq.indexOf(14) >= 0 && uniq.indexOf(5) >= 0 && uniq.indexOf(4) >= 0 && uniq.indexOf(3) >= 0 && uniq.indexOf(2) >= 0) {
      return 5;
    }
    return null;
  }

  function evaluateFive(cards) {
    var rs = sortedRanksDesc(cards);
    var counts = rankCounts(cards);
    var entries = [];
    for (var k in counts) entries.push({ r: +k, n: counts[k] });
    entries.sort(function (a, b) {
      if (b.n !== a.n) return b.n - a.n;
      return b.r - a.r;
    });
    var flush = true;
    var s0 = cards[0].s;
    for (var i = 1; i < cards.length; i++) {
      if (cards[i].s !== s0) { flush = false; break; }
    }
    var stHigh = isStraight(rs);
    if (flush && stHigh) {
      var royal = stHigh === 14 && rs.indexOf(14) >= 0;
      return { cat: royal ? 9 : 8, tie: [stHigh] };
    }
    if (entries[0].n === 4) {
      return { cat: 7, tie: [entries[0].r, entries[1].r] };
    }
    if (entries[0].n === 3 && entries[1].n === 2) {
      return { cat: 6, tie: [entries[0].r, entries[1].r] };
    }
    if (flush) {
      return { cat: 5, tie: rs.slice(0, 5) };
    }
    if (stHigh) {
      return { cat: 4, tie: [stHigh] };
    }
    if (entries[0].n === 3) {
      var kick = [];
      for (var j = 0; j < entries.length; j++) {
        if (entries[j].n !== 3) kick.push(entries[j].r);
      }
      return { cat: 3, tie: [entries[0].r].concat(kick.slice(0, 2)) };
    }
    if (entries[0].n === 2 && entries[1].n === 2) {
      var hi = Math.max(entries[0].r, entries[1].r);
      var lo = Math.min(entries[0].r, entries[1].r);
      var kicker = 0;
      for (var x = 0; x < entries.length; x++) {
        if (entries[x].n === 1) kicker = entries[x].r;
      }
      return { cat: 2, tie: [hi, lo, kicker] };
    }
    if (entries[0].n === 2) {
      var ks = [];
      for (var y = 0; y < entries.length; y++) {
        if (entries[y].n === 1) ks.push(entries[y].r);
      }
      return { cat: 1, tie: [entries[0].r].concat(ks.slice(0, 3)) };
    }
    return { cat: 0, tie: rs.slice(0, 5) };
  }

  function compareEval(a, b) {
    if (a.cat !== b.cat) return a.cat > b.cat ? 1 : -1;
    var n = Math.max(a.tie.length, b.tie.length);
    for (var i = 0; i < n; i++) {
      var av = a.tie[i] || 0;
      var bv = b.tie[i] || 0;
      if (av !== bv) return av > bv ? 1 : -1;
    }
    return 0;
  }

  function enrichEval(ev, cards) {
    return {
      cat: ev.cat,
      tie: ev.tie.slice(),
      label: HAND_NAMES[ev.cat] || "高牌",
      cards: cards.map(cloneCard),
      cardsLabel: cards.map(cardLabel).join(" ")
    };
  }

  function bestHand(hole, board) {
    var all = hole.slice().concat(board);
    if (all.length < 5) {
      return enrichEval({ cat: 0, tie: sortedRanksDesc(all).slice(0, 5) }, all.slice());
    }
    var combos = combinations(all, 5);
    var best = null;
    var bestCombo = null;
    for (var i = 0; i < combos.length; i++) {
      var ev = evaluateFive(combos[i]);
      if (!best || compareEval(ev, best) > 0) {
        best = ev;
        bestCombo = combos[i];
      }
    }
    return enrichEval(best, bestCombo);
  }

  function finalizeHandReport(state, report) {
    state.handReport = report;
    var S = root.FsTexasSummary;
    if (S) {
      var packed = S.formatSummaryLines(report);
      state.handSummaryLines = packed.display;
      state.handSummaryFullLines = packed.full;
      state.handSummaryTruncated = packed.truncated;
    } else {
      state.handSummaryLines = [];
      state.handSummaryFullLines = [];
      state.handSummaryTruncated = false;
    }
  }

  function playerEvalEntry(state, pid, ev) {
    return {
      id: pid,
      name: state.players[pid].name,
      eval: ev,
      handLabel: ev.label,
      cardsLabel: ev.cardsLabel
    };
  }

  function potLabelFor(index, total) {
    if (total <= 1) return "底池";
    if (index === 0) return "主池";
    return "边池" + index;
  }

  function buildFoldReport(state, winnerIdx, potAmount) {
    var foldedNames = [];
    for (var i = 0; i < state.players.length; i++) {
      if (state.players[i].folded && !state.players[i].eliminated && i !== winnerIdx) {
        foldedNames.push(state.players[i].name);
      }
    }
    return {
      reason: "fold",
      winner: { id: winnerIdx, name: state.players[winnerIdx].name, amount: potAmount },
      foldedPlayers: foldedNames,
      board: state.board.map(cardLabel),
      street: state.street
    };
  }

  function collectAllInNames(state) {
    var names = [];
    for (var i = 0; i < state.players.length; i++) {
      var p = state.players[i];
      if (p.allIn && !p.folded && !p.eliminated) names.push(p.name);
    }
    return names;
  }

  var AI_NAMES = [
    "James", "Emma", "Liam", "Olivia", "Noah", "Ava", "Ethan", "Sophia",
    "Mason", "Isabella", "Lucas", "Mia", "Oliver", "Charlotte", "Elijah",
    "Amelia", "Aiden", "Harper", "Jackson", "Evelyn", "Logan", "Abigail",
    "Carter", "Emily", "Jayden", "Elizabeth", "Sebastian", "Sofia", "Henry", "Avery"
  ];

  function pickAiNames(count) {
    var pool = AI_NAMES.slice();
    for (var i = pool.length - 1; i > 0; i--) {
      var j = Math.floor(Math.random() * (i + 1));
      var tmp = pool[i];
      pool[i] = pool[j];
      pool[j] = tmp;
    }
    return pool.slice(0, count);
  }

  function defaultOpts(ops) {
    ops = ops || {};
    var n = ops.playerCount || 4;
    if (n < 4) n = 4;
    if (n > 6) n = 6;
    return {
      playerCount: n,
      startingChips: ops.startingChips || 2000,
      smallBlind: ops.smallBlind || 10,
      bigBlind: ops.bigBlind || 20,
      maxRaisesPerRound: ops.maxRaisesPerRound != null ? ops.maxRaisesPerRound : 2
    };
  }

  function makePlayer(i, chips, name) {
    return {
      id: i,
      name: name || (i === 0 ? "你" : "Player"),
      isHuman: i === 0,
      chips: chips,
      hole: [],
      folded: false,
      allIn: false,
      eliminated: false,
      betThisRound: 0,
      betTotal: 0,
      acted: false,
      lastAction: null,
      lastActionAmt: 0
    };
  }

  function recordAction(p, kind, amt) {
    p.lastAction = kind;
    p.lastActionAmt = amt || 0;
  }

  function createGame(opts) {
    opts = defaultOpts(opts);
    var n = opts.playerCount;
    var aiNames = pickAiNames(n - 1);
    var players = [];
    for (var i = 0; i < n; i++) {
      players.push(makePlayer(i, opts.startingChips, i === 0 ? "你" : aiNames[i - 1]));
    }
    return {
      opts: opts,
      players: players,
      deck: [],
      board: [],
      pot: 0,
      street: "idle",
      dealer: 0,
      actor: 0,
      currentBet: 0,
      raisesThisRound: 0,
      lastAggressor: -1,
      handNo: 0,
      message: "",
      winners: [],
      showHole: false,
      handReport: null,
      handSummaryLines: [],
      handSummaryFullLines: [],
      handSummaryTruncated: false
    };
  }

  function notFoldedPlayers(state) {
    var out = [];
    for (var i = 0; i < state.players.length; i++) {
      if (!state.players[i].folded && !state.players[i].eliminated) out.push(i);
    }
    return out;
  }

  function nextSeat(state, from) {
    var n = state.players.length;
    for (var step = 1; step <= n; step++) {
      var idx = (from + step) % n;
      var p = state.players[idx];
      if (!p.eliminated && !p.folded) return idx;
    }
    return from;
  }

  function nextActorSeat(state, from) {
    var n = state.players.length;
    for (var step = 1; step <= n; step++) {
      var idx = (from + step) % n;
      var p = state.players[idx];
      if (!p.eliminated && !p.folded && !p.allIn) return idx;
    }
    return -1;
  }

  function resetRoundBets(state) {
    state.currentBet = 0;
    state.raisesThisRound = 0;
    state.lastAggressor = -1;
    for (var i = 0; i < state.players.length; i++) {
      state.players[i].betThisRound = 0;
      state.players[i].acted = false;
      state.players[i].lastAction = null;
      state.players[i].lastActionAmt = 0;
    }
  }

  function postBlind(state, idx, amount, blindKind) {
    var p = state.players[idx];
    var pay = Math.min(amount, p.chips);
    p.chips -= pay;
    p.betThisRound += pay;
    p.betTotal += pay;
    state.pot += pay;
    if (p.chips === 0) p.allIn = true;
    if (pay >= state.currentBet) state.currentBet = pay;
    recordAction(p, blindKind, pay);
  }

  function startHand(state) {
    state.handNo += 1;
    state.deck = shuffle(makeDeck());
    state.board = [];
    state.pot = 0;
    state.street = "preflop";
    state.winners = [];
    state.showHole = false;
    state.message = "";
    state.handReport = null;
    state.handSummaryLines = [];
    state.handSummaryFullLines = [];
    state.handSummaryTruncated = false;
    resetRoundBets(state);

    for (var i = 0; i < state.players.length; i++) {
      var p = state.players[i];
      p.hole = [];
      p.folded = p.eliminated;
      p.allIn = false;
      p.betTotal = 0;
      p.lastAction = null;
      p.lastActionAmt = 0;
      if (p.eliminated) continue;
    }

    var alive = [];
    for (var j = 0; j < state.players.length; j++) {
      if (!state.players[j].eliminated) alive.push(j);
    }
    if (alive.length < 2) {
      state.street = "done";
      state.message = "对局结束，请重新开始";
      return state;
    }

    while (state.players[state.dealer].eliminated) {
      state.dealer = (state.dealer + 1) % state.players.length;
    }
    var sb = nextSeat(state, state.dealer);
    var bb = nextSeat(state, sb);
    state.sb = sb;
    state.bb = bb;
    postBlind(state, sb, state.opts.smallBlind, "sb");
    postBlind(state, bb, state.opts.bigBlind, "bb");
    state.lastAggressor = bb;

    for (var k = 0; k < state.players.length; k++) {
      if (state.players[k].eliminated) continue;
      state.players[k].hole.push(state.deck.pop());
      state.players[k].hole.push(state.deck.pop());
    }

    state.actor = nextActorSeat(state, bb);
    if (state.actor < 0) advanceStreet(state);
    return state;
  }

  function toCall(state, idx) {
    var p = state.players[idx];
    return Math.max(0, state.currentBet - p.betThisRound);
  }

  function allInWouldRaise(state, idx) {
    var p = state.players[idx];
    return p.betThisRound + p.chips > state.currentBet;
  }

  function legalActions(state, idx) {
    var p = state.players[idx];
    var acts = [];
    if (state.street === "done" || state.street === "showdown" || state.street === "hand_end" || state.street === "idle") return acts;
    if (p.folded || p.eliminated || p.allIn) return acts;
    if (state.actor !== idx) return acts;

    var callAmt = toCall(state, idx);
    acts.push("fold");
    if (callAmt === 0) acts.push("check");
    else acts.push("call");
    if (p.chips > callAmt && state.raisesThisRound < state.opts.maxRaisesPerRound) {
      acts.push("raise");
    }
    if (p.chips > 0) {
      if (!allInWouldRaise(state, idx) || state.raisesThisRound < state.opts.maxRaisesPerRound) {
        acts.push("allin");
      }
    }
    return acts;
  }

  function payChips(state, idx, amount) {
    var p = state.players[idx];
    var pay = Math.min(amount, p.chips);
    p.chips -= pay;
    p.betThisRound += pay;
    p.betTotal += pay;
    state.pot += pay;
    if (p.chips === 0) p.allIn = true;
    if (p.betThisRound > state.currentBet) state.currentBet = p.betThisRound;
    return pay;
  }

  function bettingRoundComplete(state) {
    var alive = notFoldedPlayers(state);
    if (alive.length <= 1) return true;
    var pending = false;
    for (var i = 0; i < alive.length; i++) {
      var p = state.players[alive[i]];
      if (p.allIn) continue;
      if (!p.acted) { pending = true; break; }
      if (p.betThisRound < state.currentBet) { pending = true; break; }
    }
    if (pending) return false;
    return true;
  }

  function applyAction(state, idx, action, extra) {
    extra = extra || {};
    var p = state.players[idx];
    var legal = legalActions(state, idx);
    if (legal.indexOf(action) < 0) return false;

    if (action === "fold") {
      p.folded = true;
      p.acted = true;
      recordAction(p, "fold", 0);
    } else if (action === "check") {
      p.acted = true;
      recordAction(p, "check", 0);
    } else if (action === "call") {
      var callPaid = payChips(state, idx, toCall(state, idx));
      p.acted = true;
      recordAction(p, "call", callPaid);
    } else if (action === "raise") {
      var mult = extra.raiseMult || 1;
      if (mult < 1) mult = 1;
      if (mult > 9) mult = 9;
      var raiseTo = state.currentBet + mult * state.opts.bigBlind;
      if (raiseTo <= state.currentBet) raiseTo = state.currentBet + mult * state.opts.bigBlind;
      var need = raiseTo - p.betThisRound;
      if (need > p.chips) need = p.chips;
      var raisePaid = payChips(state, idx, need);
      p.acted = true;
      recordAction(p, "raise", raisePaid);
      state.raisesThisRound += 1;
      state.lastAggressor = idx;
      for (var i = 0; i < state.players.length; i++) {
        if (i !== idx && !state.players[i].folded && !state.players[i].allIn) {
          state.players[i].acted = false;
        }
      }
    } else if (action === "allin") {
      var prevBet = state.currentBet;
      var allinPaid = payChips(state, idx, p.chips);
      p.acted = true;
      recordAction(p, "allin", allinPaid);
      if (p.betThisRound > prevBet && state.raisesThisRound < state.opts.maxRaisesPerRound) {
        state.raisesThisRound += 1;
        state.lastAggressor = idx;
        for (var j = 0; j < state.players.length; j++) {
          if (j !== idx && !state.players[j].folded && !state.players[j].allIn) {
            state.players[j].acted = false;
          }
        }
      }
    }

    var alive = notFoldedPlayers(state);
    if (alive.length === 1) {
      awardPotToSingle(state, alive[0]);
      return true;
    }

    if (bettingRoundComplete(state)) {
      advanceStreet(state);
    } else {
      state.actor = nextActorSeat(state, state.actor);
      if (state.actor < 0) advanceStreet(state);
    }
    return true;
  }

  function dealBoard(state, n) {
    for (var i = 0; i < n; i++) state.board.push(state.deck.pop());
  }

  function advanceStreet(state) {
    resetRoundBets(state);
    var alive = notFoldedPlayers(state);
    if (alive.length <= 1) {
      if (alive.length === 1) awardPotToSingle(state, alive[0]);
      return;
    }

    if (state.street === "preflop") {
      state.street = "flop";
      dealBoard(state, 3);
    } else if (state.street === "flop") {
      state.street = "turn";
      dealBoard(state, 1);
    } else if (state.street === "turn") {
      state.street = "river";
      dealBoard(state, 1);
    } else if (state.street === "river") {
      resolveShowdown(state);
      return;
    }

    var canBet = false;
    for (var i = 0; i < alive.length; i++) {
      if (!state.players[alive[i]].allIn) { canBet = true; break; }
    }
    if (!canBet) {
      if (state.street === "flop") {
        state.street = "turn";
        dealBoard(state, 1);
        advanceStreet(state);
        return;
      }
      if (state.street === "turn") {
        state.street = "river";
        dealBoard(state, 1);
        advanceStreet(state);
        return;
      }
      if (state.street === "river") {
        resolveShowdown(state);
        return;
      }
    }

    state.actor = nextActorSeat(state, state.dealer);
    if (state.actor < 0) advanceStreet(state);
  }

  function awardPotToSingle(state, idx) {
    var potAmt = state.pot;
    state.players[idx].chips += potAmt;
    state.message = state.players[idx].name + " 赢得底池 " + potAmt;
    state.winners = [{ id: idx, amount: potAmt, hand: null }];
    finalizeHandReport(state, buildFoldReport(state, idx, potAmt));
    state.pot = 0;
    state.street = "showdown";
    endHandCleanup(state);
  }

  function buildPots(state) {
    var contribs = [];
    for (var i = 0; i < state.players.length; i++) {
      if (state.players[i].betTotal > 0) {
        contribs.push({ id: i, amt: state.players[i].betTotal, folded: state.players[i].folded });
      }
    }
    if (!contribs.length) return [{ amount: state.pot, eligible: notFoldedPlayers(state) }];
    contribs.sort(function (a, b) { return a.amt - b.amt; });
    var pots = [];
    var prev = 0;
    for (var level = 0; level < contribs.length; level++) {
      var slice = contribs[level].amt - prev;
      if (slice <= 0) continue;
      var contributors = 0;
      var eligible = [];
      for (var j = 0; j < state.players.length; j++) {
        if (state.players[j].betTotal >= contribs[level].amt) {
          contributors++;
          if (!state.players[j].folded) eligible.push(j);
        }
      }
      pots.push({ amount: slice * contributors, eligible: eligible });
      prev = contribs[level].amt;
    }
    return pots;
  }

  function resolveShowdown(state) {
    state.street = "showdown";
    state.showHole = true;
    var pots = buildPots(state);
    var winMap = {};
    var comparisons = [];
    for (var p = 0; p < pots.length; p++) {
      var pot = pots[p];
      if (!pot.eligible.length) continue;
      var evals = {};
      var bestIds = [];
      var bestEv = null;
      for (var e = 0; e < pot.eligible.length; e++) {
        var pid = pot.eligible[e];
        var ev = bestHand(state.players[pid].hole, state.board);
        evals[pid] = ev;
        if (!bestEv) {
          bestEv = ev;
          bestIds = [pid];
        } else {
          var cmp = compareEval(ev, bestEv);
          if (cmp > 0) { bestEv = ev; bestIds = [pid]; }
          else if (cmp === 0) bestIds.push(pid);
        }
      }
      var winners = [];
      var wi;
      for (wi = 0; wi < bestIds.length; wi++) winners.push(playerEvalEntry(state, bestIds[wi], evals[bestIds[wi]]));
      var losers = [];
      for (e = 0; e < pot.eligible.length; e++) {
        pid = pot.eligible[e];
        if (bestIds.indexOf(pid) < 0) losers.push(playerEvalEntry(state, pid, evals[pid]));
      }
      comparisons.push({
        potLabel: potLabelFor(p, pots.length),
        amount: pot.amount,
        winners: winners,
        losers: losers
      });
      var share = Math.floor(pot.amount / bestIds.length);
      var remainder = pot.amount - share * bestIds.length;
      for (var w = 0; w < bestIds.length; w++) {
        var wid = bestIds[w];
        var gain = share + (w < remainder ? 1 : 0);
        state.players[wid].chips += gain;
        winMap[wid] = winMap[wid] || { id: wid, amount: 0, hand: bestEv };
        winMap[wid].amount += gain;
        winMap[wid].hand = bestEv;
      }
    }
    state.pot = 0;
    var winners = [];
    for (var k in winMap) winners.push(winMap[k]);
    state.winners = winners;
    if (winners.length === 1) {
      state.message = winners[0].hand ? winners[0].hand.label + " · " + state.players[winners[0].id].name + " 获胜" : state.players[winners[0].id].name + " 获胜";
    } else {
      state.message = "平分底池";
    }
    finalizeHandReport(state, {
      reason: "showdown",
      board: state.board.map(cardLabel),
      potCount: comparisons.length,
      allInPlayers: collectAllInNames(state),
      comparisons: comparisons
    });
    endHandCleanup(state);
  }

  function endHandCleanup(state) {
    for (var i = 0; i < state.players.length; i++) {
      if (!state.players[i].eliminated && state.players[i].chips <= 0) {
        state.players[i].eliminated = true;
        state.players[i].folded = true;
      }
    }
    var alive = [];
    for (var j = 0; j < state.players.length; j++) {
      if (!state.players[j].eliminated) alive.push(j);
    }
    if (alive.length <= 1) {
      state.street = "done";
      if (alive.length === 1) {
        var champ = alive[0];
        var champInWinners = false;
        for (var w = 0; w < state.winners.length; w++) {
          if (state.winners[w].id === champ) { champInWinners = true; break; }
        }
        if (!champInWinners) {
          state.winners = [{ id: champ, amount: 0, hand: null }];
        }
        state.message = state.players[champ].name + " 赢得比赛！";
        if (state.handReport) {
          state.handReport.matchEnd = true;
          state.handReport.champion = state.players[champ].name;
          finalizeHandReport(state, state.handReport);
        } else {
          var champLine = state.players[champ].name + " 赢得整场比赛！";
          state.handSummaryLines = [champLine];
          state.handSummaryFullLines = [champLine];
          state.handSummaryTruncated = false;
        }
      }
      return;
    }
    state.dealer = (state.dealer + 1) % state.players.length;
    state.actor = -1;
    state.currentBet = 0;
    state.street = "hand_end";
  }

  function streetLabel(street) {
    if (street === "preflop") return "翻牌前";
    if (street === "flop") return "翻牌";
    if (street === "turn") return "转牌";
    if (street === "river") return "河牌";
    if (street === "showdown") return "摊牌";
    if (street === "hand_end") return "本手结束";
    if (street === "done") return "对局结束";
    return "准备";
  }

  function estimateStrength(hole, board, samples) {
    if (!hole || hole.length < 2) return 0;
    samples = samples || 80;
    var wins = 0;
    var deck = makeDeck();
    var used = {};
    used[cardKey(hole[0])] = 1;
    used[cardKey(hole[1])] = 1;
    for (var i = 0; i < board.length; i++) used[cardKey(board[i])] = 1;
    var rest = [];
    for (var d = 0; d < deck.length; d++) {
      if (!used[cardKey(deck[d])]) rest.push(deck[d]);
    }
    for (var t = 0; t < samples; t++) {
      var sh = shuffle(rest);
      var b = board.slice();
      var need = 5 - b.length;
      for (var n = 0; n < need; n++) b.push(sh[n]);
      var mine = bestHand(hole, b);
      var oppHole = [sh[need], sh[need + 1]];
      var opp = bestHand(oppHole, b);
      if (compareEval(mine, opp) >= 0) wins++;
    }
    return wins / samples;
  }

  function activeSeatIndices(state) {
    var out = [];
    for (var i = 0; i < state.players.length; i++) {
      if (!state.players[i].eliminated) out.push(i);
    }
    return out;
  }

  function seatPositionFromDealer(state, idx) {
    var seats = activeSeatIndices(state);
    var n = seats.length;
    if (!n) return 0;
    var di = seats.indexOf(state.dealer);
    if (di < 0) di = 0;
    var si = seats.indexOf(idx);
    if (si < 0) return 0;
    return (si - di + n) % n;
  }

  function serializePublic(state) {
    var humanFolded = state.players[0] && state.players[0].folded;
    var inHand = state.street !== "idle";
    var handEnded = state.street === "hand_end" || state.street === "done" || state.street === "showdown";
    var revealAllAi = humanFolded && inHand && !handEnded;
    return {
      playerCount: state.players.length,
      street: state.street,
      pot: state.pot,
      board: state.board.map(cloneCard),
      currentBet: state.currentBet,
      actor: state.actor,
      dealer: state.dealer,
      bb: state.bb != null ? state.bb : -1,
      message: state.message,
      handSummaryLines: state.handSummaryLines || [],
      handSummaryFullLines: state.handSummaryFullLines || state.handSummaryLines || [],
      handSummaryTruncated: !!state.handSummaryTruncated,
      showHole: state.showHole,
      revealAiHole: revealAllAi,
      winners: (state.winners || []).map(function (w) {
        var pl = state.players[w.id];
        return {
          id: w.id,
          amount: w.amount,
          name: pl ? pl.name : "",
          handLabel: w.hand ? w.hand.label : null
        };
      }),
      players: state.players.map(function (p) {
        var showFace = state.showHole || handEnded || p.isHuman
          || (inHand && !p.isHuman && (p.folded || revealAllAi));
        return {
          id: p.id,
          name: p.name,
          isHuman: p.isHuman,
          chips: p.chips,
          folded: p.folded,
          allIn: p.allIn,
          eliminated: p.eliminated,
          betThisRound: p.betThisRound,
          betTotal: p.betTotal,
          lastAction: p.lastAction,
          lastActionAmt: p.lastActionAmt,
          hole: showFace ? p.hole.map(cloneCard) : p.hole.length ? [{ hidden: true }, { hidden: true }] : []
        };
      })
    };
  }

  root.FsTexasRules = {
    SUITS: SUITS,
    SUIT_COLORS: SUIT_COLORS,
    RANK_CHARS: RANK_CHARS,
    HAND_NAMES: HAND_NAMES,
    cardLabel: cardLabel,
    createGame: createGame,
    startHand: startHand,
    legalActions: legalActions,
    applyAction: applyAction,
    bestHand: bestHand,
    compareEval: compareEval,
    toCall: toCall,
    seatPositionFromDealer: seatPositionFromDealer,
    streetLabel: streetLabel,
    estimateStrength: estimateStrength,
    serializePublic: serializePublic,
    defaultOpts: defaultOpts
  };
})(typeof self !== "undefined" ? self : window);

/* 德州扑克结算说明排版（依赖 FsTexasRules） */
(function (root) {
  "use strict";

  var R = root.FsTexasRules;

  function rankChar(r) {
    if (!R) return String(r);
    return R.RANK_CHARS[r] || String(r);
  }

  function formatEvalDetail(ev) {
    if (!ev) return "";
    var cat = ev.cat;
    var tie = ev.tie || [];
    if (cat === 1) return rankChar(tie[0]);
    if (cat === 2) return rankChar(tie[0]) + "与" + rankChar(tie[1]);
    if (cat === 3) return rankChar(tie[0]);
    if (cat === 4) return "至" + rankChar(tie[0]);
    if (cat === 5 || cat === 8 || cat === 9) return rankChar(tie[0]) + "高";
    if (cat === 6) return rankChar(tie[0]) + "带" + rankChar(tie[1]);
    if (cat === 7) return rankChar(tie[0]);
    if (cat === 0 && tie.length) return rankChar(tie[0]) + "踢脚";
    return "";
  }

  function shortHand(ev) {
    if (!ev) return "未知";
    var detail = formatEvalDetail(ev);
    if (detail) return (ev.label || "高牌") + "(" + detail + ")";
    return ev.label || "高牌";
  }

  function formatComparisonWhy(winner, loser) {
    if (!winner || !loser) return "";
    if (winner.cat !== loser.cat) {
      return (winner.label || "") + "大于" + (loser.label || "");
    }
    var tie = winner.tie || [];
    var ltie = loser.tie || [];
    for (var i = 0; i < Math.max(tie.length, ltie.length); i++) {
      var av = tie[i] || 0;
      var bv = ltie[i] || 0;
      if (av === bv) continue;
      if (winner.cat === 1) {
        if (i === 0) return "对" + rankChar(av) + "大于对" + rankChar(bv);
        return "对子相同，" + rankChar(av) + "踢脚更大";
      }
      if (winner.cat === 2) {
        if (i === 0) return "大对" + rankChar(av) + "大于" + rankChar(bv);
        if (i === 1) return "大对相同，小对" + rankChar(av) + "更大";
        return "两对相同，" + rankChar(av) + "踢脚更大";
      }
      if (winner.cat === 0) return rankChar(av) + "踢脚大于" + rankChar(bv);
      return rankChar(av) + "大于" + rankChar(bv);
    }
    return "牌力相同";
  }

  function streetEndLabel(street) {
    if (street === "preflop") return "翻牌前";
    if (street === "flop") return "翻牌";
    if (street === "turn") return "转牌";
    if (street === "river") return "河牌";
    return "摊牌";
  }

  function formatPotWinPart(comp) {
    var winText = comp.winners.map(function (w) {
      return w.name + " " + shortHand(w.eval);
    }).join("/");
    var potDesc = comp.potLabel + " " + comp.amount;
    var totalPlayers = comp.winners.length + comp.losers.length;
    if (comp.winners.length > 1 && comp.winners.length === totalPlayers) {
      return potDesc + " 平分：" + winText;
    }
    return potDesc + " → " + winText + " 胜";
  }

  function formatPotParticipants(comp) {
    var names = [];
    var seen = {};
    var all = comp.winners.concat(comp.losers);
    for (var i = 0; i < all.length; i++) {
      var n = all[i].name;
      if (!seen[n]) {
        seen[n] = 1;
        names.push(n);
      }
    }
    return names.join("、");
  }

  function formatPotDetailLine(comp) {
    var players = formatPotParticipants(comp);
    var potDesc = comp.potLabel + " " + comp.amount + " 筹码";
    var totalPlayers = comp.winners.length + comp.losers.length;
    if (comp.winners.length > 1 && comp.winners.length === totalPlayers) {
      var tieHands = comp.winners.map(function (w) {
        return w.name + " " + shortHand(w.eval);
      }).join("/");
      return potDesc + "：" + players + " 比牌 → 平分（" + tieHands + "，牌力相同）";
    }
    var winText = comp.winners.map(function (w) {
      return w.name + " " + shortHand(w.eval);
    }).join("/");
    var detail = potDesc + "：" + players + " 比牌 → " + winText + " 胜";
    if (comp.losers.length && comp.winners[0] && comp.losers[0]) {
      detail += "（" + formatComparisonWhy(comp.winners[0].eval, comp.losers[0].eval);
      if (comp.losers.length > 1) {
        detail += " · 优于 " + comp.losers.slice(1).map(function (l) {
          return l.name + " " + shortHand(l.eval);
        }).join("、");
      }
      detail += "）";
    } else if (comp.winners[0] && comp.winners[0].eval && comp.winners[0].eval.cardsLabel) {
      detail += "（最佳五张 " + comp.winners[0].eval.cardsLabel + "）";
    }
    return detail;
  }

  function formatPotSplitLine(report) {
    if (report.potCount <= 1) return "";
    if (report.allInPlayers && report.allInPlayers.length) {
      return "因 " + report.allInPlayers.join("/") + " 梭哈，分 " + report.potCount + " 池各比 1 次";
    }
    return "筹码分 " + report.potCount + " 池各比 1 次";
  }

  function formatSinglePotWhyLine(comp) {
    if (!comp.losers.length || !comp.winners[0] || !comp.losers[0]) {
      if (comp.winners[0] && comp.winners[0].eval && comp.winners[0].eval.cardsLabel) {
        return comp.winners[0].name + " 最佳五张 " + comp.winners[0].eval.cardsLabel;
      }
      return "";
    }
    var why = comp.winners[0].name + " " + formatComparisonWhy(comp.winners[0].eval, comp.losers[0].eval);
    if (comp.losers.length > 1) {
      why += " · 击败 " + comp.losers.slice(1).map(function (l) {
        return l.name + "(" + shortHand(l.eval) + ")";
      }).join("、");
    }
    return why;
  }

  function formatSummaryLines(report, maxLines) {
    maxLines = maxLines || 3;
    if (!report) return { display: [], full: [], truncated: false };
    var lines = [];

    if (report.matchEnd && report.champion) {
      var endLine = report.champion + " 赢得整场比赛！";
      return packSummary([endLine]);
    }

    if (report.reason === "fold") {
      lines.push(report.winner.name + " 赢得底池 " + report.winner.amount + " 筹码");
      if (report.foldedPlayers && report.foldedPlayers.length) {
        var names = report.foldedPlayers.join("、");
        lines.push("其余 " + report.foldedPlayers.length + " 人弃牌（" + names + "），无需摊牌");
      } else {
        lines.push("其余玩家均已弃牌，无需摊牌比牌");
      }
      if (report.board && report.board.length) {
        lines.push("进行至 " + streetEndLabel(report.street) + " · 公牌 " + report.board.join(" "));
      } else if (report.street && report.street !== "preflop") {
        lines.push("进行至 " + streetEndLabel(report.street));
      }
      return packSummary(lines.slice(0, maxLines));
    }

    if (report.reason === "showdown") {
      var comparisons = report.comparisons || [];
      var multiPot = comparisons.length > 1 || report.potCount > 1;

      if (comparisons.length) {
        lines.push(comparisons.map(formatPotWinPart).join(" · "));
      }

      if (multiPot) {
        var splitLine = formatPotSplitLine(report);
        if (splitLine) lines.push(splitLine);
        for (var pi = 0; pi < comparisons.length; pi++) {
          lines.push(formatPotDetailLine(comparisons[pi]));
        }
      } else if (comparisons.length === 1) {
        var whySingle = formatSinglePotWhyLine(comparisons[0]);
        if (whySingle) lines.push(whySingle);
      }

      if (report.board && report.board.length) {
        lines.push("公牌 " + report.board.join(" "));
      }

      return packSummary(lines);
    }

    return packSummary(lines.slice(0, maxLines));
  }

  function packSummary(rawLines) {
    return { display: rawLines.slice(), full: rawLines.slice(), truncated: false };
  }

  root.FsTexasSummary = {
    formatEvalDetail: formatEvalDetail,
    shortHand: shortHand,
    formatComparisonWhy: formatComparisonWhy,
    formatPotDetailLine: formatPotDetailLine,
    formatSummaryLines: formatSummaryLines
  };
})(typeof window !== "undefined" ? window : self);

/* 德州扑克牌桌 Canvas 绘制 */
(function (root) {
  "use strict";

  var R = root.FsTexasRules;
  if (!R) return;

  function BoardView(canvas, opts) {
    this.canvas = canvas;
    this.ctx = canvas.getContext("2d");
    this.opts = opts || {};
    this.summaryEl = this.opts.summaryEl || null;
    this.dpr = 1;
    this.w = 720;
    this.h = 520;
    this.state = null;
    this.cardBackImg = null;
    this.cardBackReady = false;
    this.pub = "";
    this.animTime = 0;
  }

  BoardView.prototype.setAnimTime = function (t) {
    this.animTime = t || 0;
  };

  BoardView.prototype.setState = function (state) {
    this.state = state;
  };

  BoardView.prototype.setPub = function (pub) {
    this.pub = pub || "";
  };

  BoardView.prototype.loadCardBack = function (pub) {
    var self = this;
    self.pub = pub || "";
    var meta = document.querySelector('meta[name="fs-texas-card-back"]');
    if (!meta || !meta.content) {
      self.cardBackReady = true;
      return Promise.resolve();
    }
    var img = new Image();
    var url = (pub || "") + meta.content;
    return new Promise(function (resolve) {
      img.onload = function () {
        self.cardBackImg = img;
        self.cardBackReady = true;
        resolve();
      };
      img.onerror = function () {
        self.cardBackReady = true;
        resolve();
      };
      img.src = url;
    });
  };

  BoardView.prototype.resize = function () {
    var section = this.canvas.closest ? this.canvas.closest(".texas-page") : null;
    var maxW = section ? section.clientWidth - 8 : 720;
    if (maxW < 280) maxW = 280;
    var scale = Math.min(1, maxW / this.w);
    this.dpr = window.devicePixelRatio || 1;
    var cssW = Math.floor(this.w * scale);
    var cssH = Math.floor(this.h * scale);
    this.canvas.style.width = cssW + "px";
    this.canvas.style.height = cssH + "px";
    this.canvas.width = Math.floor(cssW * this.dpr);
    this.canvas.height = Math.floor(cssH * this.dpr);
    this.scale = scale * this.dpr;
    var inner = this.canvas.parentElement;
    if (inner && inner.classList && inner.classList.contains("board-inner")) {
      inner.style.width = cssW + "px";
      inner.style.maxWidth = "100%";
    }
    if (this.summaryEl) this._summaryCacheWidth = 0;
  };

  BoardView.prototype.seatPos = function (idx, total) {
    var cx = this.w / 2;
    var cy = this.h / 2 - 6;
    var rx = this.w * 0.35;
    var ry = this.h * 0.28;
    var angle = Math.PI / 2 + (idx / total) * Math.PI * 2;
    if (idx === 0) angle = Math.PI / 2;
    else {
      var others = total - 1;
      var slot = idx - 1;
      angle = Math.PI + (slot / Math.max(1, others - 1)) * Math.PI;
      if (others === 1) angle = Math.PI;
    }
    return { x: cx + Math.cos(angle) * rx, y: cy + Math.sin(angle) * ry, angle: angle, cx: cx, cy: cy };
  };

  BoardView.prototype.roundRectPath = function (ctx, x, y, w, h, r) {
    ctx.beginPath();
    if (typeof ctx.roundRect === "function") {
      ctx.roundRect(x, y, w, h, r);
      return;
    }
    ctx.moveTo(x + r, y);
    ctx.arcTo(x + w, y, x + w, y + h, r);
    ctx.arcTo(x + w, y + h, x, y + h, r);
    ctx.arcTo(x, y + h, x, y, r);
    ctx.arcTo(x, y, x + w, y, r);
    ctx.closePath();
  };

  BoardView.prototype.drawCodeCardBack = function (ctx, x, y, w, h) {
    var r = 6;
    ctx.fillStyle = "#1e3a5f";
    this.roundRectPath(ctx, x, y, w, h, r);
    ctx.fill();
    ctx.strokeStyle = "#8ab0d0";
    ctx.lineWidth = 1.5;
    ctx.stroke();
    ctx.save();
    this.roundRectPath(ctx, x + 3, y + 3, w - 6, h - 6, r - 2);
    ctx.clip();
    ctx.strokeStyle = "rgba(180,210,240,.35)";
    ctx.lineWidth = 1;
    var step = 10;
    for (var i = -h; i < w + h; i += step) {
      ctx.beginPath();
      ctx.moveTo(x + i, y);
      ctx.lineTo(x + i + h, y + h);
      ctx.stroke();
      ctx.beginPath();
      ctx.moveTo(x + i, y + h);
      ctx.lineTo(x + i + h, y);
      ctx.stroke();
    }
    ctx.restore();
  };

  BoardView.prototype.drawCardBack = function (ctx, x, y, w, h) {
    if (this.cardBackImg) {
      ctx.drawImage(this.cardBackImg, x, y, w, h);
      ctx.strokeStyle = "#8ab0d0";
      ctx.lineWidth = 1;
      ctx.strokeRect(x + 0.5, y + 0.5, w - 1, h - 1);
    } else {
      this.drawCodeCardBack(ctx, x, y, w, h);
    }
  };

  BoardView.prototype.holeFaceUp = function (p) {
    return !!(p.hole && p.hole.length && p.hole[0] && !p.hole[0].hidden);
  };

  BoardView.prototype.seatLayout = function (idx, px, py, cardW, cardH, pos) {
    var human = idx === 0;
    var cx = pos.cx;
    var cy = pos.cy;
    var dx = cx - px;
    var dy = cy - py;
    var dlen = Math.sqrt(dx * dx + dy * dy) || 1;
    dx /= dlen;
    dy /= dlen;
    var perpX = -dy;
    var perpY = dx;
    var cardSpanW = cardW * 2 + 4;
    var cardInset = human ? 30 : 24;
    var cardMidX = px + dx * cardInset;
    var cardMidY = py + dy * cardInset;
    var hx = cardMidX - cardSpanW / 2;
    var cardTop = cardMidY - cardH / 2;
    var labelX = px;
    var chipR = 14;
    var nameY;
    var chipY;
    if (human) {
      nameY = cardTop - 28;
      chipY = cardTop + cardH + chipR + 10;
    } else {
      labelX = cardMidX;
      nameY = cardTop - 40;
      chipY = cardTop + cardH + chipR + 10;
    }
    var betX = cardMidX;
    var betY = cardTop + cardH + chipR + 10;
    var actionX = cardMidX;
    var actionY = betY;
    var edgeDx = -dx;
    var edgeDy = -dy;
    return {
      human: human,
      hx: hx,
      cardTop: cardTop,
      cardMidX: cardMidX,
      cardSpanW: cardSpanW,
      labelX: labelX,
      nameY: nameY,
      chipY: chipY,
      betX: betX,
      betY: betY,
      betR: chipR,
      actionX: actionX,
      actionY: actionY,
      edgeDx: edgeDx,
      edgeDy: edgeDy
    };
  };

  BoardView.prototype.computeBBPos = function (ctx, labelX, nameY, nameText, edgeDx, edgeDy, bbR) {
    bbR = bbR || 10;
    ctx.font = "13px system-ui,sans-serif";
    var tw = ctx.measureText(nameText).width;
    var th = 13;
    var left = labelX - tw / 2;
    var right = labelX + tw / 2;
    var top = nameY - th;
    var bottom = nameY;
    var corners = [[left, top], [right, top], [left, bottom], [right, bottom]];
    var bestDot = -Infinity;
    var bestX = labelX;
    var bestY = nameY - th / 2;
    for (var c = 0; c < corners.length; c++) {
      var dot = corners[c][0] * edgeDx + corners[c][1] * edgeDy;
      if (dot > bestDot) {
        bestDot = dot;
        bestX = corners[c][0];
        bestY = corners[c][1];
      }
    }
    var gap = bbR + 5;
    return { x: bestX + edgeDx * gap, y: bestY + edgeDy * gap };
  };

  BoardView.prototype.formatChipAmt = function (n) {
    if (n >= 10000) return Math.round(n / 1000) + "k";
    return String(n);
  };

  BoardView.prototype.formatLastAction = function (p) {
    if (!p.lastAction) return "";
    var labels = { fold: "弃牌", check: "过牌", call: "跟注", raise: "加注", allin: "梭哈", sb: "小盲", bb: "大盲" };
    var t = labels[p.lastAction] || p.lastAction;
    if (p.lastActionAmt > 0) t += " " + p.lastActionAmt;
    return t;
  };

  BoardView.prototype.layoutBetRow = function (ctx, cardMidX, cardTop, cardH, actionTxt, chipTxt, betTotal) {
    var gap = 8;
    var betR = 0;
    if (betTotal > 0) {
      betR = betTotal >= 1000 ? 16 : betTotal >= 100 ? 14 : 12;
    }
    var actionW = 0;
    if (actionTxt) {
      ctx.font = "11px system-ui,sans-serif";
      actionW = ctx.measureText(actionTxt).width + 14;
    }
    ctx.font = "12px system-ui,sans-serif";
    var balW = chipTxt ? ctx.measureText(chipTxt).width : 0;
    var chipW = betR ? betR * 2 : 0;
    var rowW = balW;
    if (balW && chipW) rowW += gap;
    rowW += chipW;
    if (actionW && (balW || chipW)) rowW += gap;
    rowW += actionW;
    var y = cardTop + cardH + (betR || 14) + 10;
    var x = cardMidX - rowW / 2;
    var out = { betY: y, betR: betR || 14 };
    if (balW) {
      out.balX = x;
      out.balY = y;
      x += balW;
      if (chipW || actionW) x += gap;
    }
    if (chipW) {
      out.betX = x + betR;
      x += chipW;
      if (actionW) x += gap;
    }
    if (actionW) {
      out.actionX = x + actionW / 2;
      out.actionY = y;
    }
    return out;
  };

  BoardView.prototype.drawBetChip = function (ctx, x, y, amount, r) {
    if (!amount) return;
    r = r || (amount >= 1000 ? 16 : amount >= 100 ? 14 : 12);
    ctx.save();
    ctx.fillStyle = "#c62828";
    ctx.strokeStyle = "#7f1010";
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.arc(x, y, r, 0, Math.PI * 2);
    ctx.fill();
    ctx.stroke();
    ctx.strokeStyle = "rgba(255,220,220,.45)";
    ctx.lineWidth = 1.5;
    ctx.beginPath();
    ctx.arc(x, y, r - 3.5, 0, Math.PI * 2);
    ctx.stroke();
    ctx.fillStyle = "#fff8f0";
    ctx.font = "bold " + (amount >= 1000 ? 8 : amount >= 100 ? 9 : 10) + "px system-ui,sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText(this.formatChipAmt(amount), x, y);
    ctx.restore();
  };

  BoardView.prototype.drawActionLabel = function (ctx, x, y, text, folded) {
    if (!text) return;
    ctx.save();
    ctx.font = "11px system-ui,sans-serif";
    var tw = ctx.measureText(text).width;
    var padX = 7;
    var padY = 4;
    var bw = tw + padX * 2;
    var bh = 11 + padY * 2;
    ctx.fillStyle = "rgba(10,24,16,.78)";
    this.roundRectPath(ctx, x - bw / 2, y - bh / 2, bw, bh, 5);
    ctx.fill();
    ctx.strokeStyle = folded ? "rgba(160,175,160,.45)" : "rgba(255,230,180,.35)";
    ctx.lineWidth = 1;
    ctx.stroke();
    ctx.fillStyle = folded ? "rgba(200,210,200,.8)" : "rgba(255,248,220,.95)";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText(text, x, y);
    ctx.restore();
  };

  BoardView.prototype.resolveWinnerIds = function (players, winList, street) {
    var winnerIds = {};
    var wi;
    for (wi = 0; wi < winList.length; wi++) winnerIds[winList[wi].id] = true;
    if (!Object.keys(winnerIds).length && street === "done") {
      for (wi = 0; wi < players.length; wi++) {
        if (!players[wi].eliminated) winnerIds[players[wi].id] = true;
      }
    }
    return winnerIds;
  };

  BoardView.prototype.drawBreathCardGlow = function (ctx, x, y, w, h, t, style) {
    var breath = 0.5 + 0.5 * Math.sin(t * 0.006);
    var pad = 8;
    var bx = x - pad;
    var by = y - pad;
    var bw = w + pad * 2;
    var bh = h + pad * 2;
    var r = 12;
    var grd;
    var shadowHue;

    ctx.save();
    if (style === "actor") {
      var shift = breath * 18;
      grd = ctx.createLinearGradient(bx, by, bx + bw, by + bh);
      grd.addColorStop(0, "hsl(205, 92%, " + (52 + shift) + "%)");
      grd.addColorStop(0.5, "hsl(220, 95%, " + (58 + shift) + "%)");
      grd.addColorStop(1, "hsl(195, 90%, " + (48 + shift) + "%)");
      shadowHue = "205";
    } else {
      var hue = (t * 0.1) % 360;
      grd = ctx.createLinearGradient(bx, by, bx + bw, by + bh);
      grd.addColorStop(0, "hsl(" + hue + ", 95%, 55%)");
      grd.addColorStop(0.33, "hsl(" + ((hue + 120) % 360) + ", 95%, 55%)");
      grd.addColorStop(0.66, "hsl(" + ((hue + 240) % 360) + ", 95%, 55%)");
      grd.addColorStop(1, "hsl(" + hue + ", 95%, 55%)");
      shadowHue = String(hue);
    }

    ctx.shadowColor = "hsla(" + shadowHue + ", 95%, 55%, " + (0.35 + breath * 0.35) + ")";
    ctx.shadowBlur = 10 + breath * 10;

    ctx.globalAlpha = 0.14 + breath * 0.18;
    ctx.fillStyle = grd;
    this.roundRectPath(ctx, bx, by, bw, bh, r);
    ctx.fill();

    ctx.globalAlpha = 0.85 + breath * 0.15;
    ctx.strokeStyle = grd;
    ctx.lineWidth = 3 + breath * 3;
    this.roundRectPath(ctx, bx, by, bw, bh, r);
    ctx.stroke();
    ctx.restore();
  };

  BoardView.prototype.syncHandSummaryOverflow = function (el, full) {
    if (!el || !el.clientWidth) return;
    var ps = el.querySelectorAll("p");
    var truncated = false;
    var i;
    for (i = 0; i < ps.length; i++) {
      if (ps[i].scrollWidth > ps[i].clientWidth + 1 || ps[i].scrollHeight > ps[i].clientHeight + 1) {
        truncated = true;
        break;
      }
    }
    var had = el.classList.contains("has-truncated");
    if (truncated === had) return;
    if (truncated) {
      el.classList.add("has-truncated");
      el.setAttribute("tabindex", "0");
      el.setAttribute("title", full.join("\n"));
      var tipHtml = '<div class="texas-hand-summary-tip" role="tooltip">';
      for (i = 0; i < full.length; i++) {
        tipHtml += "<p>" + escapeHtml(full[i]) + "</p>";
      }
      tipHtml += "</div>";
      el.insertAdjacentHTML("beforeend", tipHtml);
    } else {
      el.classList.remove("has-truncated");
      el.removeAttribute("tabindex");
      el.removeAttribute("title");
      var tip = el.querySelector(".texas-hand-summary-tip");
      if (tip) tip.remove();
    }
  };

  BoardView.prototype.syncHandSummary = function () {
    var el = this.summaryEl;
    if (!el) return;
    var showWinFx = this.state && (this.state.street === "hand_end" || this.state.street === "done");
    var lines = showWinFx ? (this.state.handSummaryLines || []) : [];
    if (!lines.length) {
      el.hidden = true;
      el.innerHTML = "";
      el.classList.remove("has-truncated");
      this._summaryCacheKey = "";
      this._summaryCacheWidth = 0;
      return;
    }
    var full = this.state.handSummaryFullLines || lines;
    var cacheKey = lines.join("\x00") + "\x01" + full.join("\x00");
    var needRender = cacheKey !== this._summaryCacheKey;
    el.hidden = false;
    if (needRender) {
      this._summaryCacheKey = cacheKey;
      this._summaryCacheWidth = 0;
      var html = "";
      var i;
      for (i = 0; i < lines.length; i++) {
        html += "<p>" + escapeHtml(lines[i]) + "</p>";
      }
      el.classList.remove("has-truncated");
      el.removeAttribute("tabindex");
      el.removeAttribute("title");
      el.innerHTML = html;
    }
    var w = el.clientWidth;
    if (needRender || w !== this._summaryCacheWidth) {
      this._summaryCacheWidth = w;
      this.syncHandSummaryOverflow(el, full);
    }
  };

  function escapeHtml(text) {
    return String(text)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  BoardView.prototype.drawCardFace = function (ctx, card, x, y, w, h) {
    if (!card || card.hidden) {
      this.drawCardBack(ctx, x, y, w, h);
      return;
    }
    var r = 6;
    ctx.fillStyle = "#fafcf8";
    this.roundRectPath(ctx, x, y, w, h, r);
    ctx.fill();
    ctx.strokeStyle = "#b8c9b8";
    ctx.lineWidth = 1.5;
    ctx.stroke();
    var col = R.SUIT_COLORS[card.s] || "#1a2a3a";
    var label = R.RANK_CHARS[card.r] + R.SUITS[card.s];
    ctx.fillStyle = col;
    ctx.font = "bold " + Math.floor(h * 0.28) + "px system-ui,sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText(label, x + w / 2, y + h / 2);
  };

  BoardView.prototype.draw = function () {
    if (!this.state) return;
    this.resize();
    var ctx = this.ctx;
    ctx.setTransform(this.scale, 0, 0, this.scale, 0, 0);
    ctx.clearRect(0, 0, this.w, this.h);

    var grd = ctx.createLinearGradient(0, 0, 0, this.h);
    grd.addColorStop(0, "#2a4a32");
    grd.addColorStop(1, "#1e3828");
    ctx.fillStyle = grd;
    ctx.fillRect(0, 0, this.w, this.h);

    var cx = this.w / 2;
    var cy = this.h / 2 - 2;
    ctx.fillStyle = "#245830";
    ctx.beginPath();
    ctx.ellipse(cx, cy, this.w * 0.42, this.h * 0.33, 0, 0, Math.PI * 2);
    ctx.fill();
    ctx.strokeStyle = "#6a8a5a";
    ctx.lineWidth = 3;
    ctx.stroke();

    ctx.textAlign = "center";
    ctx.textBaseline = "middle";

    if (this.state.street === "idle") {
      ctx.fillStyle = "rgba(255,255,240,.75)";
      ctx.font = "15px system-ui,sans-serif";
      ctx.fillText("点击「开始」发牌", cx, cy);
    } else {
      var cw = 44;
      var ch = 62;
      var gap = 6;
      var board = this.state.board || [];
      var boardY = cy - ch / 2 - 6;
      var bx = cx - ((board.length * (cw + gap) - gap) / 2);
      for (var b = 0; b < board.length; b++) {
        this.drawCardFace(ctx, board[b], bx + b * (cw + gap), boardY, cw, ch);
      }
      if (!board.length) {
        ctx.fillStyle = "rgba(255,255,240,.45)";
        ctx.font = "13px system-ui,sans-serif";
        ctx.fillText("公共牌", cx, boardY + ch / 2);
      }
      ctx.fillStyle = "rgba(255,255,240,.92)";
      ctx.font = "600 15px system-ui,sans-serif";
      ctx.fillText("底池 " + this.state.pot, cx, boardY + ch + 16);
    }

    var players = this.state.players || [];
    var n = players.length;
    var cardW = n > 5 ? 34 : 38;
    var cardH = n > 5 ? 48 : 54;

    var winList = this.state.winners || [];
    var showWinFx = this.state.street === "hand_end" || this.state.street === "done";
    var winnerIds = showWinFx ? this.resolveWinnerIds(players, winList, this.state.street) : {};
    var actorId = showWinFx ? -1 : this.state.actor;

    for (var i = 0; i < n; i++) {
      var p = players[i];
      var pos = this.seatPos(i, n);
      var px = pos.x;
      var py = pos.y;
      if (i === 0) py -= 18;
      var lay = this.seatLayout(i, px, py, cardW, cardH, pos);
      var faceUp = this.holeFaceUp(p);
      var hasHole = !!(p.hole && p.hole.length);

      ctx.textAlign = "center";
      ctx.textBaseline = "bottom";
      ctx.font = "12px system-ui,sans-serif";
      ctx.fillStyle = "rgba(220,235,220,.85)";
      var chipTxt = String(p.chips);
      var name = "";
      if (i !== 0) {
        name = p.name + (p.eliminated ? " (出局)" : "") + (p.folded ? " · 弃牌" : "");
        ctx.fillStyle = p.folded ? "rgba(200,210,200,.55)" : "rgba(255,255,240,.95)";
        ctx.font = "13px system-ui,sans-serif";
        ctx.fillText(name, lay.labelX, lay.nameY);
        ctx.font = "12px system-ui,sans-serif";
      } else {
        if (p.eliminated) chipTxt += " (出局)";
        if (p.folded) chipTxt += " · 弃牌";
      }

      if (this.state.bb >= 0 && p.id === this.state.bb) {
        var bbLabel = i !== 0 ? name : chipTxt;
        var bbAnchorY = i !== 0 ? lay.nameY : lay.chipY;
        var bbPos = this.computeBBPos(ctx, lay.labelX, bbAnchorY, bbLabel, lay.edgeDx, lay.edgeDy, 10);
        ctx.fillStyle = "#f0e8c0";
        ctx.beginPath();
        ctx.arc(bbPos.x, bbPos.y, 10, 0, Math.PI * 2);
        ctx.fill();
        ctx.fillStyle = "#3d3020";
        ctx.font = "bold 8px system-ui,sans-serif";
        ctx.textAlign = "center";
        ctx.textBaseline = "middle";
        ctx.fillText("BB", bbPos.x, bbPos.y);
      }

      var isActor = p.id === actorId;
      var showCards = hasHole && (!p.eliminated || showWinFx);
      if (showCards) {
        var hx = lay.hx;
        var holeCount = p.hole.length;
        var cardSpanW = lay.cardSpanW;
        var isWinner = !!winnerIds[p.id];
        for (var h = 0; h < holeCount; h++) {
          var card = p.hole[h];
          if (faceUp && card && !card.hidden) {
            this.drawCardFace(ctx, card, hx + h * (cardW + 4), lay.cardTop, cardW, cardH);
          } else {
            this.drawCardBack(ctx, hx + h * (cardW + 4), lay.cardTop, cardW, cardH);
          }
        }
        if (isWinner) {
          this.drawBreathCardGlow(ctx, hx, lay.cardTop, cardSpanW, cardH, this.animTime, "winner");
        } else if (isActor && !p.folded) {
          this.drawBreathCardGlow(ctx, hx, lay.cardTop, cardSpanW, cardH, this.animTime, "actor");
        }
      }

      var showSeatInfo = this.state.street !== "idle" && (!p.eliminated || showWinFx);
      if (showSeatInfo) {
        var actionTxt = this.formatLastAction(p);
        var row = this.layoutBetRow(ctx, lay.cardMidX, lay.cardTop, cardH, actionTxt, chipTxt, p.betTotal);
        ctx.font = "12px system-ui,sans-serif";
        ctx.fillStyle = "rgba(220,235,220,.85)";
        ctx.textAlign = "left";
        ctx.textBaseline = "middle";
        if (row.balX != null) {
          ctx.fillText(chipTxt, row.balX, row.balY);
        }
        if (p.betTotal > 0 && row.betX != null) {
          this.drawBetChip(ctx, row.betX, row.betY, p.betTotal, row.betR);
        }
        if (actionTxt && row.actionX != null) {
          this.drawActionLabel(ctx, row.actionX, row.actionY, actionTxt, p.folded);
        }
      }
    }

    this.syncHandSummary();

  };

  root.FsTexasBoard = BoardView;
})(typeof window !== "undefined" ? window : self);

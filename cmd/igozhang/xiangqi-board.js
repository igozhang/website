/* 象棋棋盘 Canvas 绘制 */
(function (root) {
  "use strict";

  var R = root.FsXiangqiRules;
  if (!R) return;

  function BoardView(canvas, opts) {
    this.canvas = canvas;
    this.ctx = canvas.getContext("2d");
    this.opts = opts || {};
    this.board = R.initialBoard();
    this.pad = 36;
    this.frame = 10;
    this.files = R.FILES;
    this.ranks = R.RANKS;
    this.dpr = 1;
    this.w = 360;
    this.h = 400;
    this.cellX = 36;
    this.cellY = 36;
    this.legal = [];
    this.selected = null;
    this.lastMove = null;
    this.hover = null;
    this.aiHighlight = null;
    this.anim = null;
    this.animRaf = 0;
    this.pulseT = 0;
  }

  BoardView.prototype.setBoard = function (board) {
    this.board = board;
  };

  BoardView.prototype.setLegal = function (moves) {
    this.legal = moves || [];
  };

  BoardView.prototype.setSelected = function (sq) {
    this.selected = sq;
  };

  BoardView.prototype.setLastMove = function (mv) {
    this.lastMove = mv;
  };

  BoardView.prototype.setAiHighlight = function (mv) {
    this.aiHighlight = mv && mv.from && mv.to ? { from: mv.from, to: mv.to } : null;
    if (!this.aiHighlight) {
      if (this._pulseRaf) {
        cancelAnimationFrame(this._pulseRaf);
        this._pulseRaf = 0;
      }
      return;
    }
    this.pulseT = performance.now ? performance.now() : Date.now();
    this.ensurePulseLoop();
  };

  BoardView.prototype.ensurePulseLoop = function () {
    if (this._pulseRaf || !this.aiHighlight || this.anim) return;
    var self = this;
    function loop() {
      if (!self.aiHighlight || self.anim) {
        self._pulseRaf = 0;
        return;
      }
      self.draw();
      self._pulseRaf = requestAnimationFrame(loop);
    }
    this._pulseRaf = requestAnimationFrame(loop);
  };

  BoardView.prototype.stopAnim = function () {
    if (this.animRaf) {
      cancelAnimationFrame(this.animRaf);
      this.animRaf = 0;
    }
    this.anim = null;
  };

  BoardView.prototype.setHover = function (sq) {
    this.hover = sq;
  };

  BoardView.prototype.prefersReducedMotion = function () {
    try {
      return window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    } catch (e) { return false; }
  };

  /** AI 走子：棋子从 from 滑到 to，结束后回调 */
  BoardView.prototype.playMoveAnim = function (move, piece, done) {
    var self = this;
    self.stopAnim();
    if (!move || !piece || self.prefersReducedMotion()) {
      if (done) done();
      return;
    }
    var dur = 320;
    var t0 = performance.now ? performance.now() : Date.now();
    self.anim = { from: move.from, to: move.to, piece: piece, t0: t0, dur: dur };
    function tick(now) {
      if (!self.anim) return;
      var p = Math.min(1, (now - self.anim.t0) / self.anim.dur);
      p = p * p * (3 - 2 * p);
      self.anim.p = p;
      self.draw();
      if (p < 1) {
        self.animRaf = requestAnimationFrame(tick);
      } else {
        self.anim = null;
        self.animRaf = 0;
        self.draw();
        if (self.aiHighlight) self.ensurePulseLoop();
        if (done) done();
      }
    }
    self.animRaf = requestAnimationFrame(tick);
  };

  BoardView.prototype.measureSide = function () {
    var section = this.canvas.closest ? this.canvas.closest(".xiangqi-page") : null;
    var availW = 0;
    var padX = 0;
    if (section) {
      var cs = window.getComputedStyle(section);
      padX = parseFloat(cs.paddingLeft) + parseFloat(cs.paddingRight);
      var cw = section.clientWidth;
      if (cw > 0) availW = Math.max(0, cw - padX);
    }
    var iw = window.innerWidth || availW || 320;
    var narrow = iw <= 640;
    if (narrow) {
      var shell = document.querySelector("main.shell");
      if (shell) {
        var shellCs = window.getComputedStyle(shell);
        var shellPad = parseFloat(shellCs.paddingLeft) + parseFloat(shellCs.paddingRight);
        var sw = shell.clientWidth;
        if (sw > 0) availW = Math.max(availW, sw - shellPad - (padX || 0));
      }
      if (availW < 50) availW = Math.max(200, iw - 12);
      return Math.floor(Math.min(availW, 520));
    }
    if (availW < 50) availW = Math.max(200, iw - 40);
    var vh = window.innerHeight || 480;
    var maxH = Math.floor(vh * 0.45);
    if (maxH < 220) maxH = 220;
    var aspect = this.ranks / this.files;
    var wFromH = Math.floor(maxH / aspect);
    return Math.floor(Math.min(availW, wFromH, 520));
  };

  BoardView.prototype.resize = function (force) {
    var canvas = this.canvas;
    var wrap = canvas.parentElement;
    var nextW = this.measureSide();
    if (nextW < 200) nextW = 200;
    if (!force && this.w > 0 && Math.abs(nextW - this.w) < 4) {
      this.draw();
      return;
    }

    if (wrap) { wrap.style.width = ""; wrap.style.height = ""; }
    canvas.style.width = "";
    canvas.style.height = "";
    void canvas.offsetWidth;

    this.dpr = Math.min(window.devicePixelRatio || 1, 2);
    this.w = nextW;
    this.h = Math.floor(this.w * this.ranks / this.files);
    this.cellX = (this.w - 2 * this.pad) / (this.files - 1);
    this.cellY = (this.h - 2 * this.pad) / (this.ranks - 1);
    var cellMin = Math.min(this.cellX, this.cellY);
    var needPad = Math.ceil(cellMin * 0.52) + 8;
    if (needPad > this.pad) {
      this.pad = needPad;
      this.cellX = (this.w - 2 * this.pad) / (this.files - 1);
      this.cellY = (this.h - 2 * this.pad) / (this.ranks - 1);
      cellMin = Math.min(this.cellX, this.cellY);
    }
    this.frame = Math.max(8, Math.floor(cellMin * 0.14));

    var pxW = Math.floor(this.w * this.dpr);
    var pxH = Math.floor(this.h * this.dpr);
    canvas.width = pxW;
    canvas.height = pxH;
    canvas.style.width = this.w + "px";
    canvas.style.height = this.h + "px";
    if (wrap) {
      wrap.style.width = this.w + "px";
      wrap.style.height = this.h + "px";
    }
    this.ctx.setTransform(this.dpr, 0, 0, this.dpr, 0, 0);
    this.draw();
  };

  BoardView.prototype.xy = function (file, rank) {
    return {
      x: this.pad + file * this.cellX,
      y: this.pad + rank * this.cellY
    };
  };

  BoardView.prototype.posFromEvent = function (ev) {
    var rect = this.canvas.getBoundingClientRect();
    var clientX, clientY;
    if (ev.touches && ev.touches.length) {
      clientX = ev.touches[0].clientX;
      clientY = ev.touches[0].clientY;
    } else if (ev.changedTouches && ev.changedTouches.length) {
      clientX = ev.changedTouches[0].clientX;
      clientY = ev.changedTouches[0].clientY;
    } else {
      clientX = ev.clientX;
      clientY = ev.clientY;
    }
    var x = clientX - rect.left;
    var y = clientY - rect.top;
    var scaleX = this.w / rect.width;
    var scaleY = this.h / rect.height;
    x *= scaleX;
    y *= scaleY;
    var fc = Math.round((x - this.pad) / this.cellX);
    var fr = Math.round((y - this.pad) / this.cellY);
    if (fr < 0 || fr >= this.ranks || fc < 0 || fc >= this.files) return null;
    return { r: fr, c: fc };
  };

  BoardView.prototype.isLegalTarget = function (r, c) {
    for (var i = 0; i < this.legal.length; i++) {
      var m = this.legal[i];
      if (m.to.r === r && m.to.c === c) {
        if (!this.selected || (m.from.r === this.selected.r && m.from.c === this.selected.c)) return true;
      }
    }
    return false;
  };

  BoardView.prototype.draw = function () {
    var ctx = this.ctx;
    var w = this.w, h = this.h;
    ctx.setTransform(this.dpr, 0, 0, this.dpr, 0, 0);
    ctx.clearRect(0, 0, w, h);
    ctx.fillStyle = "#b89550";
    ctx.fillRect(0, 0, w, h);

    var gL = this.pad - this.frame;
    var gT = this.pad - this.frame;
    var gW = (this.files - 1) * this.cellX + 2 * this.frame;
    var gH = (this.ranks - 1) * this.cellY + 2 * this.frame;
    ctx.fillStyle = "#c9a85c";
    ctx.fillRect(gL, gT, gW, gH);
    ctx.strokeStyle = "#4a3018";
    ctx.lineWidth = 2;
    ctx.strokeRect(gL + 0.5, gT + 0.5, gW - 1, gH - 1);

    ctx.strokeStyle = "#5c4020";
    ctx.lineWidth = 1;
    var f, r, p, p2;

    for (r = 0; r < this.ranks; r++) {
      p = this.xy(0, r);
      p2 = this.xy(this.files - 1, r);
      ctx.beginPath();
      ctx.moveTo(p.x, p.y);
      ctx.lineTo(p2.x, p2.y);
      ctx.stroke();
    }
    for (f = 0; f < this.files; f++) {
      p = this.xy(f, 0);
      p2 = this.xy(f, 4);
      ctx.beginPath();
      ctx.moveTo(p.x, p.y);
      ctx.lineTo(p2.x, p2.y);
      ctx.stroke();
      p = this.xy(f, 5);
      p2 = this.xy(f, this.ranks - 1);
      ctx.beginPath();
      ctx.moveTo(p.x, p.y);
      ctx.lineTo(p2.x, p2.y);
      ctx.stroke();
    }

    var riverY = this.pad + 4.5 * this.cellY;
    ctx.fillStyle = "rgba(92,64,32,.1)";
    ctx.fillRect(gL + 1, riverY - this.cellY * 0.45, gW - 2, this.cellY * 0.9);
    ctx.fillStyle = "#5c4020";
    ctx.font = "600 " + Math.max(11, Math.floor(this.cellX * 0.42)) + "px system-ui,sans-serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText("楚 河", this.pad + this.cellX * 2, riverY);
    ctx.fillText("汉 界", this.pad + this.cellX * 6, riverY);

    function drawPalace(rankTop, rankBot) {
      var tl = this.xy(3, rankTop);
      var br = this.xy(5, rankBot);
      var o1 = this.xy(3, rankTop);
      var o2 = this.xy(5, rankBot);
      var o3 = this.xy(5, rankTop);
      var o4 = this.xy(3, rankBot);
      ctx.strokeStyle = "#5c4020";
      ctx.lineWidth = 1;
      ctx.strokeRect(tl.x, tl.y, br.x - tl.x, br.y - tl.y);
      ctx.beginPath();
      ctx.moveTo(o1.x, o1.y);
      ctx.lineTo(o2.x, o2.y);
      ctx.stroke();
      ctx.beginPath();
      ctx.moveTo(o3.x, o3.y);
      ctx.lineTo(o4.x, o4.y);
      ctx.stroke();
    }
    drawPalace.call(this, 0, 2);
    drawPalace.call(this, 7, 9);

    var board = this.board;
    var rad = Math.min(this.cellX, this.cellY) * 0.36;

    if (this.selected) {
      p = this.xy(this.selected.c, this.selected.r);
      ctx.fillStyle = "rgba(61,107,74,.22)";
      ctx.beginPath();
      ctx.arc(p.x, p.y, rad + 4, 0, 6.283);
      ctx.fill();
    }

    for (var i = 0; i < this.legal.length; i++) {
      var m = this.legal[i];
      if (this.selected && (m.from.r !== this.selected.r || m.from.c !== this.selected.c)) continue;
      p = this.xy(m.to.c, m.to.r);
      ctx.fillStyle = "rgba(61,107,74,.35)";
      ctx.beginPath();
      ctx.arc(p.x, p.y, rad * 0.35, 0, 6.283);
      ctx.fill();
    }

    if (this.lastMove && !this.aiHighlight) {
      p = this.xy(this.lastMove.to.c, this.lastMove.to.r);
      ctx.strokeStyle = "rgba(200,60,40,.75)";
      ctx.lineWidth = 2;
      ctx.beginPath();
      ctx.arc(p.x, p.y, rad + 2, 0, 6.283);
      ctx.stroke();
    }

    if (this.aiHighlight) {
      this.drawAiMoveHint(ctx, rad);
    }

    var anim = this.anim;
    var hideFrom = anim ? anim.from : null;

    for (r = 0; r < this.ranks; r++) {
      for (f = 0; f < this.files; f++) {
        if (hideFrom && hideFrom.r === r && hideFrom.c === f) continue;
        var piece = board[r][f];
        if (!piece) continue;
        this.drawPiece(ctx, f, r, piece, rad, 1);
      }
    }

    if (anim && anim.piece) {
      var o0 = this.xy(anim.from.c, anim.from.r);
      var o1 = this.xy(anim.to.c, anim.to.r);
      var t = anim.p != null ? anim.p : 1;
      var ax = o0.x + (o1.x - o0.x) * t;
      var ay = o0.y + (o1.y - o0.y) * t;
      this.drawPieceAt(ctx, ax, ay, anim.piece, rad, 1);
    }
  };

  BoardView.prototype.drawPiece = function (ctx, f, r, piece, rad, alpha) {
    var p = this.xy(f, r);
    this.drawPieceAt(ctx, p.x, p.y, piece, rad, alpha);
  };

  BoardView.prototype.drawPieceAt = function (ctx, x, y, piece, rad, alpha) {
    var red = R.isRed(piece);
    ctx.save();
    ctx.globalAlpha = alpha != null ? alpha : 1;
    ctx.beginPath();
    ctx.arc(x, y, rad, 0, 6.283);
    ctx.fillStyle = red ? "#f8f0e4" : "#e8dcc8";
    ctx.fill();
    ctx.strokeStyle = red ? "#a03028" : "#2a2820";
    ctx.lineWidth = 1.2;
    ctx.stroke();
    ctx.fillStyle = red ? "#b03028" : "#1a1814";
    ctx.font = "700 " + Math.max(14, Math.floor(rad * 1.05)) + "px KaiTi,STKaiti,serif";
    ctx.textAlign = "center";
    ctx.textBaseline = "middle";
    ctx.fillText(R.PIECE_LABEL[piece] || piece, x, y + 1);
    ctx.restore();
  };

  BoardView.prototype.drawAiMoveHint = function (ctx, rad) {
    var h = this.aiHighlight;
    if (!h) return;
    var pFrom = this.xy(h.from.c, h.from.r);
    var pTo = this.xy(h.to.c, h.to.r);
    var now = performance.now ? performance.now() : Date.now();
    var pulse = 0.55 + 0.45 * Math.sin((now - (this.pulseT || now)) / 380);

    ctx.save();
    ctx.setLineDash([5, 4]);
    ctx.strokeStyle = "rgba(45,90,160," + (0.5 + 0.35 * pulse).toFixed(2) + ")";
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.arc(pFrom.x, pFrom.y, rad + 5, 0, 6.283);
    ctx.stroke();
    ctx.setLineDash([]);

    ctx.fillStyle = "rgba(45,90,160," + (0.12 + 0.1 * pulse).toFixed(2) + ")";
    ctx.beginPath();
    ctx.arc(pFrom.x, pFrom.y, rad + 3, 0, 6.283);
    ctx.fill();

    ctx.strokeStyle = "rgba(200,80,40," + (0.75 + 0.2 * pulse).toFixed(2) + ")";
    ctx.lineWidth = 2.5;
    ctx.beginPath();
    ctx.arc(pTo.x, pTo.y, rad + 4, 0, 6.283);
    ctx.stroke();
    ctx.fillStyle = "rgba(200,80,40,0.14)";
    ctx.beginPath();
    ctx.arc(pTo.x, pTo.y, rad + 2, 0, 6.283);
    ctx.fill();

    var dx = pTo.x - pFrom.x;
    var dy = pTo.y - pFrom.y;
    var len = Math.sqrt(dx * dx + dy * dy) || 1;
    var ux = dx / len;
    var uy = dy / len;
    var ax = pTo.x - ux * (rad + 6);
    var ay = pTo.y - uy * (rad + 6);
    var sx = pFrom.x + ux * (rad + 4);
    var sy = pFrom.y + uy * (rad + 4);

    ctx.strokeStyle = "rgba(61,107,74," + (0.55 + 0.35 * pulse).toFixed(2) + ")";
    ctx.lineWidth = 2;
    ctx.beginPath();
    ctx.moveTo(sx, sy);
    ctx.lineTo(ax, ay);
    ctx.stroke();

    var head = 7 + rad * 0.12;
    var ang = Math.atan2(dy, dx);
    ctx.fillStyle = "rgba(61,107,74," + (0.65 + 0.3 * pulse).toFixed(2) + ")";
    ctx.beginPath();
    ctx.moveTo(ax, ay);
    ctx.lineTo(ax - head * Math.cos(ang - 0.42), ay - head * Math.sin(ang - 0.42));
    ctx.lineTo(ax - head * Math.cos(ang + 0.42), ay - head * Math.sin(ang + 0.42));
    ctx.closePath();
    ctx.fill();
    ctx.restore();
  };

  root.FsXiangqiBoard = BoardView;
})(typeof window !== "undefined" ? window : this);

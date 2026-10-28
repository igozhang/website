(function (global) {
  "use strict";

  var mermaidPreviewReady = false;

  function mermaidThemeVariables() {
    return {
      background: "transparent",
      mainBkg: "#e4ebe4",
      primaryColor: "#e4ebe4",
      primaryBorderColor: "#b8c9b8",
      primaryTextColor: "#2d3b2d",
      secondaryColor: "#e8efe8",
      secondaryBorderColor: "#b8c9b8",
      secondaryTextColor: "#2d3b2d",
      tertiaryColor: "transparent",
      tertiaryBorderColor: "#b8c9b8",
      tertiaryTextColor: "#2d3b2d",
      clusterBkg: "transparent",
      clusterBorder: "#a8bba8",
      lineColor: "#6a7d6a",
      edgeLabelBackground: "transparent",
      titleColor: "#2d3b2d",
      nodeBorder: "#b8c9b8",
      nodeTextColor: "#2d3b2d",
    };
  }

  // Export canvas longest-edge cap (browsers often fail past ~8k–16k).
  var EXPORT_MAX_EDGE = 8192;

  function mermaidBaseConfig(forExport) {
    // Preview: fit pane + HTML labels. Export: absolute px; labels still sanitized for XML.
    var forPreview = !forExport;
    return {
      startOnLoad: false,
      theme: "base",
      securityLevel: "strict",
      flowchart: { useMaxWidth: forPreview, htmlLabels: forPreview },
      sequence: { useMaxWidth: forPreview },
      themeVariables: mermaidThemeVariables(),
    };
  }

  function configureMermaid(forExport) {
    if (!global.mermaid || typeof global.mermaid.initialize !== "function") {
      return false;
    }
    global.mermaid.initialize(mermaidBaseConfig(!!forExport));
    mermaidPreviewReady = !forExport;
    return true;
  }

  function ensureMermaidConfigured() {
    if (!global.mermaid || typeof global.mermaid.initialize !== "function") {
      return false;
    }
    if (!mermaidPreviewReady) {
      configureMermaid(false);
    }
    return true;
  }
  function isMermaidBlock(pre) {
    if (!pre || pre.closest(".doc-mermaid-wrap")) return false;
    var code = pre.querySelector("code");
    var cls = ((code && code.className) || "") + " " + (pre.className || "");
    return /\blanguage-mermaid\b/i.test(cls);
  }

  function collectMermaidBlocks(root) {
    var out = [];
    if (!root) return out;
    root.querySelectorAll("pre").forEach(function (pre) {
      if (!isMermaidBlock(pre)) return;
      var code = pre.querySelector("code");
      var text = ((code || pre).textContent || "").replace(/^\n+|\n+$/g, "");
      if (!text.trim()) return;
      out.push({ pre: pre, text: text });
    });
    return out;
  }

  function showMermaidError(wrap, pre, err) {
    var msg = document.createElement("p");
    msg.className = "doc-mermaid-err";
    msg.textContent =
      "Mermaid 渲染失败：" + (err && err.message ? err.message : String(err || "unknown"));
    wrap.appendChild(msg);
    if (pre) {
      if (pre.parentNode !== wrap) {
        wrap.appendChild(pre);
      }
      pre.style.display = "";
    }
  }

  function renderBlock(item, idx) {
    var pre = item.pre;
    var text = item.text;
    var wrap = document.createElement("div");
    wrap.className = "doc-mermaid-wrap";
    pre.parentNode.insertBefore(wrap, pre);
    pre.style.display = "none";

    var id = "igo-mmd-" + idx + "-" + String(Date.now()) + "-" + Math.floor(Math.random() * 1e6);
    return global.mermaid
      .render(id, text)
      .then(function (res) {
        var svg = res && res.svg ? res.svg : "";
        if (!svg) {
          throw new Error("empty svg");
        }
        wrap.innerHTML = svg;
        if (pre.parentNode) pre.parentNode.removeChild(pre);
      })
      .catch(function (err) {
        wrap.innerHTML = "";
        showMermaidError(wrap, pre, err);
      });
  }

  function runMermaid(root) {
    var blocks = collectMermaidBlocks(root);
    if (!blocks.length) return Promise.resolve();
    if (!ensureMermaidConfigured()) {
      return Promise.reject(new Error("mermaid unavailable"));
    }
    var chain = Promise.resolve();
    blocks.forEach(function (item, idx) {
      chain = chain.then(function () {
        return renderBlock(item, idx);
      });
    });
    return chain;
  }

  function initDocMermaid(root) {
    if (!root) return;
    if (!collectMermaidBlocks(root).length) return;

    function go() {
      runMermaid(root).catch(function () {
        collectMermaidBlocks(root).forEach(function (item) {
          item.pre.style.display = "";
        });
      });
    }

    if (global.mermaid) {
      go();
      return;
    }
    var n = 0;
    var timer = setInterval(function () {
      n += 1;
      if (global.mermaid) {
        clearInterval(timer);
        go();
      } else if (n >= 80) {
        clearInterval(timer);
      }
    }, 50);
  }

  function renderMermaidSource(text, targetEl) {
    if (!targetEl) return Promise.reject(new Error("no target"));
    var trimmed = (text || "").replace(/^\n+|\n+$/g, "");
    targetEl.innerHTML = "";
    if (!trimmed.trim()) {
      return Promise.resolve();
    }
    if (!ensureMermaidConfigured()) {
      return Promise.reject(new Error("mermaid unavailable"));
    }
    var id = "igo-mmd-src-" + String(Date.now()) + "-" + Math.floor(Math.random() * 1e6);
    return global.mermaid
      .render(id, trimmed)
      .then(function (res) {
        var svg = res && res.svg ? res.svg : "";
        if (!svg) {
          throw new Error("empty svg");
        }
        targetEl.innerHTML = svg;
      })
      .catch(function (err) {
        targetEl.innerHTML = "";
        showMermaidError(targetEl, null, err);
      });
  }

  function parseSvgLengthAttr(raw) {
    if (raw == null || raw === "") return NaN;
    var s = String(raw).trim();
    if (!s || /%$/.test(s) || /^auto$/i.test(s)) return NaN;
    return parseFloat(s.replace(/px$/i, ""));
  }

  function sizeFromViewBoxAttr(vbAttr) {
    if (!vbAttr) return null;
    var parts = String(vbAttr).trim().split(/[\s,]+/);
    if (parts.length !== 4) return null;
    var vbW = parseFloat(parts[2]);
    var vbH = parseFloat(parts[3]);
    if (vbW > 0 && vbH > 0) return { w: vbW, h: vbH };
    return null;
  }

  function svgExportSizeFromAttrs(viewBox, widthAttr, heightAttr) {
    var fromVb = sizeFromViewBoxAttr(viewBox);
    if (fromVb) return fromVb;
    var w = parseSvgLengthAttr(widthAttr);
    var h = parseSvgLengthAttr(heightAttr);
    if (w > 0 && h > 0) return { w: w, h: h };
    return null;
  }

  // Prefer intrinsic size (viewBox / px attrs). Display rect is CSS-shrunk and poor for PNG.
  function svgExportSizeFromNode(svg) {
    if (!svg) return null;
    var fromAttrs = svgExportSizeFromAttrs(
      svg.getAttribute && svg.getAttribute("viewBox"),
      svg.getAttribute && svg.getAttribute("width"),
      svg.getAttribute && svg.getAttribute("height")
    );
    if (fromAttrs) return fromAttrs;
    var vb = svg.viewBox && svg.viewBox.baseVal;
    if (vb && vb.width > 0 && vb.height > 0) {
      return { w: vb.width, h: vb.height };
    }
    if (svg.getBoundingClientRect) {
      var rect = svg.getBoundingClientRect();
      if (rect.width > 0 && rect.height > 0) {
        return { w: rect.width, h: rect.height };
      }
    }
    return null;
  }

  function svgExportSizeFromStringRegex(svgString) {
    if (!svgString) return null;
    var head = String(svgString).slice(0, 2500);
    var vbM = head.match(/\bviewBox\s*=\s*["']([^"']+)["']/i);
    var wM = head.match(/\bwidth\s*=\s*["']([^"']+)["']/i);
    var hM = head.match(/\bheight\s*=\s*["']([^"']+)["']/i);
    return svgExportSizeFromAttrs(vbM && vbM[1], wM && wM[1], hM && hM[1]);
  }

  function svgExportSizeFromString(svgString) {
    var fromRegex = svgExportSizeFromStringRegex(svgString);
    if (fromRegex) return fromRegex;
    try {
      var sanitized = sanitizeSvgForExport(svgString);
      var doc = new DOMParser().parseFromString(sanitized, "image/svg+xml");
      if (doc.querySelector("parsererror")) return null;
      var svg = doc.documentElement;
      if (!svg || svg.nodeName.toLowerCase() !== "svg") return null;
      return svgExportSizeFromNode(svg);
    } catch (e) {
      return null;
    }
  }

  // Mermaid foreignObject often emits HTML void tags (<br>) which break image/svg+xml.
  function sanitizeSvgForExport(svgString) {
    var out = svgString || "";
    out = out.replace(/<(br|hr|img|input|meta|link|source|area|base|col|embed|wbr)(\s[^>/]*)?\/?>/gi, function (_, tag, attrs) {
      return "<" + tag + (attrs || "") + "/>";
    });
    if (!/xmlns=/.test(out)) {
      out = out.replace(/<svg\b/i, '<svg xmlns="http://www.w3.org/2000/svg"');
    }
    return out;
  }

  function normalizeSvgString(svgString) {
    return sanitizeSvgForExport(svgString);
  }

  function clampExportScale(size, scale) {
    scale = scale || 2;
    if (!size || !(size.w > 0) || !(size.h > 0)) return 1;
    if (!(scale > 0)) scale = 1;
    var maxDim = Math.max(size.w, size.h);
    var maxScale = EXPORT_MAX_EDGE / maxDim;
    if (!(maxScale > 0)) return 1;
    // Never raise scale above maxScale (a floor would break the edge cap on huge diagrams).
    if (scale > maxScale) scale = maxScale;
    return scale;
  }

  function exportCanvasPixelSize(size, scale) {
    var s = clampExportScale(size, scale);
    var w = Math.max(1, Math.ceil(size.w * s));
    var h = Math.max(1, Math.ceil(size.h * s));
    // ceil() can nudge 1px past the cap; keep aspect by scaling both down if needed.
    var edge = Math.max(w, h);
    if (edge > EXPORT_MAX_EDGE) {
      var f = EXPORT_MAX_EDGE / edge;
      w = Math.max(1, Math.floor(w * f));
      h = Math.max(1, Math.floor(h * f));
    }
    return { w: w, h: h, scale: s };
  }

  function canvasToPngBlob(canvas) {
    return new Promise(function (resolve, reject) {
      if (canvas.toBlob) {
        canvas.toBlob(
          function (blob) {
            if (blob) resolve(blob);
            else reject(new Error("PNG 编码失败"));
          },
          "image/png"
        );
        return;
      }
      try {
        var data = canvas.toDataURL("image/png");
        var bin = atob(data.split(",")[1]);
        var arr = new Uint8Array(bin.length);
        for (var i = 0; i < bin.length; i++) arr[i] = bin.charCodeAt(i);
        resolve(new Blob([arr], { type: "image/png" }));
      } catch (err) {
        reject(err);
      }
    });
  }

  function downloadPngBlob(blob, filename) {
    var a = document.createElement("a");
    var url = URL.createObjectURL(blob);
    a.href = url;
    a.download = filename || "mermaid.png";
    a.rel = "noopener";
    document.body.appendChild(a);
    a.click();
    setTimeout(function () {
      try {
        a.remove();
      } catch (e) {}
      URL.revokeObjectURL(url);
    }, 200);
  }

  function downloadCanvasPngSync(canvas, filename) {
    var dataUrl = canvas.toDataURL("image/png");
    var a = document.createElement("a");
    a.href = dataUrl;
    a.download = filename || "mermaid.png";
    a.rel = "noopener";
    document.body.appendChild(a);
    a.click();
    setTimeout(function () {
      try {
        a.remove();
      } catch (e) {}
    }, 200);
  }

  function downloadCanvasPng(canvas, filename) {
    return new Promise(function (resolve, reject) {
      try {
        downloadCanvasPngSync(canvas, filename);
        resolve();
      } catch (err) {
        canvasToPngBlob(canvas)
          .then(function (blob) {
            downloadPngBlob(blob, filename);
            resolve();
          })
          .catch(reject);
      }
    });
  }

  function drawSvgStringToCanvas(svgString, scale) {
    var size = svgExportSizeFromString(svgString);
    if (!size) {
      return Promise.reject(new Error("无法确定图表尺寸"));
    }
    var pix = exportCanvasPixelSize(size, scale || 2);
    var canvas = document.createElement("canvas");
    canvas.width = pix.w;
    canvas.height = pix.h;
    var ctx = canvas.getContext("2d");
    if (!ctx) {
      return Promise.reject(new Error("Canvas 不可用"));
    }
    ctx.fillStyle = "#ffffff";
    ctx.fillRect(0, 0, pix.w, pix.h);
    svgString = normalizeSvgString(svgString);
    return new Promise(function (resolve, reject) {
      var img = new Image();
      var timer = setTimeout(function () {
        reject(new Error("SVG 渲染超时"));
      }, 20000);
      img.onload = function () {
        clearTimeout(timer);
        try {
          ctx.drawImage(img, 0, 0, pix.w, pix.h);
          resolve(canvas);
        } catch (err) {
          reject(err);
        }
      };
      img.onerror = function () {
        clearTimeout(timer);
        reject(new Error("SVG 渲染失败"));
      };
      img.src = "data:image/svg+xml;charset=utf-8," + encodeURIComponent(svgString);
    });
  }

  function svgElementToExportString(svg) {
    if (!svg) return "";
    try {
      if (typeof XMLSerializer !== "undefined") {
        return sanitizeSvgForExport(new XMLSerializer().serializeToString(svg));
      }
    } catch (e) {}
    return sanitizeSvgForExport(svg.outerHTML || "");
  }

  function renderMermaidSvgForExport(text) {
    var trimmed = (text || "").replace(/^\n+|\n+$/g, "");
    if (!trimmed.trim()) {
      return Promise.reject(new Error("源码为空"));
    }
    if (!global.mermaid || typeof global.mermaid.render !== "function") {
      return Promise.reject(new Error("Mermaid 未加载"));
    }
    configureMermaid(true);
    var id = "igo-mmd-export-" + String(Date.now()) + "-" + Math.floor(Math.random() * 1e6);
    return global.mermaid.render(id, trimmed).then(
      function (res) {
        configureMermaid(false);
        var svg = res && res.svg ? res.svg : "";
        if (!svg) {
          throw new Error("empty svg");
        }
        return sanitizeSvgForExport(svg);
      },
      function (err) {
        configureMermaid(false);
        throw err;
      }
    );
  }

  function exportMermaidPngFromSource(text, filename, scale, previewRoot) {
    scale = scale || 2;
    filename = filename || "mermaid-" + Date.now() + ".png";
    var trimmed = (text || "").replace(/^\n+|\n+$/g, "");
    if (!trimmed.trim()) {
      return Promise.reject(new Error("源码为空"));
    }
    if (previewRoot && previewRoot.querySelector && previewRoot.querySelector(".doc-mermaid-err")) {
      return Promise.reject(new Error("当前图表存在语法错误"));
    }

    var svg = previewRoot && previewRoot.querySelector ? previewRoot.querySelector("svg") : null;
    return renderMermaidSvgForExport(trimmed)
      .then(function (svgString) {
        return drawSvgStringToCanvas(svgString, scale);
      })
      .then(function (canvas) {
        return downloadCanvasPng(canvas, filename);
      })
      .catch(function (err) {
        if (!svg) {
          throw err;
        }
        var previewSvg = svgElementToExportString(svg);
        if (!previewSvg) {
          throw err;
        }
        return drawSvgStringToCanvas(previewSvg, scale)
          .then(function (canvas) {
            return downloadCanvasPng(canvas, filename);
          })
          .catch(function (e2) {
            // Prefer the fallback failure; it is what actually blocked the last attempt.
            throw e2 || err;
          });
      });
  }

  function exportMermaidPngFromPreview(root, filename, scale) {
    var el = typeof root === "string" ? document.querySelector(root) : root;
    if (!el) {
      return Promise.reject(new Error("找不到预览区域"));
    }
    var svg = el.querySelector("svg");
    if (!svg) {
      return Promise.reject(new Error("暂无可导出的图表"));
    }
    scale = scale || 2;
    filename = filename || "mermaid-" + Date.now() + ".png";
    var svgString = svgElementToExportString(svg);
    if (!svgString) {
      return Promise.reject(new Error("无法序列化图表"));
    }
    return drawSvgStringToCanvas(svgString, scale).then(function (canvas) {
      return downloadCanvasPng(canvas, filename);
    });
  }

  global.initDocMermaid = initDocMermaid;
  global.renderMermaidSource = renderMermaidSource;
  global.exportMermaidPngFromSource = exportMermaidPngFromSource;
  global.exportMermaidPngFromPreview = exportMermaidPngFromPreview;
  global.__igoCollectMermaidBlocks = collectMermaidBlocks;
  global.__igoIsMermaidBlock = isMermaidBlock;
  global.__igoSanitizeSvgForExport = sanitizeSvgForExport;
  global.__igoSvgExportSizeFromString = svgExportSizeFromString;
  global.__igoClampExportScale = clampExportScale;

  document.addEventListener("DOMContentLoaded", function () {
    function run() {
      document.querySelectorAll(".doc-md").forEach(initDocMermaid);
    }
    if (typeof requestIdleCallback === "function") {
      requestIdleCallback(run, { timeout: 3000 });
    } else {
      setTimeout(run, 0);
    }
  });
})(typeof window !== "undefined" ? window : globalThis);

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

  function mermaidBaseConfig(forExport) {
    return {
      startOnLoad: false,
      theme: "base",
      securityLevel: "strict",
      flowchart: { useMaxWidth: true, htmlLabels: !forExport },
      sequence: { useMaxWidth: true },
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

  function svgExportSizeFromNode(svg) {
    if (!svg) return null;
    if (svg.getBoundingClientRect) {
      var rect = svg.getBoundingClientRect();
      if (rect.width > 0 && rect.height > 0) {
        return { w: rect.width, h: rect.height };
      }
    }
    var vb = svg.viewBox && svg.viewBox.baseVal;
    if (vb && vb.width > 0 && vb.height > 0) {
      return { w: vb.width, h: vb.height };
    }
    var vbAttr = svg.getAttribute("viewBox");
    if (vbAttr) {
      var parts = vbAttr.trim().split(/[\s,]+/);
      if (parts.length === 4) {
        var vbW = parseFloat(parts[2]);
        var vbH = parseFloat(parts[3]);
        if (vbW > 0 && vbH > 0) {
          return { w: vbW, h: vbH };
        }
      }
    }
    var w = parseFloat(String(svg.getAttribute("width") || "").replace(/px$/i, ""));
    var h = parseFloat(String(svg.getAttribute("height") || "").replace(/px$/i, ""));
    if (w > 0 && h > 0) {
      return { w: w, h: h };
    }
    return null;
  }

  function svgExportSizeFromString(svgString) {
    try {
      var doc = new DOMParser().parseFromString(svgString, "image/svg+xml");
      var svg = doc.documentElement;
      if (!svg || svg.nodeName.toLowerCase() !== "svg") {
        return null;
      }
      return svgExportSizeFromNode(svg);
    } catch (e) {
      return null;
    }
  }

  function normalizeSvgString(svgString) {
    var out = svgString || "";
    if (!/xmlns=/.test(out)) {
      out = out.replace("<svg", '<svg xmlns="http://www.w3.org/2000/svg"');
    }
    return out;
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

  function buildExportCanvas(svg, scale) {
    var size = svgExportSizeFromNode(svg);
    if (!size) {
      throw new Error("无法确定图表尺寸");
    }
    scale = scale || 2;
    var w = Math.max(1, Math.ceil(size.w * scale));
    var h = Math.max(1, Math.ceil(size.h * scale));
    var canvas = document.createElement("canvas");
    canvas.width = w;
    canvas.height = h;
    var ctx = canvas.getContext("2d");
    if (!ctx) {
      throw new Error("Canvas 不可用");
    }
    ctx.fillStyle = "#ffffff";
    ctx.fillRect(0, 0, w, h);
    ctx.drawImage(svg, 0, 0, w, h);
    return canvas;
  }

  function drawSvgStringToCanvas(svgString, scale) {
    var size = svgExportSizeFromString(svgString);
    if (!size) {
      return Promise.reject(new Error("无法确定图表尺寸"));
    }
    scale = scale || 2;
    var w = Math.max(1, Math.ceil(size.w * scale));
    var h = Math.max(1, Math.ceil(size.h * scale));
    var canvas = document.createElement("canvas");
    canvas.width = w;
    canvas.height = h;
    var ctx = canvas.getContext("2d");
    if (!ctx) {
      return Promise.reject(new Error("Canvas 不可用"));
    }
    ctx.fillStyle = "#ffffff";
    ctx.fillRect(0, 0, w, h);
    svgString = normalizeSvgString(svgString);
    return new Promise(function (resolve, reject) {
      var img = new Image();
      var timer = setTimeout(function () {
        reject(new Error("SVG 渲染超时"));
      }, 12000);
      img.onload = function () {
        clearTimeout(timer);
        try {
          ctx.drawImage(img, 0, 0, w, h);
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
        return svg;
      },
      function (err) {
        configureMermaid(false);
        throw err;
      }
    );
  }

  function trySyncDomExport(svg, scale, filename) {
    var canvas = buildExportCanvas(svg, scale);
    downloadCanvasPngSync(canvas, filename);
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
        try {
          trySyncDomExport(svg, scale, filename);
          return;
        } catch (e2) {
          throw err;
        }
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
    try {
      trySyncDomExport(svg, scale, filename);
      return Promise.resolve();
    } catch (e) {
      return Promise.reject(e);
    }
  }

  global.initDocMermaid = initDocMermaid;
  global.renderMermaidSource = renderMermaidSource;
  global.exportMermaidPngFromSource = exportMermaidPngFromSource;
  global.exportMermaidPngFromPreview = exportMermaidPngFromPreview;
  global.__igoCollectMermaidBlocks = collectMermaidBlocks;
  global.__igoIsMermaidBlock = isMermaidBlock;

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

(function (global) {
  "use strict";

  var SHELL_LANGS = {
    bash: 1,
    sh: 1,
    shell: 1,
    zsh: 1,
    console: 1,
    terminal: 1,
  };

  var LANG_ALIASES = {
    yml: "yaml",
    py: "python",
    golang: "go",
  };

  function escHtml(s) {
    return String(s)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;");
  }

  /**
   * Match on raw text; escape only when assembling HTML (never alter visible content).
   * Placeholders use "\x00hN\x00" (lowercase h) so rules like /\b[A-Z]\w*\b/ cannot rematch them.
   * re() only runs on text outside existing placeholders for the same reason.
   */
  function createHighlighter(text) {
    var s = text || "",
      tok = [];

    /** Later rules (e.g. # comment) may span earlier placeholders; fold them into raw text. */
    function expandPlaceholders(str) {
      return str.replace(/\x00h(\d+)\x00/g, function (_, idx) {
        var inner = tok[+idx];
        return inner ? inner.raw : "";
      });
    }

    function mark(cls, m) {
      var raw = expandPlaceholders(m);
      var i = tok.length;
      tok.push({ cls: cls, raw: raw });
      return "\x00h" + i + "\x00";
    }

    function re(reSrc, cls, flags) {
      var rx = new RegExp(reSrc, flags || "g");
      // Never rematch inside placeholders (e.g. Go type rule eating h0/h5 as identifiers).
      s = s
        .split(/(\x00h\d+\x00)/)
        .map(function (part, idx) {
          if (idx % 2 === 1 || !part) return part;
          return part.replace(rx, function (m) {
            return mark(cls, m);
          });
        })
        .join("");
    }

    function finish() {
      var parts = s.split(/\x00h(\d+)\x00/),
        out = "",
        i;
      for (i = 0; i < parts.length; i++) {
        if (i % 2 === 0) {
          out += escHtml(parts[i]);
        } else {
          var t = tok[+parts[i]];
          out += t
            ? '<span class="' + t.cls + '">' + escHtml(t.raw) + "</span>"
            : "";
        }
      }
      return out;
    }

    return { re: re, finish: finish };
  }

  function detectCodeLang(codeEl) {
    if (!codeEl || !codeEl.className) return "";
    var m = String(codeEl.className).match(/\blanguage-([\w-]+)\b/i);
    if (!m) return "";
    var lang = m[1].toLowerCase();
    return LANG_ALIASES[lang] || lang;
  }

  function isShellLang(lang) {
    return !lang || !!SHELL_LANGS[lang];
  }

  /** Shell / terminal — VS Code Dark+ inspired, log-friendly. */
  function highlightDocShell(text) {
    var h = createHighlighter(text);

    h.re("(\"(?:\\\\.|[^\"\\\\])*\"|'(?:\\\\.|[^'\\\\])*'|`(?:\\\\.|[^`\\\\])*`)", "doc-hl-string");
    h.re("^\\s*#.*$", "doc-hl-comment", "gm");
    h.re("\\$\\{[^}]+\\}", "doc-hl-var");
    h.re("\\$[\\w@#?!*\\-]+", "doc-hl-var");
    h.re("(?<![\\w-])--[\\w-]+", "doc-hl-flag");
    h.re("(?<![\\w-])-[a-zA-Z](?![\\w.-])", "doc-hl-flag");
    h.re(
      "\\b(?:if|then|else|elif|fi|for|while|do|done|case|esac|in|function|return|export|local|readonly|unset|source|select|until|declare|typeset)\\b",
      "doc-hl-keyword"
    );
    h.re(
      "\\b(?:sudo|cd|echo|printf|cat|grep|awk|sed|curl|wget|chmod|chown|mkdir|rm|mv|cp|ln|tar|zip|unzip|systemctl|journalctl|docker|kubectl|helm|git|ssh|scp|rsync|apt|yum|dnf|pip|npm|make|test|\\[)\\b",
      "doc-hl-cmd"
    );
    h.re(
      "(?:/(?:[\\w.-]+/)+[\\w.-]+|/(?:etc|var|usr|opt|home|tmp|proc|sys|dev|run|root|srv|mnt|media|lib(?:64)?)(?:/[\\w.-]+)*)",
      "doc-hl-path"
    );
    h.re(
      "\\b(?:(?:25[0-5]|2[0-4]\\d|[01]?\\d\\d?)\\.){3}(?:25[0-5]|2[0-4]\\d|[01]?\\d\\d?)(?::\\d{1,5})?\\b",
      "doc-hl-ip"
    );
    h.re(
      "\\d{4}-\\d{2}-\\d{2}[T ]\\d{2}:\\d{2}:\\d{2}(?:\\.\\d+)?(?:Z|[+-]\\d{2}:?\\d{2})?",
      "doc-hl-time"
    );
    h.re("\\b[A-Z][a-z]{2}\\s+\\d{1,2}\\s+\\d{2}:\\d{2}:\\d{2}\\b", "doc-hl-time");
    h.re(
      "\\b[\\w@.-]+\\.(?:service|socket|timer|mount|target|slice|path|device)\\b",
      "doc-hl-unit"
    );
    h.re("\\b[\\w.-]+@[\\w.-]+\\b", "doc-hl-user");
    h.re("\\b(?:pid|PID)\\s*[=:]?\\s*\\d+\\b", "doc-hl-pid");
    h.re("\\[\\d{2,}\\]", "doc-hl-pid");
    h.re("\\b\\d+(?:\\.\\d+)?(?:%|\\s*(?:K|M|G|T|P|E)?i?B|\\s*(?:ms|secs?|mins?|hours?|days?))\\b", "doc-hl-num");
    h.re("\\b(?:ERROR|ERR|FATAL|CRITICAL|SEVERE|PANIC|ALERT|EMERG)\\b", "doc-hl-err");
    h.re(
      "\\b(?:failed|failure|denied|refused|not found|segfault|oom|killed|invalid|cannot|can't|unable to)\\b",
      "doc-hl-err"
    );
    h.re("\\b(?:WARN(?:ING)?|DEPRECATED|TIMEOUT|timed out)\\b", "doc-hl-warn");
    h.re("\\b(?:inactive|disabled|masked|stopped|degraded)\\b", "doc-hl-warn");
    h.re("\\bactive\\s*\\([^)]+\\)", "doc-hl-ok");
    h.re("\\b(?:running|enabled|succeeded|healthy|ok|OK|done|completed)\\b", "doc-hl-ok");
    h.re("\\b(?:INFO|NOTICE|DEBUG|TRACE)\\b", "doc-hl-dim");
    h.re("[|;&]{1,2}|>>?|<<|<(?![\\w/!?])|2>&1|>&", "doc-hl-op");

    return h.finish();
  }

  function highlightDocYAML(text) {
    var h = createHighlighter(text);
    h.re("^\\s*#.*$", "doc-hl-comment", "gm");
    h.re("(\"(?:\\\\.|[^\"\\\\])*\"|'(?:\\\\.|[^'\\\\])*')", "doc-hl-string");
    h.re("^\\s*[\\w.-]+(?=\\s*:)", "doc-hl-key", "gm");
    h.re("\\b(?:true|false|null|yes|no|on|off)\\b", "doc-hl-bool");
    h.re("\\b-?\\d+(?:\\.\\d+)?\\b", "doc-hl-num");
    return h.finish();
  }

  function highlightDocJSON(text) {
    var h = createHighlighter(text);
    h.re("(\"(?:\\\\.|[^\"\\\\])*\")\\s*(?=:)", "doc-hl-key");
    h.re("(\"(?:\\\\.|[^\"\\\\])*\")", "doc-hl-string");
    h.re("\\b(?:true|false|null)\\b", "doc-hl-bool");
    h.re("\\b-?\\d+(?:\\.\\d+)?\\b", "doc-hl-num");
    return h.finish();
  }

  function highlightDocGo(text) {
    var h = createHighlighter(text);
    h.re("//.*$", "doc-hl-comment", "gm");
    h.re("/\\*[\\s\\S]*?\\*/", "doc-hl-comment");
    h.re("`(?:\\\\.|[^`\\\\])*`", "doc-hl-string");
    h.re("(\"(?:\\\\.|[^\"\\\\])*\")", "doc-hl-string");
    h.re(
      "\\b(?:package|import|func|var|const|type|struct|interface|map|chan|go|defer|return|if|else|for|range|switch|case|default|break|continue|select|fallthrough|nil)\\b",
      "doc-hl-keyword"
    );
    h.re("\\b(?:int|string|bool|byte|rune|float\\d+|uint\\d+|error|any)\\b", "doc-hl-type");
    h.re("\\b[A-Z]\\w*\\b", "doc-hl-type");
    h.re("\\b\\d+(?:\\.\\d+)?\\b", "doc-hl-num");
    return h.finish();
  }

  function highlightDocPython(text) {
    var h = createHighlighter(text);
    h.re("#.*$", "doc-hl-comment", "gm");
    h.re("(\"\"\"[\\s\\S]*?\"\"\"|'''[\\s\\S]*?'''|\"(?:\\\\.|[^\"\\\\])*\"|'(?:\\\\.|[^'\\\\])*')", "doc-hl-string");
    h.re(
      "\\b(?:def|class|import|from|as|if|elif|else|for|while|return|yield|with|try|except|finally|raise|pass|break|continue|lambda|global|nonlocal|True|False|None|and|or|not|in|is|async|await)\\b",
      "doc-hl-keyword"
    );
    h.re("^\\s*@[\\w.]+", "doc-hl-func", "gm");
    h.re("\\b\\d+(?:\\.\\d+)?\\b", "doc-hl-num");
    return h.finish();
  }

  function highlightDocXML(text) {
    var h = createHighlighter(text);
    h.re("<!--[\\s\\S]*?-->", "doc-hl-comment");
    h.re("</?[\\w:-]+", "doc-hl-tag");
    h.re("\\b[\\w:-]+(?=\\s*=)", "doc-hl-attr");
    h.re("(\"(?:\\\\.|[^\"\\\\])*\"|'(?:\\\\.|[^'\\\\])*')", "doc-hl-string");
    return h.finish();
  }

  function highlightDocCode(text, lang) {
    switch (lang) {
      case "yaml":
        return highlightDocYAML(text);
      case "json":
        return highlightDocJSON(text);
      case "go":
        return highlightDocGo(text);
      case "python":
        return highlightDocPython(text);
      case "xml":
        return highlightDocXML(text);
      default:
        return highlightDocShell(text);
    }
  }

  /** @deprecated use highlightDocShell */
  function highlightDocTerminal(text) {
    return highlightDocShell(text);
  }

  function docCopyText(t) {
    if (!t) return Promise.reject(new Error("empty"));
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(t);
    }
    return new Promise(function (ok, no) {
      var ta = document.createElement("textarea");
      ta.value = t;
      ta.setAttribute("readonly", "");
      ta.style.position = "fixed";
      ta.style.left = "-9999px";
      document.body.appendChild(ta);
      ta.select();
      try {
        if (document.execCommand("copy")) ok();
        else no(new Error("copy"));
      } catch (e) {
        no(e);
      } finally {
        document.body.removeChild(ta);
      }
    });
  }

  function initDocCodeChrome(root) {
    if (!root) return;
    root.querySelectorAll("pre").forEach(function (pre) {
      if (pre.closest(".doc-code-wrap")) return;
      if (pre.closest(".doc-mermaid-wrap")) return;
      var code = pre.querySelector("code");
      var text = (code || pre).textContent || "";
      var lang = detectCodeLang(code);
      // Mermaid blocks are rendered by doc-mermaid.js — do not shell-highlight.
      if (lang === "mermaid") return;
      var html = highlightDocCode(text, lang);
      if (code) {
        code.innerHTML = html;
      } else {
        pre.innerHTML = html;
      }
      var wrap = document.createElement("div");
      wrap.className = "doc-code-wrap";
      if (lang) {
        wrap.classList.add("doc-code-lang-" + lang);
        wrap.setAttribute("data-lang", lang);
      } else if (isShellLang(lang)) {
        wrap.classList.add("doc-code-lang-shell");
        wrap.setAttribute("data-lang", "shell");
      }
      var btn = document.createElement("button");
      btn.type = "button";
      btn.className = "doc-copy-btn";
      btn.textContent = "复制";
      btn.title = "复制代码";
      btn.setAttribute("aria-label", "复制代码");
      btn.addEventListener("click", function (ev) {
        ev.preventDefault();
        docCopyText(text)
          .then(function () {
            btn.textContent = "已复制";
            setTimeout(function () {
              btn.textContent = "复制";
            }, 1600);
          })
          .catch(function () {
            btn.textContent = "失败";
            setTimeout(function () {
              btn.textContent = "复制";
            }, 1600);
          });
      });
      pre.parentNode.insertBefore(wrap, pre);
      wrap.appendChild(btn);
      wrap.appendChild(pre);
    });
  }

  global.initDocCodeChrome = initDocCodeChrome;
  global.highlightDocTerminal = highlightDocTerminal;
  global.highlightDocShell = highlightDocShell;
  global.highlightDocCode = highlightDocCode;

  document.addEventListener("DOMContentLoaded", function () {
    function run() {
      document.querySelectorAll(".doc-md").forEach(initDocCodeChrome);
    }
    if (typeof requestIdleCallback === "function") {
      requestIdleCallback(run, { timeout: 2000 });
    } else {
      setTimeout(run, 0);
    }
  });
})(typeof window !== "undefined" ? window : globalThis);

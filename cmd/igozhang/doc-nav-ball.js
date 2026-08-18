(function (global) {
  "use strict";

  var MOBILE_MQ = "(max-width: 42rem)";

  function escAttr(s) {
    return String(s || "")
      .replace(/&/g, "&amp;")
      .replace(/"/g, "&quot;")
      .replace(/</g, "&lt;");
  }

  function homeURL(pub) {
    var root = (pub || "").replace(/\/$/, "");
    var u = root ? root + "/" : "/";
    try {
      var cat = new URL(global.location.href).searchParams.get("category");
      if (cat) u += "?category=" + encodeURIComponent(cat);
    } catch (e) {}
    return u;
  }

  function smoothScrollTo(y) {
    var top = Math.max(0, y | 0);
    try {
      global.scrollTo({ top: top, behavior: "smooth" });
    } catch (e) {
      global.scrollTo(0, top);
    }
  }

  function scrollMaxY() {
    var h = Math.max(
      global.document.documentElement.scrollHeight || 0,
      global.document.body.scrollHeight || 0
    );
    return Math.max(0, h - global.innerHeight);
  }

  function navBallHTML(homeHref) {
    return (
      '<nav class="doc-nav-ball" aria-label="页面导航">' +
      '<button type="button" class="doc-nav-ball-half doc-nav-ball-top" aria-label="回到顶部">' +
      '<span class="doc-nav-ball-icon" aria-hidden="true">↑</span></button>' +
      '<a class="doc-nav-ball-home" href="' +
      escAttr(homeHref) +
      '" aria-label="回到首页">home</a>' +
      '<button type="button" class="doc-nav-ball-half doc-nav-ball-bottom" aria-label="回到底部">' +
      '<span class="doc-nav-ball-icon" aria-hidden="true">↓</span></button>' +
      "</nav>"
    );
  }

  function bindNavBall(ball, pub) {
    var topBtn = ball.querySelector(".doc-nav-ball-top");
    var botBtn = ball.querySelector(".doc-nav-ball-bottom");
    var homeLink = ball.querySelector(".doc-nav-ball-home");
    if (topBtn) {
      topBtn.addEventListener("click", function () {
        smoothScrollTo(0);
      });
    }
    if (botBtn) {
      botBtn.addEventListener("click", function () {
        smoothScrollTo(scrollMaxY());
      });
    }
    if (homeLink) {
      homeLink.addEventListener("click", function (e) {
        var target = homeURL(pub);
        var here = global.location.pathname.replace(/\/$/, "") || "/";
        var there = target.split("?")[0].replace(/\/$/, "") || "/";
        if (here === there && !global.location.search && target.indexOf("?") < 0) {
          e.preventDefault();
          smoothScrollTo(0);
          return;
        }
        e.preventDefault();
        if (global.__fsMp3FlushHandoff) global.__fsMp3FlushHandoff();
        global.location.href = target;
      });
    }
  }

  function cleanupNavOnlyAside(aside) {
    if (
      aside &&
      aside.nodeType === 1 &&
      aside.classList.contains("doc-link-aside-nav-only") &&
      !aside.querySelector(".doc-link-capsule") &&
      !aside.querySelector(".doc-nav-ball")
    ) {
      aside.remove();
    }
  }

  function ensureAside(grid) {
    var aside = grid.querySelector(".doc-link-aside");
    if (aside) return aside;
    aside = global.document.createElement("aside");
    aside.className = "doc-link-aside doc-link-aside-nav-only";
    aside.setAttribute("aria-label", "页面导航");
    grid.appendChild(aside);
    return aside;
  }

  function mountFixed(wrap) {
    wrap.classList.add("doc-nav-ball-wrap-fixed");
    if (wrap.parentNode !== global.document.body) {
      var oldAside = wrap.parentNode;
      global.document.body.appendChild(wrap);
      cleanupNavOnlyAside(oldAside);
    }
  }

  function mountAside(wrap, grid) {
    wrap.classList.remove("doc-nav-ball-wrap-fixed");
    var aside = ensureAside(grid);
    if (wrap.parentNode !== aside) {
      aside.appendChild(wrap);
    }
  }

  function mount(pub) {
    var ops = global.__fsDocNavBallOps || {};
    if (ops.enabled === false) return;
    if (global.document.querySelector(".doc-nav-ball")) return;

    var grid = global.document.querySelector(".doc-home-grid");
    var wrap = global.document.createElement("div");
    wrap.className = "doc-nav-ball-wrap";
    wrap.innerHTML = navBallHTML(homeURL(pub));
    var ball = wrap.querySelector(".doc-nav-ball");
    if (!ball) return;

    var mq = global.matchMedia(MOBILE_MQ);
    function place() {
      if (grid && !mq.matches) {
        mountAside(wrap, grid);
        return;
      }
      mountFixed(wrap);
    }

    place();
    if (mq.addEventListener) mq.addEventListener("change", place);
    else if (mq.addListener) mq.addListener(place);

    bindNavBall(ball, pub);
  }

  global.__fsDocNavBallInit = function (pub) {
    mount(pub || "");
  };
})(typeof window !== "undefined" ? window : this);

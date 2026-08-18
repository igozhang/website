window.__igoAuthInit=function(opts){
  opts=opts||{};
  var apiBase=opts.apiBase||"";
  var hideMode=opts.hideMode||"class"; // "class" | "attr"
  var token=null, expiresAt=0, expiryTimer=null, sessionTimer=null, readyOnce=false;
  try{ localStorage.removeItem("igoadminToken"); }catch(e){}
  try{ localStorage.removeItem("fsAuthToken"); }catch(e){}
  var authSec=document.getElementById("igoAuthSec");
  var mainSec=document.getElementById("igoMainSec");
  var keyEl=document.getElementById("igoKey");
  var authMsg=document.getElementById("igoAuthMsg");
  var expiryEl=document.getElementById("igoKeyExpiry");
  var mainKeyExpiryEl=document.getElementById("igoMainKeyExpiry");
  var btn=document.getElementById("btnIgoAuth");
  function api(p){ return apiBase+p; }
  function setHidden(el, hidden){
    if(!el) return;
    if(hideMode==="attr"){
      el.hidden=!!hidden;
      el.classList.remove("hidden");
    } else {
      if(hidden) el.classList.add("hidden");
      else el.classList.remove("hidden");
      el.hidden=false;
    }
  }
  function showAuthed(on){
    setHidden(authSec, !!on);
    setHidden(mainSec, !on);
  }
  function setAuthMsg(t,ok){
    if(!authMsg) return;
    authMsg.textContent=t||"";
    var base=opts.msgClass||"igo-auth-msg";
    authMsg.className=base+(ok===true?" is-ok":ok===false?" is-err":"");
  }
  function parseJSONResponse(r){
    return r.json().then(function(j){ return {r:r,j:j}; }).catch(function(){ return {r:r,j:null}; });
  }
  function fmtBJ(ts){
    if(window.__fsFmtBeijing) return window.__fsFmtBeijing(ts);
    var ex=Number(window.__fsBeijingOffsetHours); if(!isFinite(ex)) ex=0; if(ex>12) ex=12; if(ex<-12) ex=-12; ex=ex|0;
    var u=new Date(ts*1000),ms=Date.UTC(u.getUTCFullYear(),u.getUTCMonth(),u.getUTCDate(),u.getUTCHours()+8+ex,u.getUTCMinutes(),u.getUTCSeconds());
    var b=new Date(ms),p=function(n){return(n<10?"0":"")+n};
    return b.getUTCFullYear()+"-"+p(b.getUTCMonth()+1)+"-"+p(b.getUTCDate())+" "+p(b.getUTCHours())+":"+p(b.getUTCMinutes())+":"+p(b.getUTCSeconds());
  }
  function logout(reason){
    token=null; expiresAt=0;
    if(expiryTimer){ clearInterval(expiryTimer); expiryTimer=null; }
    if(sessionTimer){ clearInterval(sessionTimer); sessionTimer=null; }
    showAuthed(false);
    if(keyEl) keyEl.value="";
    if(expiryEl){ expiryEl.textContent=""; expiryEl.className="igo-auth-expiry"; }
    if(mainKeyExpiryEl){ mainKeyExpiryEl.textContent=""; mainKeyExpiryEl.className="igo-auth-expiry"; }
    if(typeof opts.onLogout==="function"){
      try{ opts.onLogout(reason||""); }catch(e){}
    }
    if(reason) setAuthMsg(reason,false);
  }
  function handleAuthFailure(status,j){
    var reason=j&&j.reason;
    if(status===401||reason==="key_revoked"||reason==="key_expired"){
      logout((j&&j.error)||"请重新验证");
      return true;
    }
    return false;
  }
  function requireAuth(){
    if(!token){ logout("未认证，请先验证 key"); return false; }
    if(expiresAt&&Date.now()>=expiresAt*1000){ logout("key 已过期，请重新验证"); return false; }
    return true;
  }
  function authHeaders(extra){
    var h=extra||{};
    if(token) h["X-Igo-Token"]=token;
    return h;
  }
  function withAuthQuery(url){
    if(!token||!url) return url||"";
    return url+(url.indexOf("?")>=0?"&":"?")+"token="+encodeURIComponent(token);
  }
  function getToken(){ return token; }
  function checkSession(){
    if(!token) return;
    fetch(api("/api/auth/session"),{
      method:"POST",
      credentials:"same-origin",
      headers:{"Content-Type":"application/json"},
      body:JSON.stringify({token:token})
    }).then(parseJSONResponse).then(function(x){
      if(handleAuthFailure(x.r.status,x.j)) return;
      if(x.r.ok&&x.j&&x.j.ok&&typeof x.j.expiresInSec==="number"){
        expiresAt=Math.floor(Date.now()/1000)+Math.max(0,x.j.expiresInSec);
        updateExpiryLabel();
      }
    }).catch(function(){});
  }
  function updateExpiryLabel(){
    if(!expiresAt) return;
    var left=expiresAt-Math.floor(Date.now()/1000);
    if(left<=0){ logout("key 已过期，请重新验证"); return; }
    var mins=Math.max(0,Math.ceil(left/60));
    var text="密钥有效期至 "+fmtBJ(expiresAt)+"（北京时间，剩余 "+mins+" 分钟）";
    var warn=mins<5;
    var cls="igo-auth-expiry"+(warn?" is-warn":"");
    if(expiryEl){ expiryEl.textContent=text; expiryEl.className=cls; }
    if(mainKeyExpiryEl){ mainKeyExpiryEl.textContent=text; mainKeyExpiryEl.className=cls; }
  }
  function startExpiryWatch(){
    updateExpiryLabel();
    if(expiryTimer) clearInterval(expiryTimer);
    expiryTimer=setInterval(updateExpiryLabel,1000);
    checkSession();
    if(sessionTimer) clearInterval(sessionTimer);
    sessionTimer=setInterval(checkSession,15000);
  }
  function fireReady(){
    if(typeof opts.onReady==="function"){
      try{ opts.onReady({first:!readyOnce}); }catch(e){}
    }
    readyOnce=true;
  }
  if(btn) btn.onclick=function(){
    var k=(keyEl&&keyEl.value||"").trim();
    if(!k){ setAuthMsg("请输入密钥",false); return; }
    setAuthMsg("验证中…",true);
    fetch(api("/api/auth/login"),{
      method:"POST",
      credentials:"same-origin",
      headers:{"Content-Type":"application/json"},
      body:JSON.stringify({key:k})
    }).then(parseJSONResponse).then(function(x){
      if(x.r.status===429){
        var sec=(x.j&&x.j.retryAfterSec)||60;
        setAuthMsg("验证尝试过于频繁，请 "+sec+" 秒后再试",false);
        return;
      }
      if(x.r.ok&&x.j&&x.j.ok&&x.j.token){
        token=x.j.token;
        var sec=x.j.expiresInSec;
        if(typeof sec!=="number"||sec<0) sec=0;
        expiresAt=sec>0?Math.floor(Date.now()/1000)+sec:0;
        setAuthMsg("验证通过",true);
        showAuthed(true);
        startExpiryWatch();
        fireReady();
        return;
      }
      setAuthMsg((x.j&&x.j.error)||"验证失败",false);
    }).catch(function(e){
      setAuthMsg("请求失败："+(e&&e.message||e),false);
    });
  };
  if(keyEl) keyEl.addEventListener("keydown",function(ev){
    if(ev.key==="Enter"){ ev.preventDefault(); if(btn) btn.click(); }
  });
  showAuthed(false);
  return {
    getToken:getToken,
    authHeaders:authHeaders,
    withAuthQuery:withAuthQuery,
    requireAuth:requireAuth,
    handleAuthFailure:handleAuthFailure,
    logout:logout,
    api:api
  };
};

// 首页与子页唯一 MP3 逻辑（//go:embed）。双槽连播状态表见 checklist.md §3。
// | 页可见 | 有预取且副槽可播 | tryPlayPrefetched → dualSlot；hidden 时主槽 fast |
// | 页可见 | 无预取 | playTrackWithDualBuffer → dualSlotPlaySecondaryBuffer |
// | 页隐藏 | 有/无预取 | playTrackWithDualBuffer 失败 → playSpecificMp3OnMain(fast) |
window.__fsMp3Init=function(base){
  function absRoot(){
    if(typeof window.__fsAbsRoot==="function") return window.__fsAbsRoot(base);
    if(base!==undefined&&base!==null&&String(base).length) return String(base).replace(/\/$/,"");
    return "";
  }
  function urlBrowse(q){
    return absRoot()+"/api/browse?path="+encodeURIComponent(q);
  }
  function urlStreamMp3(fileName){
    return absRoot()+"/stream/mp3/"+fileName.split("/").filter(Boolean).map(encodeURIComponent).join("/");
  }
  function urlMp3Cover(fileName){
    var q=fileName.split("/").filter(Boolean).join("/");
    return absRoot()+"/api/mp3/cover?path="+encodeURIComponent(q);
  }
  function urlMp3Gate(){
    return absRoot()+"/api/mp3/gate";
  }
  var mp3GatePromise=null;
  var mp3GateOkUntil=0;
  // Cookie 默认 12h；页内缓存约 50 分钟，避免长开标签页过期后播流 403。
  var MP3_GATE_REFRESH_MS=50*60*1000;
  function ensureMp3Gate(){
    var now=Date.now();
    if(mp3GatePromise && mp3GateOkUntil && now<mp3GateOkUntil) return mp3GatePromise;
    mp3GatePromise=fetch(urlMp3Gate(),{credentials:"same-origin",cache:"no-store"}).then(function(r){
      if(!r.ok) throw new Error("mp3 gate "+r.status);
      mp3GateOkUntil=Date.now()+MP3_GATE_REFRESH_MS;
      return true;
    }).catch(function(){
      mp3GatePromise=null;
      mp3GateOkUntil=0;
      return false;
    });
    return mp3GatePromise;
  }
  function fsFaviconFromMeta(name,fallback){
    var m=document.querySelector('meta[name="'+name+'"]');
    if(m&&m.content) return m.content;
    return fallback;
  }
  function urlSiteFavicon(){
    return fsFaviconFromMeta("fs-favicon-href", absRoot()+"/favicon.svg");
  }
  function urlSiteFaviconMedia(){
    return fsFaviconFromMeta("fs-favicon-media-href", absRoot()+"/favicon.png");
  }
  function absolutePageUrl(pathFromSiteRoot){
    if(typeof location==="undefined"||!pathFromSiteRoot) return "";
    var origin=String(location.origin||"").replace(/\/$/,"");
    var p=pathFromSiteRoot.charAt(0)==="/"?pathFromSiteRoot:"/"+pathFromSiteRoot;
    try{ return new URL(p,origin+"/").href; }catch(e){ return origin+p; }
  }
  function showToast(text,isErr){
    var w=document.getElementById("toastWrap");
    if(!w||!text) return;
    var el=document.createElement("div");
    el.className="toast "+(isErr?"toast-err":"toast-ok");
    el.textContent=text;
    w.appendChild(el);
    setTimeout(function(){ try{ el.remove(); }catch(e){} }, 4200);
  }
  var mp3PlayErrToastT=null;
  function mp3NotifyPlayBlocked(){
    if(mp3PlayErrToastT) return;
    mp3PlayErrToastT=setTimeout(function(){ mp3PlayErrToastT=null; }, 2500);
    showToast("无法开始播放（请再点一次播放）",true);
  }
  function mp3PlayMaybeToast(p){
    if(p&&typeof p.catch==="function") p.catch(function(){ mp3NotifyPlayBlocked(); });
    return p;
  }
  var mp3List=[];
  var mp3Current="";
  var mp3ListLoading=false;
  var mp3ListLoaded=false;
  var mp3ListLoadQueue=[];
  var mp3LockScreenTrack="";
  var mp3FilterTag=null;
  var mp3SessionPlaying=false;
  var mp3PickToken=0;
  var mp3LoadErrors=0;
  var mp3MetaSeq=0;
  var mp3TopTags=[];
  var mp3RecentTwo=[];
  var mp3PastStack=[];
  var mp3SlotA=document.getElementById("mp3Audio");
  var mp3SlotB=document.getElementById("mp3Prefetch");
  if(!mp3SlotA||!mp3SlotB){
    return;
  }
  var mp3Audio=mp3SlotA;
  var mp3Other=mp3SlotB;
  var prefetchName="";
  var prefetchToken=0;
  function swapMp3Outputs(){
    var t=mp3Audio; mp3Audio=mp3Other; mp3Other=t;
  }
  function mp3PageIsHidden(){
    if(typeof document==="undefined") return false;
    return !!(document.hidden||document.visibilityState==="hidden");
  }
  function configureBufferRemoteMuted(){
    if(mp3Other){
      mp3Other.muted=true;
      if("disableRemotePlayback" in mp3Other) mp3Other.disableRemotePlayback=true;
    }
    if(mp3Audio){
      mp3Audio.muted=false;
      if("disableRemotePlayback" in mp3Audio) mp3Audio.disableRemotePlayback=false;
    }
  }
  if(mp3SlotB){
    mp3SlotB.muted=true;
    if("disableRemotePlayback" in mp3SlotB) mp3SlotB.disableRemotePlayback=true;
  }
  if(mp3SlotA&&"disableRemotePlayback" in mp3SlotA) mp3SlotA.disableRemotePlayback=false;
  var mediaSessionBgTimer=null;
  var lastMsPosSync=0;
  var lastBgPositionPushMs=0;
  var mp3LoadingTrack=false;
  var mp3TrackChanging=false;
  var mp3ResumeTimers=[];
  function clearPlaybackResumeTimers(){
    mp3ResumeTimers.forEach(function(id){ clearTimeout(id); });
    mp3ResumeTimers=[];
  }
  function shouldBlockMp3AutoResume(){
    return mp3DeferToOtherAudio;
  }
  function clearMp3AudioDefer(){
    mp3DeferToOtherAudio=false;
  }
  function schedulePlaybackResume(){
    if(shouldBlockMp3AutoResume()) return;
    if(typeof document!=="undefined"&&document.visibilityState!=="visible") return;
    if(!mp3SessionPlaying||!mp3Audio||!mp3Audio.src||mp3Audio.ended) return;
    if(!mp3Audio.paused) return;
    if(mp3HandoffResumePending&&handoffResumeDeadlinePassed()){ abortHandoffAutoResume(); return; }
    clearPlaybackResumeTimers();
    [50,200,500,1200,2200].forEach(function(ms){
      var id=setTimeout(function(){
        if(mp3HandoffResumePending&&handoffResumeDeadlinePassed()){ abortHandoffAutoResume(); return; }
        if(!mp3SessionPlaying||!mp3Audio||mp3Audio.ended) return;
        if(!mp3Audio.paused) return;
        mp3PlayMaybeToast(mp3Audio.play());
      }, ms);
      mp3ResumeTimers.push(id);
    });
  }
  function applyAudioSessionPlayback(){
    try{
      if(!navigator.audioSession) return;
      if(!audioSessionCoexistBound&&typeof navigator.audioSession.addEventListener==="function"){
        audioSessionCoexistBound=true;
        navigator.audioSession.addEventListener("statechange", function(){
          try{
            var st=navigator.audioSession.state;
            if(st==="interrupted"){
              mp3DeferToOtherAudio=true;
              clearPlaybackResumeTimers();
              if(mp3Audio&&!mp3Audio.paused){
                try{ mp3Audio.pause(); }catch(e){}
              }
            }else if(st==="active"&&!mp3UserPaused){
              mp3DeferToOtherAudio=false;
            }
          }catch(e){}
        });
      }
      if("type" in navigator.audioSession){
        navigator.audioSession.type="ambient";
      }
    }catch(e){}
  }
  var mp3AttachFetchCtl=null;
  var lastMsMetaKey="";
  var mp3AttachGuardTimer=null;
  var _mp3Ops=(typeof window!=="undefined"&&window.__fsMp3Ops)||{};
  var MP3_MIN_BUFFER_BYTES=_mp3Ops.minBufferBytes||512*1024;
  var MP3_PREFETCH_AHEAD_SEC=_mp3Ops.prefetchAheadSec||10;
  var mp3NextMetaCache=null;
  var lastPrefetchScheduleMs=0;
  function mp3AllowAudioPrefetch(){
    try{
      var c=navigator.connection||navigator.mozConnection||navigator.webkitConnection;
      if(!c) return true;
      if(c.saveData) return false;
      var t=String(c.effectiveType||"").toLowerCase();
      if(t==="slow-2g"||t==="2g") return false;
    }catch(e){}
    return true;
  }
  var lastPlaybackAdvanceMs=0;
  var lastPlaybackTime=0;
  var mp3StallRetryCount=0;
  var MP3_STALL_MS=_mp3Ops.stallMs||4500;
  var MP3_MAX_STALL_RETRIES=_mp3Ops.maxStallRetries!=null?_mp3Ops.maxStallRetries:4;
  var mp3BgHintShown=false;
  var mp3CoverProbeCache={};
  var mp3DeferToOtherAudio=false;
  var mp3UserPaused=false;
  var MP3_HANDOFF_KEY="fsMp3Handoff";
  var MP3_HANDOFF_MAX_MS=45000;
  var MP3_HANDOFF_RESUME_MAX_MS=12000;
  var mp3HandoffPending=null;
  var mp3HandoffFlushLastMs=0;
  var mp3HandoffGestureBound=false;
  var mp3HandoffGestureOnce=null;
  var mp3HandoffResumePending=false;
  var mp3HandoffResumeStartedMs=0;
  var mp3HandoffGiveUpTimer=null;
  function clearHandoffGiveUpTimer(){
    if(mp3HandoffGiveUpTimer){ clearTimeout(mp3HandoffGiveUpTimer); mp3HandoffGiveUpTimer=null; }
  }
  function handoffResumeDeadlinePassed(){
    return mp3HandoffResumeStartedMs>0&&Date.now()-mp3HandoffResumeStartedMs>=MP3_HANDOFF_RESUME_MAX_MS;
  }
  function unbindHandoffGestureResume(){
    if(mp3HandoffGestureOnce){
      var once=mp3HandoffGestureOnce;
      mp3HandoffGestureOnce=null;
      document.removeEventListener("touchstart", once, true);
      document.removeEventListener("touchend", once, true);
      document.removeEventListener("pointerdown", once, true);
      document.removeEventListener("click", once, true);
    }
    mp3HandoffGestureBound=false;
  }
  function clearHandoffResumeAttempt(){
    mp3HandoffResumePending=false;
    mp3HandoffResumeStartedMs=0;
    clearHandoffGiveUpTimer();
    unbindHandoffGestureResume();
  }
  function abortHandoffAutoResume(){
    if(!mp3HandoffResumePending) return;
    if(mp3Audio&&!mp3Audio.paused&&!mp3Audio.ended){
      clearHandoffResumeAttempt();
      clearPlaybackResumeTimers();
      return;
    }
    clearHandoffResumeAttempt();
    clearPlaybackResumeTimers();
    if(mp3Audio&&mp3Audio.src&&!mp3Audio.ended){
      try{ mp3Audio.pause(); }catch(e){}
    }
    mp3SessionPlaying=false;
    mp3UserPaused=true;
    syncMediaSessionPlaybackState();
    setPlayLabel();
  }
  function armHandoffResumeDeadline(){
    clearHandoffGiveUpTimer();
    if(!mp3HandoffResumePending||!mp3HandoffResumeStartedMs) return;
    if(handoffResumeDeadlinePassed()){ abortHandoffAutoResume(); return; }
    var remain=MP3_HANDOFF_RESUME_MAX_MS-(Date.now()-mp3HandoffResumeStartedMs);
    mp3HandoffGiveUpTimer=setTimeout(abortHandoffAutoResume, remain);
  }
  function scheduleHandoffPlaybackResume(){
    if(shouldBlockMp3AutoResume()) return;
    if(mp3HandoffResumePending&&handoffResumeDeadlinePassed()){ abortHandoffAutoResume(); return; }
    schedulePlaybackResume();
    armHandoffResumeDeadline();
    if(!mp3LikelyMobile()) return;
    [2800,5500,9000,11000].forEach(function(ms){
      if(ms>=MP3_HANDOFF_RESUME_MAX_MS) return;
      var id=setTimeout(function(){
        if(handoffResumeDeadlinePassed()){ abortHandoffAutoResume(); return; }
        if(!mp3HandoffResumePending||!mp3SessionPlaying||!mp3Audio||mp3Audio.ended||!mp3Audio.paused) return;
        var pr=mp3Audio.play();
        if(pr&&typeof pr.catch==="function"){
          pr.catch(function(){
            if(handoffResumeDeadlinePassed()){ abortHandoffAutoResume(); return; }
            bindHandoffGestureResume();
          });
        }
      }, ms);
      mp3ResumeTimers.push(id);
    });
  }
  function mp3LikelyMobile(){
    if(typeof navigator==="undefined") return false;
    var ua=navigator.userAgent||"";
    if(/Android|iPhone|iPad|iPod|Mobile/i.test(ua)) return true;
    try{
      if(window.matchMedia&&window.matchMedia("(pointer:coarse)").matches) return true;
    }catch(e){}
    return false;
  }
  function buildMp3HandoffPayload(){
    var ct=0;
    if(mp3Audio&&mp3Audio.src&&!mp3Audio.ended){
      ct=mp3Audio.currentTime||0;
    }
    return {
      v:1,
      ts:Date.now(),
      current:mp3Current||"",
      currentTime:ct,
      sessionPlaying:!!mp3SessionPlaying,
      userPaused:!!mp3UserPaused,
      filterTag:mp3FilterTag||"",
      loopOn:!!(mp3Audio&&mp3Audio.loop),
      recentTwo:mp3RecentTwo.slice(0,2),
      pastStack:mp3PastStack.slice(-24),
      pickToken:mp3PickToken,
      lockScreen:mp3LockScreenTrack||mp3Current||"",
      prefetchName:(prefetchName&&prefetchToken===mp3PickToken)?prefetchName:""
    };
  }
  function validMp3Handoff(o){
    if(!o||o.v!==1||!o.current) return false;
    if(Date.now()-(o.ts||0)>MP3_HANDOFF_MAX_MS) return false;
    return true;
  }
  function peekMp3HandoffStorage(){
    try{
      var raw=sessionStorage.getItem(MP3_HANDOFF_KEY);
      if(!raw) return null;
      return JSON.parse(raw);
    }catch(e){ return null; }
  }
  function clearMp3HandoffStorage(){
    try{ sessionStorage.removeItem(MP3_HANDOFF_KEY); }catch(e){}
  }
  function handoffSnapshotWins(a, b){
    if(!b) return !!a;
    if(!a) return false;
    if(a.sessionPlaying&&a.current&&(!b.sessionPlaying||!b.current)) return true;
    if(b.sessionPlaying&&b.current&&(!a.sessionPlaying||!a.current)) return false;
    return (a.ts||0)>=(b.ts||0);
  }
  function pickMp3HandoffCandidate(cached){
    var best=validMp3Handoff(cached)?cached:null;
    var stored=peekMp3HandoffStorage();
    if(validMp3Handoff(stored)){
      if(!best||handoffSnapshotWins(stored, best)) best=stored;
    }
    return best;
  }
  function flushMp3Handoff(opts){
    opts=opts||{};
    var next=buildMp3HandoffPayload();
    if(!next.current&&!next.sessionPlaying&&!(mp3Audio&&mp3Audio.src)) return;
    try{
      if(!opts.force){
        var prev=peekMp3HandoffStorage();
        if(prev&&prev.v===1){
          if(opts.fromHide){
            if(prev.sessionPlaying&&prev.current&&(Date.now()-(prev.ts||0)<5000)){
              if(!next.sessionPlaying||!next.current) return;
              if(handoffSnapshotWins(prev, next)) return;
            }
          }else if(handoffSnapshotWins(prev, next)) return;
        }
      }
      sessionStorage.setItem(MP3_HANDOFF_KEY, JSON.stringify(next));
    }catch(e){}
  }
  function maybeFlushMp3HandoffSnap(){
    if(!mp3SessionPlaying||!mp3Current) return;
    var now=Date.now();
    if(now-mp3HandoffFlushLastMs<1800) return;
    mp3HandoffFlushLastMs=now;
    flushMp3Handoff();
  }
  function isSameOriginNavLink(a){
    if(!a||a.target==="_blank"||a.hasAttribute("download")) return false;
    var href=a.getAttribute("href");
    if(!href||href.charAt(0)==="#"||/^javascript:/i.test(href)) return false;
    try{
      var u=new URL(a.href, location.href);
      return u.origin===location.origin;
    }catch(e){ return false; }
  }
  function handoffFlushFromNavTarget(t){
    if(!t||!t.closest) return;
    var a=t.closest("a[href]");
    if(isSameOriginNavLink(a)) flushMp3Handoff({force:true});
  }
  function watchHandoffResumeFallback(){
    setTimeout(function(){
      if(!mp3HandoffResumePending||!mp3SessionPlaying||!mp3Audio||!mp3Audio.paused||mp3UserPaused||mp3Audio.ended) return;
      if(handoffResumeDeadlinePassed()){ abortHandoffAutoResume(); return; }
      bindHandoffGestureResume();
    }, 1500);
  }
  function bindHandoffGestureResume(){
    if(mp3HandoffGestureBound||!mp3SessionPlaying||!mp3Audio||!mp3Audio.paused||!mp3Audio.src||mp3Audio.ended) return;
    if(mp3HandoffResumePending&&handoffResumeDeadlinePassed()) return;
    mp3HandoffGestureBound=true;
    function once(){
      mp3HandoffGestureBound=false;
      mp3HandoffGestureOnce=null;
      document.removeEventListener("touchstart", once, true);
      document.removeEventListener("touchend", once, true);
      document.removeEventListener("pointerdown", once, true);
      document.removeEventListener("click", once, true);
      clearMp3AudioDefer();
      mp3UserPaused=false;
      clearHandoffResumeAttempt();
      if(!mp3SessionPlaying||!mp3Audio||!mp3Audio.src||mp3Audio.ended) return;
      mp3PlayMaybeToast(mp3Audio.play());
      syncMediaSessionPlaybackState();
      setPlayLabel();
    }
    mp3HandoffGestureOnce=once;
    document.addEventListener("touchstart", once, {capture:true, passive:true});
    document.addEventListener("touchend", once, {capture:true, passive:true});
    document.addEventListener("pointerdown", once, {capture:true, passive:true});
    document.addEventListener("click", once, true);
  }
  function writeMp3Handoff(ev){
    if(ev&&ev.persisted) return;
    flushMp3Handoff({fromHide:true});
  }
  function applyMp3HandoffMeta(o){
    if(Array.isArray(o.recentTwo)) mp3RecentTwo=o.recentTwo.slice(0,2);
    if(Array.isArray(o.pastStack)) mp3PastStack=o.pastStack.slice(-48);
    mp3FilterTag=o.filterTag||null;
    if(typeof o.pickToken==="number"&&o.pickToken>0) mp3PickToken=o.pickToken;
    mp3SessionPlaying=!!o.sessionPlaying;
    mp3UserPaused=!!o.userPaused;
    if(mp3Audio&&o.loopOn!==undefined) mp3Audio.loop=!!o.loopOn;
    syncLoopButtonUI();
  }
  function restoreMp3Handoff(o){
    if(!o||!o.current||mp3List.indexOf(o.current)<0) return false;
    applyMp3HandoffMeta(o);
    mp3LockScreenTrack=o.lockScreen||o.current;
    syncTagButtons();
    var seek=parseFloat(o.currentTime);
    if(!isFinite(seek)||seek<0) seek=0;
    var autoPlay=!!o.sessionPlaying&&!o.userPaused;
    mp3HandoffResumePending=autoPlay;
    mp3HandoffResumeStartedMs=autoPlay?Date.now():0;
    playSpecificMp3OnMain(o.current, autoPlay, true, seek, true, true);
    if(autoPlay){
      if(!mp3Audio.loop) schedulePrefetchNext(mp3PickToken);
      scheduleHandoffPlaybackResume();
      watchHandoffResumeFallback();
    }
    return true;
  }
  mp3HandoffPending=peekMp3HandoffStorage();
  var audioSessionCoexistBound=false;
  function abortMp3AttachFetch(){
    if(mp3AttachFetchCtl){
      try{ mp3AttachFetchCtl.abort(); }catch(e){}
      mp3AttachFetchCtl=null;
    }
  }
  function clearMp3AttachGuard(){
    if(mp3AttachGuardTimer){ clearTimeout(mp3AttachGuardTimer); mp3AttachGuardTimer=null; }
  }
  function estimatedBufferedBytes(audio, contentLength){
    if(!audio||!audio.buffered||!audio.buffered.length) return 0;
    var end=audio.buffered.end(audio.buffered.length-1);
    if(contentLength>0){
      var dur=audio.duration;
      if(isFinite(dur)&&dur>0) return Math.min(contentLength, Math.floor((end/dur)*contentLength));
    }
    return Math.floor(end*16000);
  }
  function attachMp3StreamWithMinBuffer(streamUrl, tok, thenFn){
    abortMp3AttachFetch();
    clearMp3AttachGuard();
    if(tok!==mp3PickToken) return;
    var ctl=null;
    try{
      if(typeof AbortController!=="undefined"){ ctl=new AbortController(); mp3AttachFetchCtl=ctl; }
    }catch(e){}
    var sig=ctl?ctl.signal:undefined;
    var contentLength=0;
    var started=false;
    var listeners=[];
    function cleanup(){
      clearMp3AttachGuard();
      for(var i=0;i<listeners.length;i++){
        mp3Audio.removeEventListener(listeners[i].ev, listeners[i].fn);
      }
      listeners.length=0;
    }
    function fireStart(){
      if(started||tok!==mp3PickToken) return;
      started=true;
      cleanup();
      thenFn();
    }
    function onProgress(){
      if(tok!==mp3PickToken||started) return;
      if(contentLength>0&&contentLength<=MP3_MIN_BUFFER_BYTES){
        if(mp3Audio.readyState>=2) fireStart();
        return;
      }
      if(estimatedBufferedBytes(mp3Audio, contentLength)>=MP3_MIN_BUFFER_BYTES) fireStart();
      else if(mp3Audio.readyState>=3) fireStart();
    }
    function addListen(ev, fn){
      listeners.push({ev:ev, fn:fn});
      mp3Audio.addEventListener(ev, fn);
    }
    function beginLoad(clen){
      contentLength=clen||0;
      if(tok!==mp3PickToken) return;
      ensureMp3Gate().then(function(ok){
        if(!ok||tok!==mp3PickToken) return;
        mp3Audio.preload="auto";
        mp3Audio.src=streamUrl;
        addListen("progress", onProgress);
        addListen("canplay", onProgress);
        addListen("loadedmetadata", onProgress);
        addListen("canplaythrough", function(){ fireStart(); });
        try{ mp3Audio.load(); }catch(e){}
        onProgress();
        mp3AttachGuardTimer=setTimeout(function(){
          if(tok===mp3PickToken&&!started&&mp3Audio.readyState>=2) fireStart();
        }, 15000);
      });
    }
    fetch(streamUrl,{method:"HEAD",credentials:"same-origin",signal:sig}).then(function(r){
      if(tok!==mp3PickToken) return Promise.reject("skip");
      var cl=0;
      if(r.ok){
        var h=r.headers.get("Content-Length");
        if(h) cl=parseInt(h,10)||0;
      }
      return cl;
    }).then(beginLoad).catch(function(){
      if(tok!==mp3PickToken) return;
      beginLoad(0);
    });
  }
  function attachMp3StreamFast(streamUrl, tok, thenFn){
    abortMp3AttachFetch();
    clearMp3AttachGuard();
    if(tok!==mp3PickToken) return;
    ensureMp3Gate().then(function(ok){
      if(!ok||tok!==mp3PickToken) return;
      mp3Audio.preload="auto";
      mp3Audio.src=streamUrl;
      var fired=false;
      function fire(){
        if(fired||tok!==mp3PickToken) return;
        fired=true;
        clearMp3AttachGuard();
        thenFn();
      }
      function onReady(){
        mp3Audio.removeEventListener("canplay", onReady);
        mp3Audio.removeEventListener("canplaythrough", onReady);
        fire();
      }
      mp3Audio.addEventListener("canplay", onReady);
      mp3Audio.addEventListener("canplaythrough", onReady);
      try{ mp3Audio.load(); }catch(e){}
      if(mp3Audio.readyState>=2) fire();
      else{
        mp3AttachGuardTimer=setTimeout(function(){
          if(tok===mp3PickToken&&mp3Audio.readyState>=2) fire();
          else if(tok===mp3PickToken) fire();
        }, 8000);
      }
    });
  }
  function shouldUseMinBufferAttach(autoPlay){
    if(!autoPlay) return false;
    if(mp3Audio&&mp3Audio.src) return false;
    return mp3PageIsHidden();
  }
  function shouldUseFastStreamAttach(autoPlay, fastStream){
    if(!autoPlay) return true;
    if(fastStream) return true;
    if(!shouldUseMinBufferAttach(autoPlay)) return true;
    return false;
  }
  function attachMp3StreamDirect(streamUrl, tok, thenFn){
    attachMp3StreamFast(streamUrl, tok, thenFn);
  }
  function attachMp3AudioSource(streamUrl, tok, thenFn){
    if(shouldUseFastStreamAttach(true, true)) attachMp3StreamFast(streamUrl, tok, thenFn);
    else attachMp3StreamWithMinBuffer(streamUrl, tok, thenFn);
  }
  function notePlaybackProgress(){
    if(!mp3Audio||mp3Audio.paused||!mp3SessionPlaying) return;
    var t=mp3Audio.currentTime;
    if(t>lastPlaybackTime+0.04){
      lastPlaybackTime=t;
      lastPlaybackAdvanceMs=Date.now();
      mp3StallRetryCount=0;
    }
  }
  function maybeEarlyPrefetch(){
    if(!mp3SessionPlaying||!mp3Audio||mp3Audio.loop||!mp3Other) return;
    if(!mp3AllowAudioPrefetch()) return;
    if(prefetchName&&prefetchToken===mp3PickToken) return;
    var d=mp3Audio.duration;
    if(!isFinite(d)||d<=0) return;
    var rem=d-mp3Audio.currentTime;
    if(rem>MP3_PREFETCH_AHEAD_SEC||rem<0) return;
    var now=Date.now();
    if(now-lastPrefetchScheduleMs<2000) return;
    lastPrefetchScheduleMs=now;
    schedulePrefetchNext(mp3PickToken);
  }
  function checkPlaybackStall(){
    if(shouldBlockMp3AutoResume()) return;
    if(!mp3SessionPlaying||!mp3Audio||mp3Audio.paused||mp3LoadingTrack||!mp3Audio.src) return;
    if(!lastPlaybackAdvanceMs) return;
    if(Date.now()-lastPlaybackAdvanceMs<MP3_STALL_MS) return;
    if(mp3StallRetryCount>=MP3_MAX_STALL_RETRIES) return;
    mp3StallRetryCount++;
    mp3PlayMaybeToast(mp3Audio.play());
    if(mp3Audio.paused) schedulePlaybackResumeAny();
  }
  function schedulePlaybackResumeAny(){
    if(shouldBlockMp3AutoResume()) return;
    if(mp3HandoffResumePending&&handoffResumeDeadlinePassed()){ abortHandoffAutoResume(); return; }
    if(mp3HandoffResumePending&&mp3SessionPlaying&&mp3Audio&&mp3Audio.paused&&!mp3Audio.ended&&!mp3UserPaused){
      scheduleHandoffPlaybackResume();
      return;
    }
    if(typeof document!=="undefined"&&mp3PageIsHidden()) scheduleHiddenPlaybackResume();
    else schedulePlaybackResume();
  }
  function scheduleHiddenPlaybackResume(){
    if(shouldBlockMp3AutoResume()) return;
    if(!mp3SessionPlaying||!mp3Audio||!mp3Audio.src||mp3Audio.ended) return;
    if(!mp3Audio.paused) return;
    if(mp3HandoffResumePending&&handoffResumeDeadlinePassed()){ abortHandoffAutoResume(); return; }
    clearPlaybackResumeTimers();
    [400,1200,2500,4000].forEach(function(ms){
      var id=setTimeout(function(){
        if(mp3HandoffResumePending&&handoffResumeDeadlinePassed()){ abortHandoffAutoResume(); return; }
        if(!mp3SessionPlaying||!mp3Audio||mp3Audio.ended||!mp3Audio.paused) return;
        mp3PlayMaybeToast(mp3Audio.play());
      }, ms);
      mp3ResumeTimers.push(id);
    });
  }
  var mp3PlayBtn=document.getElementById("mp3Play");
  var mp3LoopBtn=document.getElementById("mp3Loop");
  var mp3SearchInput=document.getElementById("mp3Search");
  var mp3SearchResults=document.getElementById("mp3SearchResults");
  var mp3SearchTimer=null;
  var mp3SearchIndex=[];
  var mp3SearchHits=[];
  var mp3SearchActiveIdx=-1;
  var mp3Heading=document.getElementById("mp3Heading");
  var mp3Title=document.getElementById("mp3Title");
  var mp3TagsEl=document.getElementById("mp3Tags");
  var mp3MetaEl=document.getElementById("mp3Meta");
  var mp3CoverEl=document.getElementById("mp3Cover");
  function mp3Haptic(){
    try{
      if(typeof navigator.vibrate==="function") navigator.vibrate(14);
    }catch(e){}
  }
  function bustAssetUrl(base){
    if(!base) return "";
    return base+(base.indexOf("?")>=0?"&":"?")+"_="+Date.now();
  }
  // 页内封面：仅 ID3 有图时展示；无封面不占位。锁屏/通知无封面用 fs-favicon-media-href（见 buildArtworkForTrack）。
  function syncMp3Cover(fileName, forceReload){
    if(!mp3MetaEl||!mp3CoverEl) return;
    var seq=++mp3MetaSeq;
    mp3MetaEl.classList.remove("mp3-meta-visible");
    mp3CoverEl.hidden=true;
    mp3CoverEl.onload=null;
    mp3CoverEl.onerror=null;
    mp3CoverEl.removeAttribute("src");
    if(!fileName) return;
    function showPageCover(coverPath){
      if(seq!==mp3MetaSeq||!coverPath) return;
      var coverUrl=forceReload?bustAssetUrl(coverPath):coverPath;
      mp3CoverEl.onload=function(){
        mp3CoverEl.onload=null;
        mp3CoverEl.onerror=null;
        if(seq!==mp3MetaSeq) return;
        mp3CoverProbeCache[fileName]=true;
        mp3CoverEl.hidden=false;
        mp3MetaEl.classList.add("mp3-meta-visible");
      };
      mp3CoverEl.onerror=function(){
        mp3CoverEl.onload=null;
        mp3CoverEl.onerror=null;
        if(seq!==mp3MetaSeq) return;
        mp3CoverProbeCache[fileName]=false;
        mp3CoverEl.hidden=true;
        mp3MetaEl.classList.remove("mp3-meta-visible");
        mp3CoverEl.removeAttribute("src");
        ensureTrackArtwork(fileName, {syncMs:true, force:true});
      };
      mp3CoverEl.src=coverUrl;
      if(mp3CoverEl.complete&&mp3CoverEl.naturalWidth>0){
        mp3CoverEl.onload=null;
        mp3CoverEl.onerror=null;
        mp3CoverProbeCache[fileName]=true;
        mp3CoverEl.hidden=false;
        mp3MetaEl.classList.add("mp3-meta-visible");
      }
    }
    probeMp3CoverThen(fileName, function(hasCover){
      if(seq!==mp3MetaSeq) return;
      ensureTrackArtwork(fileName, {syncMs:false});
      if(!hasCover) return;
      var coverPath=urlMp3Cover(fileName);
      if(coverPath) showPageCover(coverPath);
    });
  }
  function stripMp3Ext(fn){
    return String(fn).replace(/\.mp3$/i,"");
  }
  function syncTitleDisplay(name){
    if(!mp3Title) return;
    if(name){
      mp3Title.textContent=stripMp3Ext(name);
      mp3Title.classList.add("mp3-title-visible");
      syncMp3Cover(name);
    } else {
      mp3Title.textContent="";
      mp3Title.classList.remove("mp3-title-visible");
      syncMp3Cover("");
    }
  }
  function setMp3Heading(){
    if(mp3Heading) mp3Heading.textContent="MP3";
  }
  function isMp3Name(n){ return /\.mp3$/i.test(n); }
  function parseSong(fn){
    var base=fn.replace(/\.mp3$/i,"");
    var sep=" - ";
    var i=base.indexOf(sep);
    if(i<0) return {title:base,artist:""};
    return {title:base.slice(0,i).trim(),artist:base.slice(i+sep.length).trim()};
  }
  function tokenizeArtist(raw){
    if(!raw) return [];
    var s=raw.trim();
    var parts=s.split(/[\u3001\uFF0C,、\/\|]+/).map(function(x){return x.trim();}).filter(Boolean);
    if(!parts.length&&s) parts=[s];
    return parts;
  }
  function songMatchesFilter(fn){
    if(!mp3FilterTag) return true;
    var a=parseSong(fn).artist;
    if(!a) return false;
    var toks=tokenizeArtist(a);
    for(var i=0;i<toks.length;i++){ if(toks[i]===mp3FilterTag) return true; }
    return false;
  }
  function getMp3Pool(){
    return mp3List.filter(songMatchesFilter);
  }
  function noteTrackPlayed(name){
    if(!name) return;
    mp3RecentTwo=mp3RecentTwo.filter(function(x){ return x!==name; });
    mp3RecentTwo.unshift(name);
    if(mp3RecentTwo.length>2) mp3RecentTwo.length=2;
  }
  function pushMp3PastIfLeaving(fromName, toName){
    if(!fromName||!toName||fromName===toName) return;
    mp3PastStack.push(fromName);
    while(mp3PastStack.length>48) mp3PastStack.shift();
  }
  function filterPoolExcludeRecent(pool){
    if(!pool||!pool.length) return [];
    if(pool.length<=5) return pool.slice();
    var ex={};
    for(var i=0;i<mp3RecentTwo.length;i++){
      if(mp3RecentTwo[i]) ex[mp3RecentTwo[i]]=true;
    }
    var out=pool.filter(function(n){ return !ex[n]; });
    return out.length?out:pool.slice();
  }
  function cancelPrefetchMeta(){
    prefetchName="";
    prefetchToken=0;
  }
  function cancelPrefetch(){
    cancelPrefetchMeta();
    if(mp3Other){
      mp3Other.removeAttribute("src");
      try{ mp3Other.load(); }catch(e){}
    }
  }
  function pickPrefetchName(){
    var pool=getMp3Pool();
    if(!pool.length) return "";
    var poolPick=filterPoolExcludeRecent(pool);
    if(!poolPick.length) poolPick=pool.slice();
    var idx=Math.floor(Math.random()*poolPick.length);
    if(poolPick.length>1){
      for(var g=0;g<32&&poolPick[idx]===mp3Current;g++){ idx=Math.floor(Math.random()*poolPick.length); }
    }
    return poolPick[idx];
  }
  function schedulePrefetchNext(schedTok){
    if(!mp3Other||schedTok!==mp3PickToken) return;
    if(!mp3SessionPlaying||mp3Audio.loop) return;
    if(!mp3AllowAudioPrefetch()) return;
    var pool=getMp3Pool();
    if(!pool.length) return;
    var name=pickPrefetchName();
    if(!name) return;
    prefetchName=name;
    prefetchToken=mp3PickToken;
    warmNextTrackMeta(name);
    ensureMp3Gate().then(function(ok){
      if(!ok||schedTok!==mp3PickToken||prefetchToken!==mp3PickToken) return;
      mp3Other.src=urlStreamMp3(name);
      try{ mp3Other.load(); }catch(e){}
    });
  }
  function mediaSessionAlbumLabel(){
    return (typeof location!=="undefined"&&location.host)?location.host:(typeof document!=="undefined"&&document.title?document.title:"igozhang");
  }
  function buildArtworkTriples(absUrl){
    if(!absUrl) return [];
    return [
      {src:absUrl,sizes:"512x512"},
      {src:absUrl,sizes:"256x256"},
      {src:absUrl,sizes:"96x96"}
    ];
  }
  function faviconArtworkTriples(){
    var favAbs=absolutePageUrl(urlSiteFaviconMedia());
    return favAbs?buildArtworkTriples(favAbs):[];
  }
  // 锁屏/通知：有内嵌封面则仅封面 URL；无封面则 fs-favicon-media-href（勿与页内 favicon.svg 混用）。
  function trackHasEmbeddedCover(fn){
    return !!(fn&&mp3CoverProbeCache[fn]===true);
  }
  function buildArtworkForTrack(fn){
    if(trackHasEmbeddedCover(fn)){
      var coverAbs=absolutePageUrl(urlMp3Cover(fn));
      if(coverAbs){
        var cov=buildArtworkTriples(coverAbs);
        if(cov.length) return cov;
      }
    }
    return faviconArtworkTriples();
  }
  // 仅当「当前曲」确认有内嵌封面时，才把上一首封面排在前面（Android 常只显示 artwork[0]）。
  // 当前曲无封面时 artwork 只含站点 favicon，禁止叠加上一首封面（否则锁屏看不到定制 favicon）。
  function mergeSessionArtwork(primaryFn, outgoingFn){
    var art=buildArtworkForTrack(primaryFn);
    if(outgoingFn&&outgoingFn!==primaryFn&&trackHasEmbeddedCover(primaryFn)){
      var prev=buildArtworkForTrack(outgoingFn);
      if(prev.length) art=prev.concat(art);
    }
    if(!art.length) art=faviconArtworkTriples();
    return art;
  }
  function artworkFirstSrc(art){
    return (art&&art[0]&&art[0].src)||"";
  }
  function probeMp3CoverThen(fn, cb){
    if(!fn){ cb(false); return; }
    if(Object.prototype.hasOwnProperty.call(mp3CoverProbeCache, fn)){
      cb(!!mp3CoverProbeCache[fn]);
      return;
    }
    var url=urlMp3Cover(fn);
    if(!url){ mp3CoverProbeCache[fn]=false; cb(false); return; }
    fetch(url,{method:"HEAD",credentials:"same-origin"}).then(function(r){
      mp3CoverProbeCache[fn]=r.ok;
      cb(r.ok);
    }).catch(function(){ mp3CoverProbeCache[fn]=false; cb(false); });
  }
  function ensureTrackArtwork(fn, opts){
    opts=opts||{};
    if(!fn) return;
    probeMp3CoverThen(fn, function(){
      if(mp3NextMetaCache&&mp3NextMetaCache.name===fn){
        mp3NextMetaCache.artwork=buildArtworkForTrack(fn);
      }
      if(opts.syncMs!==false&&(fn===mp3Current||fn===mp3LockScreenTrack||(mp3NextMetaCache&&mp3NextMetaCache.name===fn))){
        var msOpts=opts.msOpts||{};
        if(opts.force) msOpts.force=true;
        if(mp3LoadingTrack&&mp3NextMetaCache&&mp3NextMetaCache.name===fn){
          applyMediaSessionFromCache(mp3NextMetaCache, mp3LockScreenTrack||undefined);
          syncMediaSessionPlaybackState();
        } else {
          syncMediaSessionMetadata(msOpts);
        }
      }
    });
  }
  function warmNextTrackMeta(name){
    if(!name||!isMp3Name(name)){
      mp3NextMetaCache=null;
      return;
    }
    var song=parseSong(name);
    mp3NextMetaCache={
      name:name,
      title:stripMp3Ext(song.title||stripMp3Ext(name)),
      artist:String(song.artist||"").trim(),
      album:mediaSessionAlbumLabel(),
      artwork:buildArtworkForTrack(name),
      ready:false
    };
    var warmUrl=absolutePageUrl(urlSiteFaviconMedia());
    if(!warmUrl){ mp3NextMetaCache.ready=true; ensureTrackArtwork(name); return; }
    var img=new Image();
    var done=function(){
      if(mp3NextMetaCache&&mp3NextMetaCache.name===name) mp3NextMetaCache.ready=true;
    };
    img.onload=done;
    img.onerror=done;
    img.src=warmUrl;
    if(img.complete) done();
    ensureTrackArtwork(name, {syncMs:false});
  }
  function msMetaFingerprint(fn, extraFn, art){
    return (fn||"")+"\x1e"+(extraFn||"")+"\x1e"+artworkFirstSrc(art);
  }
  function applyMediaSessionFromCache(cache, extraArtworkTrack){
    if(!cache||typeof MediaMetadata==="undefined"||typeof navigator.mediaSession==="undefined") return;
    var outgoing=extraArtworkTrack||"";
    var art=mergeSessionArtwork(cache.name, outgoing);
    try{
      navigator.mediaSession.metadata=new MediaMetadata({
        title:cache.title||"MP3",
        artist:cache.artist||"",
        album:cache.album||mediaSessionAlbumLabel(),
        artwork:art.length?art:undefined
      });
      lastMsMetaKey=msMetaFingerprint(cache.name, outgoing, art);
    }catch(e){}
  }
  function beginTrackTransitionTo(nextName){
    if(!nextName||!isMp3Name(nextName)) return;
    var outgoing=mp3LockScreenTrack||mp3Current||"";
    if(outgoing&&outgoing!==nextName) mp3LockScreenTrack=outgoing;
    mp3LoadingTrack=true;
    warmNextTrackMeta(nextName);
    if(!mp3SessionPlaying) return;
    if(mp3NextMetaCache) applyMediaSessionFromCache(mp3NextMetaCache, mp3LockScreenTrack||undefined);
    syncMediaSessionPlaybackState();
  }
  function dualSlotPlaySecondaryBuffer(buf, tok){
    mp3SessionPlaying=true;
    buf.preload="auto";
    buf.muted=false;
    if("disableRemotePlayback" in buf) buf.disableRemotePlayback=false;
    buf.currentTime=0;
    refreshMp3Chrome();
    warmNextTrackMeta(mp3Current);
    if(mp3NextMetaCache) applyMediaSessionFromCache(mp3NextMetaCache, mp3LockScreenTrack||undefined);
    syncMediaSessionPlaybackState();
    setPlayLabel();
    buf.addEventListener("playing", function(){
      if(tok!==mp3PickToken||!mp3SessionPlaying) return;
      var prevLock=mp3LockScreenTrack;
      try{ mp3Audio.pause(); }catch(e){}
      try{ mp3Audio.removeAttribute("src"); }catch(e){}
      swapMp3Outputs();
      configureBufferRemoteMuted();
      lastPlaybackAdvanceMs=Date.now();
      lastPlaybackTime=mp3Audio.currentTime||0;
      mp3LockScreenTrack=mp3Current;
      mp3LoadingTrack=false;
      mp3LoadErrors=0;
      if(mp3Current){ syncTitleDisplay(mp3Current); }
      syncMediaSessionMetadata(prevLock&&prevLock!==mp3Current?{extraArtworkTrack:prevLock}:undefined);
      lastMsPosSync=0;
      syncMediaSessionPositionState();
      applyAudioSessionPlayback();
      if(typeof document!=="undefined"){
        if(document.visibilityState==="visible"){ startMediaSessionForegroundKeepalive(); }
        else if(mp3SessionPlaying&&!mp3Audio.paused){ startMediaSessionBackgroundKeepalive(); }
      }
      setPlayLabel();
      schedulePrefetchNext(tok);
    },{once:true});
    var p=buf.play();
    if(p&&p.catch) p.catch(function(){
      mp3NotifyPlayBlocked();
      if(tok!==mp3PickToken) return;
      var rescue=mp3Current;
      if(rescue&&isMp3Name(rescue)&&mp3List.indexOf(rescue)>=0){
        mp3LoadingTrack=true;
        syncMediaSessionPlaybackState();
        syncMediaSessionMetadata();
        playSpecificMp3OnMain(rescue, true, true);
        return;
      }
      mp3LoadingTrack=false;
      pickRandomMp3(true,true);
    });
    return true;
  }
  function tryPlayPrefetched(){
    var buf=mp3Other;
    if(!buf||!prefetchName) return false;
    if(prefetchToken!==mp3PickToken) return false;
    if(mp3Audio.loop) return false;
    if(buf.readyState<2) return false;
    if(!(buf.currentSrc||buf.src)) return false;
    if(typeof document!=="undefined"&&mp3PageIsHidden()){
      var nextOnMain=prefetchName;
      abortMp3AttachFetch();
      cancelPrefetch();
      pushMp3PastIfLeaving(mp3LockScreenTrack||mp3Current||"", nextOnMain);
      beginTrackTransitionTo(nextOnMain);
      playSpecificMp3OnMain(nextOnMain, true, true);
      return true;
    }
    clearSingleSongLoopOnTrackChange();
    var nextName=prefetchName;
    cancelPrefetchMeta();
    abortMp3AttachFetch();
    var tok=++mp3PickToken;
    mp3LoadingTrack=true;
    pushMp3PastIfLeaving(mp3LockScreenTrack||mp3Current||"", nextName);
    mp3Current=nextName;
    noteTrackPlayed(nextName);
    return dualSlotPlaySecondaryBuffer(buf, tok);
  }
  function playTrackWithDualBuffer(name, autoPlay, fastStream){
    if(typeof fastStream==="undefined") fastStream=!!autoPlay;
    fastStream=!!fastStream;
    if(!name||!isMp3Name(name)||mp3List.indexOf(name)<0) return false;
    if(!autoPlay||!mp3Other||mp3Audio.loop) return false;
    mp3TrackChanging=true;
    if(typeof document!=="undefined"&&mp3PageIsHidden()) return false;
    clearSingleSongLoopOnTrackChange();
    cancelPrefetch();
    cancelPrefetchMeta();
    abortMp3AttachFetch();
    var tok=++mp3PickToken;
    mp3LoadingTrack=true;
    if(mp3Audio && mp3Audio.src){
      var outgoingWB=mp3Current;
      if(outgoingWB){ mp3LockScreenTrack=outgoingWB; }
    }
    mp3Current=name;
    noteTrackPlayed(name);
    prefetchName=name;
    prefetchToken=tok;
    var buf=mp3Other;
    ensureMp3Gate().then(function(ok){
      if(!ok||tok!==mp3PickToken) return;
      buf.preload="auto";
      buf.muted=false;
      if("disableRemotePlayback" in buf) buf.disableRemotePlayback=false;
      buf.src=urlStreamMp3(name);
      buf.currentTime=0;
      try{ buf.load(); }catch(e){}
    });
    warmNextTrackMeta(name);
    refreshMp3Chrome();
    if(mp3NextMetaCache) applyMediaSessionFromCache(mp3NextMetaCache, mp3LockScreenTrack||undefined);
    syncMediaSessionPlaybackState();
    setPlayLabel();
    var failTimer=setTimeout(function(){
      if(tok!==mp3PickToken) return;
      playSpecificMp3OnMain(name, autoPlay, fastStream);
    }, 12000);
    function onBufReady(){
      if(tok!==mp3PickToken) return;
      clearTimeout(failTimer);
      buf.removeEventListener("canplay", onBufReady);
      buf.removeEventListener("canplaythrough", onBufReady);
      if(buf.readyState>=2){ dualSlotPlaySecondaryBuffer(buf, tok); }
      else{ playSpecificMp3OnMain(name, autoPlay, fastStream); }
    }
    buf.addEventListener("canplay", onBufReady, {once:true});
    buf.addEventListener("canplaythrough", onBufReady, {once:true});
    setTimeout(function(){
      if(tok===mp3PickToken && buf.readyState>=2) onBufReady();
    }, 0);
    return true;
  }
  function buildTopTags(files){
    var counts={};
    for(var f=0;f<files.length;f++){
      var art=parseSong(files[f]).artist;
      tokenizeArtist(art).forEach(function(tok){
        if(tok) counts[tok]=(counts[tok]||0)+1;
      });
    }
    var arr=Object.keys(counts).map(function(k){ return {t:k,c:counts[k]}; });
    arr.sort(function(a,b){ if(b.c!==a.c) return b.c-a.c; return a.t.localeCompare(b.t); });
    return arr.slice(0,3).map(function(x){ return x.t; });
  }
  function renderMp3Tags(){
    if(!mp3TagsEl) return;
    mp3TagsEl.innerHTML="";
    for(var t=0;t<mp3TopTags.length;t++){
      (function(tag){
        var b=document.createElement("button");
        b.type="button";
        b.className="phrase-next mp3-tag";
        b.textContent=tag;
        b.title="在含「"+tag+"」的歌曲里随机一首并持续连播；该标签下备选超过 5 首时避开最近 2 首";
        b.onclick=function(){
          mp3Haptic();
          mp3FilterTag=tag;
          syncTagButtons();
          mp3SessionPlaying=true;
          pickRandomMp3(true);
        };
        mp3TagsEl.appendChild(b);
      })(mp3TopTags[t]);
    }
    syncTagButtons();
  }
  function syncTagButtons(){
    if(!mp3TagsEl) return;
    var btns=mp3TagsEl.querySelectorAll(".mp3-tag");
    for(var i=0;i<btns.length;i++){
      var on=(btns[i].textContent===mp3FilterTag);
      btns[i].classList.toggle("mp3-tag-on",on);
      btns[i].setAttribute("aria-pressed",on?"true":"false");
    }
  }
  function normalizeMp3Query(q){
    return String(q||"").trim().toLowerCase();
  }
  function escMp3Html(s){
    return String(s).replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;").replace(/"/g,"&quot;");
  }
  function songSearchHaystack(fn){
    var song=parseSong(fn);
    return (song.title+" "+song.artist+" "+stripMp3Ext(fn)).toLowerCase();
  }
  function rebuildMp3SearchIndex(){
    mp3SearchIndex=[];
    for(var i=0;i<mp3List.length;i++){
      var fn=mp3List[i];
      var song=parseSong(fn);
      mp3SearchIndex.push({
        name:fn,
        title:song.title||stripMp3Ext(fn),
        artist:song.artist||"",
        haystack:songSearchHaystack(fn)
      });
    }
  }
  function mp3SearchTokens(q){
    return normalizeMp3Query(q).split(/\s+/).filter(Boolean);
  }
  function mp3SearchScore(entry, tokens){
    var score=0;
    var title=entry.title.toLowerCase();
    var artist=entry.artist.toLowerCase();
    var hay=entry.haystack;
    for(var t=0;t<tokens.length;t++){
      var tok=tokens[t];
      if(hay.indexOf(tok)<0) return -1;
      if(title.indexOf(tok)===0) score+=40;
      else if(title.indexOf(tok)>=0) score+=20;
      else if(artist.indexOf(tok)>=0) score+=10;
      else score+=5;
    }
    return score;
  }
  function filterMp3Search(q){
    var tokens=mp3SearchTokens(q);
    if(!tokens.length) return [];
    var ranked=[];
    for(var i=0;i<mp3SearchIndex.length;i++){
      var entry=mp3SearchIndex[i];
      var sc=mp3SearchScore(entry, tokens);
      if(sc<0) continue;
      ranked.push({name:entry.name, score:sc, title:entry.title, artist:entry.artist});
    }
    ranked.sort(function(a,b){
      if(b.score!==a.score) return b.score-a.score;
      return a.title.localeCompare(b.title);
    });
    if(ranked.length>12) ranked.length=12;
    return ranked;
  }
  function syncMp3SearchHighlight(){
    if(!mp3SearchResults) return;
    var btns=mp3SearchResults.querySelectorAll("button[data-mp3-hit]");
    for(var i=0;i<btns.length;i++){
      btns[i].classList.toggle("mp3-search-hit-on", i===mp3SearchActiveIdx);
    }
    if(mp3SearchActiveIdx>=0&&btns[mp3SearchActiveIdx]){
      try{ btns[mp3SearchActiveIdx].scrollIntoView({block:"nearest"}); }catch(e){}
    }
  }
  function closeMp3Search(){
    mp3SearchHits=[];
    mp3SearchActiveIdx=-1;
    if(mp3SearchResults){
      mp3SearchResults.hidden=true;
      mp3SearchResults.innerHTML="";
    }
  }
  function openMp3Search(){
    if(mp3SearchResults) mp3SearchResults.hidden=false;
  }
  function pickMp3SearchHit(idx){
    if(idx<0||idx>=mp3SearchHits.length) return;
    var hit=mp3SearchHits[idx];
    if(!hit||!hit.name) return;
    mp3Haptic();
    closeMp3Search();
    if(mp3SearchInput){
      mp3SearchInput.value=hit.title+(hit.artist?" - "+hit.artist:"");
      try{ mp3SearchInput.blur(); }catch(e){}
    }
    mp3FilterTag=null;
    syncTagButtons();
    mp3SessionPlaying=true;
    advanceToTrack(hit.name, true);
  }
  function renderMp3Search(hits){
    if(!mp3SearchResults) return;
    mp3SearchHits=hits||[];
    mp3SearchActiveIdx=mp3SearchHits.length?0:-1;
    mp3SearchResults.innerHTML="";
    if(!mp3SearchHits.length){
      var li0=document.createElement("li");
      var btn0=document.createElement("button");
      btn0.type="button";
      btn0.disabled=true;
      btn0.textContent="无匹配";
      li0.appendChild(btn0);
      mp3SearchResults.appendChild(li0);
      openMp3Search();
      return;
    }
    for(var h=0;h<mp3SearchHits.length;h++){
      (function(hit, idx){
        var li=document.createElement("li");
        var btn=document.createElement("button");
        btn.type="button";
        btn.setAttribute("data-mp3-hit","1");
        btn.innerHTML="<strong>"+escMp3Html(hit.title)+"</strong>"+(hit.artist?"<span>"+escMp3Html(hit.artist)+"</span>":"");
        btn.title="从此曲开始随机连播";
        btn.onclick=function(){ pickMp3SearchHit(idx); };
        li.appendChild(btn);
        mp3SearchResults.appendChild(li);
      })(mp3SearchHits[h], h);
    }
    openMp3Search();
    syncMp3SearchHighlight();
  }
  function goMp3Search(){
    var q=mp3SearchInput?mp3SearchInput.value:"";
    if(normalizeMp3Query(q).length<1){ closeMp3Search(); return; }
    renderMp3Search(filterMp3Search(q));
  }
  function setMp3SearchEnabled(on){
    if(!mp3SearchInput) return;
    mp3SearchInput.disabled=!on;
    if(!on){
      mp3SearchInput.value="";
      closeMp3Search();
    }
  }
  function setPlayLabel(){
    if(!mp3PlayBtn||!mp3Audio)return;
    if(mp3Audio.paused){
      mp3PlayBtn.textContent="\u25B6";
      mp3PlayBtn.title="播放";
    } else {
      mp3PlayBtn.textContent="\u23F8";
      mp3PlayBtn.title="暂停";
    }
    syncMediaSessionPlaybackState();
  }
  function syncLoopButtonUI(){
    if(!mp3LoopBtn||!mp3Audio) return;
    var on=!!mp3Audio.loop;
    mp3LoopBtn.classList.toggle("mp3-loop-on", on);
    mp3LoopBtn.classList.toggle("mp3-active", on);
    mp3LoopBtn.setAttribute("aria-pressed", on?"true":"false");
    mp3LoopBtn.textContent="one";
    mp3LoopBtn.title=on
      ? "单曲循环已开启：重复播放当前这首；再点关闭恢复连播"
      : "单曲循环：仅重复当前这首；换歌或点随机后自动关闭";
  }
  function clearSingleSongLoopOnTrackChange(){
    if(!mp3Audio||!mp3Audio.loop) return;
    mp3Audio.loop=false;
    syncLoopButtonUI();
  }
  var mediaSessionHandlersBound=false;
  function mediaSessionTrackName(){
    if(mp3LoadingTrack&&mp3LockScreenTrack){ return mp3LockScreenTrack; }
    return mp3Current;
  }
  function displayedMp3TrackName(){
    return mediaSessionTrackName();
  }
  function refreshMp3Chrome(){
    var n=displayedMp3TrackName();
    if(n){ syncTitleDisplay(n); }
    else if(mp3Title){ mp3Title.textContent=""; mp3Title.classList.remove("mp3-title-visible"); syncMp3Cover(""); }
  }
  function syncMediaSessionPlaybackState(){
    if(typeof navigator.mediaSession==="undefined") return;
    try{
      if(!mp3Audio) return;
      if(!mp3Audio.src){
        if(mp3SessionPlaying&&(mp3LoadingTrack||mp3LockScreenTrack)){
          navigator.mediaSession.playbackState="playing";
          return;
        }
        navigator.mediaSession.playbackState="none";
        return;
      }
      if(mp3LoadingTrack&&mp3SessionPlaying){
        navigator.mediaSession.playbackState="playing";
        return;
      }
      if(mp3SessionPlaying&&!mp3Audio.loop&&mp3Audio.ended&&mp3Audio.src){
        navigator.mediaSession.playbackState="playing";
        return;
      }
      navigator.mediaSession.playbackState=mp3Audio.paused?"paused":"playing";
    }catch(e){}
  }
  function syncMediaSessionMetadata(opts){
    opts=opts||{};
    if(typeof MediaMetadata==="undefined"||typeof navigator.mediaSession==="undefined") return;
    if(mp3LoadingTrack&&mp3NextMetaCache&&mp3NextMetaCache.name&&(mp3NextMetaCache.name===mp3Current||!mediaSessionTrackName())){
      applyMediaSessionFromCache(mp3NextMetaCache, opts.extraArtworkTrack||mp3LockScreenTrack||undefined);
      return;
    }
    var fn=mediaSessionTrackName();
    if(!fn&&mp3SessionPlaying&&mp3LockScreenTrack){ fn=mp3LockScreenTrack; }
    if(!fn){
      if(mp3SessionPlaying){
        try{
          var album0=(typeof location!=="undefined"&&location.host)?location.host:(typeof document!=="undefined"&&document.title?document.title:"igozhang");
          var favOnly=faviconArtworkTriples();
          navigator.mediaSession.metadata=new MediaMetadata({
            title: "MP3",
            artist: "",
            album: album0,
            artwork: favOnly.length?favOnly:undefined
          });
        }catch(e){}
        return;
      }
      try{ navigator.mediaSession.metadata=null; }catch(e){}
      return;
    }
    var song=parseSong(fn);
    var title=stripMp3Ext(song.title||stripMp3Ext(fn));
    var artist=String(song.artist||"").trim();
    var album=mediaSessionAlbumLabel();
    var extraFn=opts.extraArtworkTrack;
    var art=mergeSessionArtwork(fn, extraFn);
    var fp=msMetaFingerprint(fn, extraFn||"", art);
    if(!opts.force&&fp===lastMsMetaKey) return;
    try{
      var md={title:title||"MP3",artist:artist,album:album};
      if(art.length) md.artwork=art;
      navigator.mediaSession.metadata=new MediaMetadata(md);
      lastMsMetaKey=fp;
    }catch(e){}
  }
  function clearMediaSessionPositionState(){
    if(typeof navigator.mediaSession==="undefined"||typeof navigator.mediaSession.setPositionState!=="function") return;
    try{ navigator.mediaSession.setPositionState(null); }catch(e){}
  }
  function syncMediaSessionPositionState(){
    if(typeof navigator.mediaSession==="undefined"||typeof navigator.mediaSession.setPositionState!=="function") return;
    if(!mp3Audio||mp3Audio.paused||!mp3Audio.src) return;
    var d=mp3Audio.duration;
    if(!isFinite(d)||d<=0) return;
    var pos=Math.min(Math.max(0,mp3Audio.currentTime),Math.max(0,d-0.02));
    try{
      navigator.mediaSession.setPositionState({duration:d,playbackRate:mp3Audio.playbackRate||1,position:pos});
    }catch(e){}
  }
  function throttledSyncMediaSessionPosition(){
    var t=Date.now();
    if(t-lastMsPosSync<2200) return;
    lastMsPosSync=t;
    syncMediaSessionPositionState();
  }
  function stopMediaSessionBackgroundKeepalive(){
    if(mediaSessionBgTimer){ clearInterval(mediaSessionBgTimer); mediaSessionBgTimer=null; }
  }
  function startMediaSessionForegroundKeepalive(){
    stopMediaSessionBackgroundKeepalive();
    mediaSessionBgTimer=setInterval(function(){
      if(typeof document!=="undefined"&&document.visibilityState!=="visible") return;
      if(mp3LoadingTrack&&mp3SessionPlaying){
        syncMediaSessionPlaybackState();
        syncMediaSessionMetadata();
        return;
      }
      if(!mp3Audio||mp3Audio.paused||!mp3SessionPlaying){ stopMediaSessionBackgroundKeepalive(); return; }
      syncMediaSessionPlaybackState();
      syncMediaSessionPositionState();
    }, 22000);
  }
  function startMediaSessionBackgroundKeepalive(){
    stopMediaSessionBackgroundKeepalive();
    lastBgPositionPushMs=0;
    mediaSessionBgTimer=setInterval(function(){
      if(typeof document!=="undefined"&&document.visibilityState==="visible"){
        stopMediaSessionBackgroundKeepalive();
        return;
      }
      if(mp3LoadingTrack&&mp3SessionPlaying){
        syncMediaSessionPlaybackState();
        syncMediaSessionMetadata();
        applyAudioSessionPlayback();
        return;
      }
      if(!mp3Audio||mp3Audio.paused||!mp3SessionPlaying){ stopMediaSessionBackgroundKeepalive(); return; }
      syncMediaSessionPlaybackState();
      applyAudioSessionPlayback();
      checkPlaybackStall();
      var now=Date.now();
      if(now-lastBgPositionPushMs>6500){
        lastBgPositionPushMs=now;
        syncMediaSessionPositionState();
      }
    }, 3000);
  }
  function bindMediaSessionActions(){
    if(mediaSessionHandlersBound||typeof navigator.mediaSession==="undefined") return;
    mediaSessionHandlersBound=true;
    var safe=function(fn){
      return function(detail){
        try{ fn(detail); }catch(e){}
      };
    };
    try{ navigator.mediaSession.setActionHandler("play", safe(function(){ withMp3ListReady(function(){ mp3SessionPlaying=true; mp3UserPaused=false; clearMp3AudioDefer(); if(!mp3Audio.src&&mp3List.length){ pickRandomMp3(true); return;} mp3PlayMaybeToast(mp3Audio.play()); syncMediaSessionPlaybackState(); }); })); }catch(e){}
    try{ navigator.mediaSession.setActionHandler("pause", safe(function(){ mp3SessionPlaying=false; mp3UserPaused=true; mp3Audio.pause(); syncMediaSessionPlaybackState(); })); }catch(e){}
    try{ navigator.mediaSession.setActionHandler("nexttrack", safe(function(){ withMp3ListReady(function(){ if(!mp3List.length) return; mp3SessionPlaying=true; pickRandomMp3(true,true); }); })); }catch(e){}
    try{ navigator.mediaSession.setActionHandler("previoustrack", safe(function(){
      if(!mp3List.length) return;
      mp3SessionPlaying=true;
      if(mp3PastStack.length>0){
        var prev=mp3PastStack.pop();
        if(prev&&mp3List.indexOf(prev)>=0&&prev!==mp3Current){
          advanceToTrack(prev,true,true);
          syncMediaSessionPlaybackState();
          return;
        }
      }
      if(!mp3Audio.src) return;
      mp3Audio.currentTime=0;
      mp3PlayMaybeToast(mp3Audio.play());
      syncMediaSessionPlaybackState();
    })); }catch(e){}
    try{ navigator.mediaSession.setActionHandler("seekbackward", safe(function(){ if(mp3Audio.src) mp3Audio.currentTime=Math.max(0,mp3Audio.currentTime-10); lastMsPosSync=0; syncMediaSessionPositionState(); })); }catch(e){}
    try{ navigator.mediaSession.setActionHandler("seekforward", safe(function(){
      if(!mp3Audio.src) return;
      var d=mp3Audio.duration;
      if(isFinite(d)&&d>0) mp3Audio.currentTime=Math.min(Math.max(0,d-0.25),mp3Audio.currentTime+10);
      else mp3Audio.currentTime=mp3Audio.currentTime+10;
      lastMsPosSync=0;
      syncMediaSessionPositionState();
    })); }catch(e){}
  }
  function loadMp3List(onDone){
    if(mp3ListLoaded){
      if(onDone) onDone(true);
      return;
    }
    if(onDone) mp3ListLoadQueue.push(onDone);
    if(mp3ListLoading) return;
    mp3ListLoading=true;
    ensureMp3Gate().then(function(gateOk){
      if(!gateOk){
        mp3ListLoading=false;
        showToast("MP3 门禁失败", true);
        var qGate=mp3ListLoadQueue.slice();
        mp3ListLoadQueue=[];
        for(var gi=0;gi<qGate.length;gi++){
          try{ qGate[gi](false); }catch(e){}
        }
        return;
      }
    fetch(urlBrowse("assets/mp3"),{credentials:"same-origin"}).then(function(r){
      return r.text().then(function(t){
        if(!r.ok) throw new Error(t);
        return JSON.parse(t);
      });
    }).then(function(data){
      cancelPrefetch();
      var handoff=pickMp3HandoffCandidate(mp3HandoffPending);
      mp3HandoffPending=null;
      if(!handoff){
        mp3RecentTwo=[];
        mp3PastStack=[];
      }
      mp3CoverProbeCache={};
      lastMsMetaKey="";
      mp3List=(data||[]).filter(function(it){ return !it.isDir&&isMp3Name(it.name); }).map(function(it){ return it.name; });
      if(!handoff) mp3FilterTag=null;
      rebuildMp3SearchIndex();
      mp3TopTags=buildTopTags(mp3List);
      setMp3Heading();
      renderMp3Tags();
      if(!mp3List.length){
        var rnd0=document.getElementById("mp3Random");
        if(rnd0) rnd0.disabled=true;
        if(mp3LoopBtn) mp3LoopBtn.disabled=true;
        if(mp3PlayBtn) mp3PlayBtn.disabled=true;
        setMp3SearchEnabled(false);
        mp3ListLoaded=true;
        mp3ListLoading=false;
        var q0=mp3ListLoadQueue.slice();
        mp3ListLoadQueue=[];
        for(var j=0;j<q0.length;j++){
          try{ q0[j](true); }catch(e){}
        }
        return;
      }
      var rnd1=document.getElementById("mp3Random");
      if(rnd1) rnd1.disabled=false;
      if(mp3LoopBtn) mp3LoopBtn.disabled=false;
      if(mp3PlayBtn) mp3PlayBtn.disabled=false;
      setMp3SearchEnabled(true);
      if(handoff){
        restoreMp3Handoff(handoff);
        clearMp3HandoffStorage();
      }
      mp3ListLoaded=true;
      mp3ListLoading=false;
      var q=mp3ListLoadQueue.slice();
      mp3ListLoadQueue=[];
      for(var i=0;i<q.length;i++){
        try{ q[i](true); }catch(e){}
      }
    }).catch(function(){
      mp3HandoffPending=null;
      cancelPrefetch();
      mp3RecentTwo=[];
      mp3PastStack=[];
      mp3CoverProbeCache={};
      lastMsMetaKey="";
      mp3List=[]; mp3TopTags=[]; mp3FilterTag=null;
      rebuildMp3SearchIndex();
      setMp3Heading();
      renderMp3Tags();
      showToast("MP3 列表加载失败", true);
      var rnd2=document.getElementById("mp3Random");
      if(rnd2) rnd2.disabled=true;
      if(mp3LoopBtn) mp3LoopBtn.disabled=true;
      if(mp3PlayBtn) mp3PlayBtn.disabled=true;
      setMp3SearchEnabled(false);
      mp3ListLoading=false;
      var qf=mp3ListLoadQueue.slice();
      mp3ListLoadQueue=[];
      for(var k=0;k<qf.length;k++){
        try{ qf[k](false); }catch(e){}
      }
    });
    });
  }
  function ensureMp3ListLoaded(cb){
    if(mp3ListLoaded){
      if(cb) cb(true);
      return;
    }
    if(cb) mp3ListLoadQueue.push(cb);
    if(mp3ListLoading) return;
    loadMp3List();
  }
  function withMp3ListReady(fn){
    if(window.__fsMp3LazyList){
      ensureMp3ListLoaded(function(ok){
        if(ok) fn();
      });
      return;
    }
    fn();
  }
  function bootMp3List(){
    if(window.__fsMp3LazyList){
      if(mp3HandoffPending) loadMp3List();
      return;
    }
    loadMp3List();
  }
  function playSpecificMp3OnMain(name, autoPlay, fastStream, seekTime, preserveLoop, handoffRestore){
    if(typeof fastStream==="undefined") fastStream=!!autoPlay;
    fastStream=!!fastStream;
    handoffRestore=!!handoffRestore;
    if(!name||!isMp3Name(name)||mp3List.indexOf(name)<0) return;
    mp3TrackChanging=true;
    if(!preserveLoop) clearSingleSongLoopOnTrackChange();
    cancelPrefetch();
    var tok;
    if(preserveLoop) tok=mp3PickToken||(++mp3PickToken);
    else tok=++mp3PickToken;
    mp3LoadingTrack=!!autoPlay;
    if(autoPlay && mp3Audio && mp3Audio.src){
      var outgoing=mp3Current;
      if(outgoing){ mp3LockScreenTrack=outgoing; }
    }
    mp3Current=name;
    if(!preserveLoop) noteTrackPlayed(name);
    if(autoPlay) beginTrackTransitionTo(name);
    var streamUrl=urlStreamMp3(name);
    var seekTo=(typeof seekTime==="number"&&isFinite(seekTime)&&seekTime>0)?seekTime:0;
    function applySeekTime(){
      if(tok!==mp3PickToken||!seekTo) return;
      try{
        var d=mp3Audio.duration;
        if(isFinite(d)&&d>0) mp3Audio.currentTime=Math.min(seekTo, Math.max(0, d-0.25));
        else mp3Audio.currentTime=seekTo;
      }catch(e){}
    }
    function afterAttach(){
      if(tok!==mp3PickToken) return;
      mp3TrackChanging=false;
      if(seekTo>0){
        if(mp3Audio.readyState>=1) applySeekTime();
        else mp3Audio.addEventListener("loadedmetadata", applySeekTime, {once:true});
      }else{
        mp3Audio.currentTime=0;
      }
      refreshMp3Chrome();
      syncMediaSessionPlaybackState();
      syncMediaSessionMetadata({force:true});
      if(!autoPlay){
        mp3LoadingTrack=false;
        mp3Audio.pause();
        setPlayLabel();
        return;
      }
      mp3SessionPlaying=true;
      var started=false;
      function finishHandoffPlay(){
        if(tok!==mp3PickToken||!mp3SessionPlaying||started) return;
        started=true;
        mp3TrackChanging=false;
        if(seekTo>0) applySeekTime();
        var pr=mp3Audio.play();
        mp3PlayMaybeToast(pr);
        if(pr&&typeof pr.catch==="function"){
          pr.catch(function(){
            if(handoffResumeDeadlinePassed()){ abortHandoffAutoResume(); return; }
            syncMediaSessionPlaybackState();
            syncMediaSessionMetadata({force:true});
            bindHandoffGestureResume();
            if(mp3LikelyMobile()) showToast("跨页续播：轻触页面继续", false);
            scheduleHandoffPlaybackResume();
          });
        }
        setPlayLabel();
        schedulePrefetchNext(tok);
      }
      function start(){
        if(tok!==mp3PickToken||!mp3SessionPlaying||started) return;
        started=true;
        mp3TrackChanging=false;
        mp3PlayMaybeToast(mp3Audio.play());
        setPlayLabel();
        schedulePrefetchNext(tok);
      }
      if(handoffRestore){
        if(mp3Audio.readyState>=2) finishHandoffPlay();
        else mp3Audio.addEventListener("canplay", finishHandoffPlay, {once:true});
        mp3Audio.addEventListener("canplaythrough", finishHandoffPlay, {once:true});
        setTimeout(function(){
          if(tok!==mp3PickToken||!mp3SessionPlaying||started) return;
          if(mp3Audio.readyState>=2) finishHandoffPlay();
        }, mp3LikelyMobile()?900:350);
        return;
      }
      mp3PlayMaybeToast(mp3Audio.play());
      if(mp3Audio.readyState>=2){ start(); return; }
      mp3Audio.addEventListener("canplay", start, {once:true});
      mp3Audio.addEventListener("canplaythrough", start, {once:true});
      setTimeout(function(){
        if(tok!==mp3PickToken||!mp3SessionPlaying||started) return;
        if(mp3Audio.readyState>=2) start();
      }, 0);
    }
    if(shouldUseFastStreamAttach(autoPlay, fastStream)) attachMp3StreamFast(streamUrl, tok, afterAttach);
    else attachMp3StreamWithMinBuffer(streamUrl, tok, afterAttach);
  }
  function playSpecificMp3(name, autoPlay, fastStream){
    if(playTrackWithDualBuffer(name, autoPlay, fastStream)) return;
    playSpecificMp3OnMain(name, autoPlay, fastStream);
  }
  function pickRandomName(){
    var pool=getMp3Pool();
    if(!pool.length) return "";
    var poolPick=filterPoolExcludeRecent(pool);
    if(!poolPick.length) poolPick=pool.slice();
    var idx=Math.floor(Math.random()*poolPick.length);
    if(poolPick.length>1){
      for(var g=0;g<32&&poolPick[idx]===mp3Current;g++){ idx=Math.floor(Math.random()*poolPick.length); }
    }
    return poolPick[idx];
  }
  function advanceToTrack(name, autoPlay, fastStream){
    if(typeof fastStream==="undefined") fastStream=!!autoPlay;
    fastStream=!!fastStream;
    if(!name||!isMp3Name(name)||mp3List.indexOf(name)<0) return;
    mp3TrackChanging=true;
    if(autoPlay) beginTrackTransitionTo(name);
    if(prefetchName===name && prefetchToken===mp3PickToken && tryPlayPrefetched()){
      mp3TrackChanging=false;
      return;
    }
    pushMp3PastIfLeaving(mp3LockScreenTrack||mp3Current||"", name);
    if(playTrackWithDualBuffer(name, autoPlay, fastStream)){
      mp3TrackChanging=false;
      return;
    }
    playSpecificMp3OnMain(name, autoPlay, fastStream);
  }
  function pickRandomMp3(autoPlay, fastStream){
    var name=pickRandomName();
    if(!name) return;
    advanceToTrack(name, autoPlay, fastStream);
  }
  var mp3RandomBtn=document.getElementById("mp3Random");
  if(mp3RandomBtn) mp3RandomBtn.onclick=function(){
    withMp3ListReady(function(){
      mp3Haptic();
      mp3FilterTag=null;
      syncTagButtons();
      mp3SessionPlaying=true;
      pickRandomMp3(true);
    });
  };
  if(mp3LoopBtn) mp3LoopBtn.onclick=function(){
    mp3Haptic();
    mp3Audio.loop=!mp3Audio.loop;
    syncLoopButtonUI();
    if(mp3Audio.loop){ cancelPrefetch(); }
    else if(mp3SessionPlaying&&!mp3Audio.paused){ schedulePrefetchNext(mp3PickToken); }
  };
  if(mp3SearchInput){
    mp3SearchInput.addEventListener("input", function(){
      withMp3ListReady(function(){
        clearTimeout(mp3SearchTimer);
        mp3SearchTimer=setTimeout(goMp3Search, 200);
      });
    });
    mp3SearchInput.addEventListener("focus", function(){
      withMp3ListReady(function(){
        if(normalizeMp3Query(mp3SearchInput.value).length>0) goMp3Search();
      });
    });
    mp3SearchInput.addEventListener("keydown", function(e){
      if(e.key==="Escape"){ closeMp3Search(); mp3SearchInput.blur(); return; }
      if(e.key==="ArrowDown"){
        e.preventDefault();
        if(!mp3SearchHits.length) goMp3Search();
        else{
          mp3SearchActiveIdx=Math.min(mp3SearchHits.length-1, mp3SearchActiveIdx<0?0:mp3SearchActiveIdx+1);
          syncMp3SearchHighlight();
          openMp3Search();
        }
        return;
      }
      if(e.key==="ArrowUp"){
        if(!mp3SearchHits.length) return;
        e.preventDefault();
        mp3SearchActiveIdx=Math.max(0, mp3SearchActiveIdx<0?0:mp3SearchActiveIdx-1);
        syncMp3SearchHighlight();
        return;
      }
      if(e.key==="Enter"){
        if(mp3SearchActiveIdx>=0&&mp3SearchHits.length){
          e.preventDefault();
          pickMp3SearchHit(mp3SearchActiveIdx);
        }
      }
    });
  }
  document.addEventListener("click", function(e){
    if(!mp3SearchInput||!mp3SearchResults) return;
    var wrap=mp3SearchInput.closest(".mp3-search-wrap");
    if(wrap&&wrap.contains(e.target)) return;
    closeMp3Search();
  });
  if(mp3PlayBtn) mp3PlayBtn.onclick=function(){
    withMp3ListReady(function(){
      mp3Haptic();
      if(mp3Audio.paused){
        clearHandoffResumeAttempt();
        mp3SessionPlaying=true;
        mp3UserPaused=false;
        clearMp3AudioDefer();
        if(!mp3Audio.src&&mp3List.length){ pickRandomMp3(true); return; }
        if(mp3Audio.src){ mp3PlayMaybeToast(mp3Audio.play()); }
      } else {
        mp3SessionPlaying=false;
        mp3UserPaused=true;
        mp3Audio.pause();
      }
      syncMediaSessionPlaybackState();
    });
  };
  bindMediaSessionActions();
  document.addEventListener("freeze", function(){
    try{
      syncMediaSessionPlaybackState();
      syncMediaSessionMetadata();
    }catch(e){}
  });
  document.addEventListener("resume", function(){
    try{
      syncMediaSessionPlaybackState();
      syncMediaSessionMetadata();
      applyAudioSessionPlayback();
    }catch(e){}
    if(mp3SessionPlaying&&mp3Audio&&mp3Audio.paused&&mp3Audio.src&&!mp3Audio.ended&&!shouldBlockMp3AutoResume()){
      schedulePlaybackResumeAny();
    }
    if(typeof document!=="undefined"){
      if(document.visibilityState==="visible"){
        if(mp3Audio&&!mp3Audio.paused&&mp3SessionPlaying){
          startMediaSessionForegroundKeepalive();
        }
      } else if(mp3SessionPlaying&&mp3Audio&&!mp3Audio.paused){
        startMediaSessionBackgroundKeepalive();
      }
    }
    var rname=displayedMp3TrackName();
    if(rname){ syncMp3Cover(rname,true); }
  });
  document.addEventListener("fs:pagehidden", function(){
    clearPlaybackResumeTimers();
    stopMediaSessionBackgroundKeepalive();
    if(mp3SessionPlaying&&mp3Audio&&!mp3Audio.paused){ startMediaSessionBackgroundKeepalive(); }
    syncMediaSessionPlaybackState();
    syncMediaSessionMetadata();
    if(mp3SessionPlaying&&!mp3BgHintShown){
      mp3BgHintShown=true;
      showToast("后台连播中，若中断可点锁屏 ▶ 继续", false);
    }
  });
  document.addEventListener("fs:pagevisible", function(){
    if(mp3SessionPlaying&&!shouldBlockMp3AutoResume()){
      if(mp3Audio&&mp3Audio.paused&&!mp3Audio.ended&&!mp3UserPaused){
        schedulePlaybackResumeAny();
      }else if(mp3Audio&&mp3Audio.ended&&mp3List.length){
        pickRandomMp3(true,true);
      }
    }
    if(mp3Audio&&!mp3Audio.paused&&mp3SessionPlaying){
      startMediaSessionForegroundKeepalive();
    }
    var resumeCover=displayedMp3TrackName();
    if(resumeCover){ syncMp3Cover(resumeCover,true); }
  });
  window.addEventListener("pageshow", function(ev){
    if(ev.persisted&&document.visibilityState==="visible"&&!shouldBlockMp3AutoResume()) schedulePlaybackResumeAny();
    if(ev.persisted&&document.visibilityState==="visible"&&mp3Current){ syncMp3Cover(mp3Current,true); }
  });
  window.addEventListener("focus", function(){
    if(document.visibilityState==="visible"&&!shouldBlockMp3AutoResume()) schedulePlaybackResumeAny();
  });
  window.addEventListener("online", function(){
    if(document.visibilityState!=="visible") return;
    if(shouldBlockMp3AutoResume()) return;
    if(mp3SessionPlaying&&mp3Audio&&mp3Audio.paused&&mp3Audio.src&&!mp3Audio.ended) schedulePlaybackResumeAny();
  });
  function mp3Wire(el,type,fn){ if(el) el.addEventListener(type, fn); }
  function wireMp3Both(type, fn){
    mp3Wire(mp3SlotA, type, fn);
    mp3Wire(mp3SlotB, type, fn);
  }
  wireMp3Both("play", function(ev){
    if(ev.target!==mp3Audio) return;
    clearPlaybackResumeTimers();
    mp3LoadErrors=0;
    lastPlaybackAdvanceMs=Date.now();
    lastPlaybackTime=mp3Audio.currentTime||0;
    mp3StallRetryCount=0;
    if(!mp3UserPaused) clearMp3AudioDefer();
    setPlayLabel();
    syncMediaSessionPositionState();
  });
  wireMp3Both("pause", function(ev){
    if(ev.target!==mp3Audio) return;
    var naturalEnd=mp3SessionPlaying&&mp3Audio&&!mp3Audio.loop&&mp3Audio.ended;
    if(naturalEnd){
      mp3LoadingTrack=true;
      syncMediaSessionPlaybackState();
      syncMediaSessionMetadata();
      if(typeof document!=="undefined"&&mp3PageIsHidden()&&mp3SessionPlaying){
        startMediaSessionBackgroundKeepalive();
      }
    }
    setPlayLabel();
    if(!mp3LoadingTrack&&!mp3TrackChanging){ stopMediaSessionBackgroundKeepalive(); }
    if(!mp3LoadingTrack&&!mp3TrackChanging){ clearMediaSessionPositionState(); }
    if(mp3SessionPlaying&&mp3Audio&&mp3Audio.paused&&mp3Audio.src&&!mp3Audio.ended){
      if(!naturalEnd&&!mp3UserPaused&&!mp3LoadingTrack&&!mp3TrackChanging){
        mp3DeferToOtherAudio=true;
        clearPlaybackResumeTimers();
        return;
      }
      if(!shouldBlockMp3AutoResume()) schedulePlaybackResumeAny();
    }
  });
  wireMp3Both("emptied", function(ev){
    if(ev.target!==mp3Audio) return;
    if(!mp3SessionPlaying) return;
    if(!mp3LoadingTrack&&!mp3LockScreenTrack) return;
    syncMediaSessionPlaybackState();
    syncMediaSessionMetadata();
  });
  wireMp3Both("playing", function(ev){
    if(ev.target!==mp3Audio) return;
    clearHandoffResumeAttempt();
    clearPlaybackResumeTimers();
    var prevLock=mp3LockScreenTrack;
    mp3LockScreenTrack=mp3Current;
    mp3LoadingTrack=false;
    mp3LoadErrors=0;
    mp3TrackChanging=false;
    if(mp3Current){ syncTitleDisplay(mp3Current); ensureTrackArtwork(mp3Current, {force:true}); }
    var extraArt=null;
    if(prevLock && prevLock!==mp3Current){ extraArt=prevLock; }
    syncMediaSessionMetadata(extraArt?{extraArtworkTrack:extraArt, force:true}:undefined);
    lastMsPosSync=0;
    syncMediaSessionPositionState();
    applyAudioSessionPlayback();
    if(typeof document!=="undefined"){
      if(document.visibilityState==="visible"){
        startMediaSessionForegroundKeepalive();
      } else if(mp3SessionPlaying&&!mp3Audio.paused){
        startMediaSessionBackgroundKeepalive();
      }
    }
  });
  wireMp3Both("durationchange", function(ev){ if(ev.target!==mp3Audio) return; syncMediaSessionPositionState(); });
  wireMp3Both("seeked", function(ev){ if(ev.target!==mp3Audio) return; lastMsPosSync=0; syncMediaSessionPositionState(); });
  wireMp3Both("timeupdate", function(ev){
    if(ev.target!==mp3Audio) return;
    notePlaybackProgress();
    maybeEarlyPrefetch();
    maybeFlushMp3HandoffSnap();
    throttledSyncMediaSessionPosition();
  });
  wireMp3Both("error", function(ev){
    var el=ev.target;
    if(el===mp3Other){
      if(prefetchToken!==mp3PickToken) return;
      if(!mp3AllowAudioPrefetch()){ cancelPrefetch(); return; }
      var pool=getMp3Pool();
      if(!pool.length) return;
      var name=pickPrefetchName();
      if(!name) return;
      prefetchName=name;
      prefetchToken=mp3PickToken;
      mp3Other.src=urlStreamMp3(name);
      try{ mp3Other.load(); }catch(e){}
      return;
    }
    if(el!==mp3Audio) return;
    mp3LoadingTrack=false;
    setPlayLabel();
    syncMediaSessionPlaybackState();
    if(!mp3SessionPlaying) return;
    mp3LoadErrors++;
    if(mp3LoadErrors>10) return;
    var pool=getMp3Pool();
    if(pool.length<1) return;
    setTimeout(function(){ pickRandomMp3(true,true); }, 300);
  });
  wireMp3Both("ended", function(ev){
    if(ev.target!==mp3Audio) return;
    if(!mp3SessionPlaying||mp3Audio.loop){
      stopMediaSessionBackgroundKeepalive();
      setPlayLabel();
      return;
    }
    mp3LoadingTrack=true;
    var nextHint=(prefetchName&&prefetchToken===mp3PickToken)?prefetchName:"";
    if(!nextHint){
      var poolEnd=getMp3Pool();
      if(poolEnd.length) nextHint=pickPrefetchName();
    }
    if(nextHint) beginTrackTransitionTo(nextHint);
    syncMediaSessionPlaybackState();
    setPlayLabel();
    if(typeof document!=="undefined"&&mp3PageIsHidden()&&mp3SessionPlaying){
      startMediaSessionBackgroundKeepalive();
    }
    if(tryPlayPrefetched()) return;
    pickRandomMp3(true, true);
  });
  window.addEventListener("pagehide", function(ev){
    writeMp3Handoff(ev);
  });
  window.addEventListener("beforeunload", function(){
    flushMp3Handoff({fromHide:true});
  });
  document.addEventListener("visibilitychange", function(){
    if(typeof document!=="undefined"&&document.visibilityState==="hidden") flushMp3Handoff({fromHide:true});
  });
  document.addEventListener("click", function(e){
    handoffFlushFromNavTarget(e.target);
  }, true);
  document.addEventListener("touchstart", function(e){
    handoffFlushFromNavTarget(e.target);
  }, {capture:true, passive:true});
  document.addEventListener("touchend", function(e){
    handoffFlushFromNavTarget(e.target);
  }, {capture:true, passive:true});
  if(typeof window!=="undefined") window.__fsMp3FlushHandoff=function(){ flushMp3Handoff({force:true}); };
  setPlayLabel();
  syncLoopButtonUI();
  if(window.__fsMp3LazyList) setMp3SearchEnabled(true);
  bootMp3List();
};


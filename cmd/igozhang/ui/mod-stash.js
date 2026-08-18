window.__fsStashInit=function(){
  if(!document.getElementById('stashText')) return;
  var __fs=window.__fs;
  function absRoot(){ return __fs.absRoot(); }
  function urlStash(){ return __fs.urlStash(); }
  function urlStashExport(){ return __fs.urlStashExport(); }
  function utf8ByteLength(s){ return __fs.utf8ByteLength(s); }
  function stashSnippet100(s){ return __fs.stashSnippet100(s); }
  function showToast(t,e){ return __fs.showToast(t,e); }
  var maxStashBytes=__fs.maxStashBytes;
  var stashServerUpdatedAt=0;
  (function(){
    var ta=document.getElementById("stashText");
    var hint=document.getElementById("stashByteHint");
    var snip=document.getElementById("stashSnippet");
    var saveMsg=document.getElementById("stashSaveMsg");
    var btnSave=document.getElementById("btnStashSave");
    var btnCopy=document.getElementById("btnStashCopy");
    var btnExport=document.getElementById("btnStashExport");
    var btnClear=document.getElementById("btnStashClear");
    var conflictBox=document.getElementById("stashConflictBox");
    var conflictMsg=document.getElementById("stashConflictMsg");
    var btnConflictLoad=document.getElementById("btnStashConflictLoad");
    var btnConflictOverwrite=document.getElementById("btnStashConflictOverwrite");
    var btnConflictDismiss=document.getElementById("btnStashConflictDismiss");
    if(!ta||!btnSave) return;
    var stashInlineTimer=null;
    var stashInlineHideGen=0;
    var stashListETag="";
    var stashLastAppliedText="";
    var stashRemoteBumpNt=-1;
    var stashPollDelay=__fs.stashPollIntervalMs||3000;
    var stashPollDelayMax=__fs.stashPollIntervalMaxMs||30000;
    var stashPollT=null;
    var stashPollReady=false;
    var stashPollInFlight=false;
    function stashInlineMsg(text, cls, opts){
      opts=opts||{};
      if(!saveMsg) return;
      if(stashInlineTimer){ clearTimeout(stashInlineTimer); stashInlineTimer=null; }
      stashInlineHideGen++;
      var myGen=stashInlineHideGen;
      saveMsg.textContent=text;
      saveMsg.className=cls;
      if(opts.persistent) return;
      var hideMs=typeof opts.hideMs==="number"?opts.hideMs:(cls==="msg err"?6000:4200);
      stashInlineTimer=setTimeout(function(){
        stashInlineTimer=null;
        if(myGen!==stashInlineHideGen) return;
        saveMsg.textContent="";
        saveMsg.className="msg";
      }, hideMs);
    }
    function getStashText(){ return ta.value; }
    function stashETagClient(ts){
      var n=(typeof ts==="number"&&!isNaN(ts))?ts:0;
      return '"stash-v1-'+n+'"';
    }
    function stashTruncated(s){
      var n=0,i=0;
      for(;i<s.length&&n<100;n++){
        var cp=s.codePointAt(i);
        i+=(cp>0xffff)?2:1;
      }
      return i<s.length;
    }
    function updateStashUI(){
      var t=getStashText();
      var bytes=utf8ByteLength(t);
      if(hint) hint.textContent="约 "+bytes+" / "+maxStashBytes+" 字节（UTF-8）";
      var has=t.length>0;
      if(btnCopy) btnCopy.disabled=!has;
      if(snip){
        if(has){
          snip.textContent="前 100 字："+stashSnippet100(t)+(stashTruncated(t)?"…":"");
          snip.hidden=false;
        } else {
          snip.textContent="";
          snip.hidden=true;
        }
      }
    }
    function resetStashPollBackoff(){ stashPollDelay=__fs.stashPollIntervalMs||3000; }
    function bumpStashPollBackoff(){ stashPollDelay=Math.min(stashPollDelayMax, Math.floor(stashPollDelay*1.35)); }
    function stashScheduleNext(){
      if(stashPollT) clearTimeout(stashPollT);
      stashPollT=setTimeout(function(){
        stashPollT=null;
        if(typeof document!=="undefined"&&document.visibilityState!=="visible"){
          return;
        }
        stashPollOnce({notify:false});
      }, stashPollDelay);
    }
    function finishStashPoll(backoff){
      stashPollInFlight=false;
      if(backoff) bumpStashPollBackoff();
      else resetStashPollBackoff();
      stashScheduleNext();
    }
    function hideStashConflict(){
      if(!conflictBox) return;
      conflictBox.hidden=true;
      if(btnConflictLoad) btnConflictLoad.onclick=null;
      if(btnConflictOverwrite) btnConflictOverwrite.onclick=null;
      if(btnConflictDismiss) btnConflictDismiss.onclick=null;
    }
    function stashPollOnce(opts){
      opts=opts||{};
      function early(){
        stashScheduleNext();
      }
      if(!stashPollReady||stashPollInFlight){
        early();
        return;
      }
      if(!__fs.getAuthToken||!__fs.getAuthToken()){
        return;
      }
      if(__fs.requireAuth&&!__fs.requireAuth()){
        return;
      }
      if(conflictBox&&!conflictBox.hidden){
        early();
        return;
      }
      if(btnSave&&btnSave.disabled){
        early();
        return;
      }
      if(btnClear&&btnClear.disabled){
        early();
        return;
      }
      stashPollInFlight=true;
      var hdr=__fs.authHeaders();
      if(stashListETag) hdr["If-None-Match"]=stashListETag;
      fetch(urlStash()+"?_="+Date.now(),{cache:"no-store",credentials:"same-origin",headers:hdr})
        .then(function(r){
          if(r.status===401){
            return r.json().then(function(j){
              if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,j);
              finishStashPoll(true);
            }).catch(function(){
              if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,null);
              finishStashPoll(true);
            });
          }
          if(r.status===304){
            finishStashPoll(true);
            return null;
          }
          if(r.status===429){
            showToast("请求过于频繁，请稍后再试",true);
            finishStashPoll(true);
            return null;
          }
          if(!r.ok) throw new Error();
          var et=r.headers.get("ETag");
          if(et) stashListETag=et.trim();
          return r.json();
        })
        .then(function(j){
          if(!j) return;
          var nt=(typeof j.updatedAt==="number")?j.updatedAt:0;
          var srv=(typeof j.text==="string")?j.text:"";
          if(nt===stashServerUpdatedAt){
            finishStashPoll(true);
            return;
          }
          if(getStashText()!==stashLastAppliedText){
            if(stashRemoteBumpNt!==nt){
              stashRemoteBumpNt=nt;
              stashInlineMsg("服务端暂存已更新。您本地有未同步的编辑，请保存、导出备份，或使用冲突页「加载服务端」后再同步。","msg err",{persistent:true});
              if(opts.notify) showToast("服务端暂存已更新（与本地不一致）",true);
            }
            finishStashPoll(true);
            return;
          }
          var prev=ta.value;
          ta.value=srv;
          stashServerUpdatedAt=nt;
          stashLastAppliedText=srv;
          stashRemoteBumpNt=-1;
          hideStashConflict();
          updateStashUI();
          if(opts.notify&&prev!==srv) showToast("暂存已由服务端更新",false);
          finishStashPoll(false);
        })
        .catch(function(){ finishStashPoll(true); });
    }
    function loadStash(){
      if(__fs.requireAuth&&!__fs.requireAuth()) return;
      hideStashConflict();
      var hdr=__fs.authHeaders();
      if(stashListETag) hdr["If-None-Match"]=stashListETag;
      fetch(urlStash(),{credentials:"same-origin",headers:hdr})
        .then(function(r){
          if(r.status===401){
            return r.json().then(function(j){
              if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,j);
              return null;
            }).catch(function(){
              if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,null);
              return null;
            });
          }
          if(r.status===304){
            stashPollReady=true;
            updateStashUI();
            resetStashPollBackoff();
            stashScheduleNext();
            return null;
          }
          if(!r.ok) throw new Error();
          var et=r.headers.get("ETag");
          if(et) stashListETag=et.trim();
          return r.json();
        })
        .then(function(j){
          if(!j) return;
          if(j&&typeof j.text==="string"){ ta.value=j.text; }
          stashServerUpdatedAt=(j&&typeof j.updatedAt==="number")?j.updatedAt:0;
          stashLastAppliedText=(j&&typeof j.text==="string")?j.text:"";
          stashRemoteBumpNt=-1;
          updateStashUI();
          stashPollReady=true;
          resetStashPollBackoff();
          stashScheduleNext();
        })
        .catch(function(){
          showToast("暂存加载失败",true);
          updateStashUI();
          stashPollReady=true;
          resetStashPollBackoff();
          stashScheduleNext();
        });
    }
    function postStashWithMatch(t, matchTs){
      return fetch(urlStash(),{method:"POST",credentials:"same-origin",headers:__fs.authHeaders({"Content-Type":"application/json"}),body:JSON.stringify({text:t,ifMatchUpdatedAt:matchTs})});
    }
    function fmtConflictTime(ts){
      if(!ts) return "";
      try{
        if(window.__fsFmtBeijing) return window.__fsFmtBeijing(ts);
        var u=new Date(ts*1000),ms=Date.UTC(u.getUTCFullYear(),u.getUTCMonth(),u.getUTCDate(),u.getUTCHours()+8,u.getUTCMinutes(),u.getUTCSeconds()),b=new Date(ms),p=function(n){return(n<10?"0":"")+n};
        return b.getUTCFullYear()+"-"+p(b.getUTCMonth()+1)+"-"+p(b.getUTCDate())+" "+p(b.getUTCHours())+":"+p(b.getUTCMinutes())+":"+p(b.getUTCSeconds());
      }catch(e){ return String(ts); }
    }
    function handleStash409(r){
      return r.json().then(function(j){
        var srv=j.serverUpdatedAt||0;
        var srvText=typeof j.serverText==="string"?j.serverText:"";
        if(!conflictBox||!btnConflictLoad||!btnConflictOverwrite||!btnConflictDismiss){
          stashInlineMsg("冲突：无法显示操作区","msg err");
          return Promise.reject(new Error("stash conflict UI missing"));
        }
        if(conflictMsg){
          conflictMsg.textContent="服务端内容已在其他设备更新（"+fmtConflictTime(srv)+"），与当前编辑不一致。请选择：";
        }
        conflictBox.hidden=false;
        try{ conflictBox.scrollIntoView({behavior:"smooth",block:"nearest"}); }catch(e){}
        return new Promise(function(resolve,reject){
          function cleanup(){ hideStashConflict(); }
          btnConflictLoad.onclick=function(){
            cleanup();
            ta.value=srvText;
            stashServerUpdatedAt=srv;
            stashLastAppliedText=srvText;
            stashListETag=stashETagClient(srv);
            updateStashUI();
            stashInlineMsg("已加载服务端版本","msg ok");
            showToast("已加载服务端版本",false);
            resolve(null);
          };
          btnConflictOverwrite.onclick=function(){
            cleanup();
            postStashWithMatch(getStashText(), srv)
              .then(function(r2){
                if(!r2.ok){
                  if(r2.status===409) return handleStash409(r2);
                  return r2.text().then(function(x){ throw new Error(x.slice(0,200)); });
                }
                return r2.json();
              })
              .then(function(j){
                if(j&&typeof j.updatedAt==="number") stashServerUpdatedAt=j.updatedAt;
                if(j&&typeof j.updatedAt==="number") stashListETag=stashETagClient(j.updatedAt);
                stashLastAppliedText=getStashText();
                resolve(j);
              })
              .catch(reject);
          };
          btnConflictDismiss.onclick=function(){
            cleanup();
            stashInlineMsg("已取消保存","msg");
            resolve(null);
          };
        });
      });
    }
    ta.addEventListener("input", updateStashUI);
    function stashExportErrMsg(raw){
      var s=String(raw||"").trim();
      if(s==="stash empty") return "暂存为空，未写入文件";
      return s||"导出失败";
    }
    if(btnExport){
      btnExport.onclick=function(){
        if(__fs.requireAuth&&!__fs.requireAuth()) return;
        btnExport.disabled=true;
        fetch(urlStashExport(),{method:"POST",credentials:"same-origin",headers:__fs.authHeaders()})
          .then(function(r){
            return r.text().then(function(x){
              var j=null;
              try{ j=JSON.parse(x); }catch(e){}
              if(r.status===401){
                if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,j);
                throw new Error((j&&j.error)||"未认证");
              }
              if(!r.ok){
                var em=(j&&j.error)||x.slice(0,200);
                throw new Error(stashExportErrMsg(em));
              }
              return j||{};
            });
          })
          .then(function(j){
            var path=j&&j.saved||"";
            var nb=j&&typeof j.bytes==="number"?j.bytes:0;
            if(nb<=0||!path){
              var err=new Error("导出响应无效");
              throw err;
            }
            stashInlineMsg("已导出："+path+"（"+nb+" 字节）","msg ok");
            showToast("已导出 "+path,false);
            __fs.bus.emit('usage:refresh');
          })
          .catch(function(e){
            var msg=stashExportErrMsg(e&&e.message||e);
            showToast(msg,true);
            stashInlineMsg(msg,"msg err");
          })
          .finally(function(){ btnExport.disabled=false; });
      };
    }
    btnSave.onclick=function(){
      if(__fs.requireAuth&&!__fs.requireAuth()) return;
      var t=getStashText();
      if(utf8ByteLength(t)>maxStashBytes){
        stashInlineMsg("超过 200 KB 上限","msg err");
        showToast("暂存超过 200 KB 上限",true);
        return;
      }
      stashInlineMsg("保存中…","msg",{persistent:true});
      btnSave.disabled=true;
      postStashWithMatch(t, stashServerUpdatedAt)
        .then(function(r){
          if(r.status===401){
            return r.json().then(function(j){
              if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,j);
              throw new Error((j&&j.error)||"未认证");
            });
          }
          if(r.status===409) return handleStash409(r);
          if(!r.ok) return r.text().then(function(x){ throw new Error(x.slice(0,200)); });
          return r.json();
        })
        .then(function(j){
          if(!j) return;
          if(typeof j.updatedAt==="number"){
            stashServerUpdatedAt=j.updatedAt;
            stashListETag=stashETagClient(j.updatedAt);
          }
          stashLastAppliedText=getStashText();
          stashInlineMsg("已暂存到服务器","msg ok");
          showToast("暂存已保存",false);
        })
        .catch(function(e){
          stashInlineMsg("保存失败："+(e&&e.message||e),"msg err");
          showToast("暂存保存失败",true);
        })
        .finally(function(){ btnSave.disabled=false; });
    };
    btnCopy.onclick=function(){
      var t=getStashText();
      if(!t) return;
      if(navigator.clipboard&&navigator.clipboard.writeText){
        navigator.clipboard.writeText(t).then(function(){
          stashInlineMsg("已复制到剪贴板","msg ok");
        }).catch(function(){
          stashInlineMsg("复制失败（浏览器权限）","msg err");
        });
      } else {
        ta.select();
        try{
          document.execCommand("copy");
          stashInlineMsg("已复制到剪贴板","msg ok");
        }catch(err){
          stashInlineMsg("复制失败","msg err");
        }
      }
    };
    if(btnClear){
      btnClear.onclick=function(){
        if(__fs.requireAuth&&!__fs.requireAuth()) return;
        stashInlineMsg("清空中…","msg",{persistent:true});
        btnClear.disabled=true;
        fetch(urlStash(),{method:"DELETE",credentials:"same-origin",headers:__fs.authHeaders()})
          .then(function(r){
            if(r.status===401){
              return r.json().then(function(j){
                if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,j);
                throw new Error((j&&j.error)||"未认证");
              });
            }
            if(!r.ok) return r.text().then(function(x){ throw new Error(x.slice(0,200)); });
            return r.json();
          })
          .then(function(){
            ta.value="";
            stashServerUpdatedAt=0;
            stashLastAppliedText="";
            stashListETag=stashETagClient(0);
            updateStashUI();
            stashInlineMsg("已清空","msg ok");
            showToast("暂存已清空",false);
          })
          .catch(function(e){
            stashInlineMsg("清空失败："+(e&&e.message||e),"msg err");
            showToast("暂存清空失败",true);
          })
          .finally(function(){ btnClear.disabled=false; });
      };
    }
    document.addEventListener("fs:pagehidden", function(){
      if(stashPollT){ clearTimeout(stashPollT); stashPollT=null; }
    });
    document.addEventListener("fs:pagevisible", function(){
      resetStashPollBackoff();
      if(stashPollT){ clearTimeout(stashPollT); stashPollT=null; }
      stashPollOnce({notify:true});
    });
    window.addEventListener("online", function(){
      resetStashPollBackoff();
      if(stashPollT){ clearTimeout(stashPollT); stashPollT=null; }
      stashPollOnce({notify:false});
    });
    __fs.bus.on("auth:logout", function(){
      stashPollReady=false;
      if(stashPollT){ clearTimeout(stashPollT); stashPollT=null; }
      hideStashConflict();
    });
    __fs.bus.on("auth:ok", function(){
      stashPollReady=false;
      if(stashPollT){ clearTimeout(stashPollT); stashPollT=null; }
      loadStash();
    });
    loadStash();
  })();

};

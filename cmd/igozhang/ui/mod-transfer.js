window.__fsTransferInit=function(){
  if(!document.getElementById('btnFetch')) return;
  var __fs=window.__fs;
  function urlFsUsage(){ return __fs.urlFsUsage(); }
  function urlFetch(){ return __fs.urlFetch(); }
  function urlUpload(){ return __fs.urlUpload(); }
  function fmtSize(n){ return __fs.fmtSize(n); }
  function fmtBytesAbbr(n){ return __fs.fmtBytesAbbr(n); }
  var fetchProgTotal=-1;
  function setFetchProgRow(show, bytes){
    var row=document.getElementById("fetchProgressRow");
    var bar=document.getElementById("fetchBar");
    var lab=document.getElementById("fetchProgressLabel");
    if(!row||!lab) return;
    if(!show){ row.hidden=true; fetchProgTotal=-1; if(bar) bar.hidden=true; lab.textContent=""; return; }
    row.hidden=false;
    var total=fetchProgTotal;
    if(bar){
      if(total>0){
        bar.hidden=false;
        bar.max=100;
        bar.value=Math.min(100,Math.floor(100*bytes/total));
        lab.textContent="拉取中… "+fmtSize(bytes)+" / "+fmtSize(total)+" ("+Math.min(99,Math.floor(100*bytes/total))+"%)";
      } else {
        bar.hidden=true;
        lab.textContent="拉取中… 已接收 "+fmtSize(bytes);
      }
    }
  }
  function consumeFetchNdjson(resp){
    return new Promise(function(resolve,reject){
      if(!resp.body||typeof resp.body.getReader!=="function"){
        reject(new Error("浏览器不支持流式进度"));
        return;
      }
      var reader=resp.body.getReader();
      var dec=new TextDecoder();
      var buf="";
      var settled=false;
      function finish(err, val){
        if(settled) return;
        settled=true;
        if(err) reject(err); else resolve(val);
      }
      function pump(){
        reader.read().then(function(res){
          if(res.done){
            if(buf.trim()){
              try{
                var lj=JSON.parse(buf.trim());
                if(lj.type==="done"){ setFetchProgRow(false); finish(null,lj); return; }
              }catch(e){}
            }
            finish(new Error("连接中断"));
            return;
          }
          buf+=dec.decode(res.value,{stream:true});
          var lines=buf.split("\n");
          buf=lines.pop()||"";
          for(var i=0;i<lines.length;i++){
            var line=lines[i].trim();
            if(!line) continue;
            var j;
            try{ j=JSON.parse(line); }catch(e){ continue; }
            if(j.type==="start"){
              fetchProgTotal=(typeof j.total==="number"&&j.total>=0)?j.total:-1;
              setFetchProgRow(true,0);
            } else if(j.type==="progress"){
              setFetchProgRow(true,j.bytes||0);
            } else if(j.type==="done"){
              setFetchProgRow(false);
              finish(null,j);
              try{ reader.cancel(); }catch(e){}
              return;
            } else if(j.type==="error"){
              setFetchProgRow(false);
              finish(new Error(j.error||"拉取失败"));
              try{ reader.cancel(); }catch(e){}
              return;
            }
          }
          pump();
        }).catch(function(e){ finish(e||new Error("读取失败")); });
      }
      pump();
    });
  }
  var uploadProgTotal=-1;
  function setUploadProgRow(show, loaded){
    var row=document.getElementById("uploadProgressRow");
    var bar=document.getElementById("uploadBar");
    var lab=document.getElementById("uploadProgressLabel");
    if(!row||!lab) return;
    if(!show){ row.hidden=true; uploadProgTotal=-1; if(bar) bar.hidden=true; lab.textContent=""; return; }
    row.hidden=false;
    var total=uploadProgTotal;
    if(bar){
      if(total>0){
        bar.hidden=false;
        bar.max=100;
        bar.value=Math.min(100,Math.floor(100*loaded/total));
        lab.textContent="上传中… "+fmtSize(loaded)+" / "+fmtSize(total)+" ("+Math.min(99,Math.floor(100*loaded/total))+"%)";
      } else {
        bar.hidden=true;
        lab.textContent="上传中… 已发送 "+fmtSize(loaded);
      }
    }
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
  function loadFsUsage(){
    if(__fs.requireAuth&&!__fs.requireAuth()) return;
    fetch(urlFsUsage(),{credentials:"same-origin",headers:__fs.authHeaders()}).then(function(r){
      if(r.status===401){
        return r.json().then(function(j){
          if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,j);
        }).catch(function(){ if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,null); });
      }
      if(!r.ok) throw new Error();
      return r.json();
    }).then(function(j){
      if(!j) return;
      var block=document.getElementById("fsUsageBlock");
      var bar=document.getElementById("fsUsageBar");
      if(!j||!block||!bar) return;
      var u=Number(j.bytesUsed)||0;
      var base=Number(j.baselineBytes)||1;
      var cap=Number(j.capBytes)||0;
      var rawPct=base>0?(100*u/base):0;
      var dispPct=u<=0?0:(rawPct<1?1:Math.min(100,Math.round(rawPct)));
      var pctTip=u<=0?"0.0":(Math.round(rawPct*10)/10).toFixed(1);
      var tip=pctTip+"% "+fmtBytesAbbr(u)+" / "+fmtBytesAbbr(base)+(cap!==base?" · 上限 "+fmtBytesAbbr(cap):"");
      bar.value=isFinite(dispPct)?dispPct:0;
      bar.classList.remove("fs-usage-bar-dgreen","fs-usage-bar-lgreen","fs-usage-bar-red","fs-usage-bar-dred");
      if(rawPct>=90){
        bar.classList.add("fs-usage-bar-dred");
      } else if(rawPct>=70){
        bar.classList.add("fs-usage-bar-red");
      } else if(rawPct>=30){
        bar.classList.add("fs-usage-bar-lgreen");
      } else {
        bar.classList.add("fs-usage-bar-dgreen");
      }
      bar.title=tip;
      block.title=tip;
      block.hidden=false;
    }).catch(function(){});
  }
  var fetchMsgHideTimer=null;
  var uploadMsgHideTimer=null;
  function clearFetchMsgLater(ms){
    if(fetchMsgHideTimer) clearTimeout(fetchMsgHideTimer);
    fetchMsgHideTimer=setTimeout(function(){
      fetchMsgHideTimer=null;
      var el=document.getElementById("fetchMsg");
      if(el){ el.textContent=""; el.className="msg"; }
    }, ms);
  }
  function clearUploadMsgLater(ms){
    if(uploadMsgHideTimer) clearTimeout(uploadMsgHideTimer);
    uploadMsgHideTimer=setTimeout(function(){
      uploadMsgHideTimer=null;
      var el=document.getElementById("uploadMsg");
      if(el){ el.textContent=""; el.className="msg"; }
    }, ms);
  }
  var fetchAbort=null;
  document.getElementById("btnFetch").onclick=function(){
    if(__fs.requireAuth&&!__fs.requireAuth()) return;
    var u=document.getElementById("url").value.trim();
    var msg=document.getElementById("fetchMsg");
    if(fetchMsgHideTimer){ clearTimeout(fetchMsgHideTimer); fetchMsgHideTimer=null; }
    msg.textContent=""; msg.className="msg";
    setFetchProgRow(false);
    if(!u){ msg.textContent="请输入 URL"; msg.className="msg err"; clearFetchMsgLater(4500); return; }
    var btn=document.getElementById("btnFetch");
    var btnC=document.getElementById("btnFetchCancel");
    if(fetchAbort) try{ fetchAbort.abort(); }catch(e){}
    fetchAbort=new AbortController();
    btn.disabled=true;
    if(btnC){ btnC.hidden=false; btnC.disabled=false; }
    fetch(urlFetch(),{method:"POST",credentials:"same-origin",headers:__fs.authHeaders({"Content-Type":"application/json"}),body:JSON.stringify({url:u}),signal:fetchAbort.signal})
      .then(function(r){
        if(r.status===401){
          return r.json().then(function(j){
            if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,j);
            throw {badHttp:true,j:j||{},status:401};
          }).catch(function(e){
            if(e&&e.badHttp) throw e;
            if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,null);
            throw {badHttp:true,j:{error:"未认证"},status:401};
          });
        }
        if(!r.ok){
          return r.text().then(function(t){
            var j={};
            try{ j=t?JSON.parse(t):{}; }catch(e){ j={error:t||("HTTP "+r.status)}; }
            throw {badHttp:true,j:j,status:r.status};
          });
        }
        var ct=(r.headers.get("Content-Type")||"");
        if(ct.indexOf("ndjson")<0){
          return r.text().then(function(t){
            var j={};
            try{ j=t?JSON.parse(t):{}; }catch(e){ j={}; }
            throw new Error(j.error||t||"响应异常");
          });
        }
        return consumeFetchNdjson(r);
      })
      .then(function(j){
        msg.textContent="已保存："+(j.saved||"")+"（"+(j.bytes||0)+" 字节）";
        msg.className="msg ok";
        showToast("拉取完成",false);
        __fs.bus.emit('browse:refresh'); __fs.bus.emit('usage:refresh');
        var urlEl=document.getElementById("url");
        if(urlEl) urlEl.value="";
        clearFetchMsgLater(5200);
      })
      .catch(function(e){
        setFetchProgRow(false);
        if(e&&e.name==="AbortError"){
          msg.textContent="已取消";
          msg.className="msg";
          showToast("已取消拉取",false);
          clearFetchMsgLater(4000);
          return;
        }
        if(e&&e.badHttp){
          msg.textContent=(e.j&&e.j.error)||("失败 "+e.status);
          showToast(msg.textContent,true);
        } else {
          msg.textContent=(e&&e.message)||"网络错误";
          showToast(msg.textContent,true);
        }
        msg.className="msg err";
        clearFetchMsgLater(9000);
      })
      .finally(function(){
        btn.disabled=false;
        if(btnC){ btnC.hidden=true; }
        fetchAbort=null;
      });
  };
  var btnFetchCancelEl=document.getElementById("btnFetchCancel");
  if(btnFetchCancelEl) btnFetchCancelEl.onclick=function(){
    if(fetchAbort) fetchAbort.abort();
  };
  var uploadXhr=null;
  (function(){
    var btnUp=document.getElementById("btnUpload");
    var btnUpC=document.getElementById("btnUploadCancel");
    var inp=document.getElementById("uploadInput");
    var upMsg=document.getElementById("uploadMsg");
    if(!btnUp||!inp) return;
    btnUp.onclick=function(){ inp.click(); };
    if(btnUpC) btnUpC.onclick=function(){
      if(uploadXhr){ try{ uploadXhr.abort(); }catch(e){} uploadXhr=null; }
    };
    inp.onchange=function(){
      if(!inp.files||inp.files.length<1) return;
      if(__fs.requireAuth&&!__fs.requireAuth()){ inp.value=""; return; }
      var fd=new FormData();
      for(var i=0;i<inp.files.length;i++) fd.append("file", inp.files[i]);
      if(uploadMsgHideTimer){ clearTimeout(uploadMsgHideTimer); uploadMsgHideTimer=null; }
      if(upMsg){ upMsg.textContent=""; upMsg.className="msg"; }
      uploadProgTotal=-1;
      setUploadProgRow(true,0);
      btnUp.disabled=true;
      if(btnUpC){ btnUpC.hidden=false; }
      var xhr=new XMLHttpRequest();
      uploadXhr=xhr;
      xhr.open("POST", urlUpload());
      xhr.withCredentials=true;
      var tok=__fs.getAuthToken&&__fs.getAuthToken();
      if(tok) xhr.setRequestHeader("X-Igo-Token",tok);
      xhr.responseType="json";
      xhr.upload.onprogress=function(ev){
        if(ev.lengthComputable&&ev.total>0) uploadProgTotal=ev.total;
        setUploadProgRow(true,ev.loaded||0);
      };
      xhr.onerror=function(){
        setUploadProgRow(false);
        if(upMsg){ upMsg.textContent="网络错误"; upMsg.className="msg err"; }
        showToast("上传失败",true);
        clearUploadMsgLater(9000);
        btnUp.disabled=false;
        if(btnUpC) btnUpC.hidden=true;
        uploadXhr=null;
        inp.value="";
      };
      xhr.onabort=function(){
        setUploadProgRow(false);
        if(upMsg){ upMsg.textContent="已取消"; upMsg.className="msg"; }
        showToast("已取消上传",false);
        clearUploadMsgLater(4000);
        btnUp.disabled=false;
        if(btnUpC) btnUpC.hidden=true;
        uploadXhr=null;
        inp.value="";
      };
      xhr.onload=function(){
        setUploadProgRow(false);
        btnUp.disabled=false;
        if(btnUpC) btnUpC.hidden=true;
        uploadXhr=null;
        inp.value="";
        if(!upMsg) return;
        if(xhr.status===401){
          var aj=xhr.response;
          if(aj&&typeof aj==="string"){ try{ aj=JSON.parse(aj); }catch(e){ aj=null; } }
          if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,aj);
          upMsg.textContent=(aj&&aj.error)||"未认证";
          upMsg.className="msg err";
          showToast(upMsg.textContent,true);
          clearUploadMsgLater(9000);
          return;
        }
        if(xhr.status>=200&&xhr.status<300){
          var j=xhr.response;
          if(j&&typeof j==="string"){ try{ j=JSON.parse(j); }catch(e){ j={}; } }
          if(!j) j={};
          var paths=j.saved;
          var line=Array.isArray(paths)?paths.join("，"):(paths||"");
          upMsg.textContent="已上传："+line+"（共 "+(j.bytes||0)+" 字节）";
          upMsg.className="msg ok";
          showToast("上传完成",false);
          __fs.bus.emit('browse:refresh'); __fs.bus.emit('usage:refresh');
          clearUploadMsgLater(5200);
        } else {
          var err;
          if(xhr.responseType!=="json"||typeof xhr.response!=="object"){
            try{ err=JSON.parse(xhr.responseText||"{}"); }catch(e){ err={error:xhr.responseText||("HTTP "+xhr.status)}; }
          } else {
            err=xhr.response;
          }
          var et=(err&&err.error)||("失败 "+xhr.status);
          upMsg.textContent=et;
          upMsg.className="msg err";
          showToast(et,true);
          clearUploadMsgLater(9000);
        }
      };
      xhr.send(fd);
    };
  })();
  __fs.bus.emit('browse:refresh');
  __fs.bus.emit('usage:refresh');

  __fs.bus.on('usage:refresh', function(){ loadFsUsage(); });
  __fs.bus.on('auth:ok', function(){ loadFsUsage(); });
  loadFsUsage();
};

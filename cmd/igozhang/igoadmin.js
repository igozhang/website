
(function(){
  var apiBase=window.__IGOADMIN_BASE__||"";
  function api(p){ return apiBase+p; }
  var auth=window.__igoAuthInit({
    apiBase: apiBase,
    onReady: function(){ loadCategories(); newSaveBaseline=newFormSnapshot(); newCreatedSlug=""; startAutoSave(); },
    onLogout: function(){ stopAutoSave(); newCreatedSlug=""; }
  });
  function getToken(){ return auth.getToken(); }
  function requireAuth(){ return auth.requireAuth(); }
  function authHeaders(extra){ return auth.authHeaders(extra); }
  function handleAuthFailure(status,j){ return auth.handleAuthFailure(status,j); }
  var newMsg=document.getElementById("newMsg");
  var editMsg=document.getElementById("editMsg");
  var batchMsg=document.getElementById("batchMsg");
  function msg(el,s,err){
    if(!el) return;
    el.textContent=s||"";
    var backup=el.id==="backupMsg";
    el.className="igoadmin-msg "+(err?"err":"ok")+(backup?" igoadmin-backup-msg":"");
    if(s&&!backup) try{ el.scrollIntoView({block:"nearest",behavior:"smooth"}); }catch(e){}
  }
  function msgTa(ta,s,err){
    msg(ta&&ta.id==="editMd"?editMsg:newMsg,s,err);
  }
  function fmtBJ(ts){
    if(window.__fsFmtBeijing) return window.__fsFmtBeijing(ts);
    var ex=Number(window.__fsBeijingOffsetHours); if(!isFinite(ex)) ex=0; if(ex>12) ex=12; if(ex<-12) ex=-12; ex=ex|0;
    var u=new Date(ts*1000),ms=Date.UTC(u.getUTCFullYear(),u.getUTCMonth(),u.getUTCDate(),u.getUTCHours()+8+ex,u.getUTCMinutes(),u.getUTCSeconds());
    var b=new Date(ms),p=function(n){return(n<10?"0":"")+n};
    return b.getUTCFullYear()+"-"+p(b.getUTCMonth()+1)+"-"+p(b.getUTCDate())+" "+p(b.getUTCHours())+":"+p(b.getUTCMinutes())+":"+p(b.getUTCSeconds());
  }
  var articleSaving=false;
  var newSaveBaseline="", newCreatedSlug="";
  function modeMsgEl(){
    if(articleMode==="new") return newMsg;
    if(articleMode==="edit") return editMsg;
    if(articleMode==="batch") return batchMsg;
    return editMsg;
  }
  function modeMsg(text,err,warn){
    var el=modeMsgEl();
    if(!el) return;
    el.textContent=text||"";
    el.className="igoadmin-msg "+(err?"err":warn?"is-warn":"ok");
    if(text&&!err) try{ el.scrollIntoView({block:"nearest",behavior:"smooth"}); }catch(e){}
  }
  function parseJSONResponse(r){
    return r.json().then(function(j){ return {r:r,j:j}; }).catch(function(){ return {r:r,j:null}; });
  }
  function apiFetchJSON(p,opt){
    if(!requireAuth()) return Promise.reject(new Error("auth"));
    opt=opt||{};
    var headers=authHeaders(opt.headers);
    if(opt.body&&!(opt.body instanceof FormData)&&!headers["Content-Type"]){
      headers["Content-Type"]="application/json";
    }
    return fetch(api(p),Object.assign({},opt,{credentials:"same-origin",headers:headers}))
      .then(parseJSONResponse)
      .then(function(x){
        if(handleAuthFailure(x.r.status,x.j)) return Promise.reject(new Error("auth"));
        return x;
      });
  }
  function apiFetchRaw(p,opt){
    if(!requireAuth()) return Promise.reject(new Error("auth"));
    opt=opt||{};
    return fetch(api(p),Object.assign({},opt,{credentials:"same-origin",headers:authHeaders(opt.headers)}))
      .then(function(r){
        if(r.status===401){
          return parseJSONResponse(r).then(function(x){
            handleAuthFailure(x.r.status,x.j);
            return Promise.reject(new Error("auth"));
          });
        }
        return r;
      });
  }
  function apiXhrJSON(method,p,body,onProgress){
    if(!requireAuth()) return Promise.reject(new Error("auth"));
    return new Promise(function(resolve,reject){
      var xhr=new XMLHttpRequest();
      xhr.open(method,api(p));
      xhr.withCredentials=true;
      var tok=getToken();
      if(tok) xhr.setRequestHeader("X-Igo-Token",tok);
      if(body instanceof FormData){
        xhr.upload.onprogress=function(e){
          if(!onProgress) return;
          if(e.lengthComputable) onProgress(e.loaded/e.total*100);
          else onProgress(-1);
        };
      }
      xhr.onload=function(){
        var j=null;
        try{ j=JSON.parse(xhr.responseText||"null"); }catch(e){}
        var r={ok:xhr.status>=200&&xhr.status<300,status:xhr.status};
        if(r.status===401&&handleAuthFailure(r.status,j)) return reject(new Error("auth"));
        resolve({r:r,j:j});
      };
      xhr.onerror=function(){ reject(new Error("network")); };
      xhr.send(body||null);
    });
  }
  function apiXhrDownload(p,onProgress){
    if(!requireAuth()) return Promise.reject(new Error("auth"));
    return new Promise(function(resolve,reject){
      var xhr=new XMLHttpRequest();
      xhr.open("GET",api(p));
      xhr.withCredentials=true;
      var tok=getToken();
      if(tok) xhr.setRequestHeader("X-Igo-Token",tok);
      xhr.responseType="blob";
      xhr.onprogress=function(e){
        if(!onProgress) return;
        if(e.lengthComputable) onProgress(e.loaded/e.total*100);
        else onProgress(-1);
      };
      xhr.onload=function(){
        if(xhr.status===401){
          var j=null;
          try{ j=JSON.parse(xhr.responseText||"null"); }catch(e){}
          if(handleAuthFailure(401,j)) return reject(new Error("auth"));
        }
        if(xhr.status>=200&&xhr.status<300){
          resolve({
            blob:xhr.response,
            header:function(k){ return xhr.getResponseHeader(k); }
          });
          return;
        }
        reject(new Error("HTTP "+xhr.status));
      };
      xhr.onerror=function(){ reject(new Error("network")); };
      xhr.send();
    });
  }
  var draftCache={}, activeTa=null, draftImgMax=200*1024;
  var draftRefRE=/!\[[^\]]*\]\(__draft__:([a-zA-Z0-9_-]+)\)/g;
  function insertAtCursor(ta,text){
    if(!ta) return;
    var start=ta.selectionStart,end=ta.selectionEnd,val=ta.value;
    ta.value=val.slice(0,start)+text+val.slice(end);
    var pos=start+text.length;
    ta.selectionStart=ta.selectionEnd=pos;
    ta.focus();
  }
  function newDraftID(){
    return "d"+Date.now().toString(36)+Math.random().toString(36).slice(2,8);
  }
  function clearDraftCache(){ draftCache={}; }
  function collectDraftImages(md){
    var out=[], seen={}, m;
    draftRefRE.lastIndex=0;
    while((m=draftRefRE.exec(md))){
      var id=m[1];
      if(seen[id]) continue;
      seen[id]=1;
      if(draftCache[id]) out.push({id:id,data:draftCache[id]});
    }
    return out;
  }
  function articlePayload(base){
    return Object.assign({},base,{draftImages:collectDraftImages(base.markdown||"")});
  }
  function editSlugSlugs(){
    var origEl=document.getElementById("editSlugOriginal");
    var slugEl=document.getElementById("editSlug");
    var oldSlug=origEl?origEl.value.trim():"";
    var slug=slugEl?(applySlugField(slugEl)||slugEl.value.trim()):"";
    if(!oldSlug) oldSlug=slug;
    return {slug:slug, oldSlug:oldSlug};
  }
  function syncEditSlugAfterSave(savedSlug){
    if(!savedSlug) return;
    var slugEl=document.getElementById("editSlug");
    var origEl=document.getElementById("editSlugOriginal");
    if(slugEl) slugEl.value=savedSlug;
    if(origEl) origEl.value=savedSlug;
  }
  function applySavedMarkdown(ta,md,sentMd){
    if(!ta||typeof md!=="string") return false;
    if(ta.value===md){
      clearDraftCache();
      return true;
    }
    if(typeof sentMd==="string"&&ta.value!==sentMd){
      if(ta.value.indexOf("__draft__:")<0) clearDraftCache();
      return false;
    }
    var start=ta.selectionStart,end=ta.selectionEnd;
    ta.value=md;
    var len=ta.value.length;
    ta.selectionStart=Math.min(start,len);
    ta.selectionEnd=Math.min(end,len);
    clearDraftCache();
    return true;
  }
  function pasteImageBlob(blob,ta){
    if(!ta){ msgTa(ta,"请先在 Markdown 编辑框内点击定位光标",true); return; }
    if(blob.size>draftImgMax){ msgTa(ta,"图片超过 200KB",true); return; }
    var mime=blob.type||"";
    if(mime.indexOf("png")<0&&mime.indexOf("jpeg")<0&&mime.indexOf("jpg")<0){
      msgTa(ta,"仅支持 png/jpeg/jpg",true); return;
    }
    var reader=new FileReader();
    reader.onload=function(){
      var data=reader.result;
      if(typeof data!=="string"||data.indexOf("data:image/")!==0){
        msgTa(ta,"读取图片失败",true); return;
      }
      var id=newDraftID();
      draftCache[id]=data;
      insertAtCursor(ta,"![]("+"__draft__:"+id+")");
      markTaTouched(ta);
      msgTa(ta,"已插入草稿图片（保存后落盘）",false);
    };
    reader.readAsDataURL(blob);
  }
  function readClipboardImage(ta){
    if(!requireAuth()) return;
    if(!ta){ msgTa(ta,"请先在 Markdown 编辑框内点击定位光标",true); return; }
    navigator.clipboard.read().then(function(items){
      for(var i=0;i<items.length;i++){
        for(var t=0;t<items[i].types.length;t++){
          if(items[i].types[t].indexOf("image")===0){
            return items[i].getType(items[i].types[t]).then(function(blob){ pasteImageBlob(blob,ta); });
          }
        }
      }
      msgTa(ta,"剪贴板无图片",true);
    }).catch(function(){ msgTa(ta,"无法读取剪贴板",true); });
  }
  function onMdPaste(e){
    var items=e.clipboardData&&e.clipboardData.items;
    if(!items) return;
    for(var i=0;i<items.length;i++){
      if(items[i].type&&items[i].type.indexOf("image")===0){
        var blob=items[i].getAsFile();
        if(blob){
          e.preventDefault();
          pasteImageBlob(blob,e.target);
          return;
        }
      }
    }
  }
  document.querySelectorAll("[data-md-editor]").forEach(function(ta){
    ta.addEventListener("focus",function(){ activeTa=ta; });
    ta.addEventListener("paste",onMdPaste);
  });
  function previewHasMermaid(root){
    if(!root) return false;
    var nodes=root.querySelectorAll("pre");
    for(var i=0;i<nodes.length;i++){
      var pre=nodes[i];
      if(pre.closest&&pre.closest(".doc-mermaid-wrap")) continue;
      var code=pre.querySelector("code");
      var cls=((code&&code.className)||"")+" "+(pre.className||"");
      if(/\blanguage-mermaid\b/i.test(cls)) return true;
    }
    return false;
  }
  var mermaidBundlePromise=null;
  var mermaidScriptLoads={};
  function loadScriptOnce(src){
    return new Promise(function(resolve,reject){
      if(!src){ reject(new Error("empty script src")); return; }
      var entry=mermaidScriptLoads[src];
      if(entry){
        if(entry.ready){ resolve(); return; }
        entry.waiters.push({resolve:resolve,reject:reject});
        return;
      }
      entry={ready:false,waiters:[{resolve:resolve,reject:reject}]};
      mermaidScriptLoads[src]=entry;
      var s=document.createElement("script");
      s.src=src;
      s.async=true;
      s.onload=function(){
        entry.ready=true;
        entry.waiters.forEach(function(w){ w.resolve(); });
        entry.waiters=[];
      };
      s.onerror=function(){
        delete mermaidScriptLoads[src];
        entry.waiters.forEach(function(w){ w.reject(new Error("failed to load "+src)); });
      };
      (document.head||document.documentElement).appendChild(s);
    });
  }
  function ensureMermaidBundle(){
    if(typeof window.initDocMermaid==="function"&&window.mermaid){
      return Promise.resolve();
    }
    if(mermaidBundlePromise) return mermaidBundlePromise;
    var lib=window.__IGO_MERMAID_LIB__||"";
    var thin=window.__IGO_MERMAID_JS__||"";
    mermaidBundlePromise=loadScriptOnce(lib).then(function(){ return loadScriptOnce(thin); }).catch(function(err){
      mermaidBundlePromise=null;
      throw err;
    });
    return mermaidBundlePromise;
  }
  function initMdPreviewCodeChrome(root){
    if(typeof window.initDocCodeChrome==="function") window.initDocCodeChrome(root);
    if(!previewHasMermaid(root)) return Promise.resolve();
    return ensureMermaidBundle().then(function(){
      if(typeof window.initDocMermaid==="function") window.initDocMermaid(root);
    });
  }
  function formatSaveStatus(base,j){
    var s=base||"";
    if(j&&j.validationHint) s+=(s?" · ":"")+j.validationHint;
    return s;
  }
  function setEditStatus(text,kind){
    if(!editMsg) return;
    editMsg.textContent=text||"";
    editMsg.className="igoadmin-msg "+(kind==="err"?"err":kind==="warn"?"is-warn":"ok");
    if(text&&!kind) try{ editMsg.scrollIntoView({block:"nearest",behavior:"smooth"}); }catch(e){}
  }
  function setNewStatus(text,kind){
    if(!newMsg) return;
    newMsg.textContent=text||"";
    newMsg.className="igoadmin-msg "+(kind==="err"?"err":kind==="warn"?"is-warn":"ok");
  }
  function setPreviewMsg(el,text,kind){
    if(!el) return;
    el.textContent=text||"";
    el.className="igoadmin-preview-msg"+(kind==="warn"?" is-warn":kind==="err"?" is-err":"");
  }
  var previewEpochByPanel={};
  function fetchMdPreview(md,panel,btn,msgEl){
    if(!md.trim()){
      if(panel){ panel.innerHTML="<p class=\"igoadmin-meta-hint\">正文为空</p>"; panel.classList.add("is-open"); }
      setPreviewMsg(msgEl,"");
      return;
    }
    var panelKey=(panel&&panel.id)||"preview";
    var epoch=(previewEpochByPanel[panelKey]=(previewEpochByPanel[panelKey]|0)+1);
    if(btn) btn.disabled=true;
    apiFetchJSON("/api/igoadmin/preview",{method:"POST",body:JSON.stringify({markdown:md})})
      .then(function(x){
        if(epoch!==previewEpochByPanel[panelKey]) return;
        if(!x.r.ok){
          if(panel) panel.innerHTML="<p class=\"igoadmin-meta-hint\">"+escHtml((x.j&&x.j.error)||"预览失败")+"</p>";
          setPreviewMsg(msgEl,"预览失败","err");
          return;
        }
        var statusText=x.j&&x.j.hint?"格式提示："+x.j.hint:"与前台渲染一致（正文首行 # 标题已省略）";
        var statusKind=x.j&&x.j.hint?"warn":"";
        if(!panel){
          setPreviewMsg(msgEl,statusText,statusKind);
          return;
        }
        panel.innerHTML=x.j&&x.j.html?x.j.html:"";
        panel.classList.add("is-open");
        if(typeof window.__fsFixFragmentHrefs==="function") try{ window.__fsFixFragmentHrefs(panel); }catch(e){}
        var needMermaid=previewHasMermaid(panel);
        if(needMermaid) setPreviewMsg(msgEl,"图表加载中…","");
        else setPreviewMsg(msgEl,statusText,statusKind);
        return initMdPreviewCodeChrome(panel).then(function(){
          if(epoch!==previewEpochByPanel[panelKey]) return;
          if(needMermaid) setPreviewMsg(msgEl,statusText,statusKind);
        }).catch(function(){
          if(epoch!==previewEpochByPanel[panelKey]) return;
          if(needMermaid) setPreviewMsg(msgEl,"图表库加载失败，源码已保留","warn");
          else setPreviewMsg(msgEl,statusText,statusKind);
        });
      })
      .catch(function(e){
        if(epoch!==previewEpochByPanel[panelKey]) return;
        if(e&&e.message!=="auth"&&panel) panel.innerHTML="<p class=\"igoadmin-meta-hint\">预览失败</p>";
        if(e&&e.message!=="auth") setPreviewMsg(msgEl,"预览失败","err");
      })
      .then(function(){
        if(epoch===previewEpochByPanel[panelKey]&&btn) btn.disabled=false;
      });
  }
  function wireMdPreview(btnId,taId,panelId,msgId){
    var btn=document.getElementById(btnId);
    var ta=document.getElementById(taId);
    var panel=document.getElementById(panelId);
    var msgEl=document.getElementById(msgId);
    if(!btn||!ta||!panel) return;
    var open=false, inputTimer=null;
    btn.addEventListener("click",function(){
      open=!open;
      if(!open){
        panel.classList.remove("is-open");
        btn.textContent="预览";
        setPreviewMsg(msgEl,"");
        return;
      }
      btn.textContent="收起预览";
      fetchMdPreview(ta.value,panel,btn,msgEl);
    });
    ta.addEventListener("input",function(){
      if(!open) return;
      clearTimeout(inputTimer);
      inputTimer=setTimeout(function(){
        fetchMdPreview(ta.value,panel,btn,msgEl);
      },600);
    });
  }
  function closeMdPreview(btnId,panelId,msgId){
    var btn=document.getElementById(btnId);
    var panel=document.getElementById(panelId);
    var msgEl=document.getElementById(msgId);
    if(panel){
      panel.classList.remove("is-open");
      panel.innerHTML="";
    }
    if(btn) btn.textContent="预览";
    setPreviewMsg(msgEl,"");
  }
  function hasNewEditorContent(){
    var slugEl=document.getElementById("newSlug");
    var mdEl=document.getElementById("newMd");
    var tagsEl=document.getElementById("newTags");
    var catEl=document.getElementById("newCat");
    return!!((slugEl&&slugEl.value.trim())||(mdEl&&mdEl.value.trim())||(tagsEl&&tagsEl.value.trim())||(catEl&&catEl.value.trim()));
  }
  function hasEditEditorSession(){
    return!!(document.getElementById("editSlug")||{}).value.trim();
  }
  function hasActiveEditorSession(mode){
    if(mode==="edit") return hasEditEditorSession();
    if(mode==="new") return hasNewEditorContent();
    return false;
  }
  function confirmClearArticle(mode){
    if(articleSaving){
      modeMsg("保存进行中，请稍候",false,true);
      return false;
    }
    if(mode==="edit"&&editUserModified){
      return confirm("当前内容有未保存的修改，确定退出当前编辑？\n未保存的内容将丢失。");
    }
    if(mode==="new"&&newUserModified&&isNewDirty()){
      return confirm("当前内容有未保存的修改，确定退出当前编辑？\n未保存的内容将丢失。");
    }
    return true;
  }
  function clearNewEditor(){
    document.getElementById("newSlug").value="";
    document.getElementById("newCat").value="";
    document.getElementById("newTags").value="";
    document.getElementById("newAuthor").value="igozhang";
    document.getElementById("newDate").value="";
    document.getElementById("newMd").value="";
    closeMdPreview("btnNewPreview","newMdPreview","newPreviewMsg");
    newCreatedSlug="";
    clearDraftCache();
    setNewStatus("");
    updateNewTitleHint();
    newSaveBaseline=newFormSnapshot();
    newUserModified=false;
  }
  function clearEditEditor(){
    editSaveBaseline="";
    editUserModified=false;
    setEditCategoriesLoaded("");
    document.getElementById("editSlug").value="";
    document.getElementById("editSlugOriginal").value="";
    document.getElementById("editSearch").value="";
    document.getElementById("editMd").value="";
    document.getElementById("editTitle").value="";
    document.getElementById("editAuthor").value="";
    document.getElementById("editCat").value="";
    if(editCatEl) editCatEl.dataset.articleCat="";
    document.getElementById("editTags").value="";
    document.getElementById("editDate").value="";
    closeMdPreview("btnEditPreview","editMdPreview","editPreviewMsg");
    closeEditSearch();
    clearDraftCache();
    setEditStatus("");
    resetEditAutoSaveBaseline();
  }
  function exitEditor(mode,skipConfirm){
    if(!skipConfirm&&!confirmClearArticle(mode)) return false;
    if(mode==="new") clearNewEditor();
    else clearEditEditor();
    touchEditorActivity();
    return true;
  }
  function wireMdClear(btnId,mode){
    var btn=document.getElementById(btnId);
    if(!btn) return;
    btn.addEventListener("click",function(){ exitEditor(mode,false); });
  }
  wireMdPreview("btnNewPreview","newMd","newMdPreview","newPreviewMsg");
  wireMdPreview("btnEditPreview","editMd","editMdPreview","editPreviewMsg");
  wireMdClear("btnNewClear","new");
  wireMdClear("btnEditClear","edit");
  var autoSaveTimer=null, editSaveBaseline="";
  var idleClearTimer=null, lastEditorActivity=Date.now(), editIdleClearMs=5*60*1000;
  var editUserModified=false, newUserModified=false;
  function touchEditorActivity(){
    lastEditorActivity=Date.now();
  }
  function markTaTouched(ta){
    if(!ta) return;
    if(ta.id==="editMd") editUserModified=true;
    else if(ta.id==="newMd") newUserModified=true;
  }
  function runIdleEditorClear(){
    if(articleMode!=="new"&&articleMode!=="edit") return;
    if(articleSaving) return;
    if(articleMode==="edit"&&editUserModified) return;
    if(articleMode==="new"&&isNewDirty()) return;
    if(!hasActiveEditorSession(articleMode)) return;
    if(Date.now()-lastEditorActivity<editIdleClearMs) return;
    exitEditor(articleMode,true);
  }
  function editFormSnapshot(){
    return [
      document.getElementById("editSlug").value,
      document.getElementById("editTitle").value,
      document.getElementById("editCat").value,
      editCategoriesLoaded,
      document.getElementById("editTags").value,
      document.getElementById("editAuthor").value,
      document.getElementById("editDate").value,
      document.getElementById("editMd").value
    ].join("\x1e");
  }
  function resetEditAutoSaveBaseline(){
    editSaveBaseline=editFormSnapshot();
  }
  function stopAutoSave(){
    if(autoSaveTimer){ clearInterval(autoSaveTimer); autoSaveTimer=null; }
    if(idleClearTimer){ clearInterval(idleClearTimer); idleClearTimer=null; }
  }
  function beginArticleSave(){
    if(articleSaving) return false;
    articleSaving=true;
    var btn=document.getElementById("btnSave");
    if(btn) btn.disabled=true;
    var btnCreate=document.getElementById("btnCreate");
    if(btnCreate) btnCreate.disabled=true;
    return true;
  }
  function endArticleSave(){
    articleSaving=false;
    var btn=document.getElementById("btnSave");
    if(btn) btn.disabled=false;
    var btnCreate=document.getElementById("btnCreate");
    if(btnCreate) btnCreate.disabled=false;
  }
  function newFormSnapshot(){
    return [
      (document.getElementById("newSlug")||{}).value,
      (document.getElementById("newCat")||{}).value,
      (document.getElementById("newTags")||{}).value,
      (document.getElementById("newAuthor")||{}).value,
      (document.getElementById("newDate")||{}).value,
      (document.getElementById("newMd")||{}).value
    ].join("\x1e");
  }
  function isEditDirty(){
    return editFormSnapshot()!==editSaveBaseline;
  }
  function isNewDirty(){
    return newFormSnapshot()!==newSaveBaseline;
  }
  function hasUnsavedChanges(){
    if(articleMode==="edit"){
      var slug=editSlugSlugs().slug;
      return !!slug&&isEditDirty();
    }
    if(articleMode==="new") return isNewDirty();
    return false;
  }
  function confirmLeaveTab(){
    if(!hasUnsavedChanges()) return true;
    return confirm("当前内容有未保存的修改，确定离开？\n（新建/编辑 Tab 下每分钟会自动保存）");
  }
  function autoSaveTimestamp(){
    var t=new Date();
    return ("0"+t.getHours()).slice(-2)+":"+("0"+t.getMinutes()).slice(-2);
  }
  function runEditAutoSave(){
    if(articleMode!=="edit") return;
    if(editArticleLoading) return;
    if(articleSaving) return;
    var slugs=editSlugSlugs();
    if(!slugs.slug) return;
    var snap=editFormSnapshot();
    if(snap===editSaveBaseline) return;
    var ta=document.getElementById("editMd");
    if(!ta) return;
    articleSaving=true;
    var payload=articlePayload({
      slug:slugs.slug,
      oldSlug:slugs.oldSlug,
      title:document.getElementById("editTitle").value.trim(),
      author:document.getElementById("editAuthor").value.trim(),
      categories:editCategoriesPayload(),
      tags:document.getElementById("editTags").value.trim(),
      publishedAt:document.getElementById("editDate").value.trim(),
      markdown:ta.value,
      autoSave:true
    });
    apiFetchJSON("/api/igoadmin/article",{method:"PUT",body:JSON.stringify(payload)})
      .then(function(x){
        if(x.r.ok){
          var synced=applySavedMarkdown(ta,x.j&&x.j.markdown,payload.markdown);
          if(x.j&&x.j.meta) applyEditMeta(x.j.meta);
          syncEditSlugAfterSave(x.j&&x.j.slug);
          if(synced){
            editSaveBaseline=editFormSnapshot();
            editUserModified=false;
          }
          var hint=x.j&&x.j.validationHint;
          setEditStatus("已自动保存 "+autoSaveTimestamp()+(hint?" · "+hint:""), hint?"warn":"ok");
          return;
        }
        if(x.j&&x.j.hint) setEditStatus("自动保存跳过 · "+x.j.hint,"warn");
        else if(x.j&&x.j.error) setEditStatus("自动保存失败 · "+x.j.error,"warn");
      })
      .catch(function(e){
        if(e&&e.message!=="auth") setEditStatus("自动保存失败","warn");
      })
      .then(function(){ articleSaving=false; });
  }
  function finishNewAutoSave(x,ta,slug,sentMd){
    if(x.r.ok){
      var synced=applySavedMarkdown(ta,x.j&&x.j.markdown,sentMd);
      loadCategories();
      newCreatedSlug=slug;
      if(synced){
        newSaveBaseline=newFormSnapshot();
        newUserModified=false;
      }
      var hint=x.j&&x.j.validationHint;
      setNewStatus("已自动保存 "+autoSaveTimestamp()+(hint?" · "+hint:""), hint?"warn":"ok");
      return;
    }
    if(x.j&&x.j.hint) setNewStatus("自动保存跳过 · "+apiErrText(x.j,x.j.hint),"warn");
    else if(x.j&&x.j.error) setNewStatus("自动保存失败 · "+apiErrText(x.j),"warn");
  }
  function runNewAutoSave(){
    if(articleMode!=="new") return;
    if(articleSaving) return;
    var slug=normalizeSlug((document.getElementById("newSlug")||{}).value);
    if(!slug) return;
    var snap=newFormSnapshot();
    if(snap===newSaveBaseline) return;
    var ta=document.getElementById("newMd");
    if(!ta) return;
    articleSaving=true;
    var payload=attachNewArticleOldSlug(articlePayload({
      slug:slug,
      categories:document.getElementById("newCat").value.trim(),
      tags:document.getElementById("newTags").value.trim(),
      author:document.getElementById("newAuthor").value.trim(),
      publishedAt:document.getElementById("newDate").value.trim(),
      markdown:ta.value,
      autoSave:true
    }),slug);
    var method=newArticleSaveMethod(slug);
    apiFetchJSON("/api/igoadmin/article",{method:method,body:JSON.stringify(payload)})
      .then(function(x){
        if(x.r.ok){ finishNewAutoSave(x,ta,slug,payload.markdown); return; }
        if(method==="POST"&&x.r.status===409){
          return apiFetchJSON("/api/igoadmin/article",{method:"PUT",body:JSON.stringify(payload)})
            .then(function(x2){ finishNewAutoSave(x2,ta,slug,payload.markdown); });
        }
        finishNewAutoSave(x,ta,slug,payload.markdown);
      })
      .catch(function(e){
        if(e&&e.message!=="auth") setNewStatus("自动保存失败","warn");
      })
      .then(function(){ articleSaving=false; });
  }
  function startAutoSave(){
    stopAutoSave();
    touchEditorActivity();
    autoSaveTimer=setInterval(function(){
      if(articleMode==="edit") runEditAutoSave();
      else if(articleMode==="new") runNewAutoSave();
    },60000);
    idleClearTimer=setInterval(runIdleEditorClear,60000);
  }
  function normalizeSlug(raw){
    raw=String(raw||"").trim().toLowerCase();
    if(!raw) return "";
    var out="", prevSep=false;
    for(var i=0;i<raw.length;i++){
      var c=raw.charAt(i);
      var isAlnum=(c>="a"&&c<="z")||(c>="0"&&c<="9");
      if(isAlnum){ out+=c; prevSep=false; }
      else if(c===" "||c==="-"||c==="_"){
        if(out.length&&!prevSep){ out+="_"; prevSep=true; }
      }else return "";
    }
    out=out.replace(/_+$/,"");
    if(!out||out.length>64||!(out.charAt(0)>="a"&&out.charAt(0)<="z")) return "";
    return out;
  }
  function applySlugField(el){
    if(!el) return "";
    var s=normalizeSlug(el.value);
    if(s) el.value=s;
    return s;
  }
  /** 已通过上传/保存建文后一律 PUT；改 slug 时带 oldSlug 走改名，避免再 POST 出第二篇。 */
  function newArticleSaveMethod(slug){
    return newCreatedSlug?"PUT":"POST";
  }
  function attachNewArticleOldSlug(payload,slug){
    if(newCreatedSlug&&newCreatedSlug!==slug) payload.oldSlug=newCreatedSlug;
    return payload;
  }
  function apiErrText(j,fallback){
    var err=(j&&j.error)||"";
    var hint=(j&&j.hint)||"";
    var map={
      "invalid slug":"slug 无效（须以字母开头；仅小写字母/数字/下划线，不能用中文文件名）",
      "invalid old slug":"原 slug 无效",
      "slug already exists":"slug 已存在，请更换或改为更新已有文章",
      "slug exists; use PUT to update":"slug 已存在，请更换 slug 或到「编辑」更新该文",
      "slug reserved":"slug 为系统保留字，请更换",
      "not found":"文章不存在",
      "bad json":"请求格式错误",
      "save failed":"保存失败",
      "check slug":"检查 slug 失败",
      "rename failed":"改名失败",
      "no file":"未选择文件",
      "invalid multipart":"上传表单无效",
      "method not allowed":"请求方法不允许"
    };
    if(err&&map[err]) return map[err];
    if(hint) return hint;
    if(err) return err;
    return fallback||"操作失败";
  }
  function inferArticleTitle(slug,md){
    var lines=(md||"").split("\n");
    for(var i=0;i<lines.length;i++){
      var line=lines[i].trim();
      if(!line) continue;
      if(line.indexOf("# ")===0&&line.indexOf("## ")!==0) return line.slice(2).trim();
      break;
    }
    slug=(slug||"").trim();
    return slug?slug.replace(/_/g," "):"";
  }
  function updateNewTitleHint(){
    var el=document.getElementById("newTitleHint");
    if(!el) return;
    var t=inferArticleTitle(document.getElementById("newSlug").value,document.getElementById("newMd").value);
    el.textContent=t?("标题预览："+t):"标题将随正文首行「# 标题」或 slug 自动生成";
  }
  var newSlugEl=document.getElementById("newSlug");
  var newMdEl=document.getElementById("newMd");
  if(newSlugEl){
    newSlugEl.addEventListener("input",updateNewTitleHint);
    newSlugEl.addEventListener("blur",function(){ applySlugField(newSlugEl); updateNewTitleHint(); });
  }
  if(newMdEl){ newMdEl.addEventListener("input",updateNewTitleHint); }
  function wireEditorField(id,onTouch){
    var el=document.getElementById(id);
    if(!el) return;
    el.addEventListener("input",onTouch);
    el.addEventListener("change",onTouch);
  }
  ["newSlug","newCat","newTags","newAuthor","newDate","newMd"].forEach(function(id){
    wireEditorField(id,function(){
      newUserModified=true;
      touchEditorActivity();
    });
  });
  ["editSlug","editTitle","editCat","editTags","editAuthor","editDate","editMd"].forEach(function(id){
    wireEditorField(id,function(){
      editUserModified=true;
      touchEditorActivity();
    });
  });
  var editSlugEl=document.getElementById("editSlug");
  if(editSlugEl){
    editSlugEl.addEventListener("blur",function(){ applySlugField(editSlugEl); });
  }
  wireEditorField("editSearch",touchEditorActivity);
  var catList=[], editSearchEl=document.getElementById("editSearch");
  var editSearchResults=document.getElementById("editSearchResults");
  var editSearchTimer=null, editArticleLoading=false;
  function catSelects(){
    return [
      document.getElementById("newCat"),
      document.getElementById("editCat"),
      document.getElementById("batchCat")
    ];
  }
  function fillBatchFilterCat(cats){
    var sel=document.getElementById("batchFilterCat");
    if(!sel) return;
    var keep=sel.value||"";
    sel.innerHTML='<option value="">全部分类</option>';
    (cats||[]).forEach(function(c){
      var o=document.createElement("option");
      o.value=c; o.textContent=c;
      sel.appendChild(o);
    });
    if(keep) sel.value=keep;
  }
  function fillBatchCatSelect(cats,value){
    var sel=document.getElementById("batchCat");
    if(!sel) return;
    var keep=value!=null?value:catKeepValue(sel);
    sel.innerHTML='<option value="">— 清空分类 —</option>';
    (cats||[]).forEach(function(c){
      var o=document.createElement("option");
      o.value=c; o.textContent=c;
      sel.appendChild(o);
    });
    if(keep){
      if(cats&&cats.indexOf(keep)<0){
        var o=document.createElement("option");
        o.value=keep; o.textContent=keep;
        sel.appendChild(o);
      }
      sel.value=keep;
    }
  }
  function catKeepValue(sel){
    if(!sel) return "";
    return sel.dataset.articleCat||sel.value||"";
  }
  function fillCatSelect(sel,cats,value){
    if(!sel) return;
    var keep=value||catKeepValue(sel)||"";
    sel.innerHTML='<option value="">— 选择分类 —</option>';
    (cats||[]).forEach(function(c){
      var o=document.createElement("option");
      o.value=c; o.textContent=c;
      sel.appendChild(o);
    });
    if(keep){
      if(!cats||cats.indexOf(keep)<0){
        var o=document.createElement("option");
        o.value=keep; o.textContent=keep;
        sel.appendChild(o);
      }
      sel.value=keep;
    }
  }
  function refreshCatSelects(cats){
    catList=cats||[];
    catSelects().forEach(function(sel){
      if(sel&&sel.id==="batchCat") fillBatchCatSelect(catList,catKeepValue(sel));
      else fillCatSelect(sel,catList,catKeepValue(sel));
    });
    fillBatchFilterCat(catList);
  }
  function loadCategories(){
    return apiFetchJSON("/api/igoadmin/categories").then(function(x){
      if(x.r.ok&&Array.isArray(x.j)) refreshCatSelects(x.j);
    }).catch(function(){});
  }
  function primaryCategory(raw){
    var p=(raw||"").split(",");
    for(var i=0;i<p.length;i++){ var s=p[i].trim(); if(s) return s; }
    return "";
  }
  var editCategoriesLoaded="";
  function setEditCategoriesLoaded(raw){
    editCategoriesLoaded=String(raw||"").trim();
  }
  function editCategoriesPayload(){
    if(!editCatEl) return "";
    var selected=editCatEl.value.trim();
    if(!editCategoriesLoaded) return selected;
    return selected===primaryCategory(editCategoriesLoaded)?editCategoriesLoaded:selected;
  }
  function setCatValue(sel,raw){
    if(!sel) return;
    var cat=primaryCategory(raw);
    sel.dataset.articleCat=cat;
    fillCatSelect(sel,catList,cat);
  }
  function promptCategoryName(title,initial){
    var n=prompt(title,initial||"");
    if(n==null) return "";
    return n.trim();
  }
  function addCategoryToSelects(name){
    name=(name||"").trim();
    if(!name||name.indexOf(",")>=0){ modeMsg("分类名无效",true,false); return; }
    if(catList.indexOf(name)<0){
      catList=catList.slice().concat([name]).sort();
      refreshCatSelects(catList);
    }
    catSelects().forEach(function(sel){ if(sel) sel.value=name; });
    modeMsg("已添加分类「"+name+"」",false,false);
  }
  function runCategoryRename(sel){
    var oldName=(sel&&sel.value||"").trim();
    if(!oldName){ modeMsg("请先选择要改名的分类",true,false); return; }
    var newName=promptCategoryName("将「"+oldName+"」改名为：",oldName);
    if(!newName||newName===oldName) return;
    apiFetchJSON("/api/igoadmin/category",{method:"PUT",body:JSON.stringify({old:oldName,new:newName})})
      .then(function(x){
        if(x.r.ok){ loadCategories().then(function(){ if(sel) sel.value=newName; }); modeMsg("分类已改名",false,false); }
        else modeMsg((x.j&&x.j.error)||"改名失败",true,false);
      }).catch(function(){});
  }
  function runCategoryDelete(sel){
    var name=(sel&&sel.value||"").trim();
    if(!name){ modeMsg("请先选择要删除的分类",true,false); return; }
    if(!confirm("从全部文章中移除分类「"+name+"」？\n文章本身不会删除。")) return;
    apiFetchJSON("/api/igoadmin/category?name="+encodeURIComponent(name),{method:"DELETE"})
      .then(function(x){
        if(x.r.ok){ loadCategories(); modeMsg("分类已删除",false,false); }
        else modeMsg((x.j&&x.j.error)||"删除失败",true,false);
      }).catch(function(){});
  }
  function wireCatButtons(prefix,selId){
    document.getElementById(prefix+"CatAdd").onclick=function(){
      var name=promptCategoryName("新分类名称：");
      if(name) addCategoryToSelects(name);
    };
    document.getElementById(prefix+"CatRename").onclick=function(){
      runCategoryRename(document.getElementById(selId));
    };
    document.getElementById(prefix+"CatDel").onclick=function(){
      runCategoryDelete(document.getElementById(selId));
    };
  }
  wireCatButtons("btnNew","newCat");
  wireCatButtons("btnEdit","editCat");
  var btnBatchCatAdd=document.getElementById("btnBatchCatAdd");
  if(btnBatchCatAdd){
    btnBatchCatAdd.onclick=function(){
      var name=promptCategoryName("新分类名称：");
      if(name) addCategoryToSelects(name);
    };
  }
  var editCatEl=document.getElementById("editCat");
  if(editCatEl){
    editCatEl.addEventListener("change",function(){ editCatEl.dataset.articleCat=editCatEl.value; });
  }
  function closeEditSearch(){
    if(editSearchResults) editSearchResults.classList.remove("is-open");
  }
  function openEditSearch(){
    if(editSearchResults) editSearchResults.classList.add("is-open");
  }
  var tabNew=document.getElementById("tabNew");
  var tabEdit=document.getElementById("tabEdit");
  var tabBatch=document.getElementById("tabBatch");
  var panelNew=document.getElementById("panelNew");
  var panelEdit=document.getElementById("panelEdit");
  var panelBatch=document.getElementById("panelBatch");
  var articleMode="new";
  function showModePanel(panel,on){
    if(!panel) return;
    panel.classList.toggle("hidden",!on);
    if(on) panel.removeAttribute("hidden"); else panel.setAttribute("hidden","");
  }
  function setArticleMode(mode){
    if(mode!=="new"&&mode!=="edit"&&mode!=="batch") mode="new";
    articleMode=mode;
    var isNew=mode==="new", isEdit=mode==="edit", isBatch=mode==="batch";
    if(tabNew){
      tabNew.classList.toggle("is-active",isNew);
      tabNew.setAttribute("aria-selected",isNew?"true":"false");
    }
    if(tabEdit){
      tabEdit.classList.toggle("is-active",isEdit);
      tabEdit.setAttribute("aria-selected",isEdit?"true":"false");
    }
    if(tabBatch){
      tabBatch.classList.toggle("is-active",isBatch);
      tabBatch.setAttribute("aria-selected",isBatch?"true":"false");
    }
    showModePanel(panelNew,isNew);
    showModePanel(panelEdit,isEdit);
    showModePanel(panelBatch,isBatch);
    if(isNew) closeEditSearch();
    if(isBatch) loadCategories();
    if(isNew||isEdit) startAutoSave();
    else stopAutoSave();
  }
  function trySetArticleMode(mode){
    if(mode===articleMode) return;
    if(!confirmLeaveTab()) return;
    setArticleMode(mode);
    if(mode==="new"){ newSaveBaseline=newFormSnapshot(); newCreatedSlug=""; }
    if(mode==="edit"&&document.getElementById("editSlug").value.trim()) resetEditAutoSaveBaseline();
  }
  if(tabNew) tabNew.onclick=function(){ trySetArticleMode("new"); };
  if(tabEdit) tabEdit.onclick=function(){ trySetArticleMode("edit"); };
  if(tabBatch) tabBatch.onclick=function(){ trySetArticleMode("batch"); };
  function activeMdTa(){
    if(articleMode==="edit") return document.getElementById("editMd");
    return document.getElementById("newMd");
  }
  var batchItems=[], batchTotal=0, batchOffset=0, batchLimit=50;
  var batchSelected={}, batchFilterCat="", batchSearchQ="";
  var batchChkAllPage=document.getElementById("batchChkAllPage");
  var batchArticleList=document.getElementById("batchArticleList");
  var batchListSummary=document.getElementById("batchListSummary");
  var batchSelCount=document.getElementById("batchSelCount");
  var batchPageInfo=document.getElementById("batchPageInfo");
  var btnBatchPrev=document.getElementById("btnBatchPrev");
  var btnBatchNext=document.getElementById("btnBatchNext");
  var batchChkCat=document.getElementById("batchChkCat");
  var batchCatEl=document.getElementById("batchCat");
  var batchChkTags=document.getElementById("batchChkTags");
  var batchTagsModeEl=document.getElementById("batchTagsMode");
  var batchTagsEl=document.getElementById("batchTags");
  function batchSelectedSlugs(){
    return Object.keys(batchSelected).filter(function(s){ return batchSelected[s]; });
  }
  function updateBatchSelCount(){
    var n=batchSelectedSlugs().length;
    if(batchSelCount) batchSelCount.textContent="已选 "+n+" 篇";
  }
  function syncBatchChkAllPage(){
    if(!batchChkAllPage||!batchItems.length){
      if(batchChkAllPage) batchChkAllPage.checked=false;
      return;
    }
    var all=true;
    batchItems.forEach(function(it){
      if(!batchSelected[it.slug]) all=false;
    });
    batchChkAllPage.checked=all;
  }
  function renderBatchList(){
    if(!batchArticleList) return;
    batchArticleList.innerHTML="";
    if(!batchItems.length){
      var li=document.createElement("li");
      li.textContent="无文章";
      li.style.padding=".5rem .55rem";
      li.style.color="var(--muted)";
      batchArticleList.appendChild(li);
      syncBatchChkAllPage();
      return;
    }
    batchItems.forEach(function(it){
      var li=document.createElement("li");
      var label=document.createElement("label");
      var chk=document.createElement("input");
      chk.type="checkbox";
      chk.checked=!!batchSelected[it.slug];
      chk.addEventListener("change",function(){
        if(chk.checked) batchSelected[it.slug]=true;
        else delete batchSelected[it.slug];
        updateBatchSelCount();
        syncBatchChkAllPage();
      });
      var main=document.createElement("span");
      main.className="batch-item-main";
      var t=document.createElement("span");
      t.className="batch-item-title";
      t.textContent=it.title||it.slug;
      var m=document.createElement("span");
      m.className="batch-item-meta";
      m.textContent=it.slug+(it.categories?" · "+it.categories:"")+(it.tags?" · "+it.tags:"");
      main.appendChild(t);
      main.appendChild(m);
      label.appendChild(chk);
      label.appendChild(main);
      li.appendChild(label);
      batchArticleList.appendChild(li);
    });
    syncBatchChkAllPage();
    updateBatchSelCount();
  }
  function updateBatchPageInfo(){
    if(batchListSummary){
      batchListSummary.textContent=batchTotal
        ?("共 "+batchTotal+" 篇匹配"+(batchFilterCat?" · 分类「"+batchFilterCat+"」":"")+(batchSearchQ?" · 关键词「"+batchSearchQ+"」":""))
        :"未找到匹配文章";
    }
    if(batchPageInfo){
      var pages=batchTotal?Math.ceil(batchTotal/batchLimit):0;
      var cur=pages?Math.floor(batchOffset/batchLimit)+1:0;
      batchPageInfo.textContent=pages?("第 "+cur+"/"+pages+" 页"):("");
    }
    if(btnBatchPrev) btnBatchPrev.disabled=batchOffset<=0;
    if(btnBatchNext) btnBatchNext.disabled=batchOffset+batchLimit>=batchTotal;
  }
  function syncBatchFilterFromDOM(){
    var catEl=document.getElementById("batchFilterCat");
    var qEl=document.getElementById("batchSearchQ");
    batchFilterCat=catEl?catEl.value.trim():"";
    batchSearchQ=qEl?qEl.value.trim():"";
  }
  function loadBatchList(offset){
    if(offset==null) offset=0;
    batchOffset=offset;
    syncBatchFilterFromDOM();
    msg(batchMsg,"加载中…",false);
    var url="/api/igoadmin/articles?limit="+batchLimit+"&offset="+batchOffset;
    if(batchFilterCat) url+="&category="+encodeURIComponent(batchFilterCat);
    if(batchSearchQ) url+="&q="+encodeURIComponent(batchSearchQ);
    return apiFetchJSON(url).then(function(x){
      if(!x.r.ok){
        msg(batchMsg,(x.j&&x.j.error)||("加载失败 (HTTP "+x.r.status+")"),true);
        return;
      }
      var j=x.j||{};
      batchItems=Array.isArray(j.items)?j.items:[];
      batchTotal=typeof j.total==="number"?j.total:batchItems.length;
      renderBatchList();
      updateBatchPageInfo();
      var hint="已加载 "+batchItems.length+" / "+batchTotal;
      if(j.truncated) hint+="（关键词命中过多，仅展示前 5000 条匹配）";
      msg(batchMsg,hint,false);
    }).catch(function(e){
      if(e&&e.message!=="auth") msg(batchMsg,"加载失败",true);
    });
  }
  function batchListURL(off,pageLimit){
    syncBatchFilterFromDOM();
    var lim=pageLimit||200;
    var url="/api/igoadmin/articles?limit="+lim+"&offset="+(off|0);
    if(batchFilterCat) url+="&category="+encodeURIComponent(batchFilterCat);
    if(batchSearchQ) url+="&q="+encodeURIComponent(batchSearchQ);
    return url;
  }
  function batchSelectAllResults(){
    syncBatchFilterFromDOM();
    msg(batchMsg,"正在全选…",false);
    function fetchPage(off,acc){
      return apiFetchJSON(batchListURL(off)).then(function(x){
        if(!x.r.ok) throw new Error("load");
        var j=x.j||{};
        var items=Array.isArray(j.items)?j.items:[];
        var total=typeof j.total==="number"?j.total:0;
        items.forEach(function(it){ batchSelected[it.slug]=true; });
        acc+=items.length;
        if(acc<total&&items.length===200){
          return fetchPage(off+200,acc);
        }
        return total;
      });
    }
    fetchPage(0,0).then(function(total){
      renderBatchList();
      updateBatchSelCount();
      msg(batchMsg,"已全选 "+batchSelectedSlugs().length+" 篇（筛选共 "+total+" 篇）",false);
    }).catch(function(){
      msg(batchMsg,"全选失败",true);
    });
  }
  if(document.getElementById("btnBatchLoad")){
    document.getElementById("btnBatchLoad").onclick=function(){
      batchSelected={};
      loadBatchList(0);
    };
  }
  var batchSearchQEl=document.getElementById("batchSearchQ");
  if(batchSearchQEl){
    batchSearchQEl.addEventListener("keydown",function(e){
      if(e.key==="Enter"){ e.preventDefault(); batchSelected={}; loadBatchList(0); }
    });
  }
  if(btnBatchPrev) btnBatchPrev.onclick=function(){ if(batchOffset>0) loadBatchList(Math.max(0,batchOffset-batchLimit)); };
  if(btnBatchNext) btnBatchNext.onclick=function(){
    if(batchOffset+batchLimit<batchTotal) loadBatchList(batchOffset+batchLimit);
  };
  if(batchChkAllPage){
    batchChkAllPage.onchange=function(){
      batchItems.forEach(function(it){
        if(batchChkAllPage.checked) batchSelected[it.slug]=true;
        else delete batchSelected[it.slug];
      });
      renderBatchList();
    };
  }
  if(document.getElementById("btnBatchSelectAll")){
    document.getElementById("btnBatchSelectAll").onclick=batchSelectAllResults;
  }
  if(document.getElementById("btnBatchSelectNone")){
    document.getElementById("btnBatchSelectNone").onclick=function(){
      batchSelected={};
      renderBatchList();
      msg(batchMsg,"已清空选择",false);
    };
  }
  if(batchChkCat&&batchCatEl){
    batchChkCat.onchange=function(){ batchCatEl.disabled=!batchChkCat.checked; };
  }
  if(batchChkTags&&batchTagsModeEl&&batchTagsEl){
    batchChkTags.onchange=function(){
      var on=batchChkTags.checked;
      batchTagsModeEl.disabled=!on;
      batchTagsEl.disabled=!on;
    };
  }
  function buildBatchPayload(dryRun){
    var slugs=batchSelectedSlugs();
    if(!slugs.length){ msg(batchMsg,"请先勾选文章",true); return null; }
    var updCat=batchChkCat&&batchChkCat.checked;
    var updTags=batchChkTags&&batchChkTags.checked;
    if(!updCat&&!updTags){ msg(batchMsg,"请勾选要修改的分类或标签",true); return null; }
    if(updTags){
      var tagMode=batchTagsModeEl?batchTagsModeEl.value:"replace";
      var tagVal=batchTagsEl?batchTagsEl.value.trim():"";
      if(tagMode!=="replace"&&!tagVal){
        msg(batchMsg,"追加/移除标签时请填写标签内容",true); return null;
      }
    }
    return {
      slugs:slugs,
      updateCategories:updCat,
      setCategories:batchCatEl?batchCatEl.value.trim():"",
      updateTags:updTags,
      setTags:batchTagsEl?batchTagsEl.value.trim():"",
      tagsMode:batchTagsModeEl?batchTagsModeEl.value:"replace",
      dryRun:!!dryRun
    };
  }
  if(document.getElementById("btnBatchPreview")){
    document.getElementById("btnBatchPreview").onclick=function(){
      var payload=buildBatchPayload(true);
      if(!payload) return;
      msg(batchMsg,"预览中…",false);
      apiFetchJSON("/api/igoadmin/articles/batch",{method:"PATCH",body:JSON.stringify(payload)})
        .then(function(x){
          if(!x.r.ok){ msg(batchMsg,(x.j&&x.j.error)||"预览失败",true); return; }
          var items=(x.j&&x.j.items)||[];
          var skipped=(x.j&&x.j.skipped)||[];
          if(!items.length&&!skipped.length){ msg(batchMsg,"无预览数据",true); return; }
          var lines=items.slice(0,8).map(function(p){
            var s=p.slug+": ";
            if(payload.updateCategories) s+="分类「"+(p.categoriesBefore||"—")+"」→「"+(p.categoriesAfter||"—")+"」 ";
            if(payload.updateTags) s+="标签「"+(p.tagsBefore||"—")+"」→「"+(p.tagsAfter||"—")+"」";
            return s;
          });
          var more=items.length>8?("\n… 另有 "+(items.length-8)+" 篇"):("");
          var skipHint=skipped.length?("\n\n无法预览 "+skipped.length+" 篇（可能已删除）"):"";
          if(payload.slugs.length>items.length+skipped.length){
            skipHint+="\n另有 "+(payload.slugs.length-items.length-skipped.length)+" 篇未返回";
          }
          alert("预览 "+items.length+" 篇：\n\n"+lines.join("\n")+more+skipHint);
          msg(batchMsg,"预览完成",false);
        }).catch(function(e){
          if(e&&e.message!=="auth") msg(batchMsg,"预览失败",true);
        });
    };
  }
  if(document.getElementById("btnBatchApply")){
    document.getElementById("btnBatchApply").onclick=function(){
      var payload=buildBatchPayload(false);
      if(!payload) return;
      var n=payload.slugs.length;
      var hint="将对 "+n+" 篇文章";
      if(payload.updateCategories) hint+="设置分类为「"+(payload.setCategories||"（清空）")+"」";
      if(payload.updateTags){
        var modeLabel=payload.tagsMode==="append"?"追加":(payload.tagsMode==="remove"?"移除":"覆盖");
        hint+=(payload.updateCategories?"，":"")+modeLabel+"标签「"+payload.setTags+"」";
      }
      if(!confirm(hint+"？\n此操作不改正文。")) return;
      msg(batchMsg,"更新中…",false);
      apiFetchJSON("/api/igoadmin/articles/batch",{method:"PATCH",body:JSON.stringify(payload)})
        .then(function(x){
          if(!x.r.ok){ msg(batchMsg,(x.j&&x.j.error)||"更新失败",true); return; }
          var ok=(x.j&&x.j.ok)||0;
          var failed=(x.j&&x.j.failed)||[];
          loadCategories();
          loadBatchList(batchOffset);
          if(failed.length){
            var failSlugs=failed.slice(0,3).map(function(f){ return f.slug; }).join(", ");
            var extra=failed.length>3?" 等":"";
            msg(batchMsg,"成功 "+ok+" 篇，失败 "+failed.length+" 篇（"+failSlugs+extra+"）",true);
          }else{
            msg(batchMsg,"已更新 "+ok+" 篇",false);
          }
        }).catch(function(e){
          if(e&&e.message!=="auth") msg(batchMsg,"更新失败",true);
        });
    };
  }
  function formatPublished(p){
    if(typeof p==="string") return p.slice(0,19).replace("T"," ");
    if(typeof p==="number"&&p>0) return fmtBJ(p);
    if(p) try{ return new Date(p).toISOString().slice(0,19).replace("T"," "); }catch(e){}
    return "";
  }
  function applyEditMeta(meta){
    if(!meta) return;
    if(meta.title!=null) document.getElementById("editTitle").value=meta.title||"";
    if(meta.tags!=null) document.getElementById("editTags").value=meta.tags||"";
    if(meta.author!=null) document.getElementById("editAuthor").value=meta.author||"igozhang";
    if(meta.publishedAt!=null) document.getElementById("editDate").value=formatPublished(meta.publishedAt);
    if(meta.categories!=null){
      setEditCategoriesLoaded(meta.categories||"");
      if(editCatEl) setCatValue(editCatEl,meta.categories||"");
    }
  }
  function loadEditArticle(slug){
    if(!slug) return;
    editArticleLoading=true;
    clearDraftCache();
    msg(editMsg,"加载中…",false);
    apiFetchJSON("/api/igoadmin/article?slug="+encodeURIComponent(slug))
      .then(function(x){
        var j=x.j;
        if(!j||!j.meta){
          document.getElementById("editSlug").value="";
          document.getElementById("editSlugOriginal").value="";
          msg(editMsg,"未找到",true);
          return null;
        }
        var loadedSlug=j.meta.slug||slug;
        document.getElementById("editSlug").value=loadedSlug;
        document.getElementById("editSlugOriginal").value=loadedSlug;
        document.getElementById("editTags").value=j.meta.tags||"";
        document.getElementById("editAuthor").value=j.meta.author||"igozhang";
        document.getElementById("editMd").value=j.markdown||"";
        document.getElementById("editTitle").value=j.meta.title||"";
        document.getElementById("editDate").value=formatPublished(j.meta.publishedAt);
        setEditCategoriesLoaded(j.meta.categories||"");
        if(editCatEl) setCatValue(editCatEl,j.meta.categories||"");
        if(editSearchEl) editSearchEl.value=(j.meta.title||slug)+"  ·  "+slug;
        closeEditSearch();
        editUserModified=false;
        resetEditAutoSaveBaseline();
        touchEditorActivity();
        return j.meta.categories||"";
      })
      .then(function(categories){
        if(categories===null){ editArticleLoading=false; return; }
        return loadCategories().then(function(){
          if(editCatEl) setCatValue(editCatEl,categories);
          resetEditAutoSaveBaseline();
          editUserModified=false;
          touchEditorActivity();
          startAutoSave();
          msg(editMsg,"已加载",false);
        }).finally(function(){ editArticleLoading=false; });
      })
      .catch(function(e){
        editArticleLoading=false;
        document.getElementById("editSlug").value="";
        document.getElementById("editSlugOriginal").value="";
        if(e&&e.message!=="auth") msg(editMsg,"加载失败",true);
      });
  }
  function renderEditSearch(hits){
    if(!editSearchResults) return;
    editSearchResults.innerHTML="";
    if(!hits||!hits.length){
      var li=document.createElement("li");
      var btn=document.createElement("button");
      btn.type="button"; btn.textContent="无匹配";
      btn.disabled=true;
      li.appendChild(btn);
      editSearchResults.appendChild(li);
      openEditSearch();
      return;
    }
    hits.forEach(function(it){
      var li=document.createElement("li");
      var btn=document.createElement("button");
      btn.type="button";
      btn.innerHTML="<strong>"+escHtml(it.title||it.slug)+"</strong><span>"+escHtml(it.slug+"  ·  "+(it.categories||"")+"  "+(it.tags||""))+"</span>";
      btn.addEventListener("click",function(){ loadEditArticle(it.slug); });
      li.appendChild(btn);
      editSearchResults.appendChild(li);
    });
    openEditSearch();
  }
  function escHtml(s){
    return String(s).replace(/&/g,"&amp;").replace(/</g,"&lt;").replace(/>/g,"&gt;");
  }
  if(editSearchEl){
    editSearchEl.addEventListener("input",function(){
      clearTimeout(editSearchTimer);
      var q=editSearchEl.value.trim();
      if(q.length<1){ closeEditSearch(); return; }
      editSearchTimer=setTimeout(function(){
        apiFetchJSON("/api/igoadmin/articles?q="+encodeURIComponent(q)+"&limit=50")
          .then(function(x){
            var hits=(x.r.ok&&x.j&&Array.isArray(x.j.items))?x.j.items:[];
            renderEditSearch(hits);
          })
          .catch(function(){ closeEditSearch(); });
      }, 200);
    });
    editSearchEl.addEventListener("focus",function(){
      if(editSearchEl.value.trim().length>0) editSearchEl.dispatchEvent(new Event("input"));
    });
  }
  document.addEventListener("click",function(e){
    if(!editSearchEl||!editSearchResults) return;
    if(editSearchEl.contains(e.target)||editSearchResults.contains(e.target)) return;
    closeEditSearch();
  });
  function finishNewManualSave(x,ta,slug,method,sentMd){
    if(x.r.ok){
      var synced=applySavedMarkdown(ta,x.j&&x.j.markdown,sentMd);
      loadCategories();
      newCreatedSlug=slug;
      if(synced){
        newSaveBaseline=newFormSnapshot();
        newUserModified=false;
      }
      var hint=x.j&&x.j.validationHint;
      var label=method==="PUT"?"已更新":"已创建";
      setNewStatus(formatSaveStatus(label+" "+x.j.slug+(x.j.title?" · "+x.j.title:""),x.j), hint?"warn":"ok");
      return;
    }
    setNewStatus(apiErrText(x.j,"失败 (HTTP "+x.r.status+")"),"err");
  }
  document.getElementById("btnCreate").onclick=function(){
    if(articleSaving){ modeMsg("保存进行中，请稍候",false,true); return; }
    if(!beginArticleSave()) return;
    var ta=document.getElementById("newMd");
    var slug=applySlugField(document.getElementById("newSlug"));
    if(!slug){
      endArticleSave();
      msg(newMsg,"slug 无效（须以字母开头；大写会自动转小写）",true);
      return;
    }
    var payload=attachNewArticleOldSlug(articlePayload({
      slug:slug,
      categories:document.getElementById("newCat").value.trim(),
      tags:document.getElementById("newTags").value.trim(),
      author:document.getElementById("newAuthor").value.trim(),
      publishedAt:document.getElementById("newDate").value.trim(),
      markdown:ta.value
    }),slug);
    var method=newArticleSaveMethod(slug);
    msg(newMsg,"保存中…",false);
    apiFetchJSON("/api/igoadmin/article",{method:method,body:JSON.stringify(payload)})
      .then(function(x){
        if(x.r.ok){ finishNewManualSave(x,ta,slug,method,payload.markdown); return; }
        if(method==="POST"&&x.r.status===409){
          return apiFetchJSON("/api/igoadmin/article",{method:"PUT",body:JSON.stringify(payload)})
            .then(function(x2){ finishNewManualSave(x2,ta,slug,"PUT",payload.markdown); });
        }
        finishNewManualSave(x,ta,slug,method,payload.markdown);
      }).catch(function(e){
        if(e&&e.message==="auth") msg(newMsg,"登录已失效，请重新验证",true);
        else msg(newMsg,"保存失败："+(e&&e.message||"网络错误"),true);
      })
      .then(function(){ endArticleSave(); });
  };
  document.getElementById("btnSave").onclick=function(){
    var btnSave=document.getElementById("btnSave");
    var btnLabel=btnSave?btnSave.textContent:"保存更新";
    function resetBtn(){ endArticleSave(); if(btnSave){ btnSave.textContent=btnLabel; } }
    try{
      if(articleSaving){ setEditStatus("保存进行中，请稍候","warn"); return; }
      if(!beginArticleSave()) return;
      var slugs=editSlugSlugs();
      if(!slugs.slug){ resetBtn(); msg(editMsg,"请先搜索并选择文章，或填写有效 slug",true); return; }
      var ta=document.getElementById("editMd");
      if(!ta){ msg(editMsg,"编辑区未就绪",true); return; }
      var payload=articlePayload({
        slug:slugs.slug,
        oldSlug:slugs.oldSlug,
        title:document.getElementById("editTitle").value.trim(),
        author:document.getElementById("editAuthor").value.trim(),
        categories:editCategoriesPayload(),
        tags:document.getElementById("editTags").value.trim(),
        publishedAt:document.getElementById("editDate").value.trim(),
        markdown:ta.value
      });
      msg(editMsg,"保存中…",false);
      if(btnSave){ btnSave.textContent="保存中…"; btnSave.disabled=true; }
      apiFetchJSON("/api/igoadmin/article",{method:"PUT",body:JSON.stringify(payload)})
        .then(function(x){
          if(x.r.ok){
            var synced=applySavedMarkdown(ta,x.j&&x.j.markdown,payload.markdown);
            if(x.j&&x.j.meta) applyEditMeta(x.j.meta);
            syncEditSlugAfterSave(x.j&&x.j.slug);
            if(synced){
              resetEditAutoSaveBaseline();
              editUserModified=false;
            }
            var hint=x.j&&x.j.validationHint;
            setEditStatus(formatSaveStatus("已保存",x.j), hint?"warn":"ok");
            if(x.j&&x.j.slug&&editSearchEl) editSearchEl.value=(x.j.meta&&x.j.meta.title||x.j.slug)+"  ·  "+x.j.slug;
            loadCategories().then(function(){
              if(x.j&&x.j.meta&&editCatEl) setCatValue(editCatEl,x.j.meta.categories||"");
            });
            return;
          }
          var eh=x.j&&x.j.hint;
          setEditStatus(apiErrText(x.j,eh||("保存失败 (HTTP "+x.r.status+")")),"err");
        })
        .catch(function(e){
          if(e&&e.message==="auth") msg(editMsg,"登录已失效，请重新验证",true);
          else msg(editMsg,"保存失败："+(e&&e.message||"网络错误"),true);
        })
        .then(function(){ resetBtn(); }, function(){ resetBtn(); });
    }catch(err){
      msg(editMsg,"保存出错",true);
      resetBtn();
    }
  };
  document.getElementById("btnDelete").onclick=function(){
    var slug=(document.getElementById("editSlugOriginal")||{}).value.trim()||(document.getElementById("editSlug")||{}).value.trim();
    if(!slug){ msg(editMsg,"请先搜索并选择文章",true); return; }
    if(!confirm("确定删除文章「"+slug+"」？\n将删除正文及未被其他文章引用的 doc-pic 图片。")) return;
    apiFetchJSON("/api/igoadmin/article?slug="+encodeURIComponent(slug),{method:"DELETE"})
      .then(function(x){
        if(x.r.ok){
          editSaveBaseline="";
          clearDraftCache();
          setEditCategoriesLoaded("");
          document.getElementById("editSlug").value="";
          document.getElementById("editSlugOriginal").value="";
          document.getElementById("editSearch").value="";
          document.getElementById("editMd").value="";
          document.getElementById("editTitle").value="";
          document.getElementById("editAuthor").value="";
          document.getElementById("editCat").value="";
          document.getElementById("editTags").value="";
          document.getElementById("editDate").value="";
          msg(editMsg,"已删除 "+slug,false);
        }else msg(editMsg,(x.j&&x.j.error)||"删除失败",true);
      }).catch(function(){});
  };
  var backupMsgEl=document.getElementById("backupMsg");
  var backupProgressEl=document.getElementById("backupProgress");
  var backupProgressBarEl=document.getElementById("backupProgressBar");
  var btnExportEl=document.getElementById("btnExport");
  var btnImportPickEl=document.getElementById("btnImportPick");
  function backupSetProgress(pct){
    if(!backupProgressEl||!backupProgressBarEl) return;
    backupProgressEl.classList.remove("hidden");
    backupProgressEl.setAttribute("aria-hidden","false");
    if(pct<0){
      backupProgressEl.classList.add("is-indeterminate");
      backupProgressBarEl.style.width="";
      return;
    }
    backupProgressEl.classList.remove("is-indeterminate");
    backupProgressBarEl.style.width=Math.max(0,Math.min(100,pct))+"%";
  }
  function backupHideProgress(){
    if(!backupProgressEl||!backupProgressBarEl) return;
    backupProgressEl.classList.add("hidden");
    backupProgressEl.setAttribute("aria-hidden","true");
    backupProgressEl.classList.remove("is-indeterminate");
    backupProgressBarEl.style.width="0";
  }
  function backupSetBusy(busy){
    if(btnExportEl) btnExportEl.disabled=busy;
    if(btnImportPickEl) btnImportPickEl.disabled=busy;
  }
  function backupStart(text){
    backupSetBusy(true);
    msg(backupMsgEl,text,false);
    backupSetProgress(0);
  }
  function backupDone(text,err){
    backupSetProgress(100);
    msg(backupMsgEl,text,!!err);
    setTimeout(backupHideProgress,700);
    backupSetBusy(false);
  }
  function backupFail(text){
    backupHideProgress();
    msg(backupMsgEl,text,true);
    backupSetBusy(false);
  }
  document.getElementById("btnExport").onclick=function(){
    backupStart("正在导出…");
    apiXhrDownload("/api/igoadmin/export",backupSetProgress).then(function(res){
      var a=document.createElement("a");
      var name="doc-export.zip";
      var cd=res.header("Content-Disposition")||"";
      var m=/filename="([^"]+)"/.exec(cd);
      if(m) name=m[1];
      a.href=URL.createObjectURL(res.blob);
      a.download=name;
      a.click();
      setTimeout(function(){ try{ URL.revokeObjectURL(a.href); }catch(e){} },4000);
      backupDone("导出已开始下载",false);
    }).catch(function(e){
      if(e&&e.message==="auth") backupFail("登录已失效，请重新验证");
      else backupFail("导出失败："+(e&&e.message||"网络错误"));
    });
  };
  document.getElementById("btnImportPick").onclick=function(){document.getElementById("importZip").click();};
  document.getElementById("importZip").onchange=function(){
    var f=this.files[0]; if(!f) return;
    var fd=new FormData(); fd.append("file",f);
    backupStart("正在导入…");
    apiXhrJSON("POST","/api/igoadmin/import",fd,backupSetProgress).then(function(x){
      var ok=x.r.ok&&x.j&&x.j.ok;
      if(ok){
        backupDone("导入完成，共 "+(x.j.count||0)+" 篇",false);
        return;
      }
      backupFail("导入失败："+((x.j&&x.j.error)||("HTTP "+x.r.status)));
    }).catch(function(e){
      if(e&&e.message==="auth") backupFail("登录已失效，请重新验证");
      else backupFail("导入失败："+(e&&e.message||"网络错误"));
    });
    this.value="";
  };
  document.getElementById("btnUploadPick").onclick=function(){document.getElementById("uploadMd").click();};
  document.getElementById("uploadMd").onchange=function(){
    var f=this.files[0]; if(!f) return;
    if(articleSaving){ modeMsg("保存进行中，请稍候",false,true); this.value=""; return; }
    var slugFromInput=document.getElementById("newSlug").value.trim();
    var reader=new FileReader();
    reader.onload=function(){
      var text=reader.result||"";
      if(typeof text!=="string") text="";
      if(!slugFromInput){
        var base=f.name.replace(/\.md$/i,"");
        var norm=normalizeSlug(base);
        if(norm) document.getElementById("newSlug").value=norm;
      }else applySlugField(document.getElementById("newSlug"));
      document.getElementById("newMd").value=text;
      newUserModified=true;
      updateNewTitleHint();
      var uploadSlug=applySlugField(document.getElementById("newSlug"));
      if(!uploadSlug){
        // 仅载入正文，不调 upload-md；请用户补合法 slug 后再保存新建
        setNewStatus("已载入正文 · 当前没有合法 slug（文件名须能规范化，或手动填写：字母开头，可含数字/下划线），请填写后再保存新建","warn");
        return;
      }
      if(!beginArticleSave()){
        setNewStatus("保存进行中，请稍候","warn");
        return;
      }
      var fd=new FormData(); fd.append("file",f);
      fd.append("slug",uploadSlug);
      fd.append("categories",document.getElementById("newCat").value.trim());
      fd.append("tags",document.getElementById("newTags").value.trim());
      fd.append("author",document.getElementById("newAuthor").value.trim());
      fd.append("publishedAt",document.getElementById("newDate").value.trim());
      msg(newMsg,"上传中…",false);
      apiFetchJSON("/api/igoadmin/upload-md",{method:"POST",body:fd})
        .then(function(x){
          if(x.r.ok){
            newCreatedSlug=(x.j&&x.j.slug)||uploadSlug;
            if(x.j&&x.j.slug){
              var slugEl=document.getElementById("newSlug");
              if(slugEl) slugEl.value=x.j.slug;
            }
            newSaveBaseline=newFormSnapshot();
            newUserModified=false;
            msg(newMsg,"上传成功 "+x.j.slug+(x.j.title?" · "+x.j.title:""),false);
          }
          else msg(newMsg,apiErrText(x.j,"失败 (HTTP "+x.r.status+")"),true);
        }).catch(function(e){
          if(e&&e.message==="auth") msg(newMsg,"登录已失效，请重新验证",true);
          else msg(newMsg,"上传失败："+(e&&e.message||"网络错误"),true);
        })
        .then(function(){ endArticleSave(); });
    };
    reader.readAsText(f);
    this.value="";
  };
  document.getElementById("btnPasteImg").onclick=function(){
    readClipboardImage(activeTa||activeMdTa()||document.getElementById("newMd"));
  };
  document.getElementById("btnEditPasteImg").onclick=function(){
    readClipboardImage(activeTa||activeMdTa()||document.getElementById("editMd"));
  };
  window.addEventListener("beforeunload",function(e){
    if(!getToken()) return;
    if(!hasUnsavedChanges()) return;
    e.preventDefault();
    e.returnValue="";
  });
})();


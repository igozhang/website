window.__fsSharedInit=function(base){
  var listeners={};
  function on(ev,fn){ if(!listeners[ev]) listeners[ev]=[]; listeners[ev].push(fn); }
  function emit(ev,detail){ (listeners[ev]||[]).forEach(function(fn){ try{ fn(detail); }catch(e){} }); }
  function absRoot(){
    if(typeof window.__fsAbsRoot==="function") return window.__fsAbsRoot(base);
    if(base) return base.replace(/\/$/,'');
    return '';
  }
  function urlBrowse(q){
    var u=absRoot() + "/api/browse?path=" + encodeURIComponent(q);
    var sortEl=document.getElementById("browseSort");
    var ordEl=document.getElementById("browseOrder");
    var extEl=document.getElementById("browseExt");
    if(sortEl&&sortEl.value) u+="&sort="+encodeURIComponent(sortEl.value);
    if(ordEl&&ordEl.value) u+="&order="+encodeURIComponent(ordEl.value);
    if(extEl&&extEl.value.trim()) u+="&ext="+encodeURIComponent(extEl.value.trim());
    return u;
  }
  function urlFsUsage(){
    return absRoot() + "/api/fs/usage";
  }
  function urlStashExport(){
    return absRoot() + "/api/stash/export";
  }
  function urlFetch(){
    return absRoot() + "/api/fetch";
  }
  function urlUpload(){
    return absRoot() + "/api/upload";
  }
  function urlFile(rel){
    return absRoot() + "/files/" + rel.split("/").map(encodeURIComponent).join("/");
  }
  function urlFileDelete(rel){
    return absRoot() + "/api/file?path=" + encodeURIComponent(rel);
  }
  function urlPreview(rel){
    return absRoot() + "/preview/" + rel.split("/").map(encodeURIComponent).join("/");
  }
  function urlStreamMp3(fileName){
    return absRoot() + "/stream/mp3/" + fileName.split("/").filter(Boolean).map(encodeURIComponent).join("/");
  }
  function urlMp3Cover(fileName){
    var q=fileName.split("/").filter(Boolean).join("/");
    return absRoot() + "/api/mp3/cover?path=" + encodeURIComponent(q);
  }
  function fsFaviconFromMeta(name, fallback){
    var m = document.querySelector('meta[name="' + name + '"]');
    if (m && m.content) return m.content;
    return fallback;
  }
  function urlSiteFavicon(){
    return fsFaviconFromMeta("fs-favicon-href", absRoot() + "/favicon.svg");
  }
  function urlSiteFaviconMedia(){
    return fsFaviconFromMeta("fs-favicon-media-href", absRoot() + "/favicon.png");
  }
  function absolutePageUrl(pathFromSiteRoot){
    if(typeof location==="undefined"||!pathFromSiteRoot) return "";
    var origin=String(location.origin||"").replace(/\/$/,"");
    var p=pathFromSiteRoot.charAt(0)==="/"?pathFromSiteRoot:"/"+pathFromSiteRoot;
    try{ return new URL(p,origin+"/").href; }catch(e){ return origin+p; }
  }
  function urlStash(){
    return absRoot() + "/api/stash";
  }
  function urlFooter(){
    return absRoot() + "/api/footer";
  }
  var _stashOps=(typeof window!=="undefined"&&window.__fsStashOps)||{};
  var maxStashBytes=_stashOps.maxBytes||200*1024;
  var stashPollIntervalMs=_stashOps.pollIntervalMs||3000;
  var stashPollIntervalMaxMs=_stashOps.pollIntervalMaxMs||30000;
  function utf8ByteLength(s){
    if(typeof TextEncoder!=="undefined") return new TextEncoder().encode(s).length;
    return unescape(encodeURIComponent(s)).length;
  }
  function stashSnippet100(s){
    var out="", n=0;
    for(var i=0;i<s.length&&n<100;n++){
      var cp=s.codePointAt(i);
      out+=String.fromCodePoint(cp);
      i+=(cp>0xffff)?2:1;
    }
    return out;
  }
  // 可预览后缀须与 preview_ext.go 的 previewKind 保持同步。
  function isPreviewable(name){
    var n=String(name||"").toLowerCase();
    var base=n.split("/").pop()||n;
    if(base==="dockerfile"||base.indexOf("dockerfile.")===0) return true;
    if(base==="makefile"||base==="gnumakefile"||base.indexOf("makefile.")===0) return true;
    if(base==="gemfile"||base==="procfile"||base==="jenkinsfile"||base==="cmakelists.txt") return true;
    var i=base.lastIndexOf(".");
    if(i<0) return false;
    var ext=base.slice(i);
    if(ext==="." ) return false;
    return /^(?:\.md|\.markdown|\.txt|\.text|\.log|\.csv|\.tsv|\.lst|\.out|\.err|\.yaml|\.yml|\.json|\.jsonc|\.json5|\.toml|\.tf|\.hcl|\.conf|\.cfg|\.ini|\.env|\.properties|\.service|\.timer|\.js|\.mjs|\.cjs|\.ts|\.jsx|\.tsx|\.vue|\.svelte|\.css|\.scss|\.sass|\.less|\.go|\.py|\.pyw|\.pyi|\.sh|\.bash|\.zsh|\.fish|\.ksh|\.ps1|\.psm1|\.bat|\.cmd|\.xml|\.svg|\.html|\.htm|\.xhtml|\.xsl|\.xsd|\.sql|\.graphql|\.gql|\.rs|\.java|\.kt|\.kts|\.scala|\.c|\.h|\.cpp|\.cc|\.cxx|\.hpp|\.hh|\.rb|\.php|\.lua|\.pl|\.pm|\.r|\.swift|\.m|\.mm|\.gitignore|\.dockerignore|\.editorconfig|\.npmrc|\.nvmrc|\.eslintrc|\.prettierrc|\.babelrc)$/.test(ext);
  }
  var path = "";
  function fmtSize(n){ n=Number(n)||0; if(n<1024) return n+" B"; if(n<1048576) return (n/1024).toFixed(1)+" KB"; return (n/1048576).toFixed(1)+" MB"; }
  function fmtBytesAbbr(n){
    n=Number(n)||0;
    var G=1073741824, M=1048576, K=1024;
    if(n>=G) return (n/G).toFixed(1)+"G";
    if(n>=M) return (n/M).toFixed(1)+"M";
    if(n>=K) return (n/K).toFixed(1)+"K";
    return n+" B";
  }
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
  function fmtTime(t){if(window.__fsFmtBeijing)return window.__fsFmtBeijing(t);var u=new Date(t*1000),ms=Date.UTC(u.getUTCFullYear(),u.getUTCMonth(),u.getUTCDate(),u.getUTCHours()+8,u.getUTCMinutes(),u.getUTCSeconds()),b=new Date(ms),p=function(n){return(n<10?"0":"")+n};return b.getUTCFullYear()+"-"+p(b.getUTCMonth()+1)+"-"+p(b.getUTCDate())+" "+p(b.getUTCHours())+":"+p(b.getUTCMinutes())+":"+p(b.getUTCSeconds());}
  function showToast(text,isErr){
    var w=document.getElementById("toastWrap");
    if(!w||!text) return;
    var el=document.createElement("div");
    el.className="toast "+(isErr?"toast-err":"toast-ok");
    el.textContent=text;
    w.appendChild(el);
    setTimeout(function(){ try{ el.remove(); }catch(e){} }, 4200);
  }
  function authHeaders(extra){
    var h=extra||{};
    return h;
  }
  function withAuthQuery(url){ return url||""; }
  function requireAuth(){ return true; }
  function handleAuthFailure(){ return false; }

  return {
    bus:{on:on,emit:emit},
    absRoot:absRoot,
    urlBrowse:urlBrowse,
    urlFsUsage:urlFsUsage,
    urlStashExport:urlStashExport,
    urlFetch:urlFetch,
    urlUpload:urlUpload,
    urlFile:urlFile,
    urlFileDelete:urlFileDelete,
    urlPreview:urlPreview,
    urlStreamMp3:urlStreamMp3,
    urlMp3Cover:urlMp3Cover,
    urlSiteFavicon:urlSiteFavicon,
    absolutePageUrl:absolutePageUrl,
    urlStash:urlStash,
    urlFooter:urlFooter,
    maxStashBytes:maxStashBytes,
    stashPollIntervalMs:stashPollIntervalMs,
    stashPollIntervalMaxMs:stashPollIntervalMaxMs,
    utf8ByteLength:utf8ByteLength,
    stashSnippet100:stashSnippet100,
    isPreviewable:isPreviewable,
    fmtSize:fmtSize,
    fmtBytesAbbr:fmtBytesAbbr,
    fmtTime:fmtTime,
    showToast:showToast,
    authHeaders:authHeaders,
    withAuthQuery:withAuthQuery,
    requireAuth:requireAuth,
    handleAuthFailure:handleAuthFailure
  };
};

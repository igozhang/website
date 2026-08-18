window.__fsBrowseInit=function(){
  if(!document.getElementById('list')) return;
  var __fs=window.__fs;
  function urlBrowse(q){ return __fs.urlBrowse(q); }
  function urlFile(rel){ return __fs.urlFile(rel); }
  function urlFileDelete(rel){ return __fs.urlFileDelete(rel); }
  function urlPreview(rel){ return __fs.urlPreview(rel); }
  function isPreviewable(n){ return __fs.isPreviewable(n); }
  function showToast(t,e){ return __fs.showToast(t,e); }
  var path = "";
  function fmtSize(n){ n=Number(n)||0; if(n<1024) return n+" B"; if(n<1048576) return (n/1024).toFixed(1)+" KB"; return (n/1048576).toFixed(1)+" MB"; }
  function fmtTime(t){if(window.__fsFmtBeijing)return window.__fsFmtBeijing(t);var u=new Date(t*1000),ms=Date.UTC(u.getUTCFullYear(),u.getUTCMonth(),u.getUTCDate(),u.getUTCHours()+8,u.getUTCMinutes(),u.getUTCSeconds()),b=new Date(ms),p=function(n){return(n<10?"0":"")+n};return b.getUTCFullYear()+"-"+p(b.getUTCMonth()+1)+"-"+p(b.getUTCDate())+" "+p(b.getUTCHours())+":"+p(b.getUTCMinutes())+":"+p(b.getUTCSeconds());}
  function crumb(){
    var el = document.getElementById("crumb");
    var parts = path.split("/").filter(Boolean);
    var html = '<a href="#" data-p="">根目录</a>';
    var acc = "";
    for(var i=0;i<parts.length;i++){
      acc += (acc?"/":"")+parts[i];
      var seg = acc;
      html += ' / <a href="#" data-p="'+seg+'">'+parts[i]+"</a>";
    }
    el.innerHTML = html;
    el.querySelectorAll("a").forEach(function(a){
      a.onclick=function(ev){ ev.preventDefault(); path=a.getAttribute("data-p")||""; load(); };
    });
  }
  function deleteFile(relPath, name, btn){
    if(!confirm("确定删除「"+name+"」？\n此操作不可恢复。")) return;
    if(__fs.requireAuth&&!__fs.requireAuth()) return;
    if(btn) btn.disabled=true;
    fetch(urlFileDelete(relPath),{
      method:"DELETE",
      credentials:"same-origin",
      headers:__fs.authHeaders()
    }).then(function(r){
      return r.text().then(function(t){
        var j=null; try{ j=JSON.parse(t); }catch(e){}
        if(r.status===401){
          if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,j);
          throw new Error((j&&j.error)||"未认证");
        }
        if(!r.ok) throw new Error((j&&j.error)||("HTTP "+r.status));
        return j;
      });
    }).then(function(){
      showToast("已删除 "+name,false);
      if(__fs.bus&&__fs.bus.emit){
        __fs.bus.emit("browse:refresh");
        __fs.bus.emit("usage:refresh");
      } else {
        load();
      }
    }).catch(function(e){
      if(btn) btn.disabled=false;
      showToast("删除失败："+(e&&e.message||e),true);
    });
  }
  function load(){
    crumb();
    if(__fs.requireAuth&&!__fs.requireAuth()) return;
    fetch(urlBrowse(path),{credentials:"same-origin",headers:__fs.authHeaders()}).then(function(r){
      return r.text().then(function(t){
        if(r.status===401){
          var j=null; try{ j=JSON.parse(t); }catch(e){}
          if(__fs.handleAuthFailure) __fs.handleAuthFailure(401,j);
          throw new Error((j&&j.error)||"未认证");
        }
        if(!r.ok) throw new Error("HTTP "+r.status+" "+t.slice(0,120));
        try { return JSON.parse(t); } catch(e){ throw new Error("非JSON响应："+t.slice(0,80)); }
      });
    }).then(function(data){
      var tb = document.getElementById("list");
      tb.innerHTML = "";
      if(!Array.isArray(data)){ tb.innerHTML="<tr><td colspan=5>无法加载</td></tr>"; return; }
      data.forEach(function(it){
        var relPath=(path?path+"/":"")+it.name;
        var tr = document.createElement("tr");
        var nameLink = document.createElement("td");
        if(it.isDir){
          var a = document.createElement("a"); a.className="list-name-link"; a.href="#"; a.textContent=it.name+" (DIR)";
          a.onclick=function(ev){ ev.preventDefault(); path=relPath; load(); };
          nameLink.appendChild(a);
        } else {
          var a = document.createElement("a"); a.className="list-name-link"; a.href=urlFile(relPath);
          a.textContent=it.name;
          nameLink.appendChild(a);
        }
        tr.appendChild(nameLink);
        var sz=document.createElement("td"); sz.className="size"; sz.textContent=it.isDir?"—":fmtSize(it.size);
        var tm=document.createElement("td"); tm.className="size col-time"; tm.textContent=fmtTime(it.modTime);
        var pv=document.createElement("td"); pv.className="size";
        if(it.isDir){ pv.textContent=""; }
        else if(isPreviewable(it.name)){
          var pa=document.createElement("a"); pa.href=urlPreview(relPath);
          pa.target="_blank"; pa.rel="noopener"; pa.className="preview-tab"; pa.textContent="预览";
          pv.appendChild(pa);
        } else { pv.textContent="—"; }
        var op=document.createElement("td"); op.className="size col-ops";
        if(it.isDir){ op.textContent=""; }
        else {
          var del=document.createElement("button");
          del.type="button";
          del.className="list-del-btn";
          del.textContent="删";
          del.title="删除 "+it.name;
          del.onclick=function(){ deleteFile(relPath, it.name, del); };
          op.appendChild(del);
        }
        tr.appendChild(sz); tr.appendChild(pv); tr.appendChild(tm); tr.appendChild(op);
        tb.appendChild(tr);
      });
    }).catch(function(e){
      document.getElementById("list").innerHTML="<tr><td colspan=5>请求失败</td></tr>";
      showToast("目录加载失败："+(e&&e.message||e),true);
    });
  }
  var browseSortEl=document.getElementById("browseSort");
  var browseOrdEl=document.getElementById("browseOrder");
  var browseExtEl=document.getElementById("browseExt");
  if(browseSortEl) browseSortEl.onchange=function(){ load(); };
  if(browseOrdEl) browseOrdEl.onchange=function(){ load(); };
  if(browseExtEl) browseExtEl.onchange=function(){ load(); };

  __fs.bus.on('browse:refresh', function(){ load(); });
  __fs.bus.on('auth:ok', function(){ load(); });
  load();
};

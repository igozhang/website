// 全站公共 URL 前缀解析（header/footer/mp3/fs 共用唯一实现）。
// 优先级：window.__fsPublicBase → 调用方传入 base → footer[data-base] → meta[fs-public-base]
// → main[data-fs-pub] → base[href] → pathname 去掉已知子路径或文章 slug。
window.__fsAbsRoot=function(optBase){
  if(typeof window.__fsPublicBase==="string"&&window.__fsPublicBase.length){
    return window.__fsPublicBase.replace(/\/$/,"");
  }
  if(optBase!==undefined&&optBase!==null&&String(optBase).length){
    return String(optBase).replace(/\/$/,"");
  }
  var foot=document.querySelector(".site-foot[data-base]");
  if(foot&&foot.hasAttribute("data-base")){
    return String(foot.getAttribute("data-base")||"").replace(/\/$/,"");
  }
  var meta=document.querySelector('meta[name="fs-public-base"]');
  if(meta&&meta.hasAttribute("content")){
    return String(meta.getAttribute("content")||"").replace(/\/$/,"");
  }
  var main=document.querySelector("main.shell[data-fs-pub]");
  if(main&&main.hasAttribute("data-fs-pub")){
    return String(main.getAttribute("data-fs-pub")||"").replace(/\/$/,"");
  }
  var baseEl=document.querySelector("base[href]");
  if(baseEl){
    try{
      var bp=new URL(baseEl.getAttribute("href"),location.href).pathname.replace(/\/$/,"");
      if(bp) return bp;
    }catch(e){}
  }
  var p=location.pathname.replace(/\/$/,"");
  if(/\/(gomoku|xiangqi|texas|zhijian|igokey|ops|fs|igoadmin|mermaid)$/i.test(p)){
    return p.replace(/\/(gomoku|xiangqi|texas|zhijian|igokey|ops|fs|igoadmin|mermaid)$/i,"")||"";
  }
  if(/\/play\/(gomoku|xiangqi|texas|zhijian)$/i.test(p)){
    return p.replace(/\/play\/(gomoku|xiangqi|texas|zhijian)$/i,"")||"";
  }
  if(/\/play$/i.test(p)) return p.replace(/\/play$/i,"")||"";
  if(/^\/[a-z][a-z0-9_]*$/.test(p)) return "";
  return p;
};

// <base href="/"> 会把 href="#id" 解析成首页 /#id；改成当前路径 + hash。
window.__fsFixFragmentHrefs=function(root){
  var scope=root&&root.querySelectorAll?root:document;
  var path=location.pathname+(location.search||"");
  if(!path) path="/";
  var list=scope.querySelectorAll?scope.querySelectorAll('a[href^="#"]'):[];
  for(var i=0;i<list.length;i++){
    var a=list[i];
    var h=a.getAttribute("href");
    if(!h||h==="#") continue;
    // 已是绝对路径或带页路径的 fragment
    if(h.charAt(0)==="#"&&h.length>1){
      a.setAttribute("href",path+h);
    }
  }
};
document.addEventListener("DOMContentLoaded",function(){
  try{ window.__fsFixFragmentHrefs(document); }catch(e){}
});

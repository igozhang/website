(function(){
  if(window.__fsSiteUIGuardInit) return;
  window.__fsSiteUIGuardInit=true;

  function isTextField(el){
    if(!el||!el.tagName) return false;
    var tag=el.tagName;
    if(tag==="TEXTAREA") return true;
    if(tag==="SELECT") return true;
    if(tag==="INPUT"){
      var ty=(el.type||"text").toLowerCase();
      if(ty==="hidden"||ty==="button"||ty==="submit"||ty==="reset"||ty==="checkbox"||ty==="radio"||ty==="file"||ty==="image") return false;
      return true;
    }
    if(el.isContentEditable) return true;
    return false;
  }

  function isInteractive(el){
    if(!el||!el.closest) return false;
    return !!el.closest(
      "button, a[href], label, select, summary, [role='button'], input, textarea, [contenteditable='true'], .phrase-next, .zhijian-btn, .zhijian-key"
    );
  }

  function shouldBlurFocus(el){
    return el&&!isTextField(el)&&!isInteractive(el);
  }

  function blurStrayFocus(){
    var ae=document.activeElement;
    if(!ae||ae===document.body||ae===document.documentElement) return;
    if(!shouldBlurFocus(ae)) return;
    try{ ae.blur(); }catch(e){}
  }

  function suppressFocus(ev){
    if(!ev||!ev.target) return;
    if(!shouldBlurFocus(ev.target)) return;
    try{ ev.target.blur(); }catch(e){}
    blurStrayFocus();
  }

  document.addEventListener("focusin", suppressFocus, true);
})();

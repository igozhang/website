window.__fsPhraseInit=function(pairs){
  var phraseLast=-1;
  function showRandomPhrase(){
    if(!pairs||!pairs.length)return;
    var i=0;
    if(pairs.length===1){
      i=0;
    } else {
      do{ i=Math.floor(Math.random()*pairs.length); }while(i===phraseLast);
    }
    phraseLast=i;
    var el=document.getElementById("phraseLine");
    if(!el)return;
    el.querySelector(".phrase-en").textContent=pairs[i].en;
    el.querySelector(".phrase-zh").textContent=pairs[i].zh;
  }
  showRandomPhrase();
  var pn=document.getElementById("phraseNext");
  if(pn) pn.onclick=function(){ showRandomPhrase(); stopPhraseAuto(); startPhraseAuto(); };
  var phraseAutoInt=null;
  function phraseAutoTick(){
    if(typeof document!=="undefined"&&document.visibilityState!=="visible") return;
    showRandomPhrase();
  }
  function startPhraseAuto(){
    if(phraseAutoInt){ clearInterval(phraseAutoInt); phraseAutoInt=null; }
    if(typeof window.matchMedia==="function"&&window.matchMedia("(prefers-reduced-motion: reduce)").matches) return;
    if(!pairs||pairs.length<2) return;
    phraseAutoInt=setInterval(phraseAutoTick, 20000);
  }
  function stopPhraseAuto(){
    if(phraseAutoInt){ clearInterval(phraseAutoInt); phraseAutoInt=null; }
  }
  document.addEventListener("fs:pagehidden", stopPhraseAuto);
  document.addEventListener("fs:pagevisible", startPhraseAuto);
  startPhraseAuto();
};

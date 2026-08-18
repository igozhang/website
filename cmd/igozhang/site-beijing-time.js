(function(){
  if(window.__fsBeijingTimeInit) return;
  window.__fsBeijingTimeInit=true;

  var BJ_BASE=8;
  function bjExtraHours(){
    var n=Number(window.__fsBeijingOffsetHours);
    if(!isFinite(n)) n=0;
    if(n>12) n=12;
    if(n<-12) n=-12;
    return n|0;
  }
  function pad2(n){ return (n<10?"0":"")+n; }

  function fmtBeijingUnixSec(unixSec){
    unixSec=Math.floor(Number(unixSec)||0);
    if(!unixSec) return "";
    var u=new Date(unixSec*1000);
    var ms=Date.UTC(
      u.getUTCFullYear(),u.getUTCMonth(),u.getUTCDate(),
      u.getUTCHours()+BJ_BASE+bjExtraHours(),u.getUTCMinutes(),u.getUTCSeconds()
    );
    var b=new Date(ms);
    return b.getUTCFullYear()+"-"+pad2(b.getUTCMonth()+1)+"-"+pad2(b.getUTCDate())+" "+
      pad2(b.getUTCHours())+":"+pad2(b.getUTCMinutes())+":"+pad2(b.getUTCSeconds());
  }

  window.__fsFmtBeijing=fmtBeijingUnixSec;
  window.__fsFmtBeijingNow=function(){
    return fmtBeijingUnixSec(Math.floor(Date.now()/1000));
  };

  function tick(){
    var el=document.getElementById("siteBeijingTime");
    if(!el) return;
    var now=window.__fsFmtBeijingNow();
    el.textContent="北京时间 "+now;
    try{ el.dateTime=now; }catch(e){}
  }
  tick();
  setInterval(tick,1000);
})();

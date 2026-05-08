// import OSC from "./osc.min.js";

function main(){
    "use strict";

    const plugin = new OSC.WebsocketClientPlugin({ host: location.hostname, port: location.port });
    const osc = new OSC({ plugin: plugin });

    osc.on('open',     () => { statusLabel.innerHTML = "Server Running"; });
    osc.on('/name',    (message) => procName(message));
    osc.on('/message', (message) => procMsg(message));
    osc.on('/time',    (message) => procTime(message));
    osc.on('/refresh', () => location.reload());
    osc.on('/connect', () => resetDisplay());
    osc.on('/stop',    () => plugin.close());
    osc.on('/tminus',  (message) => procTminus(message));
    osc.on('close',    () => onClose());

    // Elements
    const clipNameEl    = document.getElementById("clipname");
    const infoElapsed   = document.getElementById("info-elapsed");
    const infoRemaining = document.getElementById("info-remaining");
    const infoTotal     = document.getElementById("info-total");
    const systemClock   = document.getElementById("system-clock");
    const clipLength    = document.getElementById("ms");
    const statusLabel   = document.getElementById("status");
    const tableBorder   = document.getElementById("tableborder");

    // System clock — ticks every second, independent of OSC
    function tickClock() {
        const now = new Date();
        const hh  = String(now.getHours()).padStart(2, '0');
        const mm  = String(now.getMinutes()).padStart(2, '0');
        const ss  = String(now.getSeconds()).padStart(2, '0');
        systemClock.innerHTML = hh + ':' + mm + ':' + ss;
    }
    tickClock();
    setInterval(tickClock, 1000);

    resetDisplay();
    osc.open();

    function onClose() {
        statusLabel.innerHTML = "Server Stopped";
        tableBorder.style.borderColor = "#ff4545";
        resetDisplay();
    }

    function procName(data) {
        clipNameEl.innerHTML = data.args[0];
    }

    // tminus is kept for compatibility but no longer drives a separate element
    function procTminus(data) { }

    function resetDisplay() {
        clipNameEl.innerHTML    = "None";
        infoElapsed.innerHTML   = "+00:00:00.000";
        infoRemaining.innerHTML = "-00:00:00.000";
        infoTotal.innerHTML     = "00:00:00.000";
        clipLength.innerHTML    = "00:00:00.000";
        tableBorder.style.borderColor = "#4b5457";
    }

    async function procMsg(data) {
        // No msg element in new layout — kept for server compatibility
    }

    function procTime(data) {
        // args: [0]=remaining string, [1]=clip length HMS, [2]=elapsed string
        const remainStr  = data.args[0];
        const lengthStr  = data.args[1];
        const elapsedStr = data.args[2];

        clipLength.innerHTML    = lengthStr;
        infoElapsed.innerHTML   = elapsedStr !== undefined ? elapsedStr : "+00:00:00.000";
        infoRemaining.innerHTML = remainStr;
        infoTotal.innerHTML     = lengthStr;

        // Flash card border red when under 10 seconds remaining
        const parts = remainStr.split(":");
        const hPart = parts[0].substring(1);
        const mPart = parts[1];
        const sPart = parts[2].split(".")[0];
        const totalSecsLeft = parseInt(hPart) * 3600 + parseInt(mPart) * 60 + parseInt(sPart);
        tableBorder.style.borderColor = (totalSecsLeft <= 10) ? "#ff4545" : "#4b5457";
    }
}

main();

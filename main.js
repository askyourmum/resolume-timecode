// import OSC from "./osc.min.js";

function main(){
    "use strict";

    const plugin = new OSC.WebsocketClientPlugin({ host: location.hostname, port: location.port });
    const osc = new OSC({ plugin: plugin });

    osc.on('open', () => {statusLabel.innerHTML = "Server Running";});
    osc.on('/name', (message) => procName(message));
    osc.on('/message', (message) => procMsg(message));
    osc.on('/time', (message) => procTime(message));
    osc.on('/refresh', () => location.reload());
    osc.on('/connect', () => reset());
    osc.on('/stop', () => plugin.close());
    osc.on('/tminus', (message) => procTminus(message));
    osc.on('close', () => close());

    const timecodeHours = document.getElementById("timecode-hours");
    const timecodeMinutes = document.getElementById("timecode-minutes");
    const timecodeSeconds = document.getElementById("timecode-seconds");
    const timecodeMS = document.getElementById("timecode-ms");
    const timecodeMinus = document.getElementsByClassName("minus");
    const timecodeClipName = document.getElementById("clipname");
    const table = document.getElementById("table");
    const tableBorder = document.getElementById("tableborder");
    const clipLength = document.getElementById("ms");
    const statusLabel = document.getElementById("status");
    const message = document.getElementById("msg");

    // New elements
    const systemClock   = document.getElementById("system-clock");
    const infoElapsed   = document.getElementById("info-elapsed");
    const infoRemaining = document.getElementById("info-remaining");
    const infoTotal     = document.getElementById("info-total");

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

    reset();
    osc.open();

    function close() {
        statusLabel.innerHTML = "Server Stopped";

        timecodeHours.innerHTML = "00";
        timecodeMinutes.innerHTML = "00";
        timecodeSeconds.innerHTML = "00";
        timecodeMS.innerHTML = "000";
        clipLength.innerHTML = "00:00:00.000";

        infoElapsed.innerHTML   = "+00:00:00.000";
        infoRemaining.innerHTML = "-00:00:00.000";
        infoTotal.innerHTML     = "00:00:00.000";

        table.style.color = "#ff4545";
        tableBorder.style.borderColor = "#ff4545";
    }

    function procName(data) {
        timecodeClipName.innerHTML = data.args[0];
    }

    function procTminus(data) {
        data.args[0] === true ? timecodeMinus[0].innerHTML = "-" : timecodeMinus[0].innerHTML = '+'
    }

    function reset() {
        timecodeHours.innerHTML = "00";
        timecodeMinutes.innerHTML = "00";
        timecodeSeconds.innerHTML = "00";
        timecodeMS.innerHTML = "000";
        clipLength.innerHTML = "00:00:00.000";

        infoElapsed.innerHTML   = "+00:00:00.000";
        infoRemaining.innerHTML = "-00:00:00.000";
        infoTotal.innerHTML     = "00:00:00.000";
    }

    async function procMsg(data) {
        data = data.args[0];
        message.innerHTML = (data === "") ? "Timecode Monitor" : data;
        if (data === "") {
            return;
        }
        for (let i = 0; i < 3; i++) {
            message.style.color = "#ff4545";
            await new Promise(r => setTimeout(r, 500));
            message.style.color = "#FDFBF7";
            await new Promise(r => setTimeout(r, 500));
        }
    }

    function procTime(data) {
        // args: [0]=remaining string, [1]=clip length string, [2]=elapsed string
        const remainStr  = data.args[0];  // e.g. "-00:01:23.456"
        const lengthStr  = data.args[1];  // e.g. "83.456s"
        const elapsedStr = data.args[2];  // e.g. "+00:00:12.000"

        // Legacy settings panel clip length
        clipLength.innerHTML = lengthStr;

        // Info panel
        infoElapsed.innerHTML   = elapsedStr  !== undefined ? elapsedStr  : "+00:00:00.000";
        infoRemaining.innerHTML = remainStr;
        infoTotal.innerHTML     = lengthStr;

        // Main timecode display (remaining, same as original)
        // remainStr format: "-HH:MM:SS.mmm"
        const parts = remainStr.split(":");
        const hPart = parts[0].substring(1);   // strip leading '-'
        const mPart = parts[1];
        const sPart = parts[2].split(".")[0];
        const msPart = parts[2].split(".")[1];

        timecodeHours.innerHTML   = hPart;
        timecodeMinutes.innerHTML = mPart;
        timecodeSeconds.innerHTML = sPart;
        timecodeMS.innerHTML      = msPart;

        // Flash red when under 10 seconds remaining
        const totalSecsLeft = parseInt(hPart) * 3600 + parseInt(mPart) * 60 + parseInt(sPart);
        if (totalSecsLeft <= 10) {
            table.style.color = "#ff4545";
            tableBorder.style.borderColor = "#ff4545";
        } else {
            table.style.color = "#45ff45";
            tableBorder.style.borderColor = "#4b5457";
        }
    }
}

main();
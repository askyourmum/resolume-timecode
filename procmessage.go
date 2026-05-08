package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/chabad360/go-osc/osc"
)

var (
	clipName         = ""
	directionForward = true

	timeLeft    string
	timeElapsed string

	clipLength  float32
	posPrev     float32
	clipSwitched = false // true for one cycle after a path switch in auto-track mode

	// autoTrackOutput: when true, procMsg will switch clipPath to whichever
	// clip sends /connected = 1, so we always follow the live output clip.
	// Enabled automatically when the configured path ends with "/connectedclip".
	autoTrackOutput = false
)

func procMsg(data *osc.Message) {
	// Auto-track output: when enabled, watch ALL clips for /connected=1
	// and switch clipPath to that clip so we always follow the live output.
	// On a crossfade/jump, we do NOT reset posPrev — we just swap the path
	// and re-query name+duration so the timer transitions cleanly.
	if autoTrackOutput && strings.HasSuffix(data.Address, "/connected") {
		if len(data.Arguments) > 0 {
			var connected int32
			switch v := data.Arguments[0].(type) {
			case int32:
				connected = v
			case float32:
				connected = int32(v)
			}
			if connected == 1 {
				// Address is e.g. /composition/layers/2/clips/3/connected
				newPath := strings.TrimSuffix(data.Address, "/connected")
				if newPath != clipPath {
					clipPath = newPath
					clipSwitched = true // skip the posPrev==pos guard for one cycle
					// Re-query name and duration for the new clip.
					// Do NOT reset posPrev — position updates from the new
					// clip arrive immediately and must not be dropped.
					lightReset()
				}
			}
		}
	}

	if strings.HasPrefix(data.Address, clipPath) {
		switch {
		case strings.HasSuffix(data.Address, "/position"):
			procPos(data)
		case strings.HasSuffix(data.Address, "direction"):
			procDirection(data)
		case strings.HasSuffix(data.Address, "/name"):
			procName(data)
		case strings.HasSuffix(data.Address, "/duration"):
			procDuration(data)
		case strings.HasSuffix(data.Address, "/connect"):
			// Only reset on /connect when NOT in auto-track mode.
			// In auto-track mode the /connected handler above already
			// manages clip switches — resetting here would blank the
			// timer mid-crossfade.
			if !autoTrackOutput {
				reset()
			}
		case strings.Contains(data.Address, "/select"):
			// Suppress /select resets in auto-track mode — clicking a clip
			// in the arena grid while another plays to output must not
			// blank the timer.
			if !autoTrackOutput {
				reset()
			}
		}
	}
}

func procDirection(data *osc.Message) {
	directionForward = data.Arguments[0].(int32) != 0
	if !directionForward {
		posPrev = 1 - posPrev
	}
}

func procName(data *osc.Message) {
	clipName = data.Arguments[0].(string)
	clipNameBinding.Set("Clip Name: " + clipName)
	broadcast.Publish(osc.NewMessage("/name", clipName))
}

func procDuration(data *osc.Message) {
	clipLength = (data.Arguments[0].(float32) * 604800) + 0.001
	clipLengthBinding.Set(fmt.Sprintf("Clip Length: %.3fs", clipLength))
	broadcast.Publish(osc.NewMessage("/duration", clipLength))
}

func reset() {
	lightReset()

	posPrev = 0
}

func lightReset() {
	message.Address = clipPath + "/name"
	message2.Address = clipPath + "/transport/position/behaviour/duration"
	if _, err := oscServer.WriteTo(osc.NewBundle(message, message2), OSCAddr+":"+OSCPort); err != nil {
		fmt.Println(err)
	}
}

func procPos(data *osc.Message) {
	pos := data.Arguments[0].(float32)

	if !directionForward {
		pos = 1 - pos
	}

	// On a clip switch, bypass the stale-position guards for one cycle
	// so the new clip's first position update is never dropped.
	if clipSwitched {
		clipSwitched = false
		posPrev = pos
		// fall through to calculate and broadcast
	} else {
		if posPrev == 0 || posPrev == pos || pos < 0.002 {
			posPrev = pos
			return
		}
		currentPosInterval := pos - posPrev
		if currentPosInterval < 0 && posPrev > 0 {
			return
		}
	}

	posPrev = pos

	// posDisplay is pos after optional invert — used for remaining (T-/T+ display)
	posDisplay := pos
	if clipInvert {
		posDisplay = 1 - pos
	}

	// remaining time: how far from the end
	tRemaining := (clipLength * 1000) * (1 - posDisplay)
	timeActual := time.UnixMilli(int64(tRemaining)).UTC()
	timeLeft = fmt.Sprintf("-%02d:%02d:%02d.%03d", timeActual.Hour(), timeActual.Minute(), timeActual.Second(), timeActual.Nanosecond()/1000000)

	// elapsed time: always based on raw forward position, never inverted
	tElapsed := (clipLength * 1000) * pos
	timeElapsedActual := time.UnixMilli(int64(tElapsed)).UTC()
	timeElapsed = fmt.Sprintf("+%02d:%02d:%02d.%03d", timeElapsedActual.Hour(), timeElapsedActual.Minute(), timeElapsedActual.Second(), timeElapsedActual.Nanosecond()/1000000)

	// broadcast: arg0=remaining, arg1=clip length string, arg2=elapsed
	broadcast.Publish(osc.NewMessage("/time", timeLeft, fmt.Sprintf("%.3fs", clipLength), timeElapsed))
	broadcast.Send()

	//fmt.Println(message, clipLength, samples, pos, currentPosInterval, currentTimeInterval, currentEstSize, posInterval, timeInterval, average(estSizeBuffer))

}

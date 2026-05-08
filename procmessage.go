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
	timeTotal   string

	clipLength float32
	posPrev    float32

	// autoTrackOutput: when true, procMsg will switch clipPath to whichever
	// clip sends /connected = 1, so we always follow the live output clip.
	// Enabled automatically when the configured path ends with "/connectedclip".
	autoTrackOutput = false
)

func procMsg(data *osc.Message) {
	// Auto-track output: when enabled, watch ALL clips for /connected=1
	// and switch clipPath to that clip so we follow the live output.
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
				// Strip the trailing /connected to get the clip base path
				newPath := strings.TrimSuffix(data.Address, "/connected")
				if newPath != clipPath {
					clipPath = newPath
					reset()
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
			reset()
		case strings.Contains(data.Address, "/select"):
			reset()
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
	t := time.UnixMilli(int64(clipLength * 1000)).UTC()
	timeTotal = fmt.Sprintf("%02d:%02d:%02d.%03d", t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1000000)
	clipLengthBinding.Set(fmt.Sprintf("Clip Length: %s", timeTotal))
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

	if posPrev == 0 || posPrev == pos || pos < 0.002 {
		posPrev = pos
		return
	}

	currentPosInterval := pos - posPrev

	if currentPosInterval < 0 && posPrev > 0 {
		return
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

	// broadcast: arg0=remaining, arg1=clip length HMS, arg2=elapsed
	broadcast.Publish(osc.NewMessage("/time", timeLeft, timeTotal, timeElapsed))
	broadcast.Send()

	//fmt.Println(message, clipLength, samples, pos, currentPosInterval, currentTimeInterval, currentEstSize, posInterval, timeInterval, average(estSizeBuffer))

}

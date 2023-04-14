package T_SR

import (
	"Project/messageHandler"
	"math/rand"
	"time"
)

func Send(
	msgToNetwork chan<- SROnNet,
	msgToSend <-chan messageHandler.NetworkPackage,
) {
	rand.Seed(time.Now().UnixNano())
	var sendingMsg SROnNet
	for {
		select {
		case msg := <-msgToSend:
			msgId := rand.Intn(90000) + 10000
			sendingMsg = SROnNet{
				Message: msg,
				MsgId:   msgId}
			msgToNetwork <- sendingMsg
		default:
			time.Sleep(10 * time.Millisecond)
			msgToNetwork <- sendingMsg
		}
	}
}

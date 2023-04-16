package networkBridge

import (
	"Project/messageHandler"
	"math/rand"
	"time"
)

func Send(
	chMsgToNetwork chan<- SROnNet,
	chMsgToNBSend <-chan messageHandler.NetworkPackage,
) {
	rand.Seed(time.Now().UnixNano())
	var sendingMsg SROnNet
	for {
		select {
		case msg := <-chMsgToNBSend:
			msgId := rand.Intn(90000) + 10000
			sendingMsg = SROnNet{
				Message: msg,
				MsgId:   msgId}
			chMsgToNetwork <- sendingMsg
		default:
			time.Sleep(10 * time.Millisecond)
			chMsgToNetwork <- sendingMsg
		}
	}
}

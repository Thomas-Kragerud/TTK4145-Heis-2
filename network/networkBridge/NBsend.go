package networkBridge

import (
	"Project/messageHandler"
	"math/rand"
	"time"
)

func NBSend(
	chMsgToNetwork chan<- SROnNet,
	chMsgToNBSend <-chan messageHandler.NetworkPackage,
) {
	rand.Seed(time.Now().UnixNano())
	var sendingMsg SROnNet
	spam := func(send SROnNet) {
		for i := 0; i < 50; i++ {
			chMsgToNetwork <- send
			time.Sleep(10 * time.Microsecond)
		}

	}
	for {
		select {
		case msg := <-chMsgToNBSend:
			msgId := rand.Intn(90000) + 10000
			sendingMsg = SROnNet{
				Message: msg,
				MsgId:   msgId}
			chMsgToNetwork <- sendingMsg
			go spam(sendingMsg)
		default:
			time.Sleep(5 * time.Millisecond)
			chMsgToNetwork <- sendingMsg
		}
	}
}

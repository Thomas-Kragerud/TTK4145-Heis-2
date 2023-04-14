package T_SR

import (
	"Project/messageHandler"
	"github.com/oleiade/lane"
)

func Recive(
	msgFromNetwork <-chan SROnNet,
	chMsgFromReciver chan<- messageHandler.NetworkPackage,
) {
	capacity := 1000
	ringBuffer := lane.NewDeque()
	var found bool
	for {
		select {
		case msg := <-msgFromNetwork:
			found = false
			for i := 0; i < ringBuffer.Size(); i++ {
				val := ringBuffer.First()
				if val == msg.MsgId {
					found = true
					break
				}
				ringBuffer.Append(ringBuffer.Shift()) // Move to end of queue
			}
			if !found {
				if ringBuffer.Size() < capacity {
					ringBuffer.Append(msg.MsgId)
				} else {
					ringBuffer.Shift()
					ringBuffer.Append(msg.MsgId)
				}
				//fmt.Printf("Recived message: %v\n", msg.Message)
				chMsgFromReciver <- msg.Message
			} else {
				//fmt.Println("Message already recieved")
			}
		}
	}
}

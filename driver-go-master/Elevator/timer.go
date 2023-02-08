package Elevator

import "time"

func Timer_start(duration float64, c chan bool) {
	timer := time.NewTimer(2 * time.Second)
	<-timer.C
	c <- true
}

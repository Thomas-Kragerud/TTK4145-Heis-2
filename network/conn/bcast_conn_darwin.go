//go:build darwin
// +build darwin

package conn

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

// DialBroadcastUDP creates a UDP packet connection that can broadcast messages to multiple recipients
// @param port:
func DialBroadcastUDP(port int) net.PacketConn {
	// Create a socket, AF_INET for IPv4,SOCK_DGRAM for UDP socket, IPPROTO_UDP for UDP
	s, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)
	if err != nil {
		fmt.Println("Error: Socket:", err)
	}

	// ** Sets socket options **
	//SO_REUSEADDR lets multiple sockets bound to same address
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	if err != nil {
		fmt.Println("Error: SetSockOpt REUSEADDR:", err)
	}
	//SO_BROADCAST allows brodcast message to be sent fromm  this socket
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)
	if err != nil {
		fmt.Println("Error: SetSockOpt BROADCAST:", err)
	}
	//SO_REUSEPORT lets multiple sockets bee bound to same port
	syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
	if err != nil {
		fmt.Println("Error: SetSockOpt REUSEPORT:", err)
	}

	// Binds socket to the specified port
	syscall.Bind(s, &syscall.SockaddrInet4{Port: port})
	if err != nil {
		fmt.Println("Error: Bind:", err)
	}

	// New fileobject from socket descriptor
	f := os.NewFile(uintptr(s), "")
	// Converts the socket descriptor to a file descriptor
	conn, err := net.FilePacketConn(f)
	if err != nil {
		fmt.Println("Error: FilePacketConn:", err)
	}
	//Close the file object
	f.Close()

	// returns net.PacketConn object, this object can be used to send
	//and receive UDP packets to and from multiple recipients
	return conn
}

package bcast

import (
	"Project/network/conn"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
)

const bufSize = 1024

// Transmitter broadcasts data to all devices on local network using UDP
// It takes a port number, which is the  port  number to use for the UDP brodcast,
// and it takes "chans", which is a variadic parameter of interface{} type.
//
// The first thing the function does is to call "checkArgs" function to verify that each of the
// channels' inn "chans" are receive only.
//
// The function then crates two arrays of "reflect.SelectCase" and "string" type.
// The "reflect.SelectCase" type is used to represent a single channel operation in a select statement.
// The "string" type is used to represent the type of the channel.
//
// The function establishes a UDP connection  using the "conn.DialBrodcastUDP(port)" method
// and resolves the UDP adress "255.255.255.255":port using "net.ResolveUDPAddr"
//
// Finaly the function  enters into an infinite loop where it continuously waits for a message from any
// of the channels' inn "chans". The selecte valueis converte to JSON string using the "json.Marshal" method
func Transmitter(port int, chans ...interface{}) {
	checkArgs(chans...) // verify that each of the channels are receive only
	typeNames := make([]string, len(chans))
	selectCases := make([]reflect.SelectCase, len(typeNames))
	for i, ch := range chans {
		selectCases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
		typeNames[i] = reflect.TypeOf(ch).Elem().String()
	}

	conn := conn.DialBroadcastUDP(port) // Establish UDP connection
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	for {
		chosen, value, _ := reflect.Select(selectCases)
		jsonstr, _ := json.Marshal(value.Interface())
		ttj, _ := json.Marshal(typeTaggedJSON{
			TypeId: typeNames[chosen],
			JSON:   jsonstr,
		})
		if len(ttj) > bufSize {
			panic(fmt.Sprintf(
				"Tried to send a message longer than the buffer size (length: %d, buffer size: %d)\n\t'%s'\n"+
					"Either send smaller packets, or go to network/bcast/bcast.go and increase the buffer size",
				len(ttj), bufSize, string(ttj)))
		}
		conn.WriteTo(ttj, addr)

	}
}

// Matches type-tagged JSON received on `port` to element types of `chans`, then
// sends the decoded value on the corresponding channel
func Receiver(port int, chans ...interface{}) {
	checkArgs(chans...)
	chansMap := make(map[string]interface{})
	for _, ch := range chans {
		chansMap[reflect.TypeOf(ch).Elem().String()] = ch
	}

	var buf [bufSize]byte
	conn := conn.DialBroadcastUDP(port)
	for {
		n, _, e := conn.ReadFrom(buf[0:])
		if e != nil {
			fmt.Printf("bcast.Receiver(%d, ...):ReadFrom() failed: \"%+v\"\n", port, e)
		}

		var ttj typeTaggedJSON
		json.Unmarshal(buf[0:n], &ttj)
		ch, ok := chansMap[ttj.TypeId]
		if !ok {
			continue
		}
		v := reflect.New(reflect.TypeOf(ch).Elem())
		json.Unmarshal(ttj.JSON, v.Interface())
		reflect.Select([]reflect.SelectCase{{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(ch),
			Send: reflect.Indirect(v),
		}})
	}
}

type typeTaggedJSON struct {
	TypeId string
	JSON   []byte
}

// Checks that args to Tx'er/Rx'er are valid:
//
//	All args must be channels
//	Element types of channels must be encodable with JSON
//	No element types are repeated
//
// Implementation note:
//   - Why there is no `isMarshalable()` function in encoding/json is a mystery,
//     so the tests on element type are hand-copied from `encoding/json/encode.go`
func checkArgs(chans ...interface{}) {
	n := 0
	for range chans {
		n++
	}
	elemTypes := make([]reflect.Type, n)

	for i, ch := range chans {
		// Must be a channel
		if reflect.ValueOf(ch).Kind() != reflect.Chan {
			panic(fmt.Sprintf(
				"Argument must be a channel, got '%s' instead (arg# %d)",
				reflect.TypeOf(ch).String(), i+1))
		}

		elemType := reflect.TypeOf(ch).Elem()

		// Element type must not be repeated
		for j, e := range elemTypes {
			if e == elemType {
				panic(fmt.Sprintf(
					"All channels must have mutually different element types, arg# %d and arg# %d both have element type '%s'",
					j+1, i+1, e.String()))
			}
		}
		elemTypes[i] = elemType

		// Element type must be encodable with JSON
		checkTypeRecursive(elemType, []int{i + 1})

	}
}

func checkTypeRecursive(val reflect.Type, offsets []int) {
	switch val.Kind() {
	case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		panic(fmt.Sprintf(
			"Channel element type must be supported by JSON, got '%s' instead (nested arg# %v)",
			val.String(), offsets))
	case reflect.Map:
		if val.Key().Kind() != reflect.String {
			panic(fmt.Sprintf(
				"Channel element type must be supported by JSON, got '%s' instead (map keys must be 'string') (nested arg# %v)",
				val.String(), offsets))
		}
		checkTypeRecursive(val.Elem(), offsets)
	case reflect.Array, reflect.Ptr, reflect.Slice:
		checkTypeRecursive(val.Elem(), offsets)
	case reflect.Struct:
		for idx := 0; idx < val.NumField(); idx++ {
			checkTypeRecursive(val.Field(idx).Type, append(offsets, idx+1))
		}
	}
}

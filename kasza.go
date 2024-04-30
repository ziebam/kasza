package main

import (
	"errors"
	"log"
	"net/http"
)

func webSocketHandle(writer http.ResponseWriter, request *http.Request) {
	webSocket, err := New(writer, request)
	if err != nil {
		log.Println(err)
		return
	}

	err = webSocket.Handshake()
	if err != nil {
		log.Println(err)
		return
	}

	defer webSocket.Close()

	firstByte, err := webSocket.bufrw.ReadByte()
	if err != nil {
		log.Println(err)
		return
	}

	fin := firstByte & 0b10000000
	if fin == 0 {
		log.Println(errors.New("no continuations support for now"))
		return
	}

	opcode := firstByte & 0b00001111
	if opcode != 1 {
		log.Println(errors.New("only opcode 1 supported for now"))
		return
	}

	secondByte, err := webSocket.bufrw.ReadByte()
	if err != nil {
		log.Println(err)
		return
	}

	isMasked := secondByte & 0b10000000
	if isMasked == 0 {
		log.Println(errors.New("according to the spec, frames should always be masked"))
		return
	}

	payloadSize := secondByte & 0b01111111
	if payloadSize >= 126 {
		log.Println(errors.New("no big boy payloads supported"))
		return
	}

	mask := make([]byte, 4)
	bytesRead, err := webSocket.bufrw.Read(mask)
	if bytesRead < 4 {
		log.Println(err)
		return
	}

	data := make([]byte, int(payloadSize))
	bytesRead, err = webSocket.bufrw.Read(data)
	if bytesRead < int(payloadSize) {
		log.Println(err)
		return
	}

	for i := 0; i < int(payloadSize); i++ {
		data[i] = data[i] ^ mask[i%4]
	}

	response := "Me too! UwU"

	output := make([]byte, len(response)+2)
	for i := 0; i < len(output); i++ {
		if i == 0 {
			output[i] = 0b10000001
			continue
		}

		if i == 1 {
			output[i] = byte(len(response))
			continue
		}

		output[i] = byte(response[i-2])
	}

	webSocket.write(output)
}

func main() {
	http.HandleFunc("/", webSocketHandle)
	log.Fatal(http.ListenAndServe(":4269", nil))
}

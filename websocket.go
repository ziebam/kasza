package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
)

type Conn interface {
	Close() error
}

type WebSocket struct {
	connection Conn
	bufrw      *bufio.ReadWriter
	header     http.Header
	status     uint16
}

func (webSocket *WebSocket) write(data []byte) error {
	if _, err := webSocket.bufrw.Write(data); err != nil {
		return err
	}

	return webSocket.bufrw.Flush()
}

const MAGIC_STRING = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func (webSocket *WebSocket) Handshake() error {
	webSocketKey := webSocket.header.Get("Sec-WebSocket-Key")

	if webSocketKey == "" {
		return errors.New("non-Web Socket connection :O")
	}

	hasher := sha1.New()
	hasher.Write([]byte(webSocketKey + MAGIC_STRING))
	webSocketKeyResponseHash := base64.StdEncoding.EncodeToString(hasher.Sum(nil))

	lines := []string{
		"HTTP/1.1 101 Web Socket Protocol Handshake",
		"Server: kasza",
		"Upgrade: WebSocket",
		"Connection: Upgrade",
		"Sec-WebSocket-Accept: " + webSocketKeyResponseHash,
		"", // required for extra CRLF
		"", // required for extra CRLF
	}

	return webSocket.write([]byte(strings.Join(lines, "\r\n")))
}

func New(writer http.ResponseWriter, request *http.Request) (*WebSocket, error) {
	hijacker, ok := writer.(http.Hijacker)
	if !ok {
		return nil, errors.New("the server doesn't support HTTP hijacking :(")
	}

	connection, bufrw, err := hijacker.Hijack()
	if err != nil {
		return nil, err
	}

	return &WebSocket{connection, bufrw, request.Header, 1000}, nil
}

func (webSocket *WebSocket) Close() error {
	return webSocket.connection.Close()
}

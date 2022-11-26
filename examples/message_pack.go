package main

import (
	"bytes"
	"strings"
)

type MessagePack struct {
}

func NewMessagePack() *MessagePack {
	return &MessagePack{}
}

func (msg *MessagePack) Input(data string) int {
	index := strings.Index(data, "\\n")
	if index > 0 {
		return index + 2
	}
	return 0
}

func (msg *MessagePack) Pack(data []byte) []byte {
	var buff = bytes.Buffer{}
	buff.Write(data)
	buff.Write([]byte("\n"))
	return buff.Bytes()
}

func (msg *MessagePack) UnPack(data []byte) []byte {
	return []byte(strings.Replace(string(data), "\\n", "", 1))
}

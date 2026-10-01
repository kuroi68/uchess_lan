package uchess

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	MsgMove   = "move"
	MsgResign = "resign"
)

type Message struct {
	Type string `json:"type"`
	Move string `json:"move,omitempty"`
}

func EncodeMessage(msg Message) ([]byte, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("Error marshal message: %v", msg)
	}

	return data, nil
}

func DecodeMessage(data []byte) (Message, error) {
	msg := Message{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return Message{}, fmt.Errorf("Error marshal message: %v", msg)
	}

	switch msg.Type {
	case MsgMove:
		if msg.Move == "" {
			return Message{}, errors.New("move message without move")
		}
	case MsgResign:
	default:
		return Message{}, fmt.Errorf("unknown message type %q", msg.Type)
	}
	return msg, nil
}

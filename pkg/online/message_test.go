package online

import (
	"encoding/json"
	"testing"
)

func TestDecodeMessage(t *testing.T) {
	tests := []struct {
		Type string
		Move string
		want string
	}{
		{Type: MsgMove, Move: "hello"},
		{Type: MsgResign, Move: "bye"},
		{Type: "go", Move: "go"},
	}
	for _, tt := range tests {
		data, err := json.Marshal(tt)
		if err != nil {
			t.Fatalf("json marshal error: %v", err)
		}

		encodedMsg, err := DecodeMessage(data)
		if err != nil {
			t.Fatalf("Decode msg error: %v", err)
		}
		p := Message{}
		if encodedMsg == p {
			t.Fatal("encoded message is empty")
		}
	}
}

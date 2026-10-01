package online

import (
	"testing"
)

func TestDecodeMessage(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{name: "valid move", json: `{"type":"move","move":"e4"}`, wantErr: false},
		{name: "valid resign", json: `{"type":"resign"}`, wantErr: false},
		{name: "move without move field", json: `{"type":"move"}`, wantErr: true},
		{name: "unknown type", json: `{"type":"go"}`, wantErr: true},
		{name: "broken json", json: `{not json`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeMessage([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Fatalf("DecodeMessage(%q) error = %v, wantErr %v", tt.json, err, tt.wantErr)
			}
		})
	}
}

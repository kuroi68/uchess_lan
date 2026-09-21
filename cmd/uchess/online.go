package main

import (
	"time"

	uchess "github.com/tmountain/uchess/pkg"
	online "github.com/tmountain/uchess/pkg/online"
)

func startHost(gs *uchess.GameState) {
	uchess.DrawMsgLabel(gs.S, "Waiting for opponent...", gs.Theme)
	uchess.Render(gs)
	var name string
	go func() {
		session, err := online.Host(":8080", name)
		// Передать результат в главный UI-поток.
		if err != nil {
			uchess.DrawMsgLabel(gs.S, "\u26A0 Error: "+err.Error(), gs.Theme)
			return
		}

		_ = session
	}()
}

func connectToHost(gs *uchess.GameState) {
	uchess.DrawMsgLabel(gs.S, "Connecting...", gs.Theme)
	uchess.Render(gs)

	servers, err := online.Discover(30 * time.Second)
	if err != nil {
		uchess.DrawMsgLabel(gs.S, "\u26A0 Error: "+err.Error(), gs.Theme)
		return
	}

	if len(servers) == 0 {
		uchess.DrawMsgLabel(gs.S, "\u26A0 Error: servers not found", gs.Theme)
		return
	}

	go func() {
		session, err := online.Connect(servers[0].Port)
		// Передать результат в главный UI-поток.
		if err != nil {
			uchess.DrawMsgLabel(gs.S, "\u26A0 Error: "+err.Error(), gs.Theme)
			return
		}
		_ = session
	}()
}

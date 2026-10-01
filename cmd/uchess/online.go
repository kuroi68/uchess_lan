package main

import (
	"errors"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/corentings/chess/v2"
	"github.com/gdamore/tcell/v2"
	uchess "github.com/tmountain/uchess/pkg"
	online "github.com/tmountain/uchess/pkg/online"
)

type networkEvent struct {
	message        uchess.Message
	err            error
	connectionLost bool
}

type sessionEvent struct {
	session   *online.Session
	isHost    bool
	localName string
	peerName  string
}

func localNickname(gs *uchess.GameState, side chess.Color) string {
	name, player, engine := gs.Config.WhiteName, gs.Config.WhitePiece, gs.Config.UCIWhite
	if side == chess.Black {
		name, player, engine = gs.Config.BlackName, gs.Config.BlackPiece, gs.Config.UCIBlack
	}
	name = strings.TrimSpace(name)
	if name != "" && (player == "human" || name != filepath.Base(engine)) {
		return name
	}

	currentUser, err := user.Current()
	if err == nil && currentUser.Username != "" {
		return currentUser.Username
	}
	if side == chess.White {
		return "White player"
	}
	return "Black player"
}

func sendHello(session *online.Session, name string) error {
	data, err := uchess.EncodeMessage(uchess.Message{Type: uchess.MsgHello, Name: name})
	if err != nil {
		return err
	}
	return session.Send(data)
}

func sendMove(screen tcell.Screen, session *online.Session, move string) {
	if session == nil {
		postEvent(screen, tcell.NewEventInterrupt(errors.New("online session is not connected")))
		return
	}

	data, err := uchess.EncodeMessage(uchess.Message{
		Type: uchess.MsgMove,
		Move: move,
	})
	if err != nil {
		postEvent(screen, tcell.NewEventInterrupt(err))
		return
	}

	go func() {
		if err := session.Send(data); err != nil {
			postEvent(screen, tcell.NewEventInterrupt(networkEvent{
				err:            err,
				connectionLost: true,
			}))
		}
	}()
}

func listenNetwork(screen tcell.Screen, session *online.Session) {
	go func() {
		for {
			data, err := session.Receive()
			if err != nil {
				postEvent(screen, tcell.NewEventInterrupt(networkEvent{
					err:            err,
					connectionLost: true,
				}))
				return
			}

			message, err := uchess.DecodeMessage(data)
			postEvent(screen, tcell.NewEventInterrupt(networkEvent{
				message: message,
				err:     err,
			}))
		}
	}()
}

func startHost(gs *uchess.GameState) {
	name := localNickname(gs, chess.White)
	gs.Online.Enabled = true
	gs.Online.Connected = false
	gs.Online.IsHost = true
	gs.Online.LocalSide = chess.White
	gs.Config.WhiteName = name
	uchess.DrawMsgLabel(gs.S, "Waiting for opponent...", gs.Theme)
	uchess.Render(gs)
	go func() {
		session, err := online.Host(":8080", name)
		if err != nil {
			postEvent(gs.S, tcell.NewEventInterrupt(err))
			return
		}
		if err := sendHello(session, name); err != nil {
			_ = session.Close()
			postEvent(gs.S, tcell.NewEventInterrupt(err))
			return
		}

		postEvent(gs.S, tcell.NewEventInterrupt(sessionEvent{
			session:   session,
			isHost:    true,
			localName: name,
		}))
	}()
}

func connectToHost(gs *uchess.GameState) {
	name := localNickname(gs, chess.Black)
	gs.Online.Enabled = true
	gs.Online.Connected = false
	gs.Online.IsHost = false
	gs.Online.LocalSide = chess.Black
	gs.Config.BlackName = name
	uchess.DrawMsgLabel(gs.S, "Connecting...", gs.Theme)
	uchess.Render(gs)

	go func() {
		servers, err := online.DiscoverFirst(30 * time.Second)
		if err != nil {
			postEvent(gs.S, tcell.NewEventInterrupt(err))
			return
		}
		if len(servers) == 0 {
			postEvent(gs.S, tcell.NewEventInterrupt(errors.New("servers not found")))
			return
		}

		session, err := online.Connect(servers[0].Port)
		if err != nil {
			postEvent(gs.S, tcell.NewEventInterrupt(err))
			return
		}
		if err := sendHello(session, name); err != nil {
			_ = session.Close()
			postEvent(gs.S, tcell.NewEventInterrupt(err))
			return
		}
		postEvent(gs.S, tcell.NewEventInterrupt(sessionEvent{
			session:   session,
			localName: name,
			peerName:  servers[0].Name,
		}))
	}()
}

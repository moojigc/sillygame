package gameserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/moojigc/sillygame/move"
	"github.com/moojigc/sillygame/pkg/randwords"
	"github.com/moojigc/sillygame/pkg/sglog"
	"github.com/moojigc/sillygame/pkg/subs"
	"github.com/moojigc/sillygame/pkg/wsroom"
	"nhooyr.io/websocket"
)

// GameServer enables broadcasting to a set of rooms.
type GameServer struct {
	// publishLimiter *rate.Limiter

	// serveMux routes the various endpoints to the appropriate handler.
	serveMux http.ServeMux
	// rooms       map[string]*wsroom.Room
	defaultRoom *wsroom.Room
}

// New creates a new GameServer.
func New() *GameServer {
	gs := &GameServer{
		defaultRoom: wsroom.New(websocket.AcceptOptions{
			Subprotocols:   []string{"sillygame"},
			OriginPatterns: []string{"*"},
		}),
	}

	gs.serveMux.HandleFunc("/subscribe", gs.subscribeHandler)
	// gs.serveMux.HandleFunc("/messages", gs.getMessages)
	gs.serveMux.HandleFunc("/subscribers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		ctx := r.Context()
		logger := sglog.Logger(ctx)

		room := gs.defaultRoom
		allSubs := room.ListSubscribers()

		bytes, err := json.Marshal(allSubs)

		if err != nil {
			logger.Err(err).Msg("json.Marshal")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	})

	gs.serveMux.Handle("/", http.FileServer(http.Dir("./static")))

	return gs
}

func (gs *GameServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gs.serveMux.ServeHTTP(w, r)
}

// subscribeHandler accepts the WebSocket connection and then subscribes
// it to all future messages.
func (gs *GameServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	room := gs.defaultRoom
	c, err := websocket.Accept(w, r, &room.Opts)
	if err != nil {
		fmt.Printf("failed to accept websocket: %v\n", err)
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	s := subs.New(randwords.RandomPhrase(), room.SubscriberMessageBuffer, func() {
		c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
	})

	for _, otherSubscriber := range room.Subscribers {
		if s.Name == otherSubscriber.Name {
			s.Name = randwords.RandomPhrase()
		}
	}

	ctx, cancel := context.WithCancel(sglog.WithRqId(r.Context(), s.Name))
	defer cancel()

	logger := sglog.Logger(ctx)

	room.AddSubscriber(s)

	logger.Info().Msgf("Accepting connection from new client @ %s.\nAssigned ID %s and Name %s", r.RemoteAddr, s.ID, s.Name)

	// Send handshake message to client
	room.Respond(ctx, s, move.New(move.HANDSHAKE).SetUserID(s.ID))

	// Send player joined message to all other clients
	room.BroadcastExclusive(ctx, s, move.New(move.PLAYER_JOINED).SetUserID(s.ID))

	// Listen for messages from client and broadcast them to all other clients
	go room.ListenToSubscriber(ctx, c, s, func(msg []byte) {
		move := &move.Move{}

		if err := json.Unmarshal(msg, move); err != nil {
			logger.Err(err).Msg("json.Unmarshal")
			return
		}

		move.SetUserID(s.ID)

		logger.Debug().Msgf("Received message from subscriber %s: %v", s.Name, move)
		logger.Debug().Msgf("Broadcasting message to %d subscribers", len(room.Subscribers))
		room.BroadcastExclusive(ctx, s, move)
	})

	// Blocks until the connection is closed or the context is otherwise canceled.
	err = room.Subscribe(ctx, c, s)

	if errors.Is(err, context.Canceled) || websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		logger.Info().Msgf("Disconnected")
	} else if err != nil {
		logger.Err(err).Msg("err = room.Subscribe")
	}

	room.DeleteSubscriber(s)

	logger.Info().Msgf("Exited subscribeHandler for client")
}

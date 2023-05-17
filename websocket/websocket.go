package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/time/rate"

	"nhooyr.io/websocket"
)

// gameServer enables broadcasting to a set of subscribers.
type gameServer struct {
	// subscriberMessageBuffer controls the max number
	// of messages that can be queued for a subscriber
	// before it is kicked.
	//
	// Defaults to 16.
	subscriberMessageBuffer int

	// publishLimiter controls the rate limit applied to the publish endpoint.
	//
	// Defaults to one publish every 100ms with a burst of 8.
	publishLimiter *rate.Limiter

	// logger controls where logs are sent.
	// Defaults to log.Printf.
	logger *zerolog.Logger

	// serveMux routes the various endpoints to the appropriate handler.
	serveMux http.ServeMux

	messages      []Move
	subscribersMu sync.Mutex
	subscribers   map[string]*subscriber
}

// newChatServer constructs a chatServer with the defaults.
func NewGameServer(logger *zerolog.Logger) *gameServer {
	gs := &gameServer{
		subscriberMessageBuffer: 16,
		logger:                  logger,
		subscribers:             make(map[string]*subscriber),
		publishLimiter:          rate.NewLimiter(rate.Every(time.Millisecond*100), 8),
		messages:                []Move{},
	}
	gs.serveMux.Handle("/", http.FileServer(http.Dir("./static")))
	gs.serveMux.HandleFunc("/subscribe", gs.subscribeHandler)
	gs.serveMux.HandleFunc("/messages", gs.getMessages)
	gs.serveMux.HandleFunc("/subscribers", gs.getSubscribers)

	return gs
}

func (gs *gameServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	gs.serveMux.ServeHTTP(w, r)
}

// subscriber represents a subscriber.
// Messages are sent on the msgs channel and if the client
// cannot keep up with the messages,  closeSlow is called.
type subscriber struct {
	ID        string `json:"id"`
	msgs      chan []byte
	closeSlow func()
}

type Move struct {
	ID                string    `json:"id"`
	UserID            string    `json:"userId"`
	Event             string    `json:"event"`
	SentAt            int64     `json:"sentAT"`
	Coords            []float32 `json:"coords"`
	MouseDown         bool      `json:"mouseDown"`
	UpPing            int64     `json:"upPing"`
	ServerRespondedAt int64     `json:"serverRespondedAt"`
}

func (m *Move) setUserID(id string) {
	if m.UserID != "" {
		return
	}
	m.UserID = id
}

func (m Move) MarshalZerologObject(e *zerolog.Event) {
	e.Msgf("UserID %s; Coords %v; MsgID %s", m.UserID, m.Coords, m.ID)
}

// subscribeHandler accepts the WebSocket connection and then subscribes
// it to all future messages.
func (gs *gameServer) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*"},
		Subprotocols:   []string{"silly-game"},
	})

	if err != nil {
		gs.logger.Err(err).Msg("websocket.Accept")
		return
	}

	s := &subscriber{
		ID:   uuid.New().String(),
		msgs: make(chan []byte, gs.subscriberMessageBuffer),
		closeSlow: func() {
			c.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
		},
	}
	gs.addSubscriber(s)
	defer gs.deleteSubscriber(s)

	gs.logger.Info().Msgf("accepting connection from new client @ %s. Assigned ID %s", r.RemoteAddr, s.ID)

	ctx := r.Context()

	defer gs.sendPlayerLeaveAlert(ctx, c, s)

	gs.sendNewPlayerAlert(ctx, c, s)
	go gs.readMovesAndPublish(ctx, c, s)

	err = gs.subscribe(ctx, c, s)

	if errors.Is(err, context.Canceled) {
		return
	}
	if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
		websocket.CloseStatus(err) == websocket.StatusGoingAway {
		return
	}
	if err != nil {
		gs.logger.Err(err).Msg("err = gs.subscribe")
		return
	}
}

func (gs *gameServer) getMessages(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	messages, _ := json.Marshal(gs.messages)

	w.Write(messages)
}

func (gs *gameServer) getSubscribers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	subs, err := json.Marshal(map[string]interface{}{
		"data": gs.listSubscribers(),
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)

	w.Write(subs)
}

func (gs *gameServer) subscribe(ctx context.Context, c *websocket.Conn, s *subscriber) error {
	for {
		select {
		case msg := <-s.msgs:
			err := writeTimeout(ctx, time.Second*10, c, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (gs *gameServer) sendPlayerLeaveAlert(ctx context.Context, c *websocket.Conn, s *subscriber) error {
	leaveAlert, err := json.Marshal(Move{
		ID:     uuid.New().String(),
		UserID: s.ID,
		Event:  PLAYER_LEFT,
		SentAt: time.Now().UnixMilli(),
	})

	if err != nil {
		return err
	}

	for sid := range gs.subscribers {
		s, err := gs.getSubscriber(sid)
		if err != nil {
			continue
		}
		select {
		case s.msgs <- leaveAlert:
		default:
			go s.closeSlow()
		}
	}

	return nil
}

func (gs *gameServer) sendNewPlayerAlert(ctx context.Context, c *websocket.Conn, s *subscriber) error {

	handshake, err := json.Marshal(Move{
		ID:     uuid.New().String(),
		UserID: s.ID,
		Event:  HANDSHAKE,
		SentAt: time.Now().UnixMilli(),
	})

	if err != nil {
		return err
	}

	s.msgs <- handshake

	playerJoined, err := json.Marshal(Move{
		ID:     uuid.New().String(),
		UserID: s.ID,
		Event:  PLAYER_JOINED,
		SentAt: time.Now().UnixMilli(),
	})

	if err != nil {
		return err
	}

	for sid := range gs.subscribers {
		s, err := gs.getSubscriber(sid)
		if err != nil {
			return err
		} else if sid == s.ID {
			continue
		}
		select {
		case s.msgs <- playerJoined:
			gs.logger.Debug().Msgf("sending msg to %s\n", s.ID)
		default:
			go s.closeSlow()
		}
	}

	return nil
}

func (gs *gameServer) readMovesAndPublish(ctx context.Context, c *websocket.Conn, s *subscriber) {
	defer gs.deleteSubscriberUnsafe(s)
	for {
		_, reader, err := c.Reader(ctx)

		if err != nil {
			gs.logger.Err(err).Msg("c.Reader")
			c.Close(websocket.CloseStatus(err), err.Error())
			return
		}

		b, err := io.ReadAll(reader)

		if err != nil {
			gs.logger.Err(err).Msg("io.ReadAll")
			continue
		}

		move := &Move{}

		if err := json.Unmarshal(b, move); err != nil {
			gs.logger.Err(err).Msg("json.Marshal")
			continue
		}

		move.ServerRespondedAt = time.Now().UnixMilli()
		move.UpPing = time.Now().UnixMilli() - move.SentAt
		move.setUserID(s.ID)

		gs.logger.Info().Msgf("[subscriber %s] PUBLISHING %v;\n\tMouseDown %v, Upload ping %d", s.ID, move.Coords, move.MouseDown, move.UpPing)
		bytes, err := json.Marshal(move)

		if err != nil {
			gs.logger.Err(err).Msg("json.Marshal Move object")
			return
		}

		gs.subscribersMu.Lock()

		gs.logger.Info().Msgf("Sending to %d subscribers", len(gs.subscribers))

		for sid := range gs.subscribers {
			s, _ := gs.getSubscriber(sid)
			select {
			case s.msgs <- bytes:
			default:
				go s.closeSlow()
			}
		}

		gs.subscribersMu.Unlock()
	}
}

func (gs *gameServer) listSubscribers() []*subscriber {
	gs.subscribersMu.Lock()
	defer gs.subscribersMu.Unlock()

	var subscriberSlice []*subscriber

	for sid := range gs.subscribers {
		s, _ := gs.getSubscriber(sid)
		subscriberSlice = append(subscriberSlice, s)
	}

	return subscriberSlice
}

func (gs *gameServer) getSubscriber(id string) (*subscriber, error) {
	subscriber, ok := gs.subscribers[id]

	if !ok {
		return nil, fmt.Errorf("subscriber %s not found", id)
	}

	return subscriber, nil
}

// addSubscriber registers a subscriber.
func (gs *gameServer) addSubscriber(s *subscriber) {
	gs.subscribersMu.Lock()
	defer gs.subscribersMu.Unlock()

	gs.subscribers[s.ID] = s
}

// deleteSubscriber deletes the given subscriber.
func (gs *gameServer) deleteSubscriber(s *subscriber) {
	gs.subscribersMu.Lock()
	defer gs.subscribersMu.Unlock()
	gs.deleteSubscriber(s)
}

// deleteSubscriber deletes the given subscriber.
func (gs *gameServer) deleteSubscriberUnsafe(s *subscriber) {
	delete(gs.subscribers, s.ID)
	gs.logger.Info().Msgf("Goodbye subscriber %s", s.ID)
}

// func (gs *gameServer) countSubscriber() int {
// 	gs.subscribersMu.Lock()
// 	defer gs.subscribersMu.Unlock()

// 	return len(gs.subscribers)
// }

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return c.Write(ctx, websocket.MessageText, msg)
}

package wsroom

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/moojigc/sillygame/pkg/randwords"
	"github.com/moojigc/sillygame/pkg/sglog"
	"github.com/moojigc/sillygame/pkg/subs"

	"nhooyr.io/websocket"
)

type Room struct {
	ID                      string
	Name                    string
	Subscribers             map[string]*subs.Subscriber
	SubscribersMu           sync.Mutex
	SubscriberMessageBuffer int
	Opts                    websocket.AcceptOptions
}

func New(opts websocket.AcceptOptions) *Room {
	return &Room{
		ID:                      uuid.New().String(),
		Name:                    fmt.Sprintf("Room %s", randwords.RandomPhrase()),
		Subscribers:             make(map[string]*subs.Subscriber),
		SubscriberMessageBuffer: 16,
		Opts:                    opts,
	}
}

// Subscribe sends all messages to the subscriber and blocks until the context is done.
func (room *Room) Subscribe(ctx context.Context, c *websocket.Conn, s *subs.Subscriber) error {
	logger := sglog.Logger(ctx)
	logger.Info().Msgf("[Subscribe] %s to room %s", s.Name, room.Name)

	for {
		select {
		case msg := <-s.Msgs:
			err := writeTimeout(ctx, time.Second*10, c, msg)
			if err != nil {
				logger.Err(err).Msg("writeTimeout")
				return err
			}
		case <-ctx.Done():
			logger.Warn().Msgf("context done for %s", s.Name)
			return ctx.Err()
		}
	}
}

// Broadcast sends a message to all subscribers in the room.
func (room *Room) Broadcast(ctx context.Context, msg any) {
	logger := sglog.Logger(ctx)

	bytes, err := json.Marshal(msg)

	if err != nil {
		logger.Err(err).Msg("json.Marshal")
		return
	}

	room.SubscribersMu.Lock()
	defer room.SubscribersMu.Unlock()

	for _, s := range room.Subscribers {
		select {
		case s.Msgs <- bytes:
		default:
			go s.CloseSlow()
		}
	}

}

// Broadcast sends a message to all subscribers in the room except the sender.
func (room *Room) BroadcastExclusive(ctx context.Context, s *subs.Subscriber, msg any) {
	logger := sglog.Logger(ctx)

	bytes, err := json.Marshal(msg)

	if err != nil {
		logger.Err(err).Msg("json.Marshal")
		return
	}

	logger.Debug().Msgf("Broadcasting to %d subscribers", len(room.Subscribers)-1)

	room.SubscribersMu.Lock()
	defer room.SubscribersMu.Unlock()

	for _, otherSubscriber := range room.Subscribers {
		if s.ID == otherSubscriber.ID {
			continue
		}

		select {
		case otherSubscriber.Msgs <- bytes:
		default:
			go otherSubscriber.CloseSlow()
		}
	}

}

// Respond sends a JSON message to the subscriber.
func (room *Room) Respond(ctx context.Context, s *subs.Subscriber, msg any) {
	logger := sglog.Logger(ctx)
	logger.Info().Msgf("Sending response %v", msg)

	bytes, err := json.Marshal(msg)
	if err != nil {
		logger.Err(err).Msg("json.Marshal")
		return
	}

	s.Msgs <- bytes
}

// ListenToSubscriber listens to a subscriber and calls the onMessage callback
func (room *Room) ListenToSubscriber(ctx context.Context, c *websocket.Conn, s *subs.Subscriber, onMessage func(bytes []byte)) {
	logger := sglog.Logger(ctx)
	logger.Info().Msgf("[ListenToSubscriber] %s to room %s", s.Name, room.Name)
	for {
		_, reader, err := c.Reader(ctx)

		if err != nil {
			logger.Err(err).Msg("c.Reader")
			if websocket.CloseStatus(err) == websocket.StatusGoingAway {
				c.Close(websocket.StatusGoingAway, err.Error())
			}
			break
		}

		bytesRead, err := io.ReadAll(reader)

		logger.Debug().Msgf("Read %d bytes", len(bytesRead))

		if err != nil {
			logger.Err(err).Msg("c.Reader")
			continue
		}

		onMessage(bytesRead)
	}

	logger.Debug().Msg("Exited ListenToSubscriber")
}

func (room *Room) ListSubscribers() []subs.Subscriber {
	room.SubscribersMu.Lock()
	defer room.SubscribersMu.Unlock()

	var subscriberSlice []subs.Subscriber

	for sid := range room.Subscribers {
		s, _ := room.getSubscriber(sid)
		subscriberSlice = append(subscriberSlice, *s)
	}

	return subscriberSlice
}

func (room *Room) getSubscriber(id string) (*subs.Subscriber, error) {
	subscriber, ok := room.Subscribers[id]

	if !ok {
		return nil, fmt.Errorf("subscriber %s not found", id)
	}

	return subscriber, nil
}

// AddSubscriber registers a subscriber.
func (room *Room) AddSubscriber(s *subs.Subscriber) {
	room.SubscribersMu.Lock()
	defer room.SubscribersMu.Unlock()

	room.Subscribers[s.ID] = s
}

// DeleteSubscriber deletes the given subscriber.
func (room *Room) DeleteSubscriber(s *subs.Subscriber) {
	room.SubscribersMu.Lock()
	defer room.SubscribersMu.Unlock()
	room.DeleteSubscriberUnsafe(s)
}

func (room *Room) DeleteSubscriberUnsafe(s *subs.Subscriber) {
	delete(room.Subscribers, s.ID)
}

func writeTimeout(ctx context.Context, timeout time.Duration, c *websocket.Conn, msg []byte) error {
	logger := sglog.Logger(ctx)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	logger.Debug().Msgf("writeTimeout: writing msg of length %d", len(msg))

	return c.Write(ctx, websocket.MessageText, msg)
}

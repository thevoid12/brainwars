package websocket

import (
	quizmodel "brainwars/pkg/quiz/model"
	"context"
	"encoding/json"
	"time"
)

// Event is the Messages sent over the websocket
// Used to differ between different actions
type Event struct {
	// Type is the message type sent
	Type string `json:"type"`
	// Payload is the data Based on the Type
	Payload json.RawMessage `json:"payload"`
}

// EventHandler is a function signature that is used to affect messages on the socket and triggered
// depending on the type
type EventHandler func(ctx context.Context, event Event, c *Client) error

const (
	// EventSendMessage is the event name for new chat messages sent
	EventSendMessage = "send_message"
	// EventNewMessage is a response to send_message
	EventNewMessage = "new_message"
	// EventChangeRoom is event when switching rooms
	EventChangeRoom = "change_room"

	EventStartGame = "start_game"
	//EventReadyGame is that the user is ready to start the game
	EventReadyGame    = "ready_game"
	EventEndGame      = "end_game"
	EventGameStatus   = "game_status"
	EventSubmitAnswer = "submit_answer"
)

type Payload struct {
	Data string    `json:"data"`
	Time time.Time `json:"time"`
}

var (
	// pongWait is how long we will await a pong response from client
	pongWait = 10 * time.Second
	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead *9 / 10 to get 90%
	// The reason why it has to be less than PingRequency is becuase otherwise it will send a new Ping before getting response
	pingInterval = (pongWait * 9) / 10
)

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessageHandler
	m.handlers[EventReadyGame] = ReadyGameMessageHandler
	m.handlers[EventStartGame] = StartGameMessageHandler
	m.handlers[EventSubmitAnswer] = SubmitAnswerHandler
}

func (m *Manager) routeEvent(ctx context.Context, event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
		return handler(ctx, event, c)
	}
	return ErrEventNotSupported
}

// Add a GameState map to Manager to track games in different rooms
func (m *Manager) setupGameState() {
	m.gameStates = make(map[string]*quizmodel.GameState)
}

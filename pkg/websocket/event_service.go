package websocket

import (
	logs "brainwars/pkg/logger"
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
	EventLobbyState  = "lobby_state"
	EventStartGame   = "start_game" // (step3)
	//EventReadyGame is that the user is ready to start the game
	EventJoinedGame   = "joined_game" // Joined into the game (step1)
	EventReadyGame    = "ready_game"  // ready to play the game (step2)
	EventLeaveRoom    = "leave_room"  // leave game room
	EventEndGame      = "end_game"
	EventBotGameOver  = "bot_game_over" // notifies the bot that the game is over so that it can stop
	EventGameStatus   = "game_status"
	EventSubmitAnswer = "submit_answer"
	EventNewQuestion  = "new_question"
	EventNextQuestion = "next_question" // user clicks next question
	EventGameError    = "game_error"
	EventLeaderBoard  = "leaderboard"
)

type Payload struct {
	UserName string    `json:"username"`
	Data     string    `json:"data"`
	Time     time.Time `json:"time"`
}

var (
	// pongWait is how long we will await a pong response from client
	pongWait = 45 * time.Second
	// pingInterval has to be less than pongWait, We cant multiply by 0.9 to get 90% of time
	// Because that can make decimals, so instead *9 / 10 to get 90%
	// The reason why it has to be less than PingRequency is becuase otherwise it will send a new Ping before getting response
	pingInterval = (pongWait * 9) / 10
	maxReadLimit = 1024 * 1024
)

func (m *Manager) setupEventHandlers() {
	m.handlers[EventSendMessage] = SendMessageHandler
	m.handlers[EventReadyGame] = ReadyGameMessageHandler
	m.handlers[EventStartGame] = StartGameMessageHandler
	m.handlers[EventSubmitAnswer] = SubmitAnswerHandler
	m.handlers[EventNextQuestion] = NextQuestionHandler
	m.handlers[EventLeaveRoom] = LeaveGameRoomHandler
}

func (m *Manager) routeEvent(ctx context.Context, event Event, c *Client) error {
	l := logs.GetLoggerctx(ctx)
	l.Sugar().Infof("Routing event type %s from user %s in room %s",
		event.Type, c.userID, c.roomCode)

	if handler, ok := m.handlers[event.Type]; ok {
		c.manager = m
		return handler(ctx, event, c)
	}
	return ErrEventNotSupported
}

type questionEvent struct {
	QuestionIndex  int                     `json:"questionIndex"`
	TotalQuestions int                     `json:"totalQuestions"`
	Question       *quizmodel.QuestionData `json:"question"`
	StartTime      time.Time               `json:"startTime"`
	TimeLimit      int                     `json:"timeLimit"`
}

package websocket

import (
	logs "brainwars/pkg/logger"
	"brainwars/pkg/quiz"
	quizmodel "brainwars/pkg/quiz/model"
	"brainwars/pkg/room"
	roommodel "brainwars/pkg/room/model"
	usermodel "brainwars/pkg/users/model"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Set up bots to be ready when a human player joins
// bots doesnt work with egres channel. egress channel is for web socket connection
// bots can use the other channel to coordinate communication
func (m *Manager) setupBotsForRoom(ctx context.Context, roomCode string, roomDetails *roommodel.Room) {
	l := logs.GetLoggerctx(ctx)

	// Get all room members including bots
	roomMembers, err := room.ListRoomMembersByRoomCode(ctx, roommodel.RoomCodeReq{
		RoomCode: roomCode,
	})
	if err != nil || len(roomMembers) == 0 {
		l.Sugar().Error("List Room member by room id failed", err)
		return
	}

	// Set all bots to ready state
	for _, member := range roomMembers {
		if member.IsBot && member.RoomMemberStatus == roommodel.ReadyQuiz {
			// Determine bot type based on member properties or some naming convention
			// For example, if bot names contain their type like "Bot-30sec", "Bot-1min", etc.
			var botType usermodel.BotType
			if strings.Contains(member.UserDetails.UserName, "Sec10") {
				botType = usermodel.Sec10
			} else if strings.Contains(member.UserDetails.UserName, "Sec15") {
				botType = usermodel.Sec15
			} else if strings.Contains(member.UserDetails.UserName, "Sec20") {
				botType = usermodel.Sec20
			} else if strings.Contains(member.UserDetails.UserName, "Sec30") {
				botType = usermodel.Sec30
			} else if strings.Contains(member.UserDetails.UserName, "Sec45") {
				botType = usermodel.Sec45
			} else if strings.Contains(member.UserDetails.UserName, "Sec1") {
				botType = usermodel.Sec1
			} else if strings.Contains(member.UserDetails.UserName, "Sec2") {
				botType = usermodel.Sec2
			} else {
				// Default bot type
				botType = usermodel.Sec30
			}

			// Create a new bot client
			botClient := NewClient(nil, m, roomCode, true, botType, member.UserID, roomDetails)

			// Initialize the bot with event channel and start its behavior handler
			m.InitializeBot(ctx, botClient)

			// Add the client to the manager
			m.addClient(botClient)

			// update the bot status to BOT READY  quiz so that we dont set up client again in multiplayer setup
			err = room.UpdateRoomMemberStatusByRoomCodeAndUserID(ctx, &roommodel.RoomCodeReq{
				UserID:   member.UserID,
				RoomCode: roomCode,
			}, roommodel.BotReadyQuiz)
			if err != nil {
				return
			}
			// Notify all clients that this bot is ready
			botReadyNotification := Payload{
				Data: fmt.Sprintf("Bot %s is ready", member.UserDetails.UserName),
				Time: time.Now(),
			}

			data, _ := json.Marshal(botReadyNotification)
			readyEvent := Event{Type: "ready_game", Payload: data}

			// Broadcast to all clients in the room
			for client := range m.clients[roomCode] {
				if client.isBot {
					client.botEvents <- readyEvent
					continue
				}
				client.egress <- readyEvent
			}

			l.Sugar().Infof("Bot %s (type: %s) added to room %s", member.UserID.String(), botType, roomCode)
		}
	}
}

// InitializeBot should be called when a new bot client is created
func (m *Manager) InitializeBot(ctx context.Context, client *Client) {
	// Create a buffered channel for bot events
	client.botEvents = make(chan Event)

	// Start the bot behavior handler
	go client.handleBotBehavior(ctx)
}

// Method to broadcast events to all bot clients in a room
func (m *Manager) broadcastToBots(ctx context.Context, roomCode string, event Event) {
	l := logs.GetLoggerctx(ctx)
	m.RLock()
	defer m.RUnlock()

	if clients, exists := m.clients[roomCode]; exists {
		for client := range clients {
			if client.isBot && client.botEvents != nil {
				// Send the event to the bot's event channel
				select {
				case client.botEvents <- event:
					// Event sent successfully from egress event to bot event
				default:
					// Channel is full or closed, log error
					l.Sugar().Error(fmt.Sprintf("Failed to send event to bot botID: %s roomcode: %s", client.userID, roomCode))
				}
			}
		}
	}
}

func (c *Client) handleBotBehavior(ctx context.Context) {
	l := logs.GetLoggerctx(ctx)

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, clean up and exit
			return

		case event, ok := <-c.botEvents:
			if !ok {
				// Channel closed, exit
				return
			}

			switch event.Type {
			case EventNewQuestion:
				// Parse the question data
				questionEvent := questionEvent{}
				if err := json.Unmarshal(event.Payload, &questionEvent); err != nil {
					l.Sugar().Error("Failed to parse question event", "error", err)
					continue
				}

				// Calculate answer delay based on bot type
				delay := time.Duration(usermodel.BotTypeMap[c.botType])
				l.Sugar().Debugf("Resolved botType %s to delay %v", c.botType, delay)
				// Ensure delay does not exceed the time limit for the question
				maxDelay := time.Duration(questionEvent.TimeLimit) * time.Minute
				if delay > maxDelay {
					delay = maxDelay
				}

				// Spawn a new goroutine for delayed answer submission
				go func(qID uuid.UUID, d time.Duration, botID uuid.UUID) {
					l.Sugar().Debugf("Bot %v will answer question %v in %v", botID, qID, d)

					select {
					case <-ctx.Done():
						return
					case <-time.After(d):
						c.submitRandomAnswer(ctx, qID)
					}
				}(questionEvent.Question.ID, delay, c.userID)

			case EventReadyGame:
				l.Sugar().Debugf("Bot %s is ready to play", c.userID)

			case EventEndGame:
				return // Exit the bot behavior goroutine when game ends
			}
		}
	}
}

func (c *Client) submitRandomAnswer(ctx context.Context, questionID uuid.UUID) {
	l := logs.GetLoggerctx(ctx)

	// Get the game state to find the question
	c.manager.RLock()
	gameState, exists := c.manager.gameStates[c.roomCode]
	if !exists {
		c.manager.RUnlock()
		return
	}

	// Find the current question
	var currentQuestion *quizmodel.QuestionData
	found := false
	for _, q := range gameState.Questions.QuestionData {
		if q.ID == questionID {
			currentQuestion = q
			found = true
			break
		}
	}
	c.manager.RUnlock()

	if !found || len(currentQuestion.Options) == 0 {
		return
	}

	// Select a random option
	randomIndex := rand.Intn(len(currentQuestion.Options))
	selectedOption := currentQuestion.Options[randomIndex]

	// Create and send the answer event
	answerPayload := struct {
		QuestionDataID uuid.UUID `json:"questionDataID"`
		AnswerOption   int       `json:"answerOption"`
		PlayerID       uuid.UUID `json:"playerID"`
	}{
		QuestionDataID: currentQuestion.ID,
		AnswerOption:   selectedOption.ID,
		PlayerID:       c.userID,
	}

	answerData, _ := json.Marshal(answerPayload)
	answerEvent := Event{Type: EventSubmitAnswer, Payload: answerData}

	// Send the answer through the client's egress channel
	select {
	case <-ctx.Done():
		// Context canceled, no need to send answer
	default:
		// Answer sent successfully
		for client := range c.manager.clients[c.roomCode] {
			if client.isBot {
				// TODO: WE NEED to write the result in db
				err := quiz.CreateAnswer(ctx, quizmodel.AnswerReq{
					RoomCode:       c.roomCode,
					UserID:         c.userID,
					QuestionID:     questionID,
					QuestionDataID: currentQuestion.ID,
					AnswerOption:   int32(selectedOption.ID),
					IsCorrect:      currentQuestion.Answer == selectedOption.ID,
					AnswerTime:     time.Now(),
					CreatedBy:      c.userID.String(),
				})
				if err != nil {
					l.Sugar().Error("submiting answer failed ", err)
				}
				fmt.Println()
				continue
			}
			client.egress <- answerEvent
		}

		// Channel is full or closed
		// logs.GetLoggerctx(ctx).Sugar().Error("Failed to send answer userID:", c.userID)
	}
}

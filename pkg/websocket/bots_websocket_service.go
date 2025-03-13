package websocket

import (
	logs "brainwars/pkg/logger"
	"brainwars/pkg/room"
	roommodel "brainwars/pkg/room/model"
	usermodel "brainwars/pkg/users/model"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Set up bots to be ready when a human player joins
func (m *Manager) setupBotsForRoom(ctx context.Context, roomCode string) {
	l := logs.GetLoggerctx(ctx)

	// Get all room members including bots
	roomMembers, err := room.ListRoomMembersByRoomID(ctx, roommodel.RoomIDReq{
		RoomID: uuid.MustParse(roomCode),
	})
	if err != nil {
		l.Sugar().Error("List Room member by room id failed", err)
		return
	}

	// Set all bots to ready state
	for _, member := range roomMembers {
		if member.IsBot {
			// Determine bot type based on member properties or some naming convention
			// For example, if bot names contain their type like "Bot-30sec", "Bot-1min", etc.
			var botType usermodel.BotType
			if strings.Contains(strings.ToLower(member.UserDetails.UserName), "Sec10") {
				botType = usermodel.Sec10
			} else if strings.Contains(strings.ToLower(member.UserDetails.UserName), "Sec15") {
				botType = usermodel.Sec15
			} else if strings.Contains(strings.ToLower(member.UserDetails.UserName), "Sec20") {
				botType = usermodel.Sec20
			} else if strings.Contains(strings.ToLower(member.UserDetails.UserName), "Sec30") {
				botType = usermodel.Sec30
			} else if strings.Contains(strings.ToLower(member.UserDetails.UserName), "Sec45") {
				botType = usermodel.Sec45
			} else if strings.Contains(strings.ToLower(member.UserDetails.UserName), "Sec1") {
				botType = usermodel.Sec1
			} else if strings.Contains(strings.ToLower(member.UserDetails.UserName), "Sec2") {
				botType = usermodel.Sec2
			} else {
				// Default bot type
				botType = usermodel.Sec30
			}

			// Create a new bot client
			botClient := NewClient(nil, m, roomCode, true, botType, member.ID)

			// Initialize the bot with event channel and start its behavior handler
			m.InitializeBot(ctx, botClient)

			// Add the client to the manager
			m.addClient(botClient)

			// Notify all clients that this bot is ready
			botReadyNotification := Payload{
				Data: fmt.Sprintf("Bot %s is ready", member.UserDetails.UserName),
				Time: time.Now(),
			}

			data, _ := json.Marshal(botReadyNotification)
			readyEvent := Event{Type: "game_status", Payload: data}

			// Broadcast to all clients in the room
			for client := range m.clients[roomCode] {
				client.egress <- readyEvent
			}

			l.Sugar().Infof("Bot %s (type: %s) added to room %s", member.ID.String(), botType, roomCode)
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

// Updated to work with the new botEvents channel
func (m *Manager) registerRoomEventListener(roomCode string, eventChan chan Event) {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.clients[roomCode]; exists {
		for client := range m.clients[roomCode] {
			if client.isBot && client.botEvents != nil {
				// Create a copy of the bot pointer for goroutine
				botClient := client

				// Start goroutine to forward events
				go func() {
					log.Printf("Registered event listener for bot %s in room %s",
						botClient.userID, roomCode)

					// Forward events from the event channel to the bot
					for event := range eventChan {
						select {
						case botClient.botEvents <- event:
							// Successfully forwarded event to bot
						default:
							// Bot event channel is full or closed
							log.Printf("Failed to send event to bot %s: channel full or closed",
								botClient.userID)
						}
					}
				}()
			}
		}
	}
}

func (m *Manager) unregisterRoomEventListener(roomCode string, eventChan chan Event) {
	// Close the event channel, which will cause all goroutines reading from it to exit
	close(eventChan)
	log.Printf("Unregistered event listener for room %s", roomCode)
}

// Method to broadcast events to all bot clients in a room
func (m *Manager) broadcastToBots(ctx context.Context, roomCode string, event Event) {
	m.RLock()
	defer m.RUnlock()

	if clients, exists := m.clients[roomCode]; exists {
		for client := range clients {
			if client.isBot && client.botEvents != nil {
				// Send the event to the bot's event channel
				select {
				case client.botEvents <- event:
					// Event sent successfully
				default:
					// Channel is full or closed, log error
					logs.GetLoggerctx(ctx).Error("Failed to send event to bot",
						"botID", client.id, "roomCode", roomCode)
				}
			}
		}
	}
}

func (c *Client) handleBotBehavior(ctx context.Context) {
	var answerTimer *time.Timer

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, clean up
			if answerTimer != nil {
				answerTimer.Stop()
			}
			return
		case event, ok := <-c.botEvents:
			if !ok {
				// Channel closed, exit
				if answerTimer != nil {
					answerTimer.Stop()
				}
				return
			}

			switch event.Type {
			case EventNewQuestion:
				if answerTimer != nil {
					answerTimer.Stop()
				}

				// Parse the question data
				var questionEvent struct {
					QuestionIndex  int       `json:"questionIndex"`
					TotalQuestions int       `json:"totalQuestions"`
					Question       Question  `json:"question"`
					StartTime      time.Time `json:"startTime"`
				}

				if err := json.Unmarshal(event.Payload, &questionEvent); err != nil {
					logs.GetLoggerctx(ctx).Error("Failed to parse question event", "error", err)
					continue
				}

				// Calculate answer delay based on bot type
				var delay time.Duration
				switch c.botType {
				case "30sec":
					delay = time.Duration(rand.Intn(25)+5) * time.Second
				case "1min":
					delay = time.Duration(rand.Intn(30)+30) * time.Second
				case "2min":
					delay = time.Duration(rand.Intn(60)+60) * time.Second
				default:
					delay = time.Duration(rand.Intn(30)+5) * time.Second
				}

				// Make sure the delay doesn't exceed the question time limit
				if int(delay.Seconds()) > questionEvent.Question.TimeLimit {
					delay = time.Duration(questionEvent.Question.TimeLimit-1) * time.Second
				}

				// Create a separate context for the answer submission
				answerCtx, cancel := context.WithCancel(ctx)

				// Schedule the answer submission
				answerTimer = time.AfterFunc(delay, func() {
					defer cancel() // Clean up the context when done
					c.submitRandomAnswer(answerCtx, questionEvent.Question.ID)
				})

			case EventEndGame:
				if answerTimer != nil {
					answerTimer.Stop()
				}
				return // Exit the bot behavior goroutine when game ends
			}
		}
	}
}

func (c *Client) submitRandomAnswer(ctx context.Context, questionID string) {
	// Get the game state to find the question
	c.manager.RLock()
	gameState, exists := c.manager.gameStates[c.roomCode]
	if !exists {
		c.manager.RUnlock()
		return
	}

	// Find the current question
	var currentQuestion Question
	found := false
	for _, q := range gameState.Questions {
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
		QuestionID string `json:"questionID"`
		AnswerID   string `json:"answerID"`
		PlayerID   string `json:"playerID"`
	}{
		QuestionID: questionID,
		AnswerID:   selectedOption.ID,
		PlayerID:   c.id,
	}

	answerData, _ := json.Marshal(answerPayload)
	answerEvent := Event{Type: EventSubmitAnswer, Payload: answerData}

	// Send the answer through the client's egress channel
	select {
	case c.egress <- answerEvent:
		// Answer sent successfully
	case <-ctx.Done():
		// Context canceled, no need to send answer
	default:
		// Channel is full or closed
		logs.GetLoggerctx(ctx).Error("Failed to send bot answer", "botID", c.id)
	}
}

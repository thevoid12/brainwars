package websocket

import (
	logs "brainwars/pkg/logger"
	"brainwars/pkg/room"
	roommodel "brainwars/pkg/room/model"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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
			err := room.UpdateRoomMemberByID(ctx, roommodel.RoomMemberReq{
				ID:               member.ID,
				UserID:           member.UserID,
				RoomID:           uuid.MustParse(roomCode),
				RoomMemberStatus: roommodel.ReadyQuiz,
			})
			if err != nil {
				l.Sugar().Error("Failed to update bot ready status:", err)
				continue
			}

			// Notify all clients that this bot is ready
			botReadyNotification := Payload{
				Data: fmt.Sprintf("Bot %s is ready", member.ID.String()),
				Time: time.Now(),
			}

			data, _ := json.Marshal(botReadyNotification)
			readyEvent := Event{Type: "game_status", Payload: data}

			// Broadcast to all clients in the room
			for client := range m.clients[roomCode] {
				client.egress <- readyEvent
			}
		}
	}
}

func NewBotClient(ctx context.Context, manager *Manager, roomCode string, botType string) (*Client, error) {
	botUserID := uuid.New()
	// roomID, err := uuid.Parse(roomCode)
	// if err != nil {
	// 	return nil, fmt.Errorf("invalid room code: %w", err)
	// }

	botClient := NewClient(nil, manager, roomCode, true, botType, botUserID)
	// Initialize bot event channel
	botClient.botEvents = make(chan Event)

	go botClient.handleBotBehavior(ctx)

	// Immediately mark the bot as ready
	// readyEvent := Event{
	// 	Type: EventReadyGame,
	// 	Payload: []byte(fmt.Sprintf(`{"data":"Bot %s is ready","time":"%s"}`,
	// 		botUserID.String(), time.Now().Format(time.RFC3339))),
	// }

	// Use a separate goroutine to avoid blocking
	// go func() {
	// 	if err := manager.routeEvent(ctx, readyEvent, botClient); err != nil {
	// 		log.Printf("Error marking bot as ready: %v", err)
	// 	}
	// }()

	return botClient, nil
}

func (c *Client) handleBotBehavior(ctx context.Context) {
	if !c.isBot || c.botEvents == nil {
		return
	}

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
			case "new_question":
				if answerTimer != nil {
					answerTimer.Stop()
				}

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

				// Create a new context for the answer timer
				answerCtx, cancel := context.WithTimeout(ctx, delay+time.Second)

				answerTimer = time.AfterFunc(delay, func() {
					c.submitRandomAnswer(answerCtx)
					cancel() // Clean up the context when done
				})

			case "game_end":
				if answerTimer != nil {
					answerTimer.Stop()
				}
				return // Exit the bot behavior goroutine when game ends
			}
		}
	}
}

func (c *Client) submitRandomAnswer(ctx context.Context) {
	randomAnswer := rand.Intn(4) // Assuming 4 answer options

	answerPayload := struct {
		Answer int       `json:"answer"`
		UserID uuid.UUID `json:"userID"`
	}{
		Answer: randomAnswer,
		UserID: c.userID,
	}

	payloadBytes, err := json.Marshal(answerPayload)
	if err != nil {
		log.Printf("Error marshalling bot answer: %v", err)
		return
	}

	answerEvent := Event{
		Type:    "submit_answer",
		Payload: payloadBytes,
	}

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		select {
		case <-ctx.Done():
			// Context cancelled or timed out
			return
		default:
			err := c.manager.routeEvent(ctx, answerEvent, c)
			if err == nil {
				log.Printf("Bot %s submitted answer: %d", c.userID, randomAnswer)
				return // Success
			}

			log.Printf("Error submitting bot answer (attempt %d/%d): %v",
				i+1, maxRetries, err)

			if i < maxRetries-1 {
				// Wait before retry
				time.Sleep(time.Millisecond * 500)
			}
		}
	}
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
				select {
				case client.botEvents <- event:
					// Successfully sent event to bot
				default:
					// Bot event channel is full or closed
					log.Printf("Failed to broadcast event to bot %s: channel full or closed",
						client.userID)
				}
			}
		}
	}
}

// // Helper method to add to Manager to check if all clients in a room are ready
// func (m *Manager) areAllClientsReady(ctx context.Context, roomCode string) (bool, error) {
// 	roomMembers, err := room.ListRoomMembersByRoomCode(ctx, roomCode)
// 	if err != nil {
// 		return false, fmt.Errorf("failed to list room members: %w", err)
// 	}

// 	for _, member := range roomMembers {
// 		if (!member.IsBot) && member.RoomMemberStatus != roommodel.ReadyQuiz {
// 			return false, nil
// 		}
// 	}

// 	return true, nil
// }

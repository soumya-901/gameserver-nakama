package matching

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

type TicTacToeMatch struct {
	ID      string
	Players map[string]any
	Symbols map[string]string
}

func (m *TicTacToeMatch) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	logger.Info("TicTacToe match initialized.")
	return nil, 2, ""
}
func (m *TicTacToeMatch) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, signal string) (interface{}, string) {
	logger.Info("Player attempting to join match.", signal)
	return state, ""
}
func (m *TicTacToeMatch) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	logger.Info("Player attempting to join match.")
	return state, true, ""
}

func (m *TicTacToeMatch) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	logger.Info("Player joined match.")

	for _, p := range presences {
		logger.Info("Player joined: %s", p.GetUsername())

		// Notify all other players
		joinNotice := map[string]interface{}{
			"type":     "player_joined",
			"username": p.GetUsername(),
			"id":       p.GetUserId(),
		}

		msg, _ := json.Marshal(joinNotice)
		dispatcher.BroadcastMessage(1, msg, nil, nil, true)
	}

	// Add new players to our match state
	for _, p := range presences {
		m.Players[p.GetUserId()] = p
	}

	return state
}

func (m *TicTacToeMatch) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	logger.Info("Player left match.")
	for _, p := range presences {
		logger.Info("Player left: %s", p.GetUsername())

		leaveNotice := map[string]interface{}{
			"type":     "player_left",
			"username": p.GetUsername(),
			"id":       p.GetUserId(),
		}
		msg, _ := json.Marshal(leaveNotice)
		dispatcher.BroadcastMessage(1, msg, nil, nil, true)

		delete(m.Players, p.GetUserId())
	}
	return state
}

func (m *TicTacToeMatch) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {

	for _, msg := range messages {
		var payload map[string]interface{}
		if err := json.Unmarshal(msg.GetData(), &payload); err != nil {
			logger.Error("Failed to parse message: %v", err)
			continue
		}

		switch msg.GetOpCode() {

		// 2 = player chose X or O
		case 2:
			symbol := payload["symbol"].(string)
			playerID := msg.GetUserId()
			m.Symbols[playerID] = symbol

			logger.Info("Player %s chose symbol %s", playerID, symbol)

			broadcast := map[string]interface{}{
				"type":     "symbol_selected",
				"user_id":  playerID,
				"symbol":   symbol,
				"username": m.Players[playerID].(runtime.Presence).GetUsername(),
			}
			data, _ := json.Marshal(broadcast)
			dispatcher.BroadcastMessage(2, data, nil, nil, true)
		}
	}
	return state
}

func (m *TicTacToeMatch) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, newti int) interface{} {
	logger.Info("Match terminated.", newti)
	return state
}

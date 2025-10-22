package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/heroiclabs/nakama-common/runtime"
)

var (
	matchStore = struct {
		sync.RWMutex
		matches map[string]*MatchState
	}{matches: make(map[string]*MatchState)}
)

// MatchState holds Tic-Tac-Toe game state
type MatchState struct {
	Presences     map[string]runtime.Presence `json:"presences"`
	Board         [3][3]string                `json:"board"`          // 3x3 grid
	CurrentPlayer string                      `json:"current_player"` // UserID of current turn
	Started       bool                        `json:"started"`        // Game active
	Winner        string                      `json:"winner"`         // UserID or "" for draw
}

// Match is the match handler
type Match struct{}

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("Initializing Tic-Tac-Toe module")
	if err := initializer.RegisterMatch("tic_tac_toe", newMatch); err != nil {
		logger.Error("[RegisterMatch] error: %v", err)
		return err
	}
	// Register RPC for matchmaking
	if err := initializer.RegisterRpc("find_or_create_match", findOrCreateMatch); err != nil {
		logger.Error("[RegisterRpc] error: %v", err)
		return err
	}

	return nil
}

func findOrCreateMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	var params struct {
		UserID string `json:"user_id"`
	}
	logger.Info("Payload for user creation ", payload)
	if err := json.Unmarshal([]byte(payload), &params); err != nil {
		return "", err
	}

	matchStore.Lock()
	defer matchStore.Unlock()

	// Check if user is already in an active match
	for exitMatchID, match := range matchStore.matches {
		if match.Started {
			return fmt.Sprintf(`{"match_id":"%s"}`, exitMatchID), nil
		}
	}

	// Create new match

	modulename := "tic_tac_toe"
	if matchId, err := nk.MatchCreate(ctx, modulename, make(map[string]interface{})); err != nil {
		logger.Error("Error on match creation", err.Error())
		return "", err
	} else {
		logger.Debug("Created new match", "match_id", matchId)
		newMatch := &MatchState{Started: true}
		matchStore.matches[matchId] = newMatch
		return fmt.Sprintf(`{"match_id":"%s"}`, matchId), nil
	}
}

func newMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
	logger.Info("[NEW MATCH] Creating Tic-Tac-Toe match")
	return &Match{}, nil
}

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	logger.Info("Initializing match instance")
	state := &MatchState{
		Presences:     make(map[string]runtime.Presence),
		Board:         [3][3]string{},
		CurrentPlayer: "",
		Started:       false,
		Winner:        "",
	}
	return state, 10, "" // 10 ticks/sec (100ms)
}
func (m *Match) MatchDataSend(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, sender runtime.Presence, opCode int64, data []byte) interface{} {
	mState, _ := state.(*MatchState)
	logger.Info(fmt.Sprintf("MatchData called with tick=%d, opCode=%d, data=%s, state=%+v", tick, opCode, string(data), state))

	stateBytes, _ := json.Marshal(mState)
	dispatcher.BroadcastMessage(4, stateBytes, nil, nil, true)
	if !mState.Started || mState.Winner != "" {
		return mState
	}

	if opCode != 4 {
		logger.Error("Unexpected opCode: %d", opCode)
		return mState
	}

	// Parse move data "row,col"
	var row, col int
	_, err := fmt.Sscanf(string(data), "%d,%d", &row, &col)
	if err != nil {
		logger.Error("Invalid move data: %v", err)
		return mState
	}

	// Validate and apply move (uncomment and adjust as needed)
	/*
		   if sender.GetUserId() != mState.CurrentPlayer {
		       logger.Error("Not this player's turn")
		       return mState
		   }
		   if mState.Board[row][col] != "" {
		       logger.Error("Cell already occupied")
		       return mState
		   }
		   mState.Board[row][col] = sender.GetUserId()
		   for userID := range mState.Presences {
			if userID != sender.GetUserId() {
				mState.CurrentPlayer = userID
				break
				}
				}
	*/

	// Broadcast updated state

	return mState
}

func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	mState, _ := state.(*MatchState)
	// Allow join only if game hasn't started or has < 2 players
	acceptUser := len(mState.Presences) < 2 && !mState.Started
	return state, acceptUser, ""
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)
	logger.Info("Player(s) joined: %v", presences)

	for _, p := range presences {
		mState.Presences[p.GetUserId()] = p
	}

	// Start game when 2 players join
	if len(mState.Presences) == 2 && !mState.Started {
		mState.Started = true
		// Assign first player randomly (first in map)
		for userID := range mState.Presences {
			mState.CurrentPlayer = userID
			break
		}
		// Broadcast initial state
		stateBytes, _ := json.Marshal(mState)
		dispatcher.BroadcastMessage(1, stateBytes, nil, nil, true)
	}

	return mState
}

func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)
	logger.Info("Player(s) left: %v", presences)

	for _, p := range presences {
		delete(mState.Presences, p.GetUserId())
	}

	// End game if a player leaves
	if mState.Started {
		mState.Winner = "" // Draw if someone leaves
		mState.Started = false
		stateBytes, _ := json.Marshal(mState)
		dispatcher.BroadcastMessage(1, stateBytes, nil, nil, true)
	}

	return mState
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, matchState []runtime.MatchData) interface{} {
	mState, _ := state.(*MatchState)

	if !mState.Started || mState.Winner != "" {
		return mState
	}

	// Check for a winner
	mState.Winner = checkWinner(mState.Board)
	if mState.Winner != "" {
		mState.Started = false
		stateBytes, _ := json.Marshal(mState)
		dispatcher.BroadcastMessage(2, stateBytes, nil, nil, true)
	}

	// Check for draw
	if mState.Winner == "" && isBoardFull(mState.Board) {
		mState.Winner = "draw"
		mState.Started = false
		stateBytes, _ := json.Marshal(mState)
		dispatcher.BroadcastMessage(2, stateBytes, nil, nil, true)
	}

	return mState
}

func checkWinner(board [3][3]string) string {
	lines := [8][3][2]int{
		{{0, 0}, {0, 1}, {0, 2}},
		{{1, 0}, {1, 1}, {1, 2}},
		{{2, 0}, {2, 1}, {2, 2}},
		{{0, 0}, {1, 0}, {2, 0}},
		{{0, 1}, {1, 1}, {2, 1}},
		{{0, 2}, {1, 2}, {2, 2}},
		{{0, 0}, {1, 1}, {2, 2}},
		{{0, 2}, {1, 1}, {2, 0}},
	}

	for _, line := range lines {
		a, b, c := line[0], line[1], line[2]
		if board[a[0]][a[1]] != "" && board[a[0]][a[1]] == board[b[0]][b[1]] && board[a[0]][a[1]] == board[c[0]][c[1]] {
			return board[a[0]][a[1]]
		}
	}
	return ""
}

func isBoardFull(board [3][3]string) bool {
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if board[i][j] == "" {
				return false
			}
		}
	}
	return true
}

/*
context.Context, "github.com/heroiclabs/nakama-common/runtime".Logger, *sql.DB, "github.com/heroiclabs/nakama-common/runtime".NakamaModule, "github.com/heroiclabs/nakama-common/runtime".MatchDispatcher, int64, interface{}, int) interface{}
*/

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, nedata int) interface{} {
	// Remove match from store
	logger.Debug("Termination match ", state, tick, dispatcher, nk, nedata)
	matchStore.Lock()
	defer matchStore.Unlock()

	// for id, match := range matchStore.matches {
	// 	if match == m {
	// 		delete(matchStore.matches, id)
	// 		logger.Debug("Terminated match", "match_id", id)
	// 		break
	// 	}
	// }
	return state
}

/*
context.Context, "github.com/heroiclabs/nakama-common/runtime".Logger, *sql.DB, "github.com/heroiclabs/nakama-common/runtime".NakamaModule, "github.com/heroiclabs/nakama-common/runtime".MatchDispatcher, int64, interface{}, string
*/

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	mState, _ := state.(*MatchState)

	logger.Debug(fmt.Sprintf("MatchSignal called with tick=%d, data=%s, state=%+v", tick, data, state))

	if !mState.Started || mState.Winner != "" {
		return mState, data
	}

	// Parse move data "row,col"
	var row, col int
	_, err := fmt.Sscanf(data, "%d,%d", &row, &col)
	if err != nil {
		logger.Error("Invalid move data: %v", err)
		return mState, data
	}

	// if sender.GetUserId() != mState.CurrentPlayer {
	// 	logger.Error("Not this player's turn")
	// 	return mState, data
	// }

	// if mState.Board[row][col] != "" {
	// 	logger.Error("Cell already occupied")
	// 	return mState, data
	// }

	// // Apply move
	// mState.Board[row][col] = sender.GetUserId()

	// // Switch turn
	// for userID := range mState.Presences {
	// 	if userID != sender.GetUserId() {
	// 		mState.CurrentPlayer = userID
	// 		break
	// 	}
	// }

	// Broadcast updated state
	stateBytes, _ := json.Marshal(mState)
	dispatcher.BroadcastMessage(3, stateBytes, nil, nil, true)

	return mState, data
}

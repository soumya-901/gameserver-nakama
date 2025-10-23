package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/heroiclabs/nakama-common/runtime"
)

// MatchState holds Tic-Tac-Toe game state
type MatchState struct {
	MatchID       string                      `json:"match_id"`
	Presences     map[string]runtime.Presence `json:"presences"`
	Board         [3][3]string                `json:"board"`          // 3x3 grid
	CurrentPlayer string                      `json:"current_player"` // UserID of current turn
	CurrentSymbol string                      `json:"current_symbol"`
	Started       bool                        `json:"started"` // Game active
	Winner        string                      `json:"winner"`  // UserID or "" for draw
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

	limit := 1
	isAuthoritative := true
	label := ""
	minSize := 0
	maxSize := 1

	matches, err := nk.MatchList(ctx, limit, isAuthoritative, label, &minSize, &maxSize, "")
	if err != nil {
		logger.WithField("err", err).Error("Match list error.")
	} else {
		for _, match := range matches {
			logger.Info("Found match with id: %s", match.GetMatchId())
			return fmt.Sprintf(`{"match_id":"%s"}`, match.GetMatchId()), nil
		}
	}

	// Create new match

	modulename := "tic_tac_toe"
	matchDetails := map[string]interface{}{"match_id": uuid.New().String()}
	if matchId, err := nk.MatchCreate(ctx, modulename, matchDetails); err != nil {
		logger.Error("Error on match creation", err.Error())
		return "", err
	} else {
		logger.Debug("Created new match", "match_id", matchId)
		return fmt.Sprintf(`{"match_id":"%s"}`, matchId), nil
	}
}

func newMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
	logger.Info("[NEW MATCH] Creating Tic-Tac-Toe match")
	return &Match{}, nil
}

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	logger.Info("Initializing match instance with params:", params)

	matchID, _ := params["match_id"].(string)
	if matchID == "" {
		matchID = uuid.New().String() // fallback if not passed
	}

	state := &MatchState{
		MatchID:       matchID,
		Presences:     make(map[string]runtime.Presence),
		Board:         [3][3]string{},
		CurrentPlayer: "",
		CurrentSymbol: "",
		Started:       false,
		Winner:        "",
	}

	logger.Info("Initialized match with ID: ", matchID)
	return state, 2, "" // 2 ticks/sec for simplicity
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
			mState.CurrentSymbol = "X"
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

	for _, pdata := range presences {
		delete(mState.Presences, pdata.GetUserId())
	}

	// End game if a player leaves
	if mState.Started {
		for useid, _ := range mState.Presences {
			mState.Winner = useid // Draw if someone leaves
			break
		}

		mState.Started = false
		stateBytes, _ := json.Marshal(mState)
		dispatcher.BroadcastMessage(2, stateBytes, nil, nil, true)

	}
	// mState.Board = [3][3]string{}
	// mState.Presences = make(map[string]runtime.Presence)
	// mState.CurrentPlayer = ""
	// mState.CurrentSymbol = ""
	// mState.Winner = ""
	stateBytes, _ := json.Marshal(mState)
	// dispatcher.MatchKick(nil)
	dispatcher.BroadcastMessage(2, stateBytes, nil, nil, true)

	return nil
}

// func terminateMatch(mState *MatchState, logger runtime.Logger) {

// }

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, matchState []runtime.MatchData) interface{} {
	mState, _ := state.(*MatchState)

	if !mState.Started || mState.Winner != "" {
		return mState
	}
	logger.Debug("Running match loop", mState, matchState)
	for _, msg := range matchState {
		switch msg.GetOpCode() {
		case 4:
			var row, col int
			_, err := fmt.Sscanf(string(msg.GetData()), "%d,%d", &row, &col)
			if err != nil {
				logger.Error("Error occure ", err)
				return nil
			}
			logger.Debug("Player %s made a move: %+v", msg.GetUserId(), msg.GetData())

			// // Apply move
			mState.Board[row][col] = mState.CurrentSymbol

			// Switch turn
			for userID := range mState.Presences {
				if userID != msg.GetUserId() {
					mState.CurrentPlayer = userID
					if mState.CurrentSymbol == "X" {
						mState.CurrentSymbol = "0"
					} else {
						mState.CurrentSymbol = "X"
					}
					break
				}
			}
			stateBytes, _ := json.Marshal(mState)
			// Process move and update game state here...

			// Optionally broadcast the move to other players:
			dispatcher.BroadcastMessage(4, stateBytes, nil, nil, true)

		}

	}

	// Check for a winner
	winnerSymbol := checkWinner(mState.Board)
	if winnerSymbol != "" {
		// Find userID by symbol
		if winnerSymbol == mState.CurrentSymbol {
			mState.Winner = mState.CurrentPlayer
		} else {
			for userID := range mState.Presences {
				if userID != mState.CurrentPlayer {
					mState.Winner = userID
					break
				}
			}
		}

		mState.Started = false
		stateBytes, _ := json.Marshal(mState)
		dispatcher.BroadcastMessage(2, stateBytes, nil, nil, true)
		logger.Info("Match terminated:", mState.MatchID)
		return nil

	}

	// // Check for draw
	if mState.Winner == "" && isBoardFull(mState.Board) {
		mState.Winner = "draw"
		mState.Started = false
		stateBytes, _ := json.Marshal(mState)
		dispatcher.BroadcastMessage(2, stateBytes, nil, nil, true)
		return nil
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
	mState, _ := state.(*MatchState)
	return mState
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
	dispatcher.BroadcastMessage(4, stateBytes, nil, nil, true)

	return mState, data
}

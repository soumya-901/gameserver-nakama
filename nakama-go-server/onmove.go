package main

// import (
// 	"context"
// 	"encoding/json"

// 	"github.com/heroiclabs/nakama-common/runtime"
// )

// const (
// 	OpPlayerJoined = 1
// 	OpPlayerMoved  = 2
// )

// type MatchState struct {
// 	Presences map[string]runtime.Presence `json:"presences"`
// 	Board     [9]string                   `json:"board"`
// 	Turn      string                      `json:"turn"`
// }

// func MatchInit(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
// 	state := &MatchState{
// 		Presences: make(map[string]runtime.Presence),
// 	}
// 	return state, 1, ""
// }

// func MatchJoinAttempt(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
// 	return state, true, ""
// }

// func MatchJoin(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
// 	s := state.(*MatchState)

// 	for _, p := range presences {
// 		s.Presences[p.GetUserId()] = p
// 	}

// 	// Notify all others that a player joined
// 	msg := map[string]interface{}{
// 		"userId": presences[0].GetUserId(),
// 		"event":  "joined",
// 	}
// 	data, _ := json.Marshal(msg)

// 	dispatcher.BroadcastMessage(OpPlayerJoined, data, presences, nil, true)

// 	logger.Info("Player joined: %v", presences[0].GetUserId())
// 	return s
// }

// func MatchLeave(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
// 	s := state.(*MatchState)

// 	for _, p := range presences {
// 		delete(s.Presences, p.GetUserId())
// 	}

// 	logger.Info("Player left match")
// 	return s
// }

// func MatchLoop(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.Message) interface{} {
// 	s := state.(*MatchState)

// 	for _, msg := range messages {
// 		switch msg.GetOpCode() {
// 		case OpPlayerMoved:
// 			var move struct {
// 				Index  int    `json:"index"`
// 				Symbol string `json:"symbol"`
// 			}

// 			if err := json.Unmarshal(msg.GetData(), &move); err != nil {
// 				logger.Error("Invalid move data: %v", err)
// 				continue
// 			}

// 			if move.Index < 0 || move.Index > 8 {
// 				logger.Warn("Invalid move index")
// 				continue
// 			}

// 			if s.Board[move.Index] == "" {
// 				s.Board[move.Index] = move.Symbol

// 				// Broadcast the move to all other players
// 				dispatcher.BroadcastMessage(OpPlayerMoved, msg.GetData(), []runtime.Presence{msg.GetSender()}, nil, true)
// 				logger.Info("Player %v moved: %v -> %v", msg.GetSender().GetUserId(), move.Index, move.Symbol)
// 			}
// 		}
// 	}

// 	return s
// }

// func MatchTerminate(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
// 	logger.Info("Match terminated")
// 	return state
// }

// func MatchSignal(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
// 	return state, ""
// }

// func RegisterTicTacToeMatchModule(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule) error {
// 	var signal string
// 	var err error
// 	params := map[string]interface{}{
// 		"some": "data",
// 	}
// 	if signal, err = nk.MatchCreate(ctx, "tic_tac_toe", params); err != nil {
// 		return err
// 	}
// 	logger.Info("TicTacToe match handler registered.", signal)
// 	return nil
// }

package matching

// GitHub Copilot

// Simple tic-tac-toe match state
type TicTacToeState struct {
	Board    [9]int            `json:"board"`    // 0 empty, 1 X, 2 O
	Players  map[string]int    `json:"players"`  // userID -> symbol (1 or 2)
	Order    []string          `json:"order"`    // player join order (userIDs)
	Turn     string            `json:"turn"`     // userID whose turn it is
	Finished bool              `json:"finished"` // true when game finished
	Winner   int               `json:"winner"`   // 0 none/draw, 1 or 2 winner
	Meta     map[string]string `json:"meta"`     // optional metadata
}

// Match handler type
type yicTacToeMatch struct{}

// // Factory
// func NewTicTacToeMatch(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, params map[string]interface{}) (runtime.Match, error) {
// 	return &TicTacToeMatch{}, nil
// }

// Helper: check win
func checkWinner(board [9]int) int {
	lines := [8][3]int{
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
		{0, 4, 8},
		{2, 4, 6},
	}
	for _, l := range lines {
		a, b, c := l[0], l[1], l[2]
		if board[a] != 0 && board[a] == board[b] && board[b] == board[c] {
			return board[a]
		}
	}
	// check draw
	for i := 0; i < 9; i++ {
		if board[i] == 0 {
			return 0 // not finished
		}
	}
	return 0 // draw signaled by Finished true in logic
}

// MatchInit: return initial state and tick rate (0 = no ticks)
// func (m *TicTacToeMatch) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, error) {
// 	logger.Info("Init match ...")
// 	state := &TicTacToeState{
// 		Board:    [9]int{},
// 		Players:  map[string]int{},
// 		Order:    []string{},
// 		Turn:     "",
// 		Finished: false,
// 		Winner:   0,
// 		Meta:     map[string]string{},
// 	}
// 	return state, 0, nil
// }

// // MatchJoin: a presence joins the match
// func (m *TicTacToeMatch) MatchJoin(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, matchID string, presences []runtime.Presence, metadata map[string]string, state interface{}) (interface{}, error) {
// 	s := state.(*TicTacToeState)

// 	for _, pres := range presences {
// 		// if already present skip
// 		if _, ok := s.Players[pres.GetUserId()]; ok {
// 			continue
// 		}
// 		// limit players to 2
// 		if len(s.Order) >= 2 {
// 			// spectator: not assigned a symbol, just allowed to join
// 			continue
// 		}
// 		// assign symbol 1 or 2
// 		symbol := 1
// 		if len(s.Order) == 1 {
// 			symbol = 2
// 		}
// 		s.Players[pres.UserID] = symbol
// 		s.Order = append(s.Order, pres.UserID)
// 		// set turn to first player when two joined
// 		if len(s.Order) == 2 && s.Turn == "" {
// 			s.Turn = s.Order[0]
// 		}
// 	}

// 	// broadcast updated state to all participants
// 	data, _ := json.Marshal(s)
// 	presences = make([]runtime.Presence, 0, len(s.Order))
// 	for _, id := range s.Order {
// 		presences = append(presences, runtime.Presence{UserID: id})
// 	}
// 	_ = nk.MatchSend(ctx, matchID, OpStateUpdate, data, presences) // opcode 0 = full-state update

// 	return s, nil
// }

// // MatchLeave: presences left
// func (m *TicTacToeMatch) MatchLeave(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, matchID string, presences []runtime.Presence, state interface{}) (interface{}, error) {
// 	s := state.(*TicTacToeState)
// 	for _, pres := range presences {
// 		if _, ok := s.Players[pres.UserID]; ok {
// 			delete(s.Players, pres.UserID)
// 			// remove from order
// 			newOrder := []string{}
// 			for _, id := range s.Order {
// 				if id != pres.UserID {
// 					newOrder = append(newOrder, id)
// 				}
// 			}
// 			s.Order = newOrder
// 			// if game not finished and player left, mark finished
// 			if !s.Finished {
// 				s.Finished = true
// 				data, _ := json.Marshal(s)
// 				presences := make([]runtime.Presence, 0, len(s.Order))
// 				for _, id := range s.Order {
// 					presences = append(presences, runtime.Presence{UserID: id})
// 				}
// 				_ = nk.MatchSend(ctx, matchID, OpStateUpdate, data, presences)
// 				return s, nil
// 			}
// 			data, _ := json.Marshal(s)
// 			_ = nk.MatchBroadcast(ctx, matchID, 0, data)
// 			return s, nil
// 		}
// 	}
// }

// // client message opcodes
// const (
// 	OpStateUpdate = 0 // server -> client full state
// 	OpMove        = 1 // client -> server: move
// 	OpMoveResult  = 2 // server -> clients: move accepted/rejected + state
// )

// // Move payload
// type MovePayload struct {
// 	Cell int `json:"cell"` // 0..8
// }

// // MatchLoop: handle messages (moves). returns updated state and boolean end flag.
// // func (m *TicTacToeMatch) MatchLoop(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, matchID string, tick int64, state interface{}, messages []runtime.MatchData) (interface{}, bool, error) {
// // 	s := state.(*TicTacToeState)

// // 	for _, msg := range messages {
// // 		// only handle OpMove from clients
// // 		if msg.OpCode != OpMove {
// // 			continue
// // 		}
// // 		// find sender user id (sender is presence.SessionId? In Go runtime MatchData.Sender is presence string "userid:sessionid")
// // 		senderPres := msg.Sender
// // 		// the Nakama Go runtime Sender format can be "userID:sessionID" — we extract userID before colon
// // 		userID := senderPres
// // 		if idx := len(senderPres); idx > 0 {
// // 			// try split at first colon
// // 			for i := 0; i < len(senderPres); i++ {
// // 				if senderPres[i] == ':' {
// // 					userID = senderPres[:i]
// // 					break
// // 				}
// // 			}
// // 		}

// // 		var p MovePayload
// // 		if err := json.Unmarshal(msg.Data, &p); err != nil {
// // 			continue
// // 		}

// // 		// validate
// // 		if s.Finished {
// // 			// ignore moves after finished
// // 			continue
// // 		}
// // 		// only allow players
// // 		sym, ok := s.Players[userID]
// // 		if !ok || sym == 0 {
// // 			// ignore spectators or unknown
// // 			continue
// // 		}
// // 		// only allow when it's player's turn
// // 		if s.Turn != userID {
// // 			// optionally send rejection to sender
// // 			resp := map[string]interface{}{"ok": false, "reason": "not your turn"}
// // 			b, _ := json.Marshal(resp)
// // 			_ = nk.MatchSend(ctx, matchID, OpMoveResult, b, []runtime.Presence{{UserID: userID}})
// // 			continue
// // 		}
// // 		// validate cell
// // 		if p.Cell < 0 || p.Cell > 8 || s.Board[p.Cell] != 0 {
// // 			resp := map[string]interface{}{"ok": false, "reason": "invalid cell"}
// // 			b, _ := json.Marshal(resp)
// // 			_ = nk.MatchSend(ctx, matchID, OpMoveResult, b, []runtime.Presence{{UserID: userID}})
// // 			continue
// // 		}
// // 		// apply move
// // 		s.Board[p.Cell] = sym

// // 		// check winner
// // 		win := checkWinner(s.Board)
// // 		if win != 0 {
// // 			s.Finished = true
// // 			s.Winner = win
// // 		} else {
// // 			// check draw
// // 			allFilled := true
// // 			for i := 0; i < 9; i++ {
// // 				if s.Board[i] == 0 {
// // 					allFilled = false
// // 					break
// // 				}
// // 			}
// // 			if allFilled {
// // 				s.Finished = true
// // 				s.Winner = 0 // draw
// // 			}
// // 		}
// // 		// update turn if not finished
// // 		if !s.Finished {
// // 			// swap turn to other player if exists
// // 			if len(s.Order) == 2 {
// // 				s.Turn = s.Order[0]
// // 			}
// // 		}
// // 		// broadcast new state to all match participants
// // 		data, _ := json.Marshal(s)
// // 		presences := make([]runtime.Presence, 0, len(s.Order))
// // 		for _, id := range s.Order {
// // 			presences = append(presences, runtime.Presence{UserID: id})
// // 		}
// // 		_ = nk.MatchSend(ctx, matchID, OpStateUpdate, data, presences)

// // 		// // broadcast new state to all match participants
// // 		// data, _ := json.Marshal(s)
// // 		// _ = nk.MatchBroadcast(ctx, matchID, OpStateUpdate, data)

// // 		// // reply to mover that move succeeded
// // 		// resp := map[string]interface{}{"ok": true, "state": s}
// // 		// b, _ := json.Marshal(resp)
// // 		// _ = nk.MatchSend(ctx, matchID, OpMoveResult, b, []runtime.Presence{{UserID: userID}})
// // 		// continue match loop; return false to not terminate by default

// // 		return s, false, nil
// // 	};
// // }

// // MatchTerminate
// func (m *TicTacToeMatch) MatchTerminate(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, matchID string, state interface{}) error {
// 	// no persistent cleanup needed here
// 	return nil
// }

// // MatchSignal (unused)
// func (m *TicTacToeMatch) MatchSignal(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, matchID string, payload string) (interface{}, error) {
// 	return nil, nil
// }

// // RPC to create a new match programmatically
// func rpcCreateTTT(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, payload string) (string, error) {
// 	// payload can be used to pass match metadata; ignoring for simplicity
// 	matchID, err := nk.MatchCreate(ctx, "tic_tac_toe", map[string]interface{}{})
// 	if err != nil {
// 		return "", err
// 	}
// 	return `{"match_id":"` + matchID + `"}`, nil
// }

// // // InitModule registers the match and RPC
// // func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
// // 	initializer.RegisterMatch("tic_tac_toe", NewTicTacToeMatch)
// // 	initializer.RegisterRpc("create_ttt", rpcCreateTTT)
// // 	return nil
// // }

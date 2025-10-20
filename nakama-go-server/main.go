package main

import (
	"context"
	"database/sql"
	"nakam-rpc-func/matching"
	"sync"

	"github.com/heroiclabs/nakama-common/runtime"
)

var (
	matchStore = struct {
		sync.RWMutex
		matches map[string]*matching.TicTacToeMatch
	}{matches: make(map[string]*matching.TicTacToeMatch)}
)

func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("Initializing Nakama Go server module...")
	logger.Info("Registering match 'tictactoe'")
	if err := initializer.RegisterMatch("tictactoe", func(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (runtime.Match, error) {
		match := &matching.TicTacToeMatch{
			ID:      "gfdgdfgfdg",         // unique match ID
			Players: make(map[string]any), // or map[string]*Player if defined
		}

		// Store match globally for later use
		matchStore.Lock()
		matchStore.matches[match.ID] = match
		matchStore.Unlock()

		logger.Info("New match created and stored with ID: %s", match.ID)

		logger.Info("Creating new TicTacToeMatch instance")
		return match, nil
	}); err != nil {
		return err
	}

	return nil
}

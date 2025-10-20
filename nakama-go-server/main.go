package main

import (
	"context"

	"github.com/heroiclabs/nakama-common/runtime"
)

const (
	OpPlayerJoined = 1
	OpPlayerMoved  = 2
)

type MatchState struct {
	Presences map[string]runtime.Presence `json:"presences"`
	Board     [9]string                   `json:"board"`
	Turn      string                      `json:"turn"`
}

func InitModule(ctx context.Context, logger runtime.Logger, nk runtime.NakamaModule, intiallizer runtime.Initializer) error {
	logger.Info("Initializing custom module ...")

	intiallizer.RegisterRpc("HealthCheck", HealthCheck)

	return nil
}

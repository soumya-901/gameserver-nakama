package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/heroiclabs/nakama-common/runtime"
)

type HealthCheckResponse struct {
	Status string `json:"status"`
}

func HealthCheck(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	logger.Debug("Health check invoked ...")
	response := HealthCheckResponse{
		Status: "ok",
	}
	res, err := json.Marshal(response)
	if err != nil {
		logger.Error("Error marshalling health check response: %v", err)
		return "", err
	}
	logger.Debug("Health check response: %s", string(res))
	return string(res), nil
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

func initLogger() {
	w := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	lvl := os.Getenv("LOG_LEVEL")
	if lvl == "" {
		lvl = "info"
	}

	level, err := zerolog.ParseLevel(lvl)
	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	log = zerolog.New(w).With().Timestamp().Logger()
}

func validateCoordinates(latStr, lonStr string) (float64, float64, error) {
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid latitude value")
	}
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid longitude value")
	}
	if lat < -90 || lat > 90 {
		return 0, 0, fmt.Errorf("latitude must be between -90 and 90")
	}
	if lon < -180 || lon > 180 {
		return 0, 0, fmt.Errorf("longitude must be between -180 and 180")
	}
	return lat, lon, nil
}

func validateQueryParam(q string) error {
	if q == "" {
		return fmt.Errorf("query parameter is required")
	}
	if len(q) < 2 {
		return fmt.Errorf("query must be at least 2 characters")
	}
	return nil
}

func validateLimit(limitStr string) (int, error) {
	if limitStr == "" {
		return 50, nil
	}
	n, err := strconv.Atoi(limitStr)
	if err != nil {
		return 0, fmt.Errorf("invalid limit value")
	}
	if n < 1 || n > 100 {
		return 0, fmt.Errorf("limit must be between 1 and 100")
	}
	return n, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, APIError{Code: status, Message: msg})
}

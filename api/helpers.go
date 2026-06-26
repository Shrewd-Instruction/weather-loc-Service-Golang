package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"weather_loc_service/models"
)

func ValidateCoordinates(latStr, lonStr string) (float64, float64, error) {
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

func ValidateQueryParam(q string) error {
	if q == "" {
		return fmt.Errorf("query parameter is required")
	}
	if len(q) < 2 {
		return fmt.Errorf("query must be at least 2 characters")
	}
	return nil
}

func ValidateLimit(limitStr string) (int, error) {
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

func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, models.APIError{Code: status, Message: msg})
}

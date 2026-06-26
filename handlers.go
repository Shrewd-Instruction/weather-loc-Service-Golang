package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

func handleGetWeather(wsvc *WeatherService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		latStr := r.URL.Query().Get("lat")
		lonStr := r.URL.Query().Get("lon")

		if latStr == "" || lonStr == "" {
			writeError(w, http.StatusBadRequest, "lat and lon query parameters are required")
			return
		}

		lat, lon, err := validateCoordinates(latStr, lonStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		data, err := wsvc.getWeather(lat, lon)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if db != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			summary := fmt.Sprintf("%.1f°C, %s", data.Temperature, data.Description)
			_, dbErr := db.ExecContext(ctx,
				"EXEC sp_InsertQuery @query_type=@p1, @query_text=@p2, @latitude=@p3, @longitude=@p4, @response_summary=@p5",
				"weather", fmt.Sprintf("%.4f,%.4f", lat, lon), lat, lon, summary,
			)
			if dbErr != nil {
				log.Error().Err(dbErr).Msg("failed to save query to db")
			}
		}

		writeJSON(w, http.StatusOK, data)
	}
}

func handleGetWeatherByCity(wsvc *WeatherService, lsvc *LocationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		if city == "" {
			writeError(w, http.StatusBadRequest, "city parameter is required")
			return
		}

		locations, err := lsvc.searchLocation(city)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(locations) == 0 {
			writeError(w, http.StatusNotFound, "location not found")
			return
		}

		lat, _ := strconv.ParseFloat(locations[0].Latitude, 64)
		lon, _ := strconv.ParseFloat(locations[0].Longitude, 64)

		data, err := wsvc.getWeather(lat, lon)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		data.Location = locations[0].DisplayName

		if db != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			summary := fmt.Sprintf("%s: %.1f°C, %s", city, data.Temperature, data.Description)
			_, dbErr := db.ExecContext(ctx,
				"EXEC sp_InsertQuery @query_type=@p1, @query_text=@p2, @latitude=@p3, @longitude=@p4, @response_summary=@p5",
				"weather", city, lat, lon, summary,
			)
			if dbErr != nil {
				log.Error().Err(dbErr).Msg("failed to save query to db")
			}
		}

		writeJSON(w, http.StatusOK, data)
	}
}

func handleSearchLocation(lsvc *LocationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		err := validateQueryParam(q)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		locations, err := lsvc.searchLocation(q)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if locations == nil {
			locations = []LocationData{}
		}

		if db != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			summary := fmt.Sprintf("found %d results", len(locations))
			_, dbErr := db.ExecContext(ctx,
				"EXEC sp_InsertQuery @query_type=@p1, @query_text=@p2, @latitude=@p3, @longitude=@p4, @response_summary=@p5",
				"location", q, 0.0, 0.0, summary,
			)
			if dbErr != nil {
				log.Error().Err(dbErr).Msg("failed to save query to db")
			}
		}

		writeJSON(w, http.StatusOK, locations)
	}
}

func handleReverseGeocode(lsvc *LocationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		latStr := r.URL.Query().Get("lat")
		lonStr := r.URL.Query().Get("lon")

		if latStr == "" || lonStr == "" {
			writeError(w, http.StatusBadRequest, "lat and lon query parameters are required")
			return
		}

		lat, lon, err := validateCoordinates(latStr, lonStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		data, err := lsvc.reverseGeocode(lat, lon)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if db != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			_, dbErr := db.ExecContext(ctx,
				"EXEC sp_InsertQuery @query_type=@p1, @query_text=@p2, @latitude=@p3, @longitude=@p4, @response_summary=@p5",
				"reverse", fmt.Sprintf("%.4f,%.4f", lat, lon), lat, lon, data.DisplayName,
			)
			if dbErr != nil {
				log.Error().Err(dbErr).Msg("failed to save query to db")
			}
		}

		writeJSON(w, http.StatusOK, data)
	}
}

func handleGetInsights(wsvc *WeatherService, lsvc *LocationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		city := r.URL.Query().Get("city")
		if city == "" {
			writeError(w, http.StatusBadRequest, "city query parameter is required")
			return
		}

		locations, err := lsvc.searchLocation(city)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(locations) == 0 {
			writeError(w, http.StatusNotFound, "location not found")
			return
		}

		loc := locations[0]
		lat, _ := strconv.ParseFloat(loc.Latitude, 64)
		lon, _ := strconv.ParseFloat(loc.Longitude, 64)

		weather, err := wsvc.getWeather(lat, lon)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		weather.Location = loc.DisplayName

		resp := InsightResponse{
			Location: loc,
			Weather:  *weather,
		}

		if db != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			summary := fmt.Sprintf("%s: %.1f°C, %s", city, weather.Temperature, weather.Description)
			_, dbErr := db.ExecContext(ctx,
				"EXEC sp_InsertQuery @query_type=@p1, @query_text=@p2, @latitude=@p3, @longitude=@p4, @response_summary=@p5",
				"insight", city, lat, lon, summary,
			)
			if dbErr != nil {
				log.Error().Err(dbErr).Msg("failed to save query to db")
			}
		}

		writeJSON(w, http.StatusOK, resp)
	}
}

func handleGetHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeError(w, http.StatusServiceUnavailable, "database not available")
			return
		}

		queryType := r.URL.Query().Get("type")
		limitStr := r.URL.Query().Get("limit")

		limit, err := validateLimit(limitStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var typeParam interface{}
		if queryType != "" {
			typeParam = queryType
		}

		rows, err := db.QueryContext(ctx,
			"EXEC sp_GetQueryHistory @query_type=@p1, @limit=@p2",
			typeParam, limit,
		)
		if err != nil {
			log.Error().Err(err).Msg("failed to query history")
			writeError(w, http.StatusInternalServerError, "failed to fetch history")
			return
		}
		defer rows.Close()

		var history []QueryHistory
		for rows.Next() {
			var row QueryHistory
			err = rows.Scan(&row.ID, &row.QueryType, &row.QueryText, &row.Latitude, &row.Longitude, &row.ResponseSummary, &row.CreatedAt)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to read data")
				return
			}
			history = append(history, row)
		}

		if history == nil {
			history = []QueryHistory{}
		}

		writeJSON(w, http.StatusOK, history)
	}
}

func handleGetStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			writeError(w, http.StatusServiceUnavailable, "database not available")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		rows, err := db.QueryContext(ctx, "EXEC sp_GetQueryStats")
		if err != nil {
			log.Error().Err(err).Msg("failed to query stats")
			writeError(w, http.StatusInternalServerError, "failed to fetch stats")
			return
		}
		defer rows.Close()

		var stats []QueryStats
		for rows.Next() {
			var row QueryStats
			err = rows.Scan(&row.QueryType, &row.TotalCount)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to read stats")
				return
			}
			stats = append(stats, row)
		}

		if stats == nil {
			stats = []QueryStats{}
		}

		writeJSON(w, http.StatusOK, stats)
	}
}

func handleHealth(cache *CacheService, weatherURL, nominatimURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{
			"status": "healthy",
		}

		dbStatus := "down"
		if db != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			if db.PingContext(ctx) == nil {
				dbStatus = "up"
			}
		}
		status["database"] = dbStatus

		redisStatus := "down"
		if cache != nil && cache.Ping() == nil {
			redisStatus = "up"
		}
		status["redis"] = redisStatus

		weatherStatus := "down"
		resp, err := http.Get(weatherURL + "/v1/forecast?latitude=0&longitude=0&current=temperature_2m")
		if err == nil && resp.StatusCode == http.StatusOK {
			weatherStatus = "up"
		}
		if resp != nil {
			resp.Body.Close()
		}
		status["weather_api"] = weatherStatus

		nominatimStatus := "down"
		req, _ := http.NewRequest("GET", nominatimURL+"/status", nil)
		if req != nil {
			req.Header.Set("User-Agent", "WeatherLocService/1.0")
			resp2, err := http.DefaultClient.Do(req)
			if err == nil && resp2.StatusCode == http.StatusOK {
				nominatimStatus = "up"
			}
			if resp2 != nil {
				resp2.Body.Close()
			}
		}
		status["nominatim_api"] = nominatimStatus

		if dbStatus == "down" || redisStatus == "down" || weatherStatus == "down" || nominatimStatus == "down" {
			status["status"] = "degraded"
		}

		writeJSON(w, http.StatusOK, status)
	}
}

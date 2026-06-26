package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"weather_loc_service/cache"
	"weather_loc_service/database"
	"weather_loc_service/logger"
	"weather_loc_service/models"
	"weather_loc_service/services"
)

func HandleGetWeather(wsvc *services.WeatherService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		latStr := r.URL.Query().Get("lat")
		lonStr := r.URL.Query().Get("lon")

		if latStr == "" || lonStr == "" {
			WriteError(w, http.StatusBadRequest, "lat and lon query parameters are required")
			return
		}

		lat, lon, err := ValidateCoordinates(latStr, lonStr)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		data, err := wsvc.GetWeather(lat, lon)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if database.DB != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			summary := fmt.Sprintf("%.1f°C, %s", data.Temperature, data.Description)
			_, dbErr := database.DB.ExecContext(ctx,
				"EXEC sp_InsertQuery @query_type=@p1, @query_text=@p2, @latitude=@p3, @longitude=@p4, @response_summary=@p5",
				"weather", fmt.Sprintf("%.4f,%.4f", lat, lon), lat, lon, summary,
			)
			if dbErr != nil {
				logger.Log.Error().Err(dbErr).Msg("failed to save query to db")
			}
		}

		WriteJSON(w, http.StatusOK, data)
	}
}

func HandleGetWeatherByCity(wsvc *services.WeatherService, lsvc *services.LocationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		if city == "" {
			WriteError(w, http.StatusBadRequest, "city parameter is required")
			return
		}

		locations, err := lsvc.SearchLocation(city)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(locations) == 0 {
			WriteError(w, http.StatusNotFound, "location not found")
			return
		}

		lat, _ := strconv.ParseFloat(locations[0].Latitude, 64)
		lon, _ := strconv.ParseFloat(locations[0].Longitude, 64)

		data, err := wsvc.GetWeather(lat, lon)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		data.Location = locations[0].DisplayName

		if database.DB != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			summary := fmt.Sprintf("%s: %.1f°C, %s", city, data.Temperature, data.Description)
			_, dbErr := database.DB.ExecContext(ctx,
				"EXEC sp_InsertQuery @query_type=@p1, @query_text=@p2, @latitude=@p3, @longitude=@p4, @response_summary=@p5",
				"weather", city, lat, lon, summary,
			)
			if dbErr != nil {
				logger.Log.Error().Err(dbErr).Msg("failed to save query to db")
			}
		}

		WriteJSON(w, http.StatusOK, data)
	}
}

func HandleSearchLocation(lsvc *services.LocationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		err := ValidateQueryParam(q)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		locations, err := lsvc.SearchLocation(q)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if locations == nil {
			locations = []models.LocationData{}
		}

		if database.DB != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			summary := fmt.Sprintf("found %d results", len(locations))
			_, dbErr := database.DB.ExecContext(ctx,
				"EXEC sp_InsertQuery @query_type=@p1, @query_text=@p2, @latitude=@p3, @longitude=@p4, @response_summary=@p5",
				"location", q, 0.0, 0.0, summary,
			)
			if dbErr != nil {
				logger.Log.Error().Err(dbErr).Msg("failed to save query to db")
			}
		}

		WriteJSON(w, http.StatusOK, locations)
	}
}

func HandleReverseGeocode(lsvc *services.LocationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		latStr := r.URL.Query().Get("lat")
		lonStr := r.URL.Query().Get("lon")

		if latStr == "" || lonStr == "" {
			WriteError(w, http.StatusBadRequest, "lat and lon query parameters are required")
			return
		}

		lat, lon, err := ValidateCoordinates(latStr, lonStr)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		data, err := lsvc.ReverseGeocode(lat, lon)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if database.DB != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			_, dbErr := database.DB.ExecContext(ctx,
				"EXEC sp_InsertQuery @query_type=@p1, @query_text=@p2, @latitude=@p3, @longitude=@p4, @response_summary=@p5",
				"reverse", fmt.Sprintf("%.4f,%.4f", lat, lon), lat, lon, data.DisplayName,
			)
			if dbErr != nil {
				logger.Log.Error().Err(dbErr).Msg("failed to save query to db")
			}
		}

		WriteJSON(w, http.StatusOK, data)
	}
}

func HandleGetInsights(wsvc *services.WeatherService, lsvc *services.LocationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		city := r.URL.Query().Get("city")
		if city == "" {
			WriteError(w, http.StatusBadRequest, "city query parameter is required")
			return
		}

		locations, err := lsvc.SearchLocation(city)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(locations) == 0 {
			WriteError(w, http.StatusNotFound, "location not found")
			return
		}

		loc := locations[0]
		lat, _ := strconv.ParseFloat(loc.Latitude, 64)
		lon, _ := strconv.ParseFloat(loc.Longitude, 64)

		weather, err := wsvc.GetWeather(lat, lon)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		weather.Location = loc.DisplayName

		resp := models.InsightResponse{
			Location: loc,
			Weather:  *weather,
		}

		if database.DB != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()
			summary := fmt.Sprintf("%s: %.1f°C, %s", city, weather.Temperature, weather.Description)
			_, dbErr := database.DB.ExecContext(ctx,
				"EXEC sp_InsertQuery @query_type=@p1, @query_text=@p2, @latitude=@p3, @longitude=@p4, @response_summary=@p5",
				"insight", city, lat, lon, summary,
			)
			if dbErr != nil {
				logger.Log.Error().Err(dbErr).Msg("failed to save query to db")
			}
		}

		WriteJSON(w, http.StatusOK, resp)
	}
}

func HandleGetHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if database.DB == nil {
			WriteError(w, http.StatusServiceUnavailable, "database not available")
			return
		}

		queryType := r.URL.Query().Get("type")
		limitStr := r.URL.Query().Get("limit")

		limit, err := ValidateLimit(limitStr)
		if err != nil {
			WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var typeParam interface{}
		if queryType != "" {
			typeParam = queryType
		}

		rows, err := database.DB.QueryContext(ctx,
			"EXEC sp_GetQueryHistory @query_type=@p1, @limit=@p2",
			typeParam, limit,
		)
		if err != nil {
			logger.Log.Error().Err(err).Msg("failed to query history")
			WriteError(w, http.StatusInternalServerError, "failed to fetch history")
			return
		}
		defer rows.Close()

		var history []models.QueryHistory
		for rows.Next() {
			var row models.QueryHistory
			err = rows.Scan(&row.ID, &row.QueryType, &row.QueryText, &row.Latitude, &row.Longitude, &row.ResponseSummary, &row.CreatedAt)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, "failed to read data")
				return
			}
			history = append(history, row)
		}

		if history == nil {
			history = []models.QueryHistory{}
		}

		WriteJSON(w, http.StatusOK, history)
	}
}

func HandleGetStats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if database.DB == nil {
			WriteError(w, http.StatusServiceUnavailable, "database not available")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		rows, err := database.DB.QueryContext(ctx, "EXEC sp_GetQueryStats")
		if err != nil {
			logger.Log.Error().Err(err).Msg("failed to query stats")
			WriteError(w, http.StatusInternalServerError, "failed to fetch stats")
			return
		}
		defer rows.Close()

		var stats []models.QueryStats
		for rows.Next() {
			var row models.QueryStats
			err = rows.Scan(&row.QueryType, &row.TotalCount)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, "failed to read stats")
				return
			}
			stats = append(stats, row)
		}

		if stats == nil {
			stats = []models.QueryStats{}
		}

		WriteJSON(w, http.StatusOK, stats)
	}
}

func HandleHealth(c *cache.CacheService, weatherURL, nominatimURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := map[string]string{
			"status": "healthy",
		}

		dbStatus := "down"
		if database.DB != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			if database.DB.PingContext(ctx) == nil {
				dbStatus = "up"
			}
		}
		status["database"] = dbStatus

		redisStatus := "down"
		if c != nil && c.Ping() == nil {
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

		WriteJSON(w, http.StatusOK, status)
	}
}

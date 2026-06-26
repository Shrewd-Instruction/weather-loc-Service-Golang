package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"weather_loc_service/cache"
	"weather_loc_service/logger"
	"weather_loc_service/models"
)

var weatherCodes = map[int]string{
	0:  "Clear sky",
	1:  "Mainly clear",
	2:  "Partly cloudy",
	3:  "Overcast",
	45: "Fog",
	48: "Depositing rime fog",
	51: "Light drizzle",
	53: "Moderate drizzle",
	55: "Dense drizzle",
	56: "Light freezing drizzle",
	57: "Dense freezing drizzle",
	61: "Slight rain",
	63: "Moderate rain",
	65: "Heavy rain",
	66: "Light freezing rain",
	67: "Heavy freezing rain",
	71: "Slight snow fall",
	73: "Moderate snow fall",
	75: "Heavy snow fall",
	77: "Snow grains",
	80: "Slight rain showers",
	81: "Moderate rain showers",
	82: "Violent rain showers",
	85: "Slight snow showers",
	86: "Heavy snow showers",
	95: "Thunderstorm",
	96: "Thunderstorm with slight hail",
	99: "Thunderstorm with heavy hail",
}

type WeatherService struct {
	baseURL string
	client  *http.Client
	cache   *cache.CacheService
}

func NewWeatherService(baseURL string, cache *cache.CacheService) *WeatherService {
	return &WeatherService{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
		cache:   cache,
	}
}

func (s *WeatherService) GetWeather(lat, lon float64) (*models.WeatherData, error) {
	cacheKey := fmt.Sprintf("weather:%.2f:%.2f", lat, lon)

	if s.cache != nil {
		ctx := context.Background()
		cached, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var data models.WeatherData
			if json.Unmarshal([]byte(cached), &data) == nil {
				logger.Log.Debug().Msgf("cache hit for %s", cacheKey)
				return &data, nil
			}
		}
	}

	url := fmt.Sprintf("%s/v1/forecast?latitude=%.4f&longitude=%.4f&current=temperature_2m,relative_humidity_2m,apparent_temperature,weather_code,wind_speed_10m,cloud_cover,precipitation&daily=temperature_2m_max,temperature_2m_min,precipitation_sum&timezone=auto",
		s.baseURL, lat, lon)

	resp, err := s.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read weather response: %v", err)
	}

	var raw models.OpenMeteoResponse
	err = json.Unmarshal(body, &raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse weather response: %v", err)
	}

	desc := weatherCodes[raw.Current.WeatherCode]
	if desc == "" {
		desc = "Unknown"
	}

	var forecast []models.DayForecast
	for i := range raw.Daily.Time {
		f := models.DayForecast{Date: raw.Daily.Time[i]}
		if i < len(raw.Daily.Temperature2mMax) {
			f.TempMax = raw.Daily.Temperature2mMax[i]
		}
		if i < len(raw.Daily.Temperature2mMin) {
			f.TempMin = raw.Daily.Temperature2mMin[i]
		}
		if i < len(raw.Daily.PrecipitationSum) {
			f.Precipitation = raw.Daily.PrecipitationSum[i]
		}
		forecast = append(forecast, f)
	}

	data := &models.WeatherData{
		Latitude:      raw.Latitude,
		Longitude:     raw.Longitude,
		Timezone:      raw.Timezone,
		Elevation:     raw.Elevation,
		Temperature:   raw.Current.Temperature2m,
		FeelsLike:     raw.Current.ApparentTemperature,
		Humidity:      raw.Current.RelativeHumidity2m,
		WindSpeed:     raw.Current.WindSpeed10m,
		CloudCover:    raw.Current.CloudCover,
		Precipitation: raw.Current.Precipitation,
		WeatherCode:   raw.Current.WeatherCode,
		Description:   desc,
		Forecast:      forecast,
	}

	if s.cache != nil {
		ctx := context.Background()
		jsonData, _ := json.Marshal(data)
		s.cache.Set(ctx, cacheKey, string(jsonData), 10*time.Minute)
	}

	logger.Log.Info().Msgf("fetched weather for %.2f,%.2f from api", lat, lon)
	return data, nil
}

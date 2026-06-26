package models

import "time"

type OpenMeteoResponse struct {
	Latitude             float64           `json:"latitude"`
	Longitude            float64           `json:"longitude"`
	GenerationtimeMs     float64           `json:"generationtime_ms"`
	UtcOffsetSeconds     int               `json:"utc_offset_seconds"`
	Timezone             string            `json:"timezone"`
	TimezoneAbbreviation string            `json:"timezone_abbreviation"`
	Elevation            float64           `json:"elevation"`
	CurrentUnits         map[string]string `json:"current_units"`
	Current              CurrentWeather    `json:"current"`
	DailyUnits           map[string]string `json:"daily_units"`
	Daily                DailyWeather      `json:"daily"`
}

type CurrentWeather struct {
	Time                string  `json:"time"`
	Interval            int     `json:"interval"`
	Temperature2m       float64 `json:"temperature_2m"`
	RelativeHumidity2m  float64 `json:"relative_humidity_2m"`
	ApparentTemperature float64 `json:"apparent_temperature"`
	WeatherCode         int     `json:"weather_code"`
	WindSpeed10m        float64 `json:"wind_speed_10m"`
	CloudCover          float64 `json:"cloud_cover"`
	Precipitation       float64 `json:"precipitation"`
}

type DailyWeather struct {
	Time             []string  `json:"time"`
	Temperature2mMax []float64 `json:"temperature_2m_max"`
	Temperature2mMin []float64 `json:"temperature_2m_min"`
	PrecipitationSum []float64 `json:"precipitation_sum"`
}

type NominatimResult struct {
	PlaceID     int64             `json:"place_id"`
	Licence     string            `json:"licence"`
	OsmType     string            `json:"osm_type"`
	OsmID       int64             `json:"osm_id"`
	BoundingBox []string          `json:"boundingbox"`
	Lat         string            `json:"lat"`
	Lon         string            `json:"lon"`
	DisplayName string            `json:"display_name"`
	Class       string            `json:"class"`
	Type        string            `json:"type"`
	Importance  float64           `json:"importance"`
	Address     map[string]string `json:"address"`
}

type WeatherData struct {
	Location      string        `json:"location"`
	Latitude      float64       `json:"latitude"`
	Longitude     float64       `json:"longitude"`
	Timezone      string        `json:"timezone"`
	Elevation     float64       `json:"elevation"`
	Temperature   float64       `json:"temperature"`
	FeelsLike     float64       `json:"feels_like"`
	Humidity      float64       `json:"humidity"`
	WindSpeed     float64       `json:"wind_speed"`
	CloudCover    float64       `json:"cloud_cover"`
	Precipitation float64       `json:"precipitation"`
	WeatherCode   int           `json:"weather_code"`
	Description   string        `json:"description"`
	Forecast      []DayForecast `json:"forecast"`
}

type DayForecast struct {
	Date          string  `json:"date"`
	TempMax       float64 `json:"temp_max"`
	TempMin       float64 `json:"temp_min"`
	Precipitation float64 `json:"precipitation"`
}

type LocationData struct {
	DisplayName string            `json:"display_name"`
	Latitude    string            `json:"latitude"`
	Longitude   string            `json:"longitude"`
	Class       string            `json:"class"`
	Type        string            `json:"type"`
	Importance  float64           `json:"importance"`
	Address     map[string]string `json:"address"`
}

type InsightResponse struct {
	Location LocationData `json:"location"`
	Weather  WeatherData  `json:"weather"`
}

type QueryHistory struct {
	ID              int       `json:"id"`
	QueryType       string    `json:"query_type"`
	QueryText       string    `json:"query_text"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	ResponseSummary string    `json:"response_summary"`
	CreatedAt       time.Time `json:"created_at"`
}

type QueryStats struct {
	QueryType  string `json:"query_type"`
	TotalCount int    `json:"total_count"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

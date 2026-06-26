package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"weather_loc_service/logger"
	"weather_loc_service/models"
)

func TestGetWeather_Service(t *testing.T) {
	logger.InitLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.OpenMeteoResponse{
			Latitude: 52.52, Longitude: 13.41, Timezone: "Europe/Berlin", Elevation: 38,
			Current: models.CurrentWeather{
				Temperature2m: 20.5, RelativeHumidity2m: 65, ApparentTemperature: 19.8,
				WeatherCode: 3, WindSpeed10m: 12.4, CloudCover: 45, Precipitation: 0,
			},
			Daily: models.DailyWeather{
				Time:             []string{"2026-06-26"},
				Temperature2mMax: []float64{22.1},
				Temperature2mMin: []float64{14.2},
				PrecipitationSum: []float64{0},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	svc := NewWeatherService(ts.URL, nil)
	data, err := svc.GetWeather(52.52, 13.41)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data.Temperature != 20.5 {
		t.Errorf("expected 20.5, got %f", data.Temperature)
	}
	if data.Description != "Overcast" {
		t.Errorf("expected 'Overcast', got '%s'", data.Description)
	}
	if len(data.Forecast) != 1 {
		t.Errorf("expected 1 forecast day, got %d", len(data.Forecast))
	}
}

func TestGetWeather_ServerError(t *testing.T) {
	logger.InitLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()

	svc := NewWeatherService(ts.URL, nil)
	_, err := svc.GetWeather(52.52, 13.41)
	if err == nil {
		t.Error("expected error for 500, got nil")
	}
}

func TestSearchLocation_Service(t *testing.T) {
	logger.InitLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := []models.NominatimResult{
			{
				PlaceID: 12345, Lat: "48.8583", Lon: "2.2945",
				DisplayName: "Eiffel Tower, Paris, France",
				Class: "tourism", Type: "attraction", Importance: 0.9,
				Address: map[string]string{"city": "Paris", "country": "France"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	svc := NewLocationService(ts.URL, nil)
	results, err := svc.SearchLocation("Eiffel Tower")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].DisplayName != "Eiffel Tower, Paris, France" {
		t.Errorf("expected 'Eiffel Tower, Paris, France', got '%s'", results[0].DisplayName)
	}
}

func TestSearchLocation_ServerError(t *testing.T) {
	logger.InitLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()

	svc := NewLocationService(ts.URL, nil)
	_, err := svc.SearchLocation("Paris")
	if err == nil {
		t.Error("expected error for 500, got nil")
	}
}

func TestReverseGeocode_Service(t *testing.T) {
	logger.InitLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := models.NominatimResult{
			PlaceID: 67890, Lat: "28.6139", Lon: "77.2090",
			DisplayName: "New Delhi, Delhi, India",
			Class: "place", Type: "city", Importance: 0.8,
			Address: map[string]string{"city": "New Delhi", "state": "Delhi", "country": "India"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	svc := NewLocationService(ts.URL, nil)
	data, err := svc.ReverseGeocode(28.6139, 77.2090)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if data.DisplayName != "New Delhi, Delhi, India" {
		t.Errorf("expected 'New Delhi, Delhi, India', got '%s'", data.DisplayName)
	}
}

func TestReverseGeocode_ServerError(t *testing.T) {
	logger.InitLogger()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()

	svc := NewLocationService(ts.URL, nil)
	_, err := svc.ReverseGeocode(28.6139, 77.2090)
	if err == nil {
		t.Error("expected error for 500, got nil")
	}
}

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockWeatherAPI() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := OpenMeteoResponse{
			Latitude:  28.61,
			Longitude: 77.23,
			Timezone:  "Asia/Kolkata",
			Elevation: 216.0,
			Current: CurrentWeather{
				Time:                "2026-06-26T12:00",
				Interval:            900,
				Temperature2m:       35.2,
				RelativeHumidity2m:  45,
				ApparentTemperature: 37.8,
				WeatherCode:         2,
				WindSpeed10m:        12.5,
				CloudCover:          30,
				Precipitation:       0,
			},
			Daily: DailyWeather{
				Time:             []string{"2026-06-26", "2026-06-27"},
				Temperature2mMax: []float64{38.0, 39.2},
				Temperature2mMin: []float64{28.5, 29.0},
				PrecipitationSum: []float64{0, 2.5},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func mockNominatimAPI() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/reverse" {
			resp := NominatimResult{
				PlaceID:     123456,
				Lat:         "28.6139",
				Lon:         "77.2090",
				DisplayName: "New Delhi, Delhi, India",
				Class:       "place",
				Type:        "city",
				Importance:  0.9,
				Address:     map[string]string{"city": "New Delhi", "state": "Delhi", "country": "India"},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
		resp := []NominatimResult{
			{
				PlaceID:     123456,
				Lat:         "28.6139",
				Lon:         "77.2090",
				DisplayName: "New Delhi, Delhi, India",
				Class:       "place",
				Type:        "city",
				Importance:  0.9,
				Address:     map[string]string{"city": "New Delhi", "state": "Delhi", "country": "India"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestHandleGetWeather(t *testing.T) {
	initLogger()
	ts := mockWeatherAPI()
	defer ts.Close()

	wsvc := newWeatherService(ts.URL, nil)
	handler := handleGetWeather(wsvc)

	req := httptest.NewRequest("GET", "/api/v1/weather?lat=28.61&lon=77.23", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp WeatherData
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.Temperature != 35.2 {
		t.Errorf("expected temperature 35.2, got %.1f", resp.Temperature)
	}
	if resp.Description != "Partly cloudy" {
		t.Errorf("expected 'Partly cloudy', got '%s'", resp.Description)
	}
}

func TestHandleGetWeather_MissingParams(t *testing.T) {
	initLogger()
	ts := mockWeatherAPI()
	defer ts.Close()

	wsvc := newWeatherService(ts.URL, nil)
	handler := handleGetWeather(wsvc)

	req := httptest.NewRequest("GET", "/api/v1/weather", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleGetWeather_InvalidCoords(t *testing.T) {
	initLogger()
	ts := mockWeatherAPI()
	defer ts.Close()

	wsvc := newWeatherService(ts.URL, nil)
	handler := handleGetWeather(wsvc)

	req := httptest.NewRequest("GET", "/api/v1/weather?lat=999&lon=77.23", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleSearchLocation(t *testing.T) {
	initLogger()
	ts := mockNominatimAPI()
	defer ts.Close()

	lsvc := newLocationService(ts.URL, nil)
	handler := handleSearchLocation(lsvc)

	req := httptest.NewRequest("GET", "/api/v1/location/search?q=Delhi", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp []LocationData
	json.NewDecoder(rr.Body).Decode(&resp)

	if len(resp) == 0 {
		t.Error("expected at least one location result")
	}
	if resp[0].DisplayName != "New Delhi, Delhi, India" {
		t.Errorf("expected 'New Delhi, Delhi, India', got '%s'", resp[0].DisplayName)
	}
}

func TestHandleSearchLocation_EmptyQuery(t *testing.T) {
	initLogger()
	ts := mockNominatimAPI()
	defer ts.Close()

	lsvc := newLocationService(ts.URL, nil)
	handler := handleSearchLocation(lsvc)

	req := httptest.NewRequest("GET", "/api/v1/location/search", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleReverseGeocode(t *testing.T) {
	initLogger()
	ts := mockNominatimAPI()
	defer ts.Close()

	lsvc := newLocationService(ts.URL, nil)
	handler := handleReverseGeocode(lsvc)

	req := httptest.NewRequest("GET", "/api/v1/location/reverse?lat=28.6139&lon=77.2090", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var resp LocationData
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.DisplayName != "New Delhi, Delhi, India" {
		t.Errorf("expected 'New Delhi, Delhi, India', got '%s'", resp.DisplayName)
	}
}

func TestHandleGetHistory_NoDB(t *testing.T) {
	initLogger()
	db = nil
	handler := handleGetHistory()

	req := httptest.NewRequest("GET", "/api/v1/history", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestHandleGetStats_NoDB(t *testing.T) {
	initLogger()
	db = nil
	handler := handleGetStats()

	req := httptest.NewRequest("GET", "/api/v1/stats", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestValidateCoordinates(t *testing.T) {
	tests := []struct {
		name    string
		lat     string
		lon     string
		wantErr bool
	}{
		{"valid", "28.6", "77.2", false},
		{"zero", "0", "0", false},
		{"negative", "-33.86", "151.20", false},
		{"lat too high", "91", "77.2", true},
		{"lat too low", "-91", "77.2", true},
		{"lon too high", "28.6", "181", true},
		{"lon too low", "28.6", "-181", true},
		{"invalid lat", "abc", "77.2", true},
		{"invalid lon", "28.6", "xyz", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := validateCoordinates(tt.lat, tt.lon)
			if (err != nil) != tt.wantErr {
				t.Errorf("got err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateQueryParam(t *testing.T) {
	tests := []struct {
		input   string
		wantErr bool
	}{
		{"Delhi", false},
		{"New York", false},
		{"AB", false},
		{"", true},
		{"A", true},
	}

	for _, tt := range tests {
		err := validateQueryParam(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("validateQueryParam(%q) err=%v, wantErr=%v", tt.input, err, tt.wantErr)
		}
	}
}

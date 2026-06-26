package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type LocationService struct {
	baseURL string
	client  *http.Client
	cache   *CacheService
	mu      sync.Mutex
	lastReq time.Time
}

func newLocationService(baseURL string, cache *CacheService) *LocationService {
	return &LocationService{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
		cache:   cache,
	}
}

func (s *LocationService) throttle() {
	s.mu.Lock()
	defer s.mu.Unlock()
	since := time.Since(s.lastReq)
	if since < time.Second {
		time.Sleep(time.Second - since)
	}
	s.lastReq = time.Now()
}

func (s *LocationService) searchLocation(query string) ([]LocationData, error) {
	cacheKey := "location:search:" + query

	if s.cache != nil {
		ctx := context.Background()
		cached, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var data []LocationData
			if json.Unmarshal([]byte(cached), &data) == nil {
				log.Debug().Msgf("cache hit for %s", cacheKey)
				return data, nil
			}
		}
	}

	s.throttle()

	reqURL := fmt.Sprintf("%s/search?q=%s&format=json&addressdetails=1&limit=5",
		s.baseURL, url.QueryEscape(query))

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "WeatherLocService/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search location: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nominatim api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read location response: %v", err)
	}

	var results []NominatimResult
	err = json.Unmarshal(body, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to parse location response: %v", err)
	}

	var locations []LocationData
	for _, r := range results {
		locations = append(locations, LocationData{
			DisplayName: r.DisplayName,
			Latitude:    r.Lat,
			Longitude:   r.Lon,
			Class:       r.Class,
			Type:        r.Type,
			Importance:  r.Importance,
			Address:     r.Address,
		})
	}

	if s.cache != nil {
		ctx := context.Background()
		jsonData, _ := json.Marshal(locations)
		s.cache.Set(ctx, cacheKey, string(jsonData), 30*time.Minute)
	}

	log.Info().Msgf("searched location for '%s' from api", query)
	return locations, nil
}

func (s *LocationService) reverseGeocode(lat, lon float64) (*LocationData, error) {
	cacheKey := fmt.Sprintf("location:reverse:%.4f:%.4f", lat, lon)

	if s.cache != nil {
		ctx := context.Background()
		cached, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var data LocationData
			if json.Unmarshal([]byte(cached), &data) == nil {
				log.Debug().Msgf("cache hit for %s", cacheKey)
				return &data, nil
			}
		}
	}

	s.throttle()

	reqURL := fmt.Sprintf("%s/reverse?lat=%.6f&lon=%.6f&format=json&addressdetails=1",
		s.baseURL, lat, lon)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "WeatherLocService/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reverse geocode: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("nominatim api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read reverse geocode response: %v", err)
	}

	var result NominatimResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reverse geocode response: %v", err)
	}

	data := &LocationData{
		DisplayName: result.DisplayName,
		Latitude:    result.Lat,
		Longitude:   result.Lon,
		Class:       result.Class,
		Type:        result.Type,
		Importance:  result.Importance,
		Address:     result.Address,
	}

	if s.cache != nil {
		ctx := context.Background()
		jsonData, _ := json.Marshal(data)
		s.cache.Set(ctx, cacheKey, string(jsonData), 30*time.Minute)
	}

	log.Info().Msgf("reverse geocoded %.4f,%.4f from api", lat, lon)
	return data, nil
}

package main

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func setupRouter(weatherURL, nominatimURL string, cache *CacheService) *chi.Mux {
	r := chi.NewRouter()

	r.Use(requestLogger)
	r.Use(corsMiddleware)
	r.Use(rateLimiter(100, time.Minute))
	r.Use(middleware.Recoverer)

	wsvc := newWeatherService(weatherURL, cache)
	lsvc := newLocationService(nominatimURL, cache)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handleHealth(cache, weatherURL, nominatimURL))
		r.Get("/weather", handleGetWeather(wsvc))
		r.Get("/weather/{city}", handleGetWeatherByCity(wsvc, lsvc))
		r.Get("/location/search", handleSearchLocation(lsvc))
		r.Get("/location/reverse", handleReverseGeocode(lsvc))
		r.Get("/insights", handleGetInsights(wsvc, lsvc))
		r.Get("/history", handleGetHistory())
		r.Get("/stats", handleGetStats())
	})

	return r
}

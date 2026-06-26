package api

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"weather_loc_service/cache"
	"weather_loc_service/services"
)

func SetupRouter(weatherURL, nominatimURL string, c *cache.CacheService) *chi.Mux {
	r := chi.NewRouter()

	r.Use(RequestLogger)
	r.Use(CorsMiddleware)
	r.Use(RateLimiter(100, time.Minute))
	r.Use(middleware.Recoverer)

	wsvc := services.NewWeatherService(weatherURL, c)
	lsvc := services.NewLocationService(nominatimURL, c)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", HandleHealth(c, weatherURL, nominatimURL))
		r.Get("/weather", HandleGetWeather(wsvc))
		r.Get("/weather/{city}", HandleGetWeatherByCity(wsvc, lsvc))
		r.Get("/location/search", HandleSearchLocation(lsvc))
		r.Get("/location/reverse", HandleReverseGeocode(lsvc))
		r.Get("/insights", HandleGetInsights(wsvc, lsvc))
		r.Get("/history", HandleGetHistory())
		r.Get("/stats", HandleGetStats())
	})

	return r
}

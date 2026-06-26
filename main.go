package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"weather_loc_service/api"
	"weather_loc_service/cache"
	"weather_loc_service/database"
	"weather_loc_service/logger"
)

var (
	serverPort    = getEnv("SERVER_PORT", "8080")
	dbHost        = getEnv("DB_HOST", "localhost")
	dbPort        = getEnv("DB_PORT", "1433")
	dbUser        = getEnv("DB_USER", "sa")
	dbPassword    = getEnv("DB_PASSWORD", "sqlPass!223!!")
	dbName        = getEnv("DB_NAME", "weather_loc_service")
	redisAddr     = getEnv("REDIS_ADDR", "localhost:6379")
	redisPwd      = getEnv("REDIS_PASSWORD", "")
	weatherAPIURL = getEnv("WEATHER_API_URL", "https://api.open-meteo.com")
	nominatimURL  = getEnv("NOMINATIM_URL", "https://nominatim.openstreetmap.org")
)

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func main() {
	logger.InitLogger()
	logger.Log.Info().Msg("starting weather & location insights service")

	err := database.ConnectDB(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to connect to database")
		logger.Log.Info().Msg("running without database - history features wont work")
	} else {
		defer database.CloseDB()
		logger.Log.Info().Msg("connected to MSSQL database")
	}

	c := cache.NewCacheService(redisAddr, redisPwd, 0)
	if c != nil {
		defer c.Close()
	}

	r := api.SetupRouter(weatherAPIURL, nominatimURL, c)

	srv := &http.Server{
		Addr:         ":" + serverPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Log.Info().Msgf("server listening on :%s", serverPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error().Err(err).Msg("server error")
			os.Exit(1)
		}
	}()

	<-quit
	fmt.Println()
	logger.Log.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error().Err(err).Msg("server forced to shutdown")
	}

	logger.Log.Info().Msg("server stopped")
}

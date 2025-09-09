//go:build !test
// +build !test

package main

import (
	"context"
	"fmt"
	"log"
	"main-service/cache"
	"main-service/events"
	"main-service/handlers"
	"main-service/store"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	port := getenv("PORT", "8080")
	pgDSN := getenv("POSTGRES_DSN", "postgres://postgres:postgres@localhost:5432/app?sslmode=disable")
	redisAddr := getenv("REDIS_ADDR", "localhost:6379")
	redisPwd := os.Getenv("REDIS_PASSWORD")
	redisDB := atoi(getenv("REDIS_DB", "0"))
	brokers := splitAndTrim(getenv("KAFKA_BROKERS", "localhost:9092"))
	clientID := getenv("KAFKA_CLIENT_ID", "main-service")

	// --- init real dependencies ---
	store, closeDB, err := store.NewPostgresStore(ctx, pgDSN, 10)
	if err != nil {
		log.Fatalf("postgres init: %v", err)
	}
	defer func() {
		log.Println("closing Postgres connection")
		_ = closeDB(context.Background())
	}()

	cache, closeRedis := cache.NewRedisCache(ctx, redisAddr, redisPwd, redisDB)
	defer func() {
		log.Println("closing Redis cache")
		_ = closeRedis()
	}()

	producer, closeKafka, err := events.NewKafkaProducer(brokers, clientID)
	if err != nil {
		log.Fatalf("kafka init: %v", err)
	}
	defer func() {
		log.Println("closing Kafka producer")
		_ = closeKafka()
	}()

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	h := handlers.NewItemHandler(store, cache, producer)
	r.Post("/items", h.CreateItem)

	srv := &http.Server{Addr: ":" + port, Handler: r, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("main-service listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()
	log.Println("stopping server")
	_ = srv.Shutdown(shutdownCtx)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
func atoi(s string) int { var n int; _, _ = fmt.Sscan(s, &n); return n }
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

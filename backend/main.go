package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq" // registers the "postgres" driver
)

func main() {
	port := env("PORT", "8080")
	dsn := env("POSTGRES_URL", "postgres://tasker:tasker@localhost:5432/tasker?sslmode=disable")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	waitForDB(db)

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           (&Server{store: newPGStore(db)}).Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Shut down cleanly on SIGTERM (what Kubernetes sends when stopping a pod).
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	log.Printf("tas-ker backend listening on :%s", port)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

// waitForDB retries for up to ~30s so the API can start before Postgres is ready.
func waitForDB(db *sql.DB) {
	for i := 1; i <= 15; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := db.PingContext(ctx)
		cancel()
		if err == nil {
			return
		}
		log.Printf("waiting for database (%d/15): %v", i, err)
		time.Sleep(2 * time.Second)
	}
	log.Fatal("database never became ready")
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.RequestURI(), time.Since(start).Round(time.Microsecond))
	})
}

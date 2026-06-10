package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/huguescodeur/oz-rest-api-go/internal/app"
	"github.com/huguescodeur/oz-rest-api-go/internal/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// @title           o'z api
// @version         1.0
// @termsOfService  http://swagger.io/terms/
// @description     API de gestion de commerce multi-tenants et multi-boutiques/magasins construite en Go.

// @contact.name   Support API
// @contact.email  goliyaohugues@gmail.com

//@host            localhost:8080
// @BasePath        /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 Type "Bearer" suivi d'un espace et du token JWT.
func main() {
	config.LoadConfig()

	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASS"),
			os.Getenv("DB_ADDR"),
			os.Getenv("DB_NAME"),
		)
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("impossible de créer le pool de connexions: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("impossible de joindre la base de données: %v", err)
	}

	myApp := app.Init(pool)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      myApp.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Server start", "addr", "http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Arrêt du serveur...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("arrêt forcé: %v", err)
	}
	slog.Info("Serveur arrêté proprement")
}

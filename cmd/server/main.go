package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/panhao/url-shortener/internal/config"
	"github.com/panhao/url-shortener/internal/handler"
	mw "github.com/panhao/url-shortener/internal/middleware"
	"github.com/panhao/url-shortener/internal/model"
	"github.com/panhao/url-shortener/internal/repository"
	"github.com/panhao/url-shortener/internal/service"
)

const clickBufferSize = 10000

func main() {
	cfg := config.Load()

	// Database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := repository.Connect(ctx, cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := repository.RunMigrations(ctx, pool); err != nil {
		log.Printf("Warning: migration error (may be ok if tables exist): %v", err)
	}

	// Repositories
	linkRepo := repository.NewLinkRepo(pool)
	clickLogRepo := repository.NewClickLogRepo(pool)
	userRepo := repository.NewUserRepo(pool)

	// Services
	codeSvc := service.NewShortCodeService(linkRepo)
	linkSvc := service.NewLinkService(linkRepo, codeSvc, cfg.BaseURL)
	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpireHours)

	// Click log pipeline
	clickChan := make(chan model.ClickLog, clickBufferSize)
	go clickLogWorker(clickChan, clickLogRepo)

	// Handlers
	redirectH := handler.NewRedirectHandler(linkSvc, clickChan)
	linkH := handler.NewLinkHandler(linkSvc)
	authH := handler.NewAuthHandler(authSvc)
	dashboardH := handler.NewDashboardHandler(linkSvc)

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RealIP)
	r.Use(mw.Logger)
	r.Use(mw.Recovery)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(mw.RateLimitGlobal(cfg.RateLimitGlobal))

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Public endpoints
		r.Post("/links", linkH.Create)
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/login", authH.Login)

		// Optional auth: create link with user context if token provided
		r.With(mw.AuthOptional(authSvc)).Post("/links", linkH.Create)

		// Authenticated endpoints
		r.Group(func(r chi.Router) {
			r.Use(mw.AuthRequired(authSvc))

			r.Get("/links", linkH.List)
			r.Get("/links/{code}", linkH.Get)
			r.Patch("/links/{code}", linkH.Update)
			r.Delete("/links/{code}", linkH.Delete)
			r.Get("/links/{code}/stats", linkH.GetStats)

			r.Get("/auth/me", authH.Me)

			r.Get("/dashboard/overview", dashboardH.Overview)
			r.Get("/dashboard/trend", dashboardH.Trend)
			r.Get("/dashboard/geo", dashboardH.Geo)
			r.Get("/dashboard/devices", dashboardH.Devices)
		})
	})

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Redirect handler — must be last (catch-all for short codes)
	r.Get("/{shortCode}", redirectH.ServeRedirect)

	// Server
	addr := cfg.ServerHost + ":" + cfg.ServerPort
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down gracefully...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	log.Printf("Server starting on %s", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
	log.Println("Server stopped")
}

// clickLogWorker consumes click logs from the channel and batch-inserts them.
func clickLogWorker(ch <-chan model.ClickLog, repo *repository.ClickLogRepo) {
	batch := make([]model.ClickLog, 0, 100)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case click, ok := <-ch:
			if !ok {
				// Channel closed, flush remaining
				if len(batch) > 0 {
					flushBatch(repo, batch)
				}
				return
			}
			batch = append(batch, click)
			if len(batch) >= 100 {
				flushBatch(repo, batch)
				batch = make([]model.ClickLog, 0, 100)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				flushBatch(repo, batch)
				batch = make([]model.ClickLog, 0, 100)
			}
		}
	}
}

func flushBatch(repo *repository.ClickLogRepo, batch []model.ClickLog) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := repo.BatchInsert(ctx, batch); err != nil {
		log.Printf("Error inserting click batch: %v", err)
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/course/backend-go/internal/handler"
	"github.com/course/backend-go/internal/telemetry"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// newRelicMiddleware adalah chi middleware yang:
//  1. Membuat New Relic Transaction per request (instrumen transaction)
//  2. Menamai transaction sesuai route pattern dari chi.RouteContext
//  3. Menyimpan txn ke request context agar handler bisa mengambilnya
//  4. Melaporkan panic sebagai NR noticed error
//  5. Mengakhiri transaction setelah response selesai ditulis
func newRelicMiddleware(app *newrelic.Application) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if app == nil {
				next.ServeHTTP(w, r)
				return
			}

			// Nama transaction: "<METHOD> <path_pattern>"
			// Setelah chi selesai routing, RouteContext().RoutePattern() berisi pola (mis: /api/v1/hello).
			// Karena middleware berjalan sebelum routing, kita gunakan r.URL.Path sebagai fallback.
			// Nilai ini akan kita update setelah chi routing selesai via chi.RouteContext.
			txnName := r.Method + " " + r.URL.Path
			txn := app.StartTransaction(txnName)
			defer txn.End()

			// Inject header W3C TraceContext agar distributed tracing bekerja
			// antara frontend, backend-go, dan external services
			txn.SetWebRequestHTTP(r)

			// Wrap ResponseWriter agar NR bisa track status code dan response size
			ww := txn.SetWebResponse(w)

			// Inject transaction ke context agar handler bisa pakai newrelic.FromContext(r.Context())
			ctx := newrelic.NewContext(r.Context(), txn)
			next.ServeHTTP(ww, r.WithContext(ctx))

			// Update nama transaction dengan route pattern setelah chi routing (lebih spesifik)
			if rctx := chi.RouteContext(r.Context()); rctx != nil && rctx.RoutePattern() != "" {
				txn.SetName(r.Method + " " + rctx.RoutePattern())
			}
		})
	}
}

func main() {
	// ── 1. Inisialisasi New Relic agent ─────────────────────────────────────
	if err := telemetry.Init(); err != nil {
		log.Printf("[newrelic] warning: %v — server tetap start\n", err)
	}
	defer telemetry.Shutdown()

	// ── 2. Port ─────────────────────────────────────────────────────────────
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// ── 3. Router ───────────────────────────────────────────────────────────
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// New Relic middleware — instrument setiap request sebagai NR Transaction
	r.Use(newRelicMiddleware(telemetry.App))

	// ── 4. Routes ───────────────────────────────────────────────────────────
	r.Get("/healthz", handler.Healthz)
	r.Get("/api/v1/hello", handler.Hello)
	r.Get("/api/v1/version", handler.Version)

	// ── 5. HTTP Server ──────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		log.Printf("Server running on port %s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v\n", err)
		}
	}()

	// ── 6. Graceful shutdown ─────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println("\nShutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	fmt.Println("Server stopped.")
}

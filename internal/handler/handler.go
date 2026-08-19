// Package handler berisi HTTP handler untuk backend-go.
// Setiap handler diinstrumentasi ke New Relic:
//   - Transaction: otomatis oleh nrchi middleware (wrap di main.go)
//   - Segment function: instrument seluruh durasi satu fungsi
//   - Nested segment: segment di dalam segment (hirarki call)
//   - Block of code segment: instrument blok kode tertentu di dalam fungsi
//   - External service segment: instrument panggilan HTTP ke third-party API
//   - Datastore segment: instrument query database
package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// ─── Helper ────────────────────────────────────────────────────────────────

// txnFromRequest mengekstrak New Relic Transaction dari request context.
// Mengembalikan nil-safe wrapper jika NR tidak aktif.
func txnFromRequest(r *http.Request) *newrelic.Transaction {
	return newrelic.FromContext(r.Context())
}

// ─── Healthz ───────────────────────────────────────────────────────────────

// Healthz — health check endpoint untuk K8s liveness/readiness.
// Tidak diinstrumentasi agar tidak mencemari apdex score dengan health poll.
func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ─── Hello ─────────────────────────────────────────────────────────────────

// Hello — contoh endpoint dengan instrumentasi lengkap:
//  1. Segment function (span seluruh proses bisnis)
//  2. Nested segment (span sub-proses di dalam proses utama)
//  3. Block of code segment (span sepotong kode spesifik)
//  4. External service segment (simulasi call ke third-party API)
//  5. Datastore segment (simulasi query DB)
func Hello(w http.ResponseWriter, r *http.Request) {
	txn := txnFromRequest(r)

	// ── 1. Function Segment ─────────────────────────────────────────────────
	// Instrument seluruh fungsi sebagai satu segment.
	// startFunctionSegment melacak nama fungsi + file melalui runtime.
	funcSeg := newrelic.StartSegment(txn, "handler.Hello/business-logic")
	defer funcSeg.End() // segment berakhir saat fungsi return

	// ── 2. Block of Code Segment ────────────────────────────────────────────
	// Instrument blok kode tertentu (mis: validasi input, enrichment).
	{
		blkSeg := newrelic.StartSegment(txn, "handler.Hello/input-validation")
		// Simulasi validasi input — bisa be any synchronous block
		_ = r.URL.Query().Get("lang")
		blkSeg.End()
	}

	// ── 3. Datastore Segment ────────────────────────────────────────────────
	// Instrument panggilan ke database (simulasi, tanpa koneksi nyata).
	// Jika sudah ada *sql.DB, gunakan nrsql / nrpgx wrapper — lihat catatan di bawah.
	greeting := fetchGreetingFromDB(txn)

	// ── 4. Nested Segment ───────────────────────────────────────────────────
	// Nested: segment di dalam segment — hirarki Business.
	// "enrichment" berjalan di dalam "business-logic" yang sudah terbuka.
	enrichSeg := newrelic.StartSegment(txn, "handler.Hello/enrichment")
	{
		// Nested lebih dalam: format message
		fmtSeg := newrelic.StartSegment(txn, "handler.Hello/enrichment/format-message")
		greeting = fmt.Sprintf("%s 🚀", greeting)
		fmtSeg.End()
	}
	enrichSeg.End()

	// ── 5. External Service Segment ─────────────────────────────────────────
	// Instrument panggilan ke external HTTP service (mis: lookup versi build).
	buildVersion := callExternalVersionAPI(txn, r)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": greeting,
		"version": buildVersion,
	})
}

// fetchGreetingFromDB mensimulasikan query DB dengan Datastore Segment.
//
// Pada implementasi nyata, gunakan salah satu:
//   - nrsql.InstrumentSQLConnector(connector, nrAgent) — untuk database/sql
//   - github.com/newrelic/go-agent/v3/integrations/nrpgx5 — untuk pgx v5
//   - github.com/newrelic/go-agent/v3/integrations/nrmysql  — untuk MySQL
//
// Di sini kita instrument manual agar kode standalone (tanpa DB nyata).
func fetchGreetingFromDB(txn *newrelic.Transaction) string {
	seg := newrelic.DatastoreSegment{
		StartTime:          txn.StartSegmentNow(),
		Product:            newrelic.DatastorePostgres,
		Collection:         "greetings",
		Operation:          "SELECT",
		ParameterizedQuery: "SELECT message FROM greetings WHERE lang = $1",
		DatabaseName:       "app_db",
		Host:               os.Getenv("DB_HOST"),
		PortPathOrID:       "5432",
	}
	defer seg.End()

	// Simulasi latency query
	time.Sleep(1 * time.Millisecond)

	// Pada implementasi nyata:
	// row := db.QueryRowContext(ctx, "SELECT message FROM greetings WHERE lang = $1", lang)
	// row.Scan(&greeting)
	_ = (*sql.DB)(nil) // import sql agar tidak unused

	return "Hello from Be A DevOps Employee course!"
}

// callExternalVersionAPI mensimulasikan panggilan ke external HTTP API
// menggunakan ExternalSegment sehingga New Relic dapat melacak:
//   - URL tujuan
//   - Metode HTTP
//   - Durasi call
//   - Status response
func callExternalVersionAPI(txn *newrelic.Transaction, r *http.Request) string {
	externalURL := "https://api.example.com/version"

	// ExternalSegment — cara manual (cocok untuk custom HTTP client)
	seg := &newrelic.ExternalSegment{
		StartTime: txn.StartSegmentNow(),
		URL:       externalURL,
		Procedure: http.MethodGet,
		Library:   "net/http",
	}
	defer seg.End()

	// Cara alternatif dengan newrelic.NewRoundTripper (otomatis wrap semua request):
	// client := &http.Client{Transport: newrelic.NewRoundTripper(txn.Application())}
	// resp, err := client.Get(externalURL)
	// ── Simulasi (skip real HTTP call agar tidak butuh network di test) ──
	_ = url.QueryEscape(externalURL) // dummy usage

	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "dev"
	}
	return version
}

// ─── Version ───────────────────────────────────────────────────────────────

// Version — menampilkan versi build.
// Transaction sudah diinstrumentasi oleh nrchi middleware (otomatis).
// Tambahan: custom attribute agar bisa di-filter di NR query.
func Version(w http.ResponseWriter, r *http.Request) {
	txn := txnFromRequest(r)

	// Tambahkan custom attribute ke transaction ini
	if txn != nil {
		txn.AddAttribute("endpoint", "version")
		txn.AddAttribute("app_env", os.Getenv("APP_ENV"))
	}

	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "dev"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version": version,
	})
}

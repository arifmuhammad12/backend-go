package telemetry

import (
	"fmt"
	"log"
	"os"

	"github.com/newrelic/go-agent/v3/newrelic"
)

// App adalah singleton New Relic application instance yang di-share ke seluruh aplikasi.
var App *newrelic.Application

// Init menginisialisasi New Relic agent.
// Dipanggil sekali di main() sebelum server start.
// Env vars yang dibutuhkan:
//   - NEW_RELIC_LICENSE_KEY  — license key New Relic (required)
//   - NEW_RELIC_APP_NAME     — nama aplikasi di New Relic UI (default: backend-go)
//   - NEW_RELIC_LOG_LEVEL    — level log NR agent: "info" / "debug" / "off" (default: "info")
func Init() error {
	licenseKey := os.Getenv("NEW_RELIC_LICENSE_KEY")
	if licenseKey == "" {
		log.Println("[newrelic] NEW_RELIC_LICENSE_KEY kosong — instrumentasi dinonaktifkan")
		return nil // tidak fatal; app tetap jalan tanpa NR
	}

	appName := os.Getenv("NEW_RELIC_APP_NAME")
	if appName == "" {
		appName = "backend-go"
	}

	logLevel := os.Getenv("NEW_RELIC_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(appName),
		newrelic.ConfigLicense(licenseKey),
		newrelic.ConfigAppLogEnabled(true),
		newrelic.ConfigAppLogForwardingEnabled(true),
		newrelic.ConfigDistributedTracerEnabled(true),
		newrelic.ConfigCodeLevelMetricsEnabled(true),
		func(cfg *newrelic.Config) {
			cfg.Logger = newrelic.NewLogger(log.Writer())
		},
	)
	if err != nil {
		return fmt.Errorf("gagal inisialisasi New Relic: %w", err)
	}

	App = app
	log.Printf("[newrelic] agent aktif — app_name=%s\n", appName)
	return nil
}

// Shutdown menutup agent NR dengan flush semua pending data.
// Dipanggil di defer setelah graceful shutdown server.
func Shutdown() {
	if App != nil {
		App.Shutdown(5e9) // 5 detik timeout
	}
}

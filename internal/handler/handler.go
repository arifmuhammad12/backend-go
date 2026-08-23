package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// Build & Git metadata — diinjeksi saat build via ldflags -X atau runtime env
var (
	BuildNumber   = "local"
	CommitSHA     = "dev"
	CommitMessage = "local development"
)

func init() {
	if val := os.Getenv("BUILD_NUMBER"); val != "" {
		BuildNumber = val
	}
	if val := os.Getenv("COMMIT_SHA"); val != "" {
		CommitSHA = val
	}
	if val := os.Getenv("COMMIT_MESSAGE"); val != "" {
		CommitMessage = val
	}
}

// ─── Healthz ───────────────────────────────────────────────────────────────

type HealthzResponse struct {
	Status        string `json:"status"`
	BuildNumber   string `json:"build_number"`
	CommitSHA     string `json:"commit_sha"`
	CommitMessage string `json:"commit_message"`
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthzResponse{
		Status:        "ok",
		BuildNumber:   BuildNumber,
		CommitSHA:     CommitSHA,
		CommitMessage: CommitMessage,
	})
}

// ─── Hello ─────────────────────────────────────────────────────────────────

func Hello(w http.ResponseWriter, r *http.Request) {
	greeting := fetchGreetingFromDB()
	greeting = fmt.Sprintf("%s 🚀", greeting)

	buildVersion := callExternalVersionAPI(r)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": greeting,
		"version": buildVersion,
	})
}

// fetchGreetingFromDB mensimulasikan query DB
func fetchGreetingFromDB() string {
	return "Hello from Be A DevOps Employee course!"
}

// callExternalVersionAPI mensimulasikan panggilan ke external HTTP API
func callExternalVersionAPI(r *http.Request) string {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "dev"
	}
	return version
}

// ─── Version ───────────────────────────────────────────────────────────────

func Version(w http.ResponseWriter, r *http.Request) {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "dev"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version": version,
	})
}

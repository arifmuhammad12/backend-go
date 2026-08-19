# apps/backend-go — Mini Backend Go (Contoh Course)

Aplikasi backend sederhana berbahasa Go. Tanpa database — fokus demonstrasi alur CI, build, deploy, dan observability.

## Endpoint
- `GET /healthz` — health check (K8s liveness/readiness)
- `GET /api/v1/hello` — contoh response
- `GET /api/v1/version` — versi build (dari env `APP_VERSION`)

## Run lokal
```bash
make run        # default port 8080
# atau: APP_PORT=3000 go run main.go
```

## Test
```bash
make test       # go test -v -cover ./...
```

## Build
```bash
make build      # output: bin/backend-go
make docker-build  # docker image lokal
```

## CI
- `Jenkinsfile` memanggil shared library (`containerPipeline`).
- `.cicd/pipeline.yaml` berisi konfigurasi CI (app_name, registry, gitops, build tool, dll).
- Build tool: Kaniko (rootless). Tag image: git short SHA.

## Pola dari produksi
Terinspirasi `dc-cinema-service-production` (layered architecture + chi router + New Relic + graceful shutdown). Disederhanakan: tanpa database, tanpa Swagger, tanpa repo/service layer — fokus DevOps pipeline bukan business logic.

Instrumentasi New Relic ditambahkan di Epic 7 (Observability).
Semangat!

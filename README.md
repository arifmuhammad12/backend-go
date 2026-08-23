# `backend-go` — Microservice Sample Application

Aplikasi microservice berbasis **Go 1.22** dan **Chi Router** yang dirancang sebagai target demonstrasi pipeline CI/CD Jenkins, GitOps ArgoCD, dan Envoy Gateway API.

---

## 🚀 Endpoint API

| Method | Endpoint | Deskripsi | Contoh Response |
|---|---|---|---|
| `GET` | `/healthz` | Health check probe & metadata build CI | `{"status":"ok","build_number":"12","commit_sha":"a1b2c3d","commit_message":"feat: update handler"}` |
| `GET` | `/api/v1/hello` | Contoh pesan respon mikroservis | `{"message":"Hello from Be A DevOps Employee course! 🚀","version":"dev"}` |
| `GET` | `/api/v1/version` | Versi aplikasi runtime | `{"version":"dev"}` |

---

## 🧩 Metadata Build & Injeksi Compile-Time

Binary Go di-compile menggunakan flag `-ldflags -X` yang menginjeksi variabel metadata secara permanen ke dalam struct handler saat proses build di Jenkins:

- `BuildNumber`: Diambil dari `$BUILD_NUMBER` Jenkins.
- `CommitSHA`: Diambil dari `git rev-parse --short HEAD`.
- `CommitMessage`: Diambil dari `git log -1 --pretty=%s`.

Hal ini memudahkan pembuktian bahwa perubahan commit di Git benar-benar ter-deploy hingga ke container yang berjalan.

---

## 🛠️ Menjalankan Aplikasi Secara Lokal

```bash
# 1. Jalankan unit test
go test -v ./...
# atau: make test

# 2. Jalankan server lokal (default port 4000)
go run main.go
# atau: make run

# 3. Build binary lokal
make build
```

---

## 🐳 Docker Build

```bash
# Build multi-stage image dengan custom build args
docker build \
  --build-arg BUILD_NUMBER=1 \
  --build-arg COMMIT_SHA=local-dev \
  --build-arg "COMMIT_MESSAGE=test build" \
  -t backend-go:1 .

# Jalankan container
docker run -d -p 4000:4000 --name backend-go backend-go:1

# Uji health check
curl -s http://localhost:4000/healthz | jq .
```

package cmd

import (
	"path/filepath"
	"testing"
)

func TestRunVerifyConfig(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=8080\nSHUTDOWN_TIMEOUT=10s\nLOG_LEVEL=info\nREQUEST_TIMEOUT=5s\nHTTP_READ_HEADER_TIMEOUT=2s\nHTTP_READ_TIMEOUT=10s\nHTTP_WRITE_TIMEOUT=15s\nHTTP_IDLE_TIMEOUT=60s\nMAX_HEADER_BYTES=1048576\nDEFAULT_PAGE_SIZE=10\nMAX_PAGE_SIZE=100\nBODY_LIMIT_BYTES=1048576\nRATE_LIMIT_PER_SECOND=20\nCORS_ALLOW_ORIGINS=*\nAUTH_TOKEN=dev-token\nDATABASE_URL=postgres://postgres:postgres@localhost:5432/app?sslmode=disable\nDATABASE_TIMEOUT=3s\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"), `package http

import "net/http"

func withMiddlewares(next http.Handler) http.Handler {
	handler := next
	handler = corsMiddleware(handler)
	handler = timeoutMiddleware(handler)
	handler = rateLimitMiddleware(handler)
	return handler
}
`)
	writeFile(t, filepath.Join(projectDir, "internal", "http", "router.go"), `package http

import "net/http"

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login", nil)
	mux.HandleFunc("/db/readyz", nil)
	return mux
}
`)
	writeFile(t, filepath.Join(projectDir, "internal", "http", "timeout.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "ratelimit.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "cors.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "auth.go"), "package http\n")
	writeFile(t, filepath.Join(projectDir, "internal", "store", "postgres.go"), "package store\n")

	if err := runVerifyConfig([]string{projectDir}); err != nil {
		t.Fatalf("配置校验本应通过: %v", err)
	}
}

func TestRunVerifyConfigInvalid(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=70000\nSHUTDOWN_TIMEOUT=0s\nLOG_LEVEL=verbose\nHTTP_READ_HEADER_TIMEOUT=2s\nHTTP_READ_TIMEOUT=10s\nHTTP_WRITE_TIMEOUT=15s\nHTTP_IDLE_TIMEOUT=60s\nMAX_HEADER_BYTES=1048576\nDEFAULT_PAGE_SIZE=10\nMAX_PAGE_SIZE=100\nBODY_LIMIT_BYTES=1048576\n")

	err := runVerifyConfig([]string{projectDir})
	if err == nil {
		t.Fatal("配置校验本应失败")
	}
}

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDoctorFix(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "router.go"), `package http

import (
	"encoding/json"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	// Golider 路由扩展锚点
	return withMiddlewares(mux)
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
`)
	writeFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"), `package http

import "net/http"

func withMiddlewares(next http.Handler) http.Handler {
	handler := next
	// Golider 中间件扩展锚点
	handler = requestLogMiddleware(handler)
	handler = recoverMiddleware(handler)
	return handler
}

func requestLogMiddleware(next http.Handler) http.Handler { return next }

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	})
}
`)

	if err := runDoctor([]string{"fix", projectDir}); err != nil {
		t.Fatalf("执行 doctor fix 失败: %v", err)
	}

	for _, path := range []string{
		".env.example",
		".gitignore",
		"Dockerfile",
		filepath.Join(".github", "workflows", "ci.yml"),
		filepath.Join("internal", "http", "errors.go"),
		filepath.Join("internal", "http", "requestid.go"),
		filepath.Join("internal", "http", "timeout.go"),
		filepath.Join("internal", "http", "metrics.go"),
		filepath.Join("internal", "http", "ratelimit.go"),
		filepath.Join("internal", "http", "cors.go"),
	} {
		if _, err := os.Stat(filepath.Join(projectDir, path)); err != nil {
			t.Fatalf("doctor fix 未生成 %s: %v", path, err)
		}
	}

	middleware := readFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"))
	for _, fragment := range []string{
		"handler = corsMiddleware(handler)",
		"handler = requestIDMiddleware(handler)",
		"handler = timeoutMiddleware(handler)",
		"handler = metricsMiddleware(handler)",
		"handler = rateLimitMiddleware(handler)",
		"writeError(w, r, http.StatusInternalServerError, \"internal_server_error\", \"internal server error\")",
	} {
		if !strings.Contains(middleware, fragment) {
			t.Fatalf("middleware.go 缺少补丁片段 %q: %s", fragment, middleware)
		}
	}

	router := readFile(t, filepath.Join(projectDir, "internal", "http", "router.go"))
	if !strings.Contains(router, "mux.HandleFunc(\"/metrics\", metricsHandler)") {
		t.Fatalf("router.go 未注入 metrics 路由: %s", router)
	}

	envFile := readFile(t, filepath.Join(projectDir, ".env.example"))
	for _, fragment := range []string{
		"REQUEST_TIMEOUT=5s",
		"RATE_LIMIT_PER_SECOND=20",
		"CORS_ALLOW_ORIGINS=*",
	} {
		if !strings.Contains(envFile, fragment) {
			t.Fatalf(".env.example 缺少片段 %q: %s", fragment, envFile)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取文件失败: %v", err)
	}
	return string(content)
}

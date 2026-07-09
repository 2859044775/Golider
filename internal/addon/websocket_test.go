package addon

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallWebsocketModule(t *testing.T) {
	projectDir := t.TempDir()
	writeFile(t, filepath.Join(projectDir, "go.mod"), "module github.com/acme/demo\n\ngo 1.20\n")
	writeFile(t, filepath.Join(projectDir, ".env.example"), "PORT=8080\n")
	writeFile(t, filepath.Join(projectDir, "internal", "http", "router.go"), `package http

import "net/http"

func registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/echo", echoHandler)
	// Golider 路由扩展锚点
}
`)

	err := Install(Options{
		ModuleName: "websocket",
		TargetDir:  projectDir,
	})
	if err != nil {
		t.Fatalf("安装 websocket 模块失败: %v", err)
	}

	routerFile := readFile(t, filepath.Join(projectDir, "internal", "http", "router.go"))
	if !strings.Contains(routerFile, `mux.HandleFunc("/ws", websocketHandler)`) {
		t.Fatalf("router.go 未注入 websocket 路由: %s", routerFile)
	}

	wsFile := readFile(t, filepath.Join(projectDir, "internal", "http", "websocket.go"))
	if !strings.Contains(wsFile, "func websocketHandler") {
		t.Fatalf("websocket.go 未生成: %s", wsFile)
	}
	if !strings.Contains(wsFile, "type Hub struct") {
		t.Fatalf("websocket.go 缺少 Hub 定义: %s", wsFile)
	}
	if !strings.Contains(wsFile, "type Client struct") {
		t.Fatalf("websocket.go 缺少 Client 定义: %s", wsFile)
	}
	if !strings.Contains(wsFile, "BroadcastToRoom") {
		t.Fatalf("websocket.go 缺少 BroadcastToRoom 函数: %s", wsFile)
	}
}

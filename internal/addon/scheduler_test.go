package addon

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallSchedulerModule(t *testing.T) {
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
	writeFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"), `package http

func applyMiddlewares(handler http.Handler) http.Handler {
	// Golider 中间件扩展锚点
	return handler
}
`)
	writeFile(t, filepath.Join(projectDir, "cmd", "api", "main.go"), `package main

import (
	"os"

	"github.com/acme/demo/internal/app"
)

func main() {
	lifecycle := app.New()
	_ = os.Getenv("PORT")
}
`)

	err := Install(Options{
		ModuleName: "scheduler",
		TargetDir:  projectDir,
	})
	if err != nil {
		t.Fatalf("安装 scheduler 模块失败: %v", err)
	}

	// 验证调度器核心文件
	schedFile := readFile(t, filepath.Join(projectDir, "internal", "scheduler", "scheduler.go"))
	if !strings.Contains(schedFile, "type Scheduler struct") {
		t.Fatalf("scheduler.go 缺少 Scheduler 定义: %s", schedFile)
	}
	if !strings.Contains(schedFile, "func (s *Scheduler) Register") {
		t.Fatalf("scheduler.go 缺少 Register 方法: %s", schedFile)
	}
	if !strings.Contains(schedFile, "func (s *Scheduler) List") {
		t.Fatalf("scheduler.go 缺少 List 方法: %s", schedFile)
	}
	if !strings.Contains(schedFile, "func (s *Scheduler) Trigger") {
		t.Fatalf("scheduler.go 缺少 Trigger 方法: %s", schedFile)
	}
	if !strings.Contains(schedFile, "func parseSchedule") {
		t.Fatalf("scheduler.go 缺少 parseSchedule 函数: %s", schedFile)
	}

	// 验证生命周期文件
	lifecycleFile := readFile(t, filepath.Join(projectDir, "internal", "scheduler", "lifecycle.go"))
	if !strings.Contains(lifecycleFile, "func (s *Scheduler) StartHook()") {
		t.Fatalf("lifecycle.go 缺少 StartHook 方法: %s", lifecycleFile)
	}
	if !strings.Contains(lifecycleFile, "func (s *Scheduler) StopHook()") {
		t.Fatalf("lifecycle.go 缺少 StopHook 方法: %s", lifecycleFile)
	}

	// 验证 HTTP 端点文件
	httpFile := readFile(t, filepath.Join(projectDir, "internal", "http", "scheduler.go"))
	if !strings.Contains(httpFile, "SchedulerInstance") {
		t.Fatalf("scheduler.go (http) 缺少 SchedulerInstance: %s", httpFile)
	}
	if !strings.Contains(httpFile, "schedulerListHandler") {
		t.Fatalf("scheduler.go (http) 缺少 schedulerListHandler: %s", httpFile)
	}
	if !strings.Contains(httpFile, "schedulerTriggerHandler") {
		t.Fatalf("scheduler.go (http) 缺少 schedulerTriggerHandler: %s", httpFile)
	}

	// 验证路由注入
	routerFile := readFile(t, filepath.Join(projectDir, "internal", "http", "router.go"))
	if !strings.Contains(routerFile, `mux.HandleFunc("/scheduler/tasks", schedulerListHandler)`) {
		t.Fatalf("router.go 未注入 scheduler tasks 路由: %s", routerFile)
	}
	if !strings.Contains(routerFile, `mux.HandleFunc("/scheduler/trigger/", schedulerTriggerHandler)`) {
		t.Fatalf("router.go 未注入 scheduler trigger 路由: %s", routerFile)
	}

	// 验证 main.go 生命周期注入
	mainFile := readFile(t, filepath.Join(projectDir, "cmd", "api", "main.go"))
	if !strings.Contains(mainFile, "scheduler.New()") {
		t.Fatalf("main.go 未注入 scheduler.New(): %s", mainFile)
	}
	if !strings.Contains(mainFile, "sched.StartHook()") {
		t.Fatalf("main.go 未注入 sched.StartHook(): %s", mainFile)
	}
	if !strings.Contains(mainFile, "sched.StopHook()") {
		t.Fatalf("main.go 未注入 sched.StopHook(): %s", mainFile)
	}
	if !strings.Contains(mainFile, "http.SchedulerInstance = sched") {
		t.Fatalf("main.go 未注入 SchedulerInstance 赋值: %s", mainFile)
	}

	// 验证日志基础文件已生成
	loggerFile := readFile(t, filepath.Join(projectDir, "internal", "observability", "logger.go"))
	if !strings.Contains(loggerFile, "func New(") {
		t.Fatalf("logger.go 未生成: %s", loggerFile)
	}
}

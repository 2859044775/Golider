package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateProjectIncludesLifecycleAndValidation(t *testing.T) {
	projectDir := filepath.Join(t.TempDir(), "demo")

	err := CreateProject(Options{
		AppName:     "demo",
		Module:      "github.com/acme/demo",
		TargetDir:   projectDir,
		DefaultPort: "8080",
	})
	if err != nil {
		t.Fatalf("创建项目失败: %v", err)
	}

	for _, path := range []string{
		filepath.Join(projectDir, "internal", "app", "app.go"),
		filepath.Join(projectDir, "internal", "app", "dependencies.go"),
		filepath.Join(projectDir, "internal", "config", "config.go"),
		filepath.Join(projectDir, "cmd", "api", "main.go"),
		filepath.Join(projectDir, "internal", "http", "binding.go"),
		filepath.Join(projectDir, "internal", "http", "binding_test.go"),
		filepath.Join(projectDir, "internal", "http", "errors.go"),
		filepath.Join(projectDir, "internal", "http", "requestid.go"),
		filepath.Join(projectDir, "internal", "http", "readiness.go"),
		filepath.Join(projectDir, "internal", "http", "query.go"),
		filepath.Join(projectDir, "internal", "http", "query_test.go"),
		filepath.Join(projectDir, "internal", "http", "timeout.go"),
		filepath.Join(projectDir, "internal", "http", "router_test.go"),
		filepath.Join(projectDir, "internal", "http", "middleware_test.go"),
		filepath.Join(projectDir, "internal", "repository", "message.go"),
		filepath.Join(projectDir, "internal", "repository", "message_test.go"),
		filepath.Join(projectDir, "internal", "service", "message.go"),
		filepath.Join(projectDir, "internal", "service", "message_test.go"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("缺少生成文件 %s: %v", path, err)
		}
	}

	configFile := readFile(t, filepath.Join(projectDir, "internal", "config", "config.go"))
	if !strings.Contains(configFile, "func Load() (Config, error)") {
		t.Fatalf("config.go 未生成新的 Load 签名: %s", configFile)
	}
	if !strings.Contains(configFile, "func validate(cfg Config) error") {
		t.Fatalf("config.go 未生成配置校验函数: %s", configFile)
	}
	if !strings.Contains(configFile, "LOG_LEVEL 必须是 debug、info、warn、error 之一") {
		t.Fatalf("config.go 未生成日志级别校验: %s", configFile)
	}
	for _, fragment := range []string{
		"ReadHeaderTimeout time.Duration",
		"ReadTimeout       time.Duration",
		"WriteTimeout      time.Duration",
		"IdleTimeout       time.Duration",
		"DefaultPageSize   int",
		"MaxPageSize       int",
		"HTTP_READ_HEADER_TIMEOUT 必须大于 0",
		"HTTP_READ_TIMEOUT 必须大于 0",
		"HTTP_WRITE_TIMEOUT 必须大于 0",
		"HTTP_IDLE_TIMEOUT 必须大于 0",
		"DEFAULT_PAGE_SIZE 必须大于 0",
		"MAX_PAGE_SIZE 不能小于 DEFAULT_PAGE_SIZE",
	} {
		if !strings.Contains(configFile, fragment) {
			t.Fatalf("config.go 缺少服务超时护栏片段 %q: %s", fragment, configFile)
		}
	}

	appFile := readFile(t, filepath.Join(projectDir, "internal", "app", "app.go"))
	if !strings.Contains(appFile, "func (a *App) OnStart") {
		t.Fatalf("app.go 未生成 OnStart: %s", appFile)
	}
	if !strings.Contains(appFile, "func (a *App) OnStop") {
		t.Fatalf("app.go 未生成 OnStop: %s", appFile)
	}

	mainFile := readFile(t, filepath.Join(projectDir, "cmd", "api", "main.go"))
	if !strings.Contains(mainFile, "lifecycle := app.New()") {
		t.Fatalf("main.go 未接入生命周期装配: %s", mainFile)
	}
	if !strings.Contains(mainFile, "cfg, err := config.Load()") {
		t.Fatalf("main.go 未接入配置校验: %s", mainFile)
	}
	if !strings.Contains(mainFile, "httptransport.MarkNotReady(\"shutting_down\")") {
		t.Fatalf("main.go 未接入停机摘流: %s", mainFile)
	}
	for _, fragment := range []string{
		"ReadHeaderTimeout: cfg.ReadHeaderTimeout",
		"ReadTimeout:       cfg.ReadTimeout",
		"WriteTimeout:      cfg.WriteTimeout",
		"IdleTimeout:       cfg.IdleTimeout",
		"deps := app.NewDependencies(cfg)",
		"handler := httptransport.NewRouter(deps)",
	} {
		if !strings.Contains(mainFile, fragment) {
			t.Fatalf("main.go 缺少服务超时配置片段 %q: %s", fragment, mainFile)
		}
	}

	middlewareFile := readFile(t, filepath.Join(projectDir, "internal", "http", "middleware.go"))
	for _, fragment := range []string{
		"handler = requestIDMiddleware(handler)",
		"handler = timeoutMiddleware(handler)",
		"writeError(w, r, http.StatusInternalServerError, \"internal_server_error\", \"internal server error\")",
	} {
		if !strings.Contains(middlewareFile, fragment) {
			t.Fatalf("middleware.go 缺少默认生产能力片段 %q: %s", fragment, middlewareFile)
		}
	}

	envFile := readFile(t, filepath.Join(projectDir, ".env.example"))
	for _, fragment := range []string{
		"REQUEST_TIMEOUT=5s",
		"HTTP_READ_HEADER_TIMEOUT=2s",
		"HTTP_READ_TIMEOUT=10s",
		"HTTP_WRITE_TIMEOUT=15s",
		"HTTP_IDLE_TIMEOUT=60s",
		"DEFAULT_PAGE_SIZE=10",
		"MAX_PAGE_SIZE=100",
	} {
		if !strings.Contains(envFile, fragment) {
			t.Fatalf(".env.example 缺少服务超时配置 %q: %s", fragment, envFile)
		}
	}

	routerTestFile := readFile(t, filepath.Join(projectDir, "internal", "http", "router_test.go"))
	if !strings.Contains(routerTestFile, "TestNewRouterDefaultEndpoints") {
		t.Fatalf("router_test.go 未生成默认接口测试: %s", routerTestFile)
	}
	if !strings.Contains(routerTestFile, "TestReadyHandlerDrainingState") {
		t.Fatalf("router_test.go 未生成就绪摘流测试: %s", routerTestFile)
	}
	for _, fragment := range []string{
		"TestCreateMessageHandlerSuccess",
		"TestCreateMessageHandlerIdempotencyReplay",
		"TestCreateMessageHandlerConflict",
		"TestGetMessageHandlerSuccess",
		"TestUpdateMessageHandlerSuccess",
		"TestUpdateMessageHandlerTransitionConflict",
		"TestDeleteMessageHandlerSuccess",
	} {
		if !strings.Contains(routerTestFile, fragment) {
			t.Fatalf("router_test.go 缺少写接口测试片段 %q: %s", fragment, routerTestFile)
		}
	}

	middlewareTestFile := readFile(t, filepath.Join(projectDir, "internal", "http", "middleware_test.go"))
	for _, fragment := range []string{
		"TestRecoverMiddlewareWritesErrorResponse",
		"TestTimeoutMiddleware",
	} {
		if !strings.Contains(middlewareTestFile, fragment) {
			t.Fatalf("middleware_test.go 缺少测试片段 %q: %s", fragment, middlewareTestFile)
		}
	}

	bindingFile := readFile(t, filepath.Join(projectDir, "internal", "http", "binding.go"))
	for _, fragment := range []string{
		"func decodeJSON",
		"DisallowUnknownFields()",
		"func writeBindingError",
	} {
		if !strings.Contains(bindingFile, fragment) {
			t.Fatalf("binding.go 缺少输入校验片段 %q: %s", fragment, bindingFile)
		}
	}

	routerFile := readFile(t, filepath.Join(projectDir, "internal", "http", "router.go"))
	for _, fragment := range []string{
		"mux.HandleFunc(\"/messages\", func(w http.ResponseWriter, r *http.Request)",
		"mux.HandleFunc(\"/messages/\", func(w http.ResponseWriter, r *http.Request)",
		"func listMessagesHandler",
		"func createMessageHandler",
		"func getMessageHandler",
		"func updateMessageHandler",
		"func deleteMessageHandler",
		"parseListQuery(r, deps.DefaultPageSize, deps.MaxPageSize)",
		"validateCreateMessageRequest",
		"validateUpdateMessageRequest",
		"Idempotency-Key",
		"messageIDFromPath",
		"func NewRouter(deps app.Dependencies) http.Handler",
		"deps = deps.WithDefaults()",
		"mux.HandleFunc(\"/echo\", echoHandler)",
		"func echoHandler",
		"message is required",
		"message title already exists",
		"message status transition is not allowed",
		"\"deleted\":    true",
	} {
		if !strings.Contains(routerFile, fragment) {
			t.Fatalf("router.go 缺少输入校验片段 %q: %s", fragment, routerFile)
		}
	}

	bindingTestFile := readFile(t, filepath.Join(projectDir, "internal", "http", "binding_test.go"))
	for _, fragment := range []string{
		"TestDecodeJSONRejectsUnknownFields",
		"TestDecodeJSONRejectsNonJSONContentType",
	} {
		if !strings.Contains(bindingTestFile, fragment) {
			t.Fatalf("binding_test.go 缺少测试片段 %q: %s", fragment, bindingTestFile)
		}
	}

	queryFile := readFile(t, filepath.Join(projectDir, "internal", "http", "query.go"))
	for _, fragment := range []string{
		"func parseListQuery",
		"func writePaginatedJSON",
		"page_size must be less than or equal to max page size",
		"status must be active or archived",
		"sort_by must be created_at or title",
		"created_from must be earlier than or equal to created_to",
	} {
		if !strings.Contains(queryFile, fragment) {
			t.Fatalf("query.go 缺少查询解析片段 %q: %s", fragment, queryFile)
		}
	}

	queryTestFile := readFile(t, filepath.Join(projectDir, "internal", "http", "query_test.go"))
	for _, fragment := range []string{
		"TestParseListQueryDefaults",
		"TestParseListQueryInvalidPageSize",
		"TestParseListQueryInvalidTimeRange",
	} {
		if !strings.Contains(queryTestFile, fragment) {
			t.Fatalf("query_test.go 缺少测试片段 %q: %s", fragment, queryTestFile)
		}
	}

	serviceFile := readFile(t, filepath.Join(projectDir, "internal", "service", "message.go"))
	for _, fragment := range []string{
		"type MessageService struct",
		"func NewMessageService(repo MessageRepository)",
		"func (s *MessageService) Create",
		"func (s *MessageService) GetByID",
		"func (s *MessageService) Update",
		"func (s *MessageService) Delete",
		"func (s *MessageService) List",
		"type CreateMessageInput struct",
		"type UpdateMessageInput struct",
		"type DeleteMessageOutput struct",
		"type MessageConflictError struct",
		"type IdempotencyConflictError struct",
		"type MessageNotFoundError struct",
		"type MessageStatusTransitionError struct",
		"DeletedAt  *time.Time",
		"ArchivedAt *time.Time",
		"CreatedAt  time.Time",
		"SortBy      string",
		"repo              MessageRepository",
		"func canTransitionMessageStatus",
		"func isDeletedMessage",
	} {
		if !strings.Contains(serviceFile, fragment) {
			t.Fatalf("message.go 缺少服务层片段 %q: %s", fragment, serviceFile)
		}
	}

	serviceTestFile := readFile(t, filepath.Join(projectDir, "internal", "service", "message_test.go"))
	for _, fragment := range []string{
		"TestMessageServiceCreateSuccess",
		"TestMessageServiceCreateConflict",
		"TestMessageServiceCreateIdempotencyReplay",
		"TestMessageServiceCreateIdempotencyConflict",
		"TestMessageServiceGetByIDSuccess",
		"TestMessageServiceUpdateSuccess",
		"TestMessageServiceUpdateTransitionConflict",
		"TestMessageServiceDeleteSuccess",
	} {
		if !strings.Contains(serviceTestFile, fragment) {
			t.Fatalf("message_test.go 缺少服务层测试片段 %q: %s", fragment, serviceTestFile)
		}
	}

	repositoryFile := readFile(t, filepath.Join(projectDir, "internal", "repository", "message.go"))
	for _, fragment := range []string{
		"type InMemoryMessageRepository struct",
		"func NewInMemoryMessageRepository()",
		"func (r *InMemoryMessageRepository) Save",
		"func (r *InMemoryMessageRepository) NextID",
	} {
		if !strings.Contains(repositoryFile, fragment) {
			t.Fatalf("repository/message.go 缺少仓储片段 %q: %s", fragment, repositoryFile)
		}
	}

	repositoryTestFile := readFile(t, filepath.Join(projectDir, "internal", "repository", "message_test.go"))
	if !strings.Contains(repositoryTestFile, "TestInMemoryMessageRepositorySaveAndFind") {
		t.Fatalf("repository/message_test.go 缺少仓储测试: %s", repositoryTestFile)
	}

	depsFile := readFile(t, filepath.Join(projectDir, "internal", "app", "dependencies.go"))
	for _, fragment := range []string{
		"type Dependencies struct",
		"func NewDependencies(cfg config.Config) Dependencies",
		"func (d Dependencies) WithDefaults() Dependencies",
		"repository.NewInMemoryMessageRepository()",
		"service.NewMessageService(repo)",
	} {
		if !strings.Contains(depsFile, fragment) {
			t.Fatalf("dependencies.go 缺少依赖装配片段 %q: %s", fragment, depsFile)
		}
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

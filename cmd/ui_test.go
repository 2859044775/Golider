package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestTerminalUIStatusLineWithColor(t *testing.T) {
	var buf bytes.Buffer
	ui := newStyledUI(&buf, true)

	ui.StatusLine("正常", "消息仓储 (internal/repository/message.go)")

	output := buf.String()
	if !strings.Contains(output, "\x1b[") {
		t.Fatalf("彩色模式输出缺少颜色控制码: %q", output)
	}
	for _, fragment := range []string{"[正常]", "消息仓储", "internal/repository/message.go"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("彩色模式输出缺少片段 %q: %q", fragment, output)
		}
	}
}

func TestTerminalUIStatusLineWithoutColor(t *testing.T) {
	var buf bytes.Buffer
	ui := newStyledUI(&buf, false)

	ui.StatusLine("缺失", "Dockerfile")

	output := buf.String()
	if strings.Contains(output, "\x1b[") {
		t.Fatalf("无色模式不应包含颜色控制码: %q", output)
	}
	for _, fragment := range []string{"[缺失]", "Dockerfile"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("无色模式输出缺少片段 %q: %q", fragment, output)
		}
	}
}

func TestTerminalUIHeaderAndKeyValue(t *testing.T) {
	var buf bytes.Buffer
	ui := newStyledUI(&buf, false)

	ui.Header("Golider 工程检查结果")
	ui.KeyValue("目标目录", "./demo")

	output := buf.String()
	for _, fragment := range []string{"== Golider 工程检查结果 ==", "目标目录", "./demo"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("输出缺少片段 %q: %q", fragment, output)
		}
	}
}

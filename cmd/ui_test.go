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

func TestTerminalUIFoldedSummary(t *testing.T) {
	var buf bytes.Buffer
	ui := newStyledUI(&buf, false)

	ui.FoldedSummary(8, 14, "项正常")
	ui.AbnormalItem("缺失", "Dockerfile")

	output := buf.String()
	for _, fragment := range []string{"8/14", "项正常", "[缺失]", "Dockerfile"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("折叠输出缺少片段 %q: %q", fragment, output)
		}
	}
}

func TestTerminalUIFoldedSummaryZero(t *testing.T) {
	var buf bytes.Buffer
	ui := newStyledUI(&buf, false)

	ui.FoldedSummary(0, 14, "项正常")

	output := buf.String()
	if output != "" {
		t.Fatalf("0 项正常时不应输出任何内容: %q", output)
	}
}

func TestTerminalUIProgressStep(t *testing.T) {
	var buf bytes.Buffer
	ui := newStyledUI(&buf, false)

	ui.ProgressStep(1, 4, "初始化目录结构")

	output := buf.String()
	if !strings.Contains(output, "初始化目录结构") {
		t.Fatalf("进度输出缺少步骤描述: %q", output)
	}
}

func TestTerminalUIConclusionOk(t *testing.T) {
	var buf bytes.Buffer
	ui := newStyledUI(&buf, false)

	ui.ConclusionOk()

	output := buf.String()
	for _, fragment := range []string{"[通过]", "所有检查均已通过"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("结论输出缺少片段 %q: %q", fragment, output)
		}
	}
}

func TestTerminalUIConclusionSummary(t *testing.T) {
	var buf bytes.Buffer
	ui := newStyledUI(&buf, false)

	ui.ConclusionSummary(23, 54, 6, 19, 6)

	output := buf.String()
	for _, fragment := range []string{"23/54", "项正常", "31", "异常"} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("结论摘要缺少片段 %q: %q", fragment, output)
		}
	}
}

func TestTerminalUISuccessAndFailure(t *testing.T) {
	var buf bytes.Buffer
	ui := newStyledUI(&buf, false)

	ui.Success("安装完成")
	ui.Failure("安装失败")

	output := buf.String()
	if !strings.Contains(output, "[完成]") {
		t.Fatalf("缺少成功标签: %q", output)
	}
	if !strings.Contains(output, "[失败]") {
		t.Fatalf("缺少失败标签: %q", output)
	}
}

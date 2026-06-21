package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	output := captureStdout(t, func() {
		if err := runVersion(nil); err != nil {
			t.Fatalf("执行 version 失败: %v", err)
		}
	})

	for _, fragment := range []string{
		"golider 0.5.0",
		"commit: dev",
		"build_date: unknown",
	} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("version 输出缺少片段 %q: %s", fragment, output)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("创建输出管道失败: %v", err)
	}

	os.Stdout = writer
	defer func() {
		os.Stdout = original
	}()

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("关闭写端失败: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("读取输出失败: %v", err)
	}

	return buf.String()
}

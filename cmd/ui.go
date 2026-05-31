package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"
)

type terminalUI struct {
	out    io.Writer
	colors bool
}

func newTerminalUI(out *os.File) terminalUI {
	return terminalUI{
		out:    out,
		colors: detectColorSupport(out),
	}
}

func newStyledUI(out io.Writer, colors bool) terminalUI {
	return terminalUI{
		out:    out,
		colors: colors,
	}
}

func detectColorSupport(out *os.File) bool {
	if strings.TrimSpace(os.Getenv("FORCE_COLOR")) != "" {
		return true
	}
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("TERM")), "dumb") {
		return false
	}
	info, err := out.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func (u terminalUI) Header(title string) {
	fmt.Fprintln(u.out, u.style("1;36", "== "+title+" =="))
}

func (u terminalUI) Section(title string) {
	fmt.Fprintln(u.out, u.style("1;34", title))
}

func (u terminalUI) Blank() {
	fmt.Fprintln(u.out)
}

func (u terminalUI) Success(message string) {
	fmt.Fprintf(u.out, "%s %s\n", u.statusLabel("完成"), message)
}

func (u terminalUI) Warning(message string) {
	fmt.Fprintf(u.out, "%s %s\n", u.statusLabel("注意"), message)
}

func (u terminalUI) Failure(message string) {
	fmt.Fprintf(u.out, "%s %s\n", u.statusLabel("失败"), message)
}

func (u terminalUI) Info(message string) {
	fmt.Fprintf(u.out, "%s %s\n", u.statusLabel("提示"), message)
}

func (u terminalUI) StatusLine(status string, content string) {
	fmt.Fprintf(u.out, "- %s %s\n", u.statusLabel(status), content)
}

func (u terminalUI) KeyValue(key string, value string) {
	fmt.Fprintf(u.out, "  %s %s\n", u.style("1;37", key), value)
}

func (u terminalUI) statusLabel(status string) string {
	switch status {
	case "正常", "完成", "通过":
		return u.style("1;32", "["+status+"]")
	case "缺失", "无效", "失败":
		return u.style("1;31", "["+status+"]")
	case "注意":
		return u.style("1;33", "["+status+"]")
	default:
		return u.style("1;36", "["+status+"]")
	}
}

func (u terminalUI) style(code string, text string) string {
	if !u.colors {
		return text
	}
	return "\x1b[" + code + "m" + text + "\x1b[0m"
}

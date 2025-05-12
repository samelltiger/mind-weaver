package htmljsvalidator

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// Validator 结构体封装检查逻辑
type Validator struct {
	browser *rod.Browser
	timeout time.Duration
}

// NewValidator 创建一个新的验证器实例
func NewValidator(timeout time.Duration) *Validator {
	// 启动浏览器（默认无头模式）
	path, _ := launcher.LookPath()
	browser := rod.New().ControlURL(launcher.New().Bin(path).
		NoSandbox(true).
		Headless(true).
		MustLaunch(),
	).MustConnect()
	return &Validator{
		browser: browser,
		timeout: timeout,
	}
}

// Close 释放浏览器资源
func (v *Validator) Close() {
	v.browser.MustClose()
}

// CheckHTMLFile 检查 HTML 文件中的 JS 错误
func (v *Validator) CheckLocalHTMLFile(filePath string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), v.timeout)
	defer cancel()

	page := v.browser.MustPage("file://" + filePath)
	defer page.MustClose()

	// 监听 JS 错误
	jsErrors := make(chan string, 10)
	wait := page.EachEvent(func(e *proto.ConsoleMessageAdded) {
		fmt.Println("Console Message: ", e)
		if e.Message.Level == "error" {
			fmt.Println("error level")
			jsErrors <- e.Message.Text
		}
	})

	// 等待页面加载完成
	page.MustWaitLoad().MustWaitIdle()

	// 收集错误
	var errors []string
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("检查超时")
	case <-time.After(1 * time.Second): // 额外等待 1 秒确保捕获异步错误
		wait() // 停止监听
		close(jsErrors)
		for err := range jsErrors {
			errors = append(errors, err)
		}
	}

	return errors, nil
}

// CheckHTMLString 检查 HTML 字符串中的 JS 错误
func (v *Validator) CheckHTMLString(htmlContent string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), v.timeout)
	defer cancel()

	page := v.browser.MustPage("about:blank").MustInsertText(htmlContent)
	defer page.MustClose()

	// 监听 JS 错误
	jsErrors := make(chan string, 10)
	wait := page.EachEvent(func(e *proto.ConsoleMessageAdded) {
		if e.Message.Level == "error" {
			jsErrors <- e.Message.Text
		}
	})

	// 等待页面加载完成
	page.MustWaitLoad().MustWaitIdle()

	// 收集错误
	var errors []string
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("检查超时")
	case <-time.After(1 * time.Second): // 额外等待 1 秒确保捕获异步错误
		wait() // 停止监听
		close(jsErrors)
		for err := range jsErrors {
			errors = append(errors, err)
		}
	}

	return errors, nil
}

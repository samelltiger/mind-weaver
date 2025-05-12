package htmljsvalidator

import (
	"fmt"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// ClickElement 在指定页面上查找并点击匹配选择器的元素。
// page: 要操作的 rod 页面对象。
// selector: CSS 选择器，用于定位要点击的元素。
// timeout: 操作的超时时间。
func ClickElement(page *rod.Page, selector string, timeout time.Duration) error {
	el, err := page.Timeout(timeout).Element(selector)
	if err != nil {
		return fmt.Errorf("查找元素 '%s' 失败: %w", selector, err)
	}
	err = el.Click(proto.InputMouseButtonLeft, 1) // 使用 proto 指定点击
	if err != nil {
		return fmt.Errorf("点击元素 '%s' 失败: %w", selector, err)
	}
	return nil
}

// TypeIntoElement 在指定页面上查找匹配选择器的元素，并向其输入文本。
// page: 要操作的 rod 页面对象。
// selector: CSS 选择器，用于定位要输入文本的元素。
// text: 要输入的文本内容。
// timeout: 操作的超时时间。
func TypeIntoElement(page *rod.Page, selector string, text string, timeout time.Duration) error {
	el, err := page.Timeout(timeout).Element(selector)
	if err != nil {
		return fmt.Errorf("查找元素 '%s' 失败: %w", selector, err)
	}
	err = el.Input(text)
	if err != nil {
		return fmt.Errorf("向元素 '%s' 输入文本失败: %w", selector, err)
	}
	return nil
}

// EvaluateJS 在指定页面的上下文中执行 JavaScript 代码片段。
// page: 要操作的 rod 页面对象。
// script: 要执行的 JavaScript 代码字符串。
// timeout: 操作的超时时间。
// 返回值: JavaScript 代码的执行结果 (proto.RuntimeRemoteObject) 或错误。
func EvaluateJS(page *rod.Page, script string, timeout time.Duration) (*proto.RuntimeRemoteObject, error) {
	res, err := page.Timeout(timeout).Evaluate(&rod.EvalOptions{
		JS: script,
	})
	if err != nil {
		return nil, fmt.Errorf("执行 JavaScript 失败: %w", err)
	}
	return res, nil
}

// WaitForElement 等待指定页面上匹配选择器的元素出现。
// page: 要操作的 rod 页面对象。
// selector: CSS 选择器，用于定位要等待的元素。
// timeout: 等待的超时时间。
func WaitForElement(page *rod.Page, selector string, timeout time.Duration) error {
	_, err := page.Timeout(timeout).Element(selector)
	if err != nil {
		return fmt.Errorf("等待元素 '%s' 超时或失败: %w", selector, err)
	}
	return nil
}

// GetElementText 获取指定页面上匹配选择器的元素的文本内容。
// page: 要操作的 rod 页面对象。
// selector: CSS 选择器，用于定位要获取文本的元素。
// timeout: 操作的超时时间。
// 返回值: 元素的文本内容或错误。
func GetElementText(page *rod.Page, selector string, timeout time.Duration) (string, error) {
	el, err := page.Timeout(timeout).Element(selector)
	if err != nil {
		return "", fmt.Errorf("查找元素 '%s' 失败: %w", selector, err)
	}
	text, err := el.Text()
	if err != nil {
		return "", fmt.Errorf("获取元素 '%s' 文本失败: %w", selector, err)
	}
	return text, nil
}

// GetPageHTML 获取指定页面的完整 HTML 内容。
// page: 要操作的 rod 页面对象。
// timeout: 操作的超时时间。
// 返回值: 页面的 HTML 字符串或错误。
func GetPageHTML(page *rod.Page, timeout time.Duration) (string, error) {
	html, err := page.Timeout(timeout).HTML()
	if err != nil {
		return "", fmt.Errorf("获取页面 HTML 失败: %w", err)
	}
	return html, nil
}

// ScreenshotPage 对指定页面进行截图并保存到文件。
// page: 要操作的 rod 页面对象。
// filePath: 截图保存的文件路径。
// timeout: 操作的超时时间。
func ScreenshotPage(page *rod.Page, fullPage bool, filePath string, timeout time.Duration) error {
	// Take the screenshot (returns []byte)
	imgData, err := page.Timeout(timeout).Screenshot(fullPage, &proto.PageCaptureScreenshot{
		Format: "png", // Options: "png" (default) | "jpeg"
	})
	if err != nil {
		return fmt.Errorf("页面截图失败: %w", err)
	}

	// Save the image data to file
	if err := os.WriteFile(filePath, imgData, 0644); err != nil {
		return fmt.Errorf("保存截图失败: %w", err)
	}

	fmt.Printf("截图成功，保存到: %s\n", filePath)
	return nil
}

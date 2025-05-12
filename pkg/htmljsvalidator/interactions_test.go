package htmljsvalidator

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer 创建一个简单的 HTTP 测试服务器
func setupTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// 用于测试点击和获取文本的页面
	mux.HandleFunc("/click", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
			<!DOCTYPE html>
			<html>
			<head><title>Click Test</title></head>
			<body>
				<p id="message">Initial Message</p>
				<button id="myButton" onclick="document.getElementById('message').innerText = 'Button Clicked!'">Click Me</button>
			</body>
			</html>
		`)
	})

	// 用于测试输入文本的页面
	mux.HandleFunc("/type", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
			<!DOCTYPE html>
			<html>
			<head><title>Type Test</title></head>
			<body>
				<input type="text" id="myInput" oninput="document.getElementById('output').innerText = this.value">
				<p id="output"></p>
			</body>
			</html>
		`)
	})

	// 用于测试 JS 执行的页面
	mux.HandleFunc("/eval", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
			<!DOCTYPE html>
			<html>
			<head><title>Eval Test</title></head>
			<body>
				<div id="data" data-value="test-data">Some Content</div>
				<script>
					window.myVar = 123;
					function getDivData() { return document.getElementById('data').getAttribute('data-value'); }
				</script>
			</body>
			</html>
		`)
	})

	// 用于测试等待元素的页面 (元素延迟出现)
	mux.HandleFunc("/wait", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
			<!DOCTYPE html>
			<html>
			<head><title>Wait Test</title></head>
			<body>
				<div id="loading">Loading...</div>
				<script>
					setTimeout(() => {
						const loadedDiv = document.createElement('div');
						loadedDiv.id = 'loadedContent';
						loadedDiv.innerText = 'Content Loaded!';
						document.body.appendChild(loadedDiv);
						document.getElementById('loading').remove();
					}, 500); // 500ms 延迟
				</script>
			</body>
			</html>
		`)
	})

	// 用于测试截图的页面
	mux.HandleFunc("/screenshot", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `
			<!DOCTYPE html>
			<html>
			<head><title>Screenshot Test</title></head>
			<body style="background-color: lightblue;">
				<h1>Screenshot Page</h1>
			</body>
			</html>
		`)
	})

	return httptest.NewServer(mux)
}

// helper function to get a browser instance for tests
func getTestBrowser(t *testing.T) *rod.Browser {
	// 使用 NewValidator 来获取浏览器实例，确保与主代码一致
	// 注意：这里的 timeout 是 Validator 级别的，交互函数有自己的 timeout
	validator := NewValidator(30 * time.Second)
	require.NotNil(t, validator)
	require.NotNil(t, validator.browser)
	// 将浏览器返回，让调用者决定何时关闭
	return validator.browser
}

func TestClickElement(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	browser := getTestBrowser(t)
	defer browser.MustClose() // 确保浏览器在测试结束后关闭

	page := browser.MustPage(ts.URL + "/click")
	defer page.MustClose()
	page.MustWaitLoad().MustWaitIdle()

	interactionTimeout := 5 * time.Second

	// 检查初始文本
	initialText, err := GetElementText(page, "#message", interactionTimeout)
	require.NoError(t, err)
	assert.Equal(t, "Initial Message", initialText)

	// 点击按钮
	err = ClickElement(page, "#myButton", interactionTimeout)
	require.NoError(t, err)

	// 等待一下让 JS 执行 (或者更好的方式是等待文本变化)
	page.MustWaitIdle() // 等待可能的 DOM 更新

	// 检查点击后的文本
	finalText, err := GetElementText(page, "#message", interactionTimeout)
	require.NoError(t, err)
	assert.Equal(t, "Button Clicked!", finalText)

	// 测试点击不存在的元素
	err = ClickElement(page, "#nonExistentButton", interactionTimeout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "查找元素 '#nonExistentButton' 失败")
}

func TestTypeIntoElement(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	browser := getTestBrowser(t)
	defer browser.MustClose()

	page := browser.MustPage(ts.URL + "/type")
	defer page.MustClose()
	page.MustWaitLoad().MustWaitIdle()

	interactionTimeout := 5 * time.Second
	inputText := "Hello, Rod!"

	// 输入文本
	err := TypeIntoElement(page, "#myInput", inputText, interactionTimeout)
	require.NoError(t, err)

	// 等待 JS 更新 output
	page.MustWaitIdle()

	// 检查输出段落的文本是否与输入一致
	outputText, err := GetElementText(page, "#output", interactionTimeout)
	require.NoError(t, err)
	assert.Equal(t, inputText, outputText)

	// 测试向不存在的元素输入
	err = TypeIntoElement(page, "#nonExistentInput", "test", interactionTimeout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "查找元素 '#nonExistentInput' 失败")
}

func TestEvaluateJS(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	browser := getTestBrowser(t)
	defer browser.MustClose()

	page := browser.MustPage(ts.URL + "/eval")
	defer page.MustClose()
	page.MustWaitLoad().MustWaitIdle()

	interactionTimeout := 5 * time.Second

	// 1. 执行简单表达式
	result, err := EvaluateJS(page, "1 + 2", interactionTimeout)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, float64(3), result.Value.Num()) // JSON 数字默认为 float64

	// 2. 获取全局变量
	result, err = EvaluateJS(page, "window.myVar", interactionTimeout)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, float64(123), result.Value.Num())

	// 3. 调用页面上的函数
	result, err = EvaluateJS(page, "getDivData()", interactionTimeout)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "test-data", result.Value.Str())

	// 4. 测试执行错误 JS
	_, err = EvaluateJS(page, "invalid code !!!", interactionTimeout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "执行 JavaScript 失败") // 检查我们包装的错误信息
}

func TestWaitForElement(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	browser := getTestBrowser(t)
	defer browser.MustClose()

	page := browser.MustPage(ts.URL + "/wait")
	defer page.MustClose()
	// 不需要 MustWaitLoad/Idle，因为我们要测试等待元素出现

	// 等待延迟加载的元素
	waitTimeout := 3 * time.Second // 超时时间应大于 JS 的 setTimeout 时间
	err := WaitForElement(page, "#loadedContent", waitTimeout)
	require.NoError(t, err, "等待 #loadedContent 时出错")

	// 验证元素确实存在并且文本正确
	text, err := GetElementText(page, "#loadedContent", 1*time.Second) // 短超时，因为元素应该已经存在
	require.NoError(t, err)
	assert.Equal(t, "Content Loaded!", text)

	// 测试等待一个永远不会出现的元素（应该超时）
	shortTimeout := 100 * time.Millisecond // 远小于 JS 的 setTimeout 时间
	err = WaitForElement(page, "#neverAppears", shortTimeout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "等待元素 '#neverAppears' 超时或失败")
}

func TestGetElementText(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	browser := getTestBrowser(t)
	defer browser.MustClose()

	page := browser.MustPage(ts.URL + "/click") // 复用 click 页面的 HTML
	defer page.MustClose()
	page.MustWaitLoad().MustWaitIdle()

	interactionTimeout := 5 * time.Second

	// 获取存在元素的文本
	text, err := GetElementText(page, "#message", interactionTimeout)
	require.NoError(t, err)
	assert.Equal(t, "Initial Message", text)

	// 获取按钮的文本
	text, err = GetElementText(page, "#myButton", interactionTimeout)
	require.NoError(t, err)
	assert.Equal(t, "Click Me", text)

	// 获取不存在元素的文本
	_, err = GetElementText(page, "#nonExistent", interactionTimeout)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "查找元素 '#nonExistent' 失败")
}

func TestGetPageHTML(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	browser := getTestBrowser(t)
	defer browser.MustClose()

	page := browser.MustPage(ts.URL + "/click") // 复用 click 页面的 HTML
	defer page.MustClose()
	page.MustWaitLoad().MustWaitIdle()

	interactionTimeout := 5 * time.Second

	html, err := GetPageHTML(page, interactionTimeout)
	require.NoError(t, err)
	assert.NotEmpty(t, html)
	assert.Contains(t, html, "<title>Click Test</title>")
	assert.Contains(t, html, `id="message"`)
	assert.Contains(t, html, `id="myButton"`)
}

func TestScreenshotPage(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()
	browser := getTestBrowser(t)
	defer browser.MustClose()

	page := browser.MustPage(ts.URL + "/screenshot")
	defer page.MustClose()
	page.MustWaitLoad().MustWaitIdle()

	interactionTimeout := 10 * time.Second // 截图可能稍慢

	// 创建临时目录存放截图
	screenshotPath := "test_screenshot.png"

	err := ScreenshotPage(page, true, screenshotPath, interactionTimeout)
	require.NoError(t, err)

	// 检查截图文件是否存在
	_, err = os.Stat(screenshotPath)
	require.NoError(t, err, "截图文件未创建")

	// 可选：更复杂的测试可以检查文件大小或使用图像库验证内容，但通常检查文件存在即可
}

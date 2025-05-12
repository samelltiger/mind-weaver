package htmljsvalidator

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestValidator 创建一个用于测试的 Validator 实例
// 使用 t.Cleanup 确保资源被释放
func setupTestValidator(t *testing.T, timeout time.Duration) *Validator {
	validator := NewValidator(timeout)
	require.NotNil(t, validator)
	t.Cleanup(func() {
		validator.Close() // Ensure browser is closed after test finishes
	})
	return validator
}

// createTempHTMLFile 创建一个包含指定内容的临时 HTML 文件
// 返回文件路径和清理函数
func createTempHTMLFile(t *testing.T, content string, filenamePattern string) string {
	t.Helper() // 标记为测试辅助函数

	// 使用 t.TempDir() 创建临时目录，测试结束后会自动清理
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, filenamePattern)

	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err, "无法创建临时 HTML 文件")

	return filePath
}

// TestCheckLocalHTMLFile_NoError 测试没有 JS 错误的 HTML 文件
func TestCheckLocalHTMLFile_NoError(t *testing.T) {
	validator := setupTestValidator(t, 10*time.Second) // 正常超时

	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head><title>No Error Test</title></head>
	<body>
		<p>Hello, World!</p>
		<script>
			console.log("This is fine.");
			const x = 10; 
			ocument.body.style.backgroundColor = 'lightgreen';
		</script>
	</body>
	</html>`
	filePath := createTempHTMLFile(t, htmlContent, "valid-*.html")

	errors, err := validator.CheckLocalHTMLFile(filePath)

	require.NoError(t, err, "检查无错误文件时不应返回错误")
	assert.Empty(t, errors, "检查无错误文件时错误列表应为空")
}

// TestCheckLocalHTMLFile_WithErrors 测试包含 JS 错误的 HTML 文件
func TestCheckLocalHTMLFile_WithErrors(t *testing.T) {
	validator := setupTestValidator(t, 10*time.Second) // 正常超时

	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head><title>Error Test</title></head>
	<body>
		<p>This page has errors.</p>
		<script>
			console.log("Script starts");
			// 故意制造错误
			console.error("This is a test error message."); // 显式错误
			nonExistentFunction(); // 引用错误
			let y; y.prop = 1; // 类型错误
			console.log("Script might not reach here");
		</script>
		<script>
		    // 另一个错误
		    const obj = null;
		    console.log(obj.property); // 另一个类型错误
		</script>
	</body>
	</html>`
	filePath := createTempHTMLFile(t, htmlContent, "error-*.html")
	// /mnt/h/code/test_project/tetris2.html

	errors, err := validator.CheckLocalHTMLFile(filePath)

	require.NoError(t, err, "检查有错误文件时不应返回函数执行错误") // 函数本身应成功执行
	require.NotEmpty(t, errors, "检查有错误文件时错误列表不应为空")

	// 断言包含预期的错误信息 (根据浏览器和环境，具体消息可能略有不同)
	// 使用 assert.Contains 更加健壮
	assert.Contains(t, errors[0], "This is a test error message.", "应包含 console.error 输出")

	// 检查是否捕获了 ReferenceError 和 TypeError
	foundReferenceError := false
	foundTypeError := false
	for _, errMsg := range errors {
		if containsAny(errMsg, "nonExistentFunction is not defined", "ReferenceError") {
			foundReferenceError = true
		}
		// TypeError 的消息可能多样，检查 "null" 或 "undefined" 相关的属性访问错误
		if containsAny(errMsg, "Cannot read propert", "TypeError") { // "Cannot read properties of undefined", "Cannot read properties of null" etc.
			foundTypeError = true
		}
	}
	assert.True(t, foundReferenceError, "应捕获到 ReferenceError")
	assert.True(t, foundTypeError, "应捕获到 TypeError")

	// 可以进一步检查错误的数量，但要小心异步错误可能导致数量不稳定
	// assert.Len(t, errors, 3) // 或更多，取决于具体错误如何报告
	assert.GreaterOrEqual(t, len(errors), 3, "至少应捕获到 3 个错误")
}

// containsAny 检查字符串是否包含子字符串列表中的任何一个
func containsAny(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if assert.Contains(nil, s, sub) { // 使用 assert.Contains 的内部逻辑，但不让它失败测试
			return true
		}
	}
	return false
}

// TestCheckLocalHTMLFile_Timeout 测试检查过程超时
func TestCheckLocalHTMLFile_Timeout(t *testing.T) {
	// 设置一个非常短的超时时间
	validator := setupTestValidator(t, 500*time.Millisecond)

	htmlContent := `
	<!DOCTYPE html>
	<html>
	<head><title>Timeout Test</title></head>
	<body>
		<p>This script will run for a long time.</p>
		<script>
			const start = Date.now();
			// 死循环或者长时间运行的任务
			// 使用 setTimeout 确保页面有机会加载，但脚本执行时间超过超时
			setTimeout(() => {
			    console.log("Starting long task..."); // 可能不会被记录，因为超时了
			    while(Date.now() - start < 2000) { // 运行 2 秒，远超 500ms 超时
			        // Busy loop
			    }
			    console.log("Long task finished (should not happen in test)");
			}, 50); // 延迟 50ms 开始长任务
		</script>
	</body>
	</html>`
	filePath := createTempHTMLFile(t, htmlContent, "timeout-*.html")

	errors, err := validator.CheckLocalHTMLFile(filePath)

	require.Error(t, err, "检查超时应返回错误")
	assert.Contains(t, err.Error(), "检查超时", "错误信息应包含 '检查超时'")
	assert.Nil(t, errors, "超时错误发生时，错误列表应为 nil") // 因为 ctx.Done() 会先生效
}

// TestCheckLocalHTMLFile_FileNotFound (预期：Panic)
// 注意：此测试验证的是 MustPage 的行为，而不是 CheckLocalHTMLFile 的错误返回值
func TestCheckLocalHTMLFile_FileNotFound(t *testing.T) {
	validator := setupTestValidator(t, 5*time.Second)
	nonExistentPath := filepath.Join(t.TempDir(), "non-existent-file.html") // 确保文件不存在

	// 因为 CheckLocalHTMLFile 内部使用了 MustPage，预期会 panic
	assert.Panics(t, func() {
		_, _ = validator.CheckLocalHTMLFile(nonExistentPath)
	}, "访问不存在的文件时，MustPage 应该导致 panic")
}

// TestCheckLocalHTMLFile_InvalidFilePath (预期：Panic)
// 如果提供了无效的路径格式（不是 rod 能理解的 file:// URL）
func TestCheckLocalHTMLFile_InvalidFilePath(t *testing.T) {
	validator := setupTestValidator(t, 5*time.Second)
	// 提供一个非文件协议的、格式可能无效的路径
	invalidPath := "invalid-protocol://some/path"

	// 同样，预期 MustPage 会 panic
	assert.Panics(t, func() {
		_, _ = validator.CheckLocalHTMLFile(invalidPath)
	}, "使用无效文件路径格式时，MustPage 应该导致 panic")

	// 另一个例子：一个在 Windows 上可能无效的路径（如果不在 Windows 上运行，行为可能不同）
	// invalidPath = "C:" // 仅盘符可能不足以构成有效 URL
	// assert.Panics(t, func() {
	//     _, _ = validator.CheckLocalHTMLFile(invalidPath)
	// }, "使用无效文件路径格式时，MustPage 应该导致 panic")
}

package util

import (
	"context"
	"mind-weaver/pkg/logger"
	"os"
	"sync"
	"time"
)

var (
	// 全局启用计时功能标志
	enableTiming bool

	// 单独启用的函数名集合
	enabledFunctions map[string]bool

	// 保护map的互斥锁
	mu sync.RWMutex
)

// 初始化计时功能
func init() {
	// 从环境变量中读取调试模式
	if os.Getenv("DEBUG") == "true" {
		enableTiming = true
	} else {
		enableTiming = false
	}

	// 初始化启用函数集合
	enabledFunctions = make(map[string]bool)
}

// EnableTimingForFunction 为特定函数启用计时功能
func EnableTimingForFunction(funcName string) {
	mu.Lock()
	enabledFunctions[funcName] = true
	mu.Unlock()
}

// DisableTimingForFunction 为特定函数禁用计时功能
func DisableTimingForFunction(funcName string) {
	mu.Lock()
	delete(enabledFunctions, funcName)
	mu.Unlock()
}

// isFunctionTimingEnabled 检查特定函数是否启用了计时
func isFunctionTimingEnabled(funcName string) bool {
	mu.RLock()
	defer mu.RUnlock()

	// 优先检查函数是否单独启用
	if enabled, exists := enabledFunctions[funcName]; exists {
		return enabled
	}

	// 否则使用全局设置
	return enableTiming
}

// TimeTrack 用于追踪函数执行时间的工具函数
func TimeTrack(ctx context.Context, start time.Time, name string) {
	if !isFunctionTimingEnabled(name) {
		return
	}
	elapsed := time.Since(start)
	logger.InfofWithCtx(ctx, "TimeTrack: %s 执行耗时: %s", name, elapsed)
}

// FuncTimer 用于函数计时
type FuncTimer struct {
	Name      string
	StartTime time.Time
	Enabled   bool
}

// NewFuncTimer 创建一个新的函数计时器
func NewFuncTimer(name string) *FuncTimer {
	enabled := isFunctionTimingEnabled(name)
	if !enabled {
		return &FuncTimer{Enabled: false}
	}
	return &FuncTimer{
		Name:      name,
		StartTime: time.Now(),
		Enabled:   true,
	}
}

// ForceTimerForFunction 强制为当前函数创建计时器(不受全局DEBUG控制)
func ForceTimerForFunction(name string) *FuncTimer {
	return &FuncTimer{
		Name:      name,
		StartTime: time.Now(),
		Enabled:   true,
	}
}

// End 结束计时并记录耗时
func (t *FuncTimer) End(ctx context.Context) {
	if !t.Enabled {
		return // 如果计时器未启用，直接返回
	}
	elapsed := time.Since(t.StartTime)
	logger.InfofWithCtx(ctx, "TimeTrack: 函数 [%s] 执行耗时: %s", t.Name, elapsed)
}

// MeasureExecutionTime 记录任意函数的执行时间
func MeasureExecutionTime(ctx context.Context, f func() error, funcName string) error {
	if !isFunctionTimingEnabled(funcName) {
		return f() // 如果未启用计时，直接执行函数
	}
	startTime := time.Now()
	err := f()
	elapsed := time.Since(startTime)
	logger.InfofWithCtx(ctx, "TimeTrack: 函数 [%s] 执行耗时: %s", funcName, elapsed)
	return err
}

// ForceMeasureExecutionTime 强制测量函数执行时间(不受全局DEBUG控制)
func ForceMeasureExecutionTime(ctx context.Context, f func() error, funcName string) error {
	startTime := time.Now()
	err := f()
	elapsed := time.Since(startTime)
	logger.InfofWithCtx(ctx, "TimeTrack: 函数 [%s] 执行耗时: %s", funcName, elapsed)
	return err
}

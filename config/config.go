package config

import (
	"fmt"
	"os"

	yaml "sigs.k8s.io/yaml/goyaml.v2"
)

var cfg Config

type Config struct {
	MaxContextSize  int
	TempStoragePath string

	Server    Server    `yaml:"server"`
	Sqliter   Sqlite    `yaml:"sqlite"`
	LLM       LLMConfig `yaml:"llm"`
	Logger    Logger    `yaml:"logger"`
	Bin       BinConfig `yaml:"bin"`
	DiffLine  int       `yaml:"diff_line"`
	DiffModel string    `yaml:"diff_model"`
}

type Server struct {
	Port            string `yaml:"port"`
	Mode            string `yaml:"mode"` // debug, release, test
	TempStoragePath string `yaml:"temp_storage_path"`
}

type LLMConfig struct {
	BaseURL     string      `yaml:"base_url" json:"base_url"`       // 基础API地址
	APIKey      string      `yaml:"api_key" json:"api_key"`         // API密钥
	Model       string      `yaml:"model" json:"model"`             // 默认模型名称
	Models      []ModelInfo `yaml:"models" json:"models"`           // 可用模型列表
	Timeout     int         `yaml:"timeout" json:"timeout"`         // 请求超时(秒)
	MaxTokens   int         `yaml:"max_tokens" json:"max_tokens"`   // 最大token数
	MaxContext  int         `yaml:"max_context" json:"max_context"` // 最大上下文长度
	Temperature float64     `yaml:"temperature" json:"temperature"` // 温度参数
	TopP        float64     `yaml:"top_p" json:"top_p"`             // Top-P采样
	EnableLog   bool        `yaml:"enable_log" json:"enable_log"`   // 是否启用日志
	Proxy       string      `yaml:"proxy" json:"proxy"`             // 代理设置
	RateLimit   RateLimit   `yaml:"rate_limit" json:"rate_limit"`   // 速率限制
}

type BinConfig struct {
	Python string `yaml:"python"`
}

type JWT struct {
	Secret    string `yaml:"secret"`
	ExpiresIn int    `yaml:"expires_in"` // 小时
}

func (l LLMConfig) GetCurrentLLMInfo(modelName string) ModelInfo {
	for _, model := range l.Models {
		if model.Name == modelName {
			return model
		}
	}

	fmt.Printf("Model %s not found, return default model.\n", modelName)
	return ModelInfo{
		Name:         l.Model,
		Description:  "",
		MaxContext:   l.MaxContext,
		MaxTokens:    l.MaxTokens,
		Capabilities: []string{"chat", "completion"},
		IsChatModel:  true,
		CostPerToken: 0.00003,
	}
}

type ModelInfo struct {
	Name         string   `yaml:"name" json:"name"`                     // 模型名称
	Description  string   `yaml:"description" json:"description"`       // 模型描述
	MaxContext   int      `yaml:"max_context" json:"max_context"`       // 最大上下文长度
	MaxTokens    int      `yaml:"max_tokens" json:"max_tokens"`         // 最大token数
	Capabilities []string `yaml:"capabilities" json:"capabilities"`     // 能力列表
	IsChatModel  bool     `yaml:"is_chat_model" json:"is_chat_model"`   // 是否为聊天模型
	CostPerToken float64  `yaml:"cost_per_token" json:"cost_per_token"` // 每token成本
	// deepseek 官方建议
	// 代码生成/数学解题   	0.0
	// 数据抽取/分析	1.0
	// 通用对话	1.3
	// 翻译	1.3
	// 创意类写作/诗歌创作	1.5
	Temperature float64 `yaml:"temperature " json:"temperature "` // 温度设置
}

type RateLimit struct {
	RequestsPerMinute int `yaml:"requests_per_minute" json:"requests_per_minute"` // 每分钟请求数
	TokensPerMinute   int `yaml:"tokens_per_minute" json:"tokens_per_minute"`     // 每分钟token数
}

type Sqlite struct {
	DBPath string `yaml:"db_path"`
}

type Logger struct {
	Level      string `yaml:"level"` // debug, info, warn, error
	Filename   string `yaml:"filename"`
	MaxSize    int    `yaml:"maxsize"`    // MB
	MaxBackups int    `yaml:"maxbackups"` // 文件个数
	MaxAge     int    `yaml:"maxage"`     // 天数
	Compress   bool   `yaml:"compress"`   // 是否压缩
}

// func LoadConfig() *Config {
func LoadConfig(file string) (*Config, error) {
	// Load 从文件加载配置
	// func Load(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	cfg = config

	return &config, nil
}

func GetConfig() *Config {
	return &cfg
}

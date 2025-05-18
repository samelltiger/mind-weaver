# My AI Coding Buddy (MindWeaver)

> 思维编织者，将想法转化为代码。

<!-- 可选: Logo -->

<!--
<p align="center">
  <img src="path/to/your/logo.png" alt="My AI Coding Buddy Logo" width="200"/>
</p>
-->

<!-- 徽章 -->

<p align="center">
  <!-- 请将 samelltiger/mind-weaver 替换为您的实际 GitHub 用户名和仓库名 -->
  <a href="https://github.com/samelltiger/mind-weaver/actions/workflows/go.yml"><img src="https://github.com/samelltiger/mind-weaver/actions/workflows/go.yml/badge.svg" alt="Build Status"></a>
  <a href="https://goreportcard.com/report/github.com/samelltiger/mind-weaver"><img src="https://goreportcard.com/badge/github.com/samelltiger/mind-weaver" alt="Go Report Card"></a>
  <a href="https://github.com/samelltiger/mind-weaver/blob/main/LICENSE"><img src="https://img.shields.io/github/license/samelltiger/mind-weaver.svg" alt="License"></a>
  <a href="https://github.com/samelltiger/mind-weaver/releases/latest"><img src="https://img.shields.io/github/v/release/samelltiger/mind-weaver.svg" alt="Latest Release"></a>
  <!-- 更多徽章 -->
</p>

**My AI Coding Buddy (MindWeaver) 是一个使用 Go 和现代 Web 技术构建的本地化 AI 编程助手，旨在通过集成强大的大型语言模型 (LLM) 来提升开发者的编码效率和代码质量。正如其名“思维编织者”，它致力于将您的想法顺畅地转化为实际的代码。**

<!-- 可选: 演示 GIF 或截图 -->

<p align="center">
  <img src="path/to/your/demo.gif" alt="My AI Coding Buddy Demo">
</p>

## ✨ 特性 (Features)

* 🧠 **智能代码辅助:**
  * 根据自然语言描述快速生成代码片段。
  * 详细解释复杂代码段的逻辑和功能（支持选中代码解释）。
  * 提供代码优化建议，辅助重构。
* 🐞 **Bug 辅助与修复:**
  * 辅助定位代码中的错误。
  * 在特定模式下（如 `single-html` 模式）尝试自动修复 HTML/JS 错误。
* 📄 **项目与文件管理:**
  * 支持创建和管理多个项目。
  * 内置文件树浏览器，方便查看和打开项目文件。
  * 多标签页编辑器，同时处理多个文件。
* 💬 **多会话与上下文管理:**
  * 为不同任务创建独立的 AI 会话。
  * 支持将选定文件或代码片段作为上下文提交给 AI。
* 🔄 **灵活的会话模式，满足不同场景需求:**
  * **手动模式 (Manual Mode):** 您可以精确选择项目中的代码文件作为 AI 的参考，AI 将基于这些已有代码进行开发、修改或优化。适合需要精细控制 AI 输入的场景。
  * **智能模式 (Smart Mode):** AI 会更智能地分析您的需求，不仅能编写代码，还能自动执行如创建文件/文件夹、运行 Shell 命令等辅助操作，以自主完成任务。
  * **单HTML模式 (Single HTML Mode):** 专注于在单个 HTML 文件中实现您的完整想法。AI 能生成内容丰富、结构完整的 HTML 页面，非常适合快速制作 DEMO 演示页或产品原型。
* 🔌 **多模型支持:**
  * 通过可配置的 `base_url`（如 `one-api`, `new-api`）支持接入多种 LLM，例如：
    * OpenAI GPT 系列
    * Anthropic Claude 系列 (示例配置中包含 `claude-3-7-sonnet-20250219`)
    * DeepSeek 模型 (示例配置中包含 `deepseek-chat`)
    * Google Gemini 模型 (示例配置中包含 `qt-gemini-2.5-pro-preview-05-06`)
  * 可轻松扩展以支持更多模型。
* ⚙️ **强大工具集成:**
  * **代码执行:** 支持直接执行代码片段（如 Python）。
  * **Shell 命令执行:** 安全地执行 Shell 命令。
  * **HTML 实时预览:** 在编辑器内直接预览 HTML 文件。
  * **代码差异合并:** 辅助处理 AI 生成代码与现有代码的合并。
* 💻 **简洁易用的 Web 界面:**
  * 基于 Gin 构建的后端 API 和原生 HTML, CSS, JavaScript 前端。
  * 集成 Monaco Editor 提供桌面级代码编辑体验。
  * Markdown 实时渲染聊天消息。
* 🔒 **本地优先与数据安全:**
  * API Key 和敏感配置存储在本地 `config.yaml` 文件中。
  * 核心逻辑在本地运行，保障代码和数据隐私。
* 🔧 **灵活配置:**
  * 通过 `config.yaml` 文件自定义服务器、LLM 参数、日志、代理等。
  * 支持不同模型的特定参数配置（如 `temperature`, `max_tokens`）。
* 📜 **API 文档:**
  * 内置 Swagger UI (`/swagger/index.html`)，方便查阅和测试 API。
* 跨平台支持 (Windows, macOS, Linux Go 编译目标均可支持)。


## 🛠️ 技术栈 (Tech Stack)

* **后端:** Go, Gin (Web 框架)
* **前端:** HTML, CSS, JavaScript (Vanilla JS, 模块化结构), Monaco Editor (代码编辑器), Marked.js (Markdown 渲染)
* **数据库:** SQLite (通过 `internal/db` 包)
* **配置:** YAML
* **AI 模型接口:** 通过 HTTP 客户端与 OpenAI 兼容的 API (如 `one-api`, `new-api`, 或直接连接 Ollama 等) 进行交互。

## 🚀 快速开始 (Quick Start)

### 先决条件 (Prerequisites)

* Go 1.23+ (安装指南: [https://golang.org/doc/install](https://golang.org/doc/install))
* 一个现代 Web 浏览器 (Chrome, Firefox, Edge, Safari)
* (可选) Python 环境 (用于执行 `scripts` 目录下的某些脚本，如 `jsinspector.py`, `code_diff_merge.py`)
* 您选择的 LLM 服务 API Key (例如，如果您使用 `one-api`，则为 `one-api` 的 Key)

### 1. 安装 (Installation)

**选项 A: 从 Release 下载 (推荐)**

1. 访问 [GitHub Releases 页面](https://github.com/samelltiger/mind-weaver/releases)。
2. 下载适用于你操作系统的最新预编译版本 (e.g., `mind-weaver-windows-amd64.exe`, `mind-weaver-linux-amd64`)。
3. 解压并将可执行文件放到你的 `PATH` 中 (可选)。

**选项 B: 从源码构建**

```bash
# 1. 克隆仓库
git clone https://github.com/samelltiger/mind-weaver.git
cd mind-weaver # 仓库名可能为 mind-weaver 或 My-AI-Coding-Buddy

# 2. 构建 Go 后端
#    可执行文件名在 Swagger 注释中为 mind-weaver
go build -o mind-weaver ./cmd/main.go

# 3. (可选) 生成 Swagger 文档 (如果需要更新)
# go run ./scripts/genswag.go
```

前端资源已包含在 `frontend` 目录中，并由 Go 程序直接提供静态文件服务，通常不需要单独构建。

### 2. 配置 (Configuration)

`My AI Coding Buddy` 通过配置文件 (`config.yaml`) 进行配置。启动时会自动在程序运行目录下寻找 `config/config.yaml`。

**创建 `config/config.yaml` 文件 (基于 `config/config.example.yaml`):**

```yaml
server:
  port: "14010" # 默认端口
  mode: "debug" # debug, release, test
  temp_storage_path: "./data/temp"

sqlite:
  db_path: "./data/mind-weaver.db" # SQLite 数据库文件路径

llm:
  base_url: "http://<your-one-api-or-llm-service-host>:<port>"  # 示例: "http://192.168.0.200:3000" (填写one-api/new-api服务的地址)
  api_key: "sk-YOUR_API_KEY"  # 强烈建议使用环境变量或安全的密钥管理方式 (填写one-api/new-api的key)
  model: "claude-3-7-sonnet-20250219" # 默认使用的模型
  timeout: 600 # 请求超时 (秒)
  max_tokens: 8192 # AI 回复的最大 token 数
  max_context: 195000 # 模型支持的最大上下文 token 数
  temperature: 0.7 # 默认温度
  top_p: 0.9
  # 可用模型列表
  models:
  - name: "deepseek-chat"
    description: "deepseek v3模型"
    max_tokens: 8192
    max_context: 128000
    capabilities: ["chat", "completion"]
    is_chat_model: true
    temperature: 0.0

  - name: "claude-3-7-sonnet-20250219"
    description: "claude-3-7-sonnet-20250219 模型"
    max_tokens: 8192
    max_context: 200000
    capabilities: ["chat", "completion"]
    is_chat_model: true
    temperature: 0.7

  - name: "qt-gemini-2.5-pro-preview-05-06"
    description: "gemini-2.5-pro-exp-03-25 模型"
    max_tokens: 8192
    max_context: 200000
    capabilities: ["chat", "completion"]
    is_chat_model: true
    temperature: 0.7

logger:
  level: "info" # 日志级别: debug, info, warn, error
  filename: "logs/app.log" # 日志文件路径
  maxsize: 100 # 单个日志文件最大 MB
  maxbackups: 10 # 最多保留的旧日志文件数
  maxage: 30 # 日志文件最长保留天数
  compress: true # 是否压缩旧日志

bin:
  python: "/opt/anaconda3/bin/python" # Python 解释器路径 (用于执行脚本)

diff_line: 20 # 代码差异比较时，上下文保留行数
diff_model: "deepseek-chat" # 用于代码差异修复的模型
```

**环境变量:**

虽然当前代码主要通过 `config.yaml` 加载配置，但您可以修改 `config/config.go` 中的 `LoadConfig` 函数以支持从环境变量读取敏感信息（如 `API_KEY`），这是一种更安全的做法。

更多配置选项请参考 `config/config.example.yaml` 文件。

### 3. 运行 (Run)

如果您从源码构建并将可执行文件命名为 `mind-weaver`：

```bash
./mind-weaver
```

或者，直接从 `cmd/main.go` 运行（主要用于开发）：

```bash
go run ./cmd/main.go
```

程序启动后，日志会提示服务运行的端口 (默认为 `14010`)。
然后打开浏览器访问 `http://localhost:14010` (或您配置的地址和端口)。
API 文档 (Swagger) 通常在 `http://localhost:14010/swagger/index.html`。

### 4. 基本用法 (Basic Usage)

1. **项目管理:**
   * 在左上角的 "项目"区域，点击 "新建项目" 或选择现有项目。
   * 项目文件会显示在下方的 "文件" 区域。
2. **文件操作:**
   * 点击文件树中的文件，在中间的编辑器区域打开。
   * 支持多标签页编辑。
3. **AI 会话:**
   * 在右上角的 "会话" 区域，点击 "新建会话" 或选择现有会话。
   * 会话可以配置不同的模式（如手动选择上下文文件、智能模式等）。
4. **与 AI 交互:**
   * 在右下角的聊天输入框中输入您的问题、代码或指令。
   * 选择希望使用的 AI 模型。
   * 在 "上下文文件" 区域，可以勾选希望 AI 参考的文件。
   * 点击 "发送" 或按回车与 AI 交互。
   * 特定功能如 "解释代码"、"HTML 预览" 可通过对应按钮触发。

## 📖 文档 (Documentation)

* **API 文档:** 启动服务后访问 `/swagger/index.html` (例如 `http://localhost:14010/swagger/index.html`)。
* [安装指南](./docs/installation.md) (请创建此文件)
* [配置详解](./docs/configuration.md) (请创建此文件，可基于 `config.example.yaml` 详细说明)
* [用户手册](./docs/usage_guide.md) (请创建此文件)
* [常见问题](./docs/troubleshooting.md) (请创建此文件)

## 🤝 贡献 (Contributing)

我们非常欢迎各种形式的贡献！请阅读我们的 [贡献指南 (CONTRIBUTING.md)](./CONTRIBUTING.md) (请创建此文件) 来了解如何参与。

在提交代码前，请确保你已阅读并同意我们的 [行为准则 (CODE_OF_CONDUCT.md)](./CODE_OF_CONDUCT.md) (请创建此文件)。

## 📜 许可证 (License)

本项目采用 [MIT 许可证](./LICENSE) (请创建此文件并填入 MIT 许可证内容)。

## 🙏 致谢 (Acknowledgements)

* 感谢 [Go 语言](https://golang.org/) 团队。
* 感谢 [Gin Web Framework](https://gin-gonic.com/)。
* 感谢 [Monaco Editor](https://microsoft.github.io/monaco-editor/)。
* 感谢 [Marked.js](https://marked.js.org/)。
* 感谢 [Swagger](https://swagger.io/) 和 `gin-swagger` / `swaggo/files`。
* 感谢所有为本项目提供灵感和帮助的开源社区及 LLM 服务提供商。


# 更多
更多信息请关注公众号： 思维编织者

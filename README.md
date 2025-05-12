# mind-weaver
思维编织者，将想法转化为代码。一款快速利用AI完成编码的功能，提供多种会话模式，利用多个大模型完成模式功能，节省成本。

## API 接口文档

### 项目管理

#### 创建项目
- **URL**: `/api/projects`
- **Method**: `POST`
- **请求体**:
  ```json
  {
    "name": "项目名称",
    "path": "/path/to/project",
    "language": "go"
  }
  ```
- **响应**: 创建的项目对象

#### 获取所有项目
- **URL**: `/api/projects`
- **Method**: `GET`
- **响应**: 项目对象数组

#### 获取单个项目
- **URL**: `/api/projects/:i
## API 接口文档（续）

### 项目管理（续）

#### 获取单个项目
- **URL**: `/api/projects/:id`
- **Method**: `GET`
- **响应**: 项目对象

#### 获取项目文件树
- **URL**: `/api/projects/:id/files`
- **Method**: `GET`
- **查询参数**:
  - `maxDepth`: 文件树最大深度（默认为3）
- **响应**: 文件树结构

### 文件操作

#### 读取文件内容
- **URL**: `/api/files/read`
- **Method**: `POST`
- **请求体**:
  ```json
  {
    "project_id": 1,
    "file_path": "src/main.go"
  }
  ```
- **响应**:
  ```json
  {
    "file_path": "src/main.go",
    "content": "文件内容..."
  }
  ```

### 会话管理

#### 创建会话
- **URL**: `/api/sessions`
- **Method**: `POST`
- **请求体**:
  ```json
  {
    "project_id": 1,
    "name": "新功能开发"
  }
  ```
- **响应**: 创建的会话对象

#### 获取项目的所有会话
- **URL**: `/api/sessions/project/:projectId`
- **Method**: `GET`
- **响应**: 会话对象数组

#### 获取单个会话
- **URL**: `/api/sessions/:id`
- **Method**: `GET`
- **响应**: 会话对象（包含消息历史）

#### 发送消息
- **URL**: `/api/sessions/:id/message`
- **Method**: `POST`
- **请求体**:
  ```json
  {
    "content": "生成一个处理HTTP请求的函数",
    "project_path": "/path/to/project",
    "context_files": ["src/main.go", "src/handlers.go"]
  }
  ```
- **响应**:
  ```json
  {
    "user_message": {
      "id": 1,
      "role": "user",
      "content": "生成一个处理HTTP请求的函数",
      "timestamp": "2023-07-01T10:00:00Z"
    },
    "ai_message": {
      "id": 2,
      "role": "ai",
      "content": "AI生成的代码...",
      "timestamp": "2023-07-01T10:00:01Z"
    }
  }
  ```

#### 流式发送消息
- **URL**: `/api/sessions/:id/stream`
- **Method**: `POST`
- **请求体**: 同上
- **响应**: 服务器发送事件(SSE)流，包含AI生成的内容

#### 更新会话上下文
- **URL**: `/api/sessions/:id/context`
- **Method**: `PUT`
- **请求体**:
  ```json
  {
    "files": ["src/main.go", "src/handlers.go"],
    "current_file": "src/handlers.go",
    "cursor_position": 150,
    "selected_code": "func handleRequest() {\n\n}"
  }
  ```
- **响应**: 状态确认

#### 获取会话上下文
- **URL**: `/api/sessions/:id/context`
- **Method**: `GET`
- **响应**: 会话上下文对象

## 部署说明

### 环境要求
- Go 1.16+
- SQLite3
- OpenAI API 密钥


### 构建和运行
1. 克隆仓库
2. 设置环境变量
3. 构建应用:
   ```bash
   go build -o codepilot ./cmd/server
   ```
4. 运行应用:
   ```bash
   ./codepilot
   ```

## 核心功能实现说明

### 1. 项目和文件管理

MindWeaver AI 后端实现了基本的项目和文件管理功能，支持：
- 创建和管理多个项目
- 浏览项目文件结构
- 读取文件内容
- 自动检测项目语言

项目路径是用户从前端设置的，后端会验证路径的有效性并在本地读取文件内容。这种设计使得系统不需要复制或上传整个代码库，而是直接在用户指定的位置工作。

### 2. 上下文分析

上下文分析是 MindWeaver AI 的核心功能之一，它能够：
- 解析当前文件内容
- 识别代码语言和结构
- 提取导入的依赖
- 收集相关文件作为上下文

这些信息被用来为 AI 提供足够的上下文，使其能够生成更加准确和相关的代码。

### 3. AI 代码生成

AI 代码生成功能通过 OpenAI API 实现，支持：
- 基于用户提示和代码上下文生成代码
- 流式响应（实时显示生成结果）
- 提示词优化以获得更好的代码质量

系统会将当前文件内容、相关文件以及用户的提示作为上下文发送给 AI，从而生成与项目相关的代码。

### 4. 会话管理

会话管理系统允许用户：
- 创建多个会话
- 在会话中保存消息历史
- 保存和更新代码上下文
- 查看历史会话

这使得用户可以在不同的编码任务之间切换，同时保持每个任务的上下文和历史记录。

## 总结

这个后端实现为 MindWeaver AI 提供了核心的功能支持，包括项目管理、文件访问、上下文分析和 AI 代码生成。系统设计简洁明了，使用 Gin 框架提供 RESTful API，使用 SQLite 数据库存储项目和会话信息，并通过 OpenAI API 实现代码生成功能。

# 更多
更多讯息请关注公众号： 思维编织者

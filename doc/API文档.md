Here's the API documentation based on the provided code:

---

# MindWeaver API Documentation

## Base URL
`/api`

## Error Codes
| Code | Description                  |
|------|------------------------------|
| 0    | Success                      |
| 10001| Internal server error        |
| 10002| Invalid parameters           |
| 10004| Resource not found           |
| 20001| Connection failed            |
| 30001| Task is not running          |
| 30002| Task execution failed        |

---

## Projects

### 1. Get All Projects
**URL**: `/projects`  
**Method**: `GET`  
**Description**: List all available projects  
**Response**:
```json
{
  "code": 0,
  "msg": "Success",
  "data": [
    {
      "id": 1,
      "name": "project1",
      "path": "/path/to/project",
      "language": "go",
      "created_at": "2023-01-01T00:00:00Z",
      "last_opened_at": "2023-01-02T00:00:00Z"
    }
  ]
}
```

### 2. Create Project
**URL**: `/projects`  
**Method**: `POST`  
**Description**: Create a new project  
**Request Body**:
```json
{
  "name": "project1",
  "path": "/path/to/project",
  "language": "go"
}
```
**Response**:
```json
{
  "code": 0,
  "msg": "Success",
  "data": {
    "id": 1,
    "name": "project1",
    "path": "/path/to/project",
    "language": "go",
    "created_at": "2023-01-01T00:00:00Z",
    "last_opened_at": "2023-01-01T00:00:00Z"
  }
}
```

### 3. Get Project Details
**URL**: `/projects/:id`  
**Method**: `GET`  
**Description**: Get details of a specific project  
**Parameters**:
- `id` (path): Project ID  
**Response**: Same as Create Project response

### 4. Get Project Files
**URL**: `/projects/:id/files`  
**Method**: `GET`  
**Description**: Get file tree for a project  
**Parameters**:
- `id` (path): Project ID  
- `maxDepth` (query, optional): Maximum depth of file tree (default: 3)  
**Response**:
```json
{
  "code": 0,
  "msg": "Success",
  "data": {
    "name": "project1",
    "path": "/path/to/project",
    "is_dir": true,
    "children": [
      {
        "name": "src",
        "path": "/path/to/project/src",
        "is_dir": true,
        "children": [...]
      }
    ]
  }
}
```

---

## Files

### 1. Read File
**URL**: `/files/read`  
**Method**: `POST`  
**Description**: Read content of a file  
**Request Body**:
```json
{
  "project_id": 1,
  "file_path": "src/main.go"
}
```
**Response**:
```json
{
  "code": 0,
  "msg": "Success",
  "data": {
    "file_path": "src/main.go",
    "content": "package main\n\nfunc main() {...}"
  }
}
```

---

## Sessions

### 1. Create Session
**URL**: `/sessions`  
**Method**: `POST`  
**Description**: Create a new coding session  
**Request Body**:
```json
{
  "project_id": 1,
  "name": "debug session"
}
```
**Response**:
```json
{
  "code": 0,
  "msg": "Success",
  "data": {
    "id": 1,
    "project_id": 1,
    "name": "debug session",
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

### 2. Get Project Sessions
**URL**: `/sessions/project/:projectId`  
**Method**: `GET`  
**Description**: List all sessions for a project  
**Parameters**:
- `projectId` (path): Project ID  
**Response**:
```json
{
  "code": 0,
  "msg": "Success",
  "data": [
    {
      "id": 1,
      "project_id": 1,
      "name": "debug session",
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z"
    }
  ]
}
```

### 3. Get Session Details
**URL**: `/sessions/:id`  
**Method**: `GET`  
**Description**: Get details of a specific session including messages  
**Parameters**:
- `id` (path): Session ID  
**Response**:
```json
{
  "code": 0,
  "msg": "Success",
  "data": {
    "id": 1,
    "project_id": 1,
    "name": "debug session",
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z",
    "messages": [
      {
        "id": 1,
        "role": "user",
        "content": "How do I fix this bug?",
        "timestamp": "2023-01-01T00:01:00Z"
      }
    ]
  }
}
```

### 4. Send Message
**URL**: `/sessions/:id/message`  
**Method**: `POST`  
**Description**: Send a message to the AI and get response  
**Parameters**:
- `id` (path): Session ID  
**Request Body**:
```json
{
  "content": "How do I fix this bug?",
  "project_path": "/path/to/project",
  "context_files": ["src/main.go"]
}
```
**Response**:
```json
{
  "code": 0,
  "msg": "Success",
  "data": {
    "user_message": {
      "id": 1,
      "role": "user",
      "content": "How do I fix this bug?",
      "timestamp": "2023-01-01T00:01:00Z"
    },
    "ai_message": {
      "id": 2,
      "role": "ai",
      "content": "You need to check the null pointer...",
      "timestamp": "2023-01-01T00:01:02Z"
    }
  }
}
```

### 5. Stream Message
**URL**: `/sessions/:id/stream`  
**Method**: `POST`  
**Description**: Stream AI response (Server-Sent Events)  
**Parameters**:
- `id` (path): Session ID  
**Request Body**: Same as Send Message  
**Response**: Stream of SSE events with AI response chunks

### 6. Update Session Context
**URL**: `/sessions/:id/context`  
**Method**: `PUT`  
**Description**: Update the context for a session  
**Parameters**:
- `id` (path): Session ID  
**Request Body**:
```json
{
  "files": ["src/main.go"],
  "current_file": "src/main.go",
  "cursor_position": 42,
  "selected_code": "func main()"
}
```
**Response**:
```json
{
  "code": 0,
  "msg": "Success",
  "data": {
    "status": "ok"
  }
}
```

### 7. Get Session Context
**URL**: `/sessions/:id/context`  
**Method**: `GET`  
**Description**: Get the current context for a session  
**Parameters**:
- `id` (path): Session ID  
**Response**:
```json
{
  "code": 0,
  "msg": "Success",
  "data": {
    "files": ["src/main.go"],
    "current_file": "src/main.go",
    "cursor_position": 42,
    "selected_code": "func main()"
  }
}
```

---

This documentation covers all the endpoints implemented in the provided code. Each endpoint includes the URL, HTTP method, description, parameters (path/query/body), and example request/response formats.
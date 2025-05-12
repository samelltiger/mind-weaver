/**
 * Messaging component
 * @module modules/sessions/messaging
 */
import { sessions } from "../../api/client.js";
import {
  showNotification,
  escapeHtml,
  createElement,
} from "../../utils/dom.js";
import { showHtmlPreview, getCurrentPreviewPath } from "./messagingPreview.js";

/**
 * Send message to session
 * @param {string} type - 消息类型 (normal, explain, retry)
 * @param {string} content - Message content
 * @param {Object} currentSession - Current session
 * @param {Set} selectedContextFiles - Selected context files
 * @param {Function} setActiveStream - Function to set active stream
 * @param {Function} getActiveStream - Function to get active stream
 */
export async function sendMessage(
  type,
  content,
  currentSession,
  selectedContextFiles,
  setActiveStream,
  getActiveStream,
  tool = null // Add tool parameter
) {
  console.log("发送消息...");
  console.log("当前会话:", currentSession);
  console.log("会话模式:", currentSession?.mode); // 调试模式
  console.log("消息类型:", type); // 添加消息类型日志

  // Trim content
  content = content?.trim();

  if (!currentSession) {
    showNotification("无活动会话", "error");
    return;
  }

  // 对于非重试类型的消息，检查内容是否为空
  if (!content && (type !== "retry" && type !== "explain")) return;

  console.log("当前会话:", currentSession);

  // Get selected model
  const modelSelector = document.getElementById("model-selector");
  const selectedModel = modelSelector?.value || "gpt-4";

  // Clear input (对于非重试类型的消息)
  if (type !== "retry") {
    const userMessageInput = document.getElementById("user-message");
    if (userMessageInput) {
      userMessageInput.value = "";
    }
  }

  const messagesContainer = document.getElementById("messages-container");
  const hasMessages =
    messagesContainer?.children.length > 0 &&
    !messagesContainer.querySelector(".empty-message");

  // If this is the first message, clear "no messages" prompt
  if (!hasMessages && messagesContainer) {
    messagesContainer.innerHTML = "";
  }

  // 获取当前项目
  const currentProject = window.ProjectsModule.getCurrentProject();

  let aiMessageElement;

  // 处理重试逻辑
  if (type === "retry") {
    // 找到最后一条AI消息，我们将更新它而不是创建新消息
    const aiMessages = document.querySelectorAll('.message[data-role="ai"], .message[data-role="assistant"]');
    if (aiMessages.length > 0) {
      aiMessageElement = aiMessages[aiMessages.length - 1];
      // 更新消息内容为"思考中..."
      aiMessageElement.innerHTML = "思考中...";
      // 保留消息操作按钮
      const messageActions = document.createElement("div");
      messageActions.className = "message-actions";
      messageActions.innerHTML = `
        <button class="msg-action-btn copy-msg-btn" title="复制消息">复制</button>
        <button class="msg-action-btn delete-msg-btn" title="删除消息">删除</button>
        <button class="msg-action-btn retry-msg-btn" title="重新生成">重试</button>
      `;
      aiMessageElement.appendChild(messageActions);
    } else {
      console.error("没有找到可重试的AI消息");
      return;
    }
  } else {
    // 对于非重试类型的消息，添加用户消息到UI
    addMessageToUI("user", content);

    // 创建AI响应占位符
    aiMessageElement = addMessageToUI("ai", "思考中...");
  }



  // Update AI status
  const aiStatus = document.getElementById("ai-status");
  if (aiStatus) {
    aiStatus.textContent = "AI: 处理中...";
  }

  // Close any existing stream
  const activeStream = getActiveStream();
  if (activeStream) {
    activeStream.close();
  }

  try {
    // Collect context files based on session mode
    let contextFiles = [];

    if (currentSession.mode === "manual") {
      // Manual mode: use user-selected files
      contextFiles = Array.from(selectedContextFiles);

      // If no files selected but current file exists, use current file
      if (
        contextFiles.length === 0 &&
        window.EditorModule &&
        window.EditorModule.getCurrentFilePath
      ) {
        const filePath = window.EditorModule.getCurrentFilePath();
        if (filePath) {
          contextFiles.push(filePath);
        }
      }
    }
    // Other modes (auto, all) are handled by backend

    let aiResponse = "";
    let toolUseData = null;
    let messageId = null; // 用于存储消息ID

    // Try using SSE streaming response
    if (typeof EventSource !== "undefined" && window.EventSource) {
      console.log("Using SSE streaming response");

      try {
        // Use streamCompletions for streaming response
        const stream = sessions.streamCompletions(
          currentSession.id,
          {
            type,
            content,
            session_id: currentSession.id,
            project_path: currentProject.path,
            context_files: contextFiles,
            model: selectedModel,
            tool_use: tool,
          },
          // Chunk handler
          (chunk) => {
            console.log("Received chunk:", chunk);
            if (
              chunk.content ||
              (chunk.parsed_line && Array.isArray(chunk.parsed_line))
            ) {
              aiResponse += chunk.content; // 累加内容

              // 如果收到了消息ID，保存它
              if (chunk.id) {
                messageId = chunk.id;
                console.log("收到消息ID:", messageId);
                // 将ID保存到消息元素中
                if (aiMessageElement) {
                  aiMessageElement.dataset.messageId = messageId;
                }
              }

              // Check for tool_use in parsed_line if available
              if (chunk.parsed_line && Array.isArray(chunk.parsed_line)) {
                const toolUseNode = chunk.parsed_line.find(
                  (node) => node.type === "tool_use" && !node.partial
                );

                // 暂存工具使用数据，等 stream 完成后处理
                if (toolUseNode && !toolUseData) {
                  toolUseData = toolUseNode;
                }
              }

              // 处理用户消息ID
              if (chunk.user_msg_id && !aiMessageElement.dataset.userMsgIdProcessed) {
                console.log("收到用户消息ID:", chunk.user_msg_id);
                // 获取最后一个用户消息元素
                const userMessages = document.querySelectorAll('.message[data-role="user"]');
                console.log("用户消息元素:", userMessages);
                if (userMessages.length > 0) {
                    const userMessageElement = userMessages[userMessages.length - 1];
                    userMessageElement.dataset.messageId = chunk.user_msg_id;
            
                    // 确保用户消息也有操作按钮
                    if (!userMessageElement.querySelector('.message-actions')) {
                        const messageActions = document.createElement("div");
                        messageActions.className = "message-actions";
                        messageActions.innerHTML = `
                            <button class="msg-action-btn copy-msg-btn" title="复制消息">复制</button>
                            <button class="msg-action-btn delete-msg-btn" title="删除消息">删除</button>
                        `;
                        userMessageElement.appendChild(messageActions);
                    }
            
                    // 获取按钮元素
                    const copyBtn = userMessageElement.querySelector('.copy-msg-btn');
                    const deleteBtn = userMessageElement.querySelector('.delete-msg-btn');
            
                    // 移除旧的事件监听器（如果有）
                    if (copyBtn) {
                        copyBtn.replaceWith(copyBtn.cloneNode(true)); // 这是移除事件监听器的简单方法
                        const newCopyBtn = userMessageElement.querySelector('.copy-msg-btn');
                        newCopyBtn.addEventListener('click', () => copyMessage(userMessageElement));
                    }
            
                    if (deleteBtn) {
                        deleteBtn.replaceWith(deleteBtn.cloneNode(true)); // 这是移除事件监听器的简单方法
                        const newDeleteBtn = userMessageElement.querySelector('.delete-msg-btn');
                        newDeleteBtn.addEventListener('click', () => deleteMessage(chunk.user_msg_id, userMessageElement));
                    }
                }
            
                // 标记为已处理，避免重复添加
                aiMessageElement.dataset.userMsgIdProcessed = "true";
            }

              if (aiMessageElement) {
                aiMessageElement.innerHTML = formatMessageContent(aiResponse);

                // 确保消息操作按钮仍然存在
                if (!aiMessageElement.querySelector('.message-actions')) {
                  const messageActions = document.createElement("div");
                  messageActions.className = "message-actions";
                  messageActions.innerHTML = `
                    <button class="msg-action-btn copy-msg-btn" title="复制消息">复制</button>
                    <button class="msg-action-btn delete-msg-btn" title="删除消息">删除</button>
                    <button class="msg-action-btn retry-msg-btn" title="重新生成">重试</button>
                  `;
                  aiMessageElement.appendChild(messageActions);

                  // 添加事件监听器
                  const copyBtn = aiMessageElement.querySelector('.copy-msg-btn');
                  const deleteBtn = aiMessageElement.querySelector('.delete-msg-btn');
                  const retryBtn = aiMessageElement.querySelector('.retry-msg-btn');

                  if (copyBtn) {
                    copyBtn.addEventListener('click', () => copyMessage(aiMessageElement));
                  }

                  if (deleteBtn) {
                    deleteBtn.addEventListener('click', () => deleteMessage(messageId || '', aiMessageElement));
                  }

                  if (retryBtn) {
                    retryBtn.addEventListener('click', () => retryMessage(messageId || ''));
                  }
                }

                const messagesContainer =
                  document.getElementById("messages-container");
                if (messagesContainer) {
                  messagesContainer.scrollTop = messagesContainer.scrollHeight;
                }
              }

              // 刷新预览功能
              if (chunk.filename !== "") {
                console.log("Tool use detected in single-html mode");
                console.log("Tchunk.filename: ", chunk.filename);
                showHtmlPreview(chunk.filename);
              }
            }

            if (chunk.error) {
              console.error("Server error:", chunk.error);
              if (aiMessageElement) {
                aiMessageElement.innerHTML += `<div class="error">错误: ${chunk.error}</div>`;
              }
            }
          },
          // Completion handler
          async () => {
            console.log("Stream completed");
            if (aiStatus) {
              aiStatus.textContent = "AI: 已连接";
            }

            // 确保消息元素包含最终的消息ID
            if (messageId && aiMessageElement) {
              aiMessageElement.dataset.messageId = messageId;

              // 更新消息操作按钮
              if (!aiMessageElement.querySelector('.message-actions')) {
                const messageActions = document.createElement("div");
                messageActions.className = "message-actions";
                messageActions.innerHTML = `
                  <button class="msg-action-btn copy-msg-btn" title="复制消息">复制</button>
                  <button class="msg-action-btn delete-msg-btn" title="删除消息">删除</button>
                  <button class="msg-action-btn retry-msg-btn" title="重新生成">重试</button>
                `;
                aiMessageElement.appendChild(messageActions);

                // 添加事件监听器
                const copyBtn = aiMessageElement.querySelector('.copy-msg-btn');
                const deleteBtn = aiMessageElement.querySelector('.delete-msg-btn');
                const retryBtn = aiMessageElement.querySelector('.retry-msg-btn');

                if (copyBtn) {
                  copyBtn.addEventListener('click', () => copyMessage(aiMessageElement));
                }

                if (deleteBtn) {
                  deleteBtn.addEventListener('click', () => deleteMessage(messageId, aiMessageElement));
                }

                if (retryBtn) {
                  retryBtn.addEventListener('click', () => retryMessage(messageId));
                }
              }
            }

            // 保存原始内容到数据属性
            if (aiMessageElement) {
              aiMessageElement.dataset.originalContent = aiResponse;
            }

            // 隐藏非最后一条消息的重试按钮
            hideRetryButtonsExceptLast();

            console.log("Tool use detected:", toolUseData);
            console.log("currentSession.mode:", currentSession.mode);

            // Handle tool_use confirmation if found
            // --- 在这里处理 single-html 模式 ---
            if (currentSession.mode === "single-html" && toolUseData) {
              console.log("Tool use detected in single-html mode");
              let previewPath = null;
              // (保持你现有的 previewPath 查找逻辑)
              if (
                toolUseData.params &&
                toolUseData.params.path &&
                toolUseData.params.path.endsWith(".html")
              ) {
                previewPath = toolUseData.params.path;
              } else if (
                toolUseData.name === "attempt_completion" &&
                toolUseData.params &&
                toolUseData.params.result &&
                typeof toolUseData.params.result === "string" &&
                toolUseData.params.result.endsWith(".html")
              ) {
                previewPath = toolUseData.params.result;
              } else if (
                toolUseData.name === "write_to_file" && // 如果是写入文件操作
                toolUseData.params &&
                toolUseData.params.path &&
                toolUseData.params.path.endsWith(".html")
              ) {
                previewPath = toolUseData.params.path;
              }
              // ... 可以添加其他查找 previewPath 的逻辑

              if (previewPath) {
                console.log("找到预览路径:", previewPath);
                // *** 调用新的 showHtmlPreview 函数 ***
                showHtmlPreview(previewPath); // <--- 修改点：调用更新后的函数
              } else {
                console.warn(
                  "single-html 模式下未找到有效的 HTML 文件路径进行预览。 Tool data:",
                  toolUseData
                );
                // 可以选择关闭旧预览或显示通知
                // closeHtmlPreview(); // 如果希望在没有新路径时关闭旧预览
                // 只显示工具确认（如果适用，例如非写入/完成操作）
                if (
                  toolUseData.name !== "write_to_file" &&
                  toolUseData.name !== "attempt_completion"
                ) {
                  // showToolUseConfirmation(toolUseData, () => { });
                }
              }
              // single-html 处理完毕，清除 toolUseData 避免后续通用逻辑再次处理
              toolUseData = null; // <--- 添加：防止重复处理
            } else if (toolUseData) {
              // Show tool use confirmation UI based on tool type
              console.log(
                "Tool use detected (non-single-html or no path):",
                toolUseData
              );
              showToolUseConfirmation(toolUseData, async (isConfirmed) => {
                // For attempt_completion, we don't need to do anything further
                if (toolUseData.name === "attempt_completion") {
                  return;
                }

                // For other tools, execute if confirmed
                try {
                  const toolResponse = await executeToolUse(
                    currentSession.id,
                    toolUseData,
                    isConfirmed,
                    currentProject.path,
                    contextFiles,
                    selectedModel,
                    currentSession,
                    selectedContextFiles,
                    setActiveStream,
                    getActiveStream
                  );
                } catch (toolError) {
                  console.error("Tool execution error:", toolError);
                  if (aiMessageElement) {
                    aiMessageElement.innerHTML += `<div class="error">工具执行错误: ${toolError.message}</div>`;
                  }
                }
              });
            }

            setActiveStream(null);

            // Auto-refresh current file after AI response
            if (window.FilesModule && window.FilesModule.refreshCurrentFile) {
              // 延迟一点刷新，确保文件写入完成（如果 AI 写入了文件）
              setTimeout(() => {
                if (window.EditorModule && window.EditorModule.getCurrentFilePath() === getCurrentPreviewPath()) {
                  // 如果当前编辑器打开的就是预览文件，则刷新编辑器内容
                  window.FilesModule.refreshCurrentFile();
                }
              }, 500);
            }
          },
          // Error handler
          (error) => {
            console.error("Stream error:", error);
            if (aiMessageElement) {
              aiMessageElement.textContent = "接收AI响应时出错";
            }
            if (aiStatus) {
              aiStatus.textContent = "AI: 错误";
            }
            setActiveStream(null);
          }
        );

        setActiveStream(stream);
      } catch (error) {
        console.error("Streaming response failed:", error);
        // Fallback to regular method
        const response = await sessions.sendMessage(
          currentSession.id,
          content,
          currentProject.path,
          contextFiles,
          selectedModel
        );

        if (response && response.ai_message) {
          if (aiMessageElement) {
            aiMessageElement.innerHTML = formatMessageContent(
              response.ai_message.content
            );
          }
        } else {
          if (aiMessageElement) {
            aiMessageElement.textContent = "未收到响应";
          }
        }

        if (aiStatus) {
          aiStatus.textContent = "AI: 已连接";
        }

        // Auto-refresh current file
        if (window.FilesModule && window.FilesModule.refreshCurrentFile) {
          window.FilesModule.refreshCurrentFile();
        }
      }
    } else {
      // EventSource not available, use traditional method
      // ... existing code ...
    }
  } catch (error) {
    console.error("发送消息失败:", error);
    if (aiMessageElement) {
      aiMessageElement.textContent = "发送消息时出错";
    }
    if (aiStatus) {
      aiStatus.textContent = "AI: 错误";
    }
  }
}

/**
 * Show tool use confirmation UI
 * @param {Object} toolUseData - Tool use data from AI
 * @param {Function} callback - Callback function that receives user choice (true/false)
 */
function showToolUseConfirmation(toolUseData, callback) {
  // For attempt_completion tools, don't show confirmation UI
  if (toolUseData.name === "attempt_completion") {
    // Show a friendly message indicating task is completed
    const aiMessageElement = document.querySelector(".message.ai:last-child");
    if (aiMessageElement) {
      const completionNote = document.createElement("div");
      completionNote.className = "completion-notification";
      completionNote.innerHTML = `<span style="font-weight:500">✓ 任务已完成</span>: ${toolUseData.params.result || ""
        }`;
      aiMessageElement.appendChild(completionNote);
    }

    // Call callback with false (no need for further action)
    if (callback) callback(false);
    return;
  }

  // Create tool confirmation UI for other tool types
  const confirmationUI = document.createElement("div");
  confirmationUI.className = "tool-use-confirmation";

  // Create content based on tool_use data
  let displayPath = toolUseData.params.path || "";
  let displayAction = "未知操作";
  let displayIcon = "🔧";
  let displayColor = "#3b82f6"; // Default blue

  // Determine action type and icon based on tool name
  switch (toolUseData.name) {
    case "write_to_file":
      displayAction = "写入文件";
      displayIcon = "📝";
      displayColor = "#0ea5e9"; // Blue
      break;
    case "read_file":
      displayAction = "读取文件";
      displayIcon = "📄";
      displayColor = "#10b981"; // Green
      break;
    case "delete_file":
      displayAction = "删除文件";
      displayIcon = "🗑️";
      displayColor = "#ef4444"; // Red
      break;
    case "execute_command":
      displayAction = "执行命令";
      displayIcon = "⌨️";
      displayColor = "#8b5cf6"; // Purple
      const commandText = toolUseData.params.command || "";
      displayPath = commandText; // Use command as the "path"
      break;
    case "insert_content":
      displayAction = "插入内容";
      displayIcon = "➕";
      displayColor = "#f59e0b"; // Amber
      break;
    case "list_files":
      displayAction = "列出文件";
      displayIcon = "📂";
      displayColor = "#64748b"; // Slate
      break;
    case "search_files":
      displayAction = "搜索文件";
      displayIcon = "🔍";
      displayColor = "#ec4899"; // Pink
      break;
    case "list_code_definition_names":
      displayAction = "列出代码定义";
      displayIcon = "📋";
      displayColor = "#14b8a6"; // Teal
      break;
    case "ask_followup_question":
      displayAction = "询问后续问题";
      displayIcon = "❓";
      displayColor = "#f97316"; // Orange
      const questionText = toolUseData.params.question || "";
      displayPath = questionText; // Use question as the "path"
      // ai咨询文问题，需要手动回复，无需同意
      return;
      break;
    case "attempt_completion":
      displayAction = "尝试完成";
      displayIcon = "✅";
      displayColor = "#22c55e"; // Emerald
      break;
    default:
      displayAction = "未知操作";
      displayIcon = "❔";
      displayColor = "#6b7280"; // Gray
      break;
  }

  // Create header
  const header = document.createElement("div");
  header.style.cssText = `
    font-weight: 500;
    margin-bottom: 8px;
    display: flex;
    justify-content: space-between;
    align-items: center;
  `;
  header.innerHTML = `
    <span style="display:flex;align-items:center;gap:8px">
      <span style="font-size:1.2em;color:${displayColor}">${displayIcon}</span>
      <span>AI 请求执行操作</span>
    </span>
    <button class="icon-btn close-btn" style="background:none;border:none;cursor:pointer;opacity:0.7;font-size:1.1em;">✕</button>
  `;
  confirmationUI.appendChild(header);

  // Create details
  const details = document.createElement("div");
  details.style.cssText = "margin-bottom: 12px;";

  if (toolUseData.name === "execute_command") {
    details.innerHTML = `
      <div><strong>操作：</strong>${displayAction}</div>
      <div style="margin-top: 5px;"><strong>命令：</strong><code style="background:#f1f5f9;padding:2px 4px;border-radius:3px;font-family:monospace;">${displayPath}</code></div>
    `;
  } else if (
    toolUseData.name === "write_to_file" &&
    toolUseData.params.content
  ) {
    const contentPreview =
      toolUseData.params.content.length > 100
        ? toolUseData.params.content.substring(0, 100) + "..."
        : toolUseData.params.content;

    details.innerHTML = `
      <div><strong>操作：</strong>${displayAction}</div>
      <div style="margin-top: 5px;"><strong>文件路径：</strong><code style="background:#f1f5f9;padding:2px 4px;border-radius:3px;font-family:monospace;">${displayPath}</code></div>
      <div style="margin-top: 5px;"><strong>内容预览：</strong></div>
      <div style="margin-top: 3px;max-height:80px;overflow-y:auto;background:#f1f5f9;padding:6px;border-radius:4px;font-family:monospace;font-size:0.85em;white-space:pre-wrap;">${contentPreview}</div>
    `;
  } else {
    details.innerHTML = `
      <div><strong>操作：</strong>${displayAction}</div>
      <div style="margin-top: 5px;"><strong>文件路径：</strong><code style="background:#f1f5f9;padding:2px 4px;border-radius:3px;font-family:monospace;">${displayPath}</code></div>
    `;
  }

  confirmationUI.appendChild(details);

  // Add buttons
  const buttonContainer = document.createElement("div");
  buttonContainer.style.cssText = `
    display: flex;
    justify-content: flex-end;
    gap: 10px;
  `;

  const rejectButton = document.createElement("button");
  rejectButton.className = "reject-btn";
  rejectButton.textContent = "拒绝";
  rejectButton.onmouseover = () => {
    rejectButton.style.backgroundColor = "#e2e8f0";
  };
  rejectButton.onmouseout = () => {
    rejectButton.style.backgroundColor = "#f1f5f9";
  };

  const acceptButton = document.createElement("button");
  acceptButton.className = "accept-btn";
  acceptButton.textContent = "接受";
  acceptButton.style.cssText = `
    padding: 6px 14px;
    background-color: var(--primary-color);
    border: none;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
    color: white;
    transition: all 0.2s;
  `;
  acceptButton.onmouseover = () => {
    acceptButton.style.filter = "brightness(1.1)";
  };
  acceptButton.onmouseout = () => {
    acceptButton.style.filter = "brightness(1)";
  };

  buttonContainer.appendChild(rejectButton);
  buttonContainer.appendChild(acceptButton);
  confirmationUI.appendChild(buttonContainer);

  // Add to the DOM - insert before the input container
  const contextFilesSection = document.querySelector(".context-files-section");
  if (contextFilesSection) {
    contextFilesSection.parentNode.insertBefore(
      confirmationUI,
      contextFilesSection
    );
  }

  // Add CSS animation to the document if it doesn't exist
  if (!document.getElementById("tool-use-confirmation-style")) {
    const style = document.createElement("style");
    style.id = "tool-use-confirmation-style";
    style.textContent = `
      @keyframes slideDown {
        from { transform: translateY(-20px); opacity: 0; }
        to { transform: translateY(0); opacity: 1; }
      }
      @keyframes slideUp {
        from { transform: translateY(0); opacity: 1; }
        to { transform: translateY(-20px); opacity: 0; }
      }
      .sliding-out {
        animation: slideUp 0.3s ease-out forwards;
      }
    `;
    document.head.appendChild(style);
  }

  // Add event listeners
  const closeButton = header.querySelector(".close-btn");
  closeButton.addEventListener("click", () => {
    removeConfirmationUI(false);
  });

  rejectButton.addEventListener("click", () => {
    removeConfirmationUI(false);
  });

  acceptButton.addEventListener("click", () => {
    removeConfirmationUI(true);
  });

  // Function to remove UI with animation and trigger callback
  function removeConfirmationUI(isAccepted) {
    confirmationUI.classList.add("sliding-out");
    setTimeout(() => {
      if (confirmationUI.parentNode) {
        confirmationUI.parentNode.removeChild(confirmationUI);
      }
      if (callback) callback(isAccepted);
    }, 300); // Match animation duration
  }
}

/**
 * Execute a tool use action
 * @param {string} sessionId - Current session ID
 * @param {Object} toolUseData - Tool use data from AI
 * @param {boolean} isConfirmed - Whether user confirmed the action
 * @param {string} projectPath - Current project path
 * @param {Array} contextFiles - Context files
 * @param {string} selectedModel - Selected model
 * @returns {Promise<Object>} - Response from API
 */
async function executeToolUse(
  sessionId,
  toolUseData,
  isConfirmed,
  projectPath,
  contextFiles,
  selectedModel,
  currentSession,
  selectedContextFiles,
  setActiveStream,
  getActiveStream
) {
  let msg;
  if (isConfirmed) {
    msg = "正在执行工具操作...";
  } else {
    msg = "已拒绝工具操作。";
    console.log("已拒绝工具操作。");
  }

  return sendMessage(
    "tool_use",
    msg,
    currentSession,
    selectedContextFiles,
    setActiveStream,
    getActiveStream,
    {
      tool_use: {
        name: toolUseData.name,
        params: toolUseData.params,
      },
      project_path: projectPath,
      context_files: contextFiles,
      model: selectedModel,
      confirmed: isConfirmed,
    }
  );
}

/**
 * Add message to UI
 * @param {string} role - Message role ('user' or 'ai')
 * @param {string} content - Message content
 * @returns {HTMLElement} Message element
 */
export function addMessageToUI(role, content) {
  const messagesContainer = document.getElementById("messages-container");
  if (!messagesContainer) return null;

  const messageElement = document.createElement("div");
  messageElement.className = `message ${role}`;
  messageElement.dataset.role = role;
  messageElement.dataset.originalContent = content; // 保存原始Markdown内容

  // 生成唯一ID用于消息操作
  const messageId = Date.now().toString();
  messageElement.dataset.messageId = messageId;

  // 渲染Markdown内容
  messageElement.innerHTML = formatMessageContent(content);

  // 消息内容
  messageElement.innerHTML = formatMessageContent(content);

  // 添加消息操作按钮
  const messageActions = document.createElement("div");
  messageActions.className = "message-actions";

  if (role === "user") {
    // 用户消息操作：复制和删除
    messageActions.innerHTML = `
      <button class="msg-action-btn copy-msg-btn" title="复制消息">复制</button>
      <button class="msg-action-btn delete-msg-btn" title="删除消息">删除</button>
    `;
    // messageActions.innerHTML ='';
  } else if (role === "ai" || role === "assistant") {
    // AI 消息操作：复制、删除、重试（仅最后一条消息显示重试）
    messageActions.innerHTML = `
      <button class="msg-action-btn copy-msg-btn" title="复制消息">复制</button>
      <button class="msg-action-btn delete-msg-btn" title="删除消息">删除</button>
      <button class="msg-action-btn retry-msg-btn" title="重新生成">重试</button>
    `;
  }

  messageElement.appendChild(messageActions);

  // 添加事件监听器
  setTimeout(() => {
    const copyBtn = messageElement.querySelector('.copy-msg-btn');
    const deleteBtn = messageElement.querySelector('.delete-msg-btn');
    const retryBtn = messageElement.querySelector('.retry-msg-btn');

    if (copyBtn) {
      copyBtn.addEventListener('click', () => copyMessage(messageElement));
    }

    if (deleteBtn) {
      deleteBtn.addEventListener('click', () => deleteMessage(messageId, messageElement));
    }

    if (retryBtn) {
      retryBtn.addEventListener('click', () => retryMessage(messageId));
    }
  }, 0);

  messagesContainer.appendChild(messageElement);
  messagesContainer.scrollTop = messagesContainer.scrollHeight;

  // 如果添加了新消息，隐藏所有之前的 AI 消息上的重试按钮
  if (role === "user") {
    hideRetryButtonsExceptLast();
  }

  return messageElement;
}

/**
 * 隐藏所有非最后一条 AI 消息的重试按钮
 */
function hideRetryButtonsExceptLast() {
  const aiMessages = document.querySelectorAll('.message[data-role="ai"], .message[data-role="assistant"]');

  if (aiMessages.length <= 1) return;

  // 隐藏所有 AI 消息的重试按钮
  aiMessages.forEach(msg => {
    const retryBtn = msg.querySelector('.retry-msg-btn');
    if (retryBtn) {
      retryBtn.style.display = 'none';
    }
  });

  // 显示最后一条 AI 消息的重试按钮
  const lastAiMessage = aiMessages[aiMessages.length - 1];
  const retryBtn = lastAiMessage.querySelector('.retry-msg-btn');
  if (retryBtn) {
    retryBtn.style.display = 'inline-block';
  }
}


/**
 * 复制消息内容到剪贴板（原始Markdown格式）
 * @param {HTMLElement} messageElement - 消息元素
 */
function copyMessage(messageElement) {
  // 获取原始消息内容（从数据属性或存储中）
  const originalContent = messageElement.dataset.originalContent ||
    messageElement.textContent;

  // 创建临时文本区域用于复制
  const textArea = document.createElement("textarea");
  textArea.value = originalContent;
  textArea.style.position = "fixed";
  textArea.style.left = "-999999px";
  textArea.style.top = "-999999px";
  document.body.appendChild(textArea);
  textArea.focus();
  textArea.select();

  try {
    const successful = document.execCommand("copy");
    if (successful) {
      showCopyFeedback(messageElement);
    } else {
      console.error("复制失败");
      alert("复制失败，请重试");
    }
  } catch (err) {
    console.error("复制失败:", err);
    alert("复制失败，请重试");
  }

  document.body.removeChild(textArea);
}

/**
 * 显示复制成功反馈
 * @param {HTMLElement} messageElement - 消息元素
 */
function showCopyFeedback(messageElement) {
  const copyBtn = messageElement.querySelector('.copy-msg-btn');
  if (!copyBtn) return;

  const originalText = copyBtn.textContent;
  copyBtn.textContent = '已复制!';
  copyBtn.disabled = true;

  setTimeout(() => {
    copyBtn.textContent = originalText;
    copyBtn.disabled = false;
  }, 1500);
}

/**
 * 删除指定消息
 * @param {string} messageId - 消息 ID
 * @param {HTMLElement} messageElement - 消息元素
 */
async function deleteMessage(messageId, messageElement) {
  try {
    const currentSession = window.SessionsModule.getCurrentSession();
    if (!currentSession || !currentSession.id) {
      throw new Error('无法获取当前会话信息');
    }

    // 使用 client.js 中的 API 方法
    await sessions.deleteMessage(currentSession.id, messageId);

    // 从 UI 中移除消息
    messageElement.remove();

    // 如果消息已经删除，可能需要重新显示最后一条消息的重试按钮
    hideRetryButtonsExceptLast();

    // 检查是否还有消息
    const messagesContainer = document.getElementById("messages-container");
    if (messagesContainer && messagesContainer.children.length === 0) {
      messagesContainer.innerHTML = '<div class="empty-message">暂无消息</div>';
    }

  } catch (error) {
    console.error('删除消息失败:', error);
    alert(`删除失败: ${error.message}`);
  }
}


/**
 * 重试 AI 消息生成
 * @param {string} messageId - 需要重试的消息 ID
 */
function retryMessage(messageId) {
  console.log('重试消息:', messageId);
  const currentSession = window.SessionsModule.getCurrentSession();
  if (!currentSession) {
    alert('无法获取当前会话信息');
    return;
  }
  console.log('当前会话:', currentSession);

  // 调用 sendMessage 发送重试请求
  sendMessage(
    "retry",
    "" + messageId, // 传递消息 ID 作为内容
    currentSession,
    window.SessionsModule.getSelectedContextFiles(),
    window.SessionsModule.setActiveStream,
    window.SessionsModule.getActiveStream
  );
}

/**
 * 清空当前会话的所有消息
 */
export async function clearAllMessages() {
  if (!confirm('确定要清空所有消息吗？此操作不可撤销！')) return;

  try {
    const currentSession = window.SessionsModule.getCurrentSession();
    if (!currentSession || !currentSession.id) {
      throw new Error('无法获取当前会话信息');
    }

    // 使用 client.js 中的 API 方法
    await sessions.clearAllMessages(currentSession.id);

    // 清空 UI
    const messagesContainer = document.getElementById("messages-container");
    if (messagesContainer) {
      messagesContainer.innerHTML = '<div class="empty-message">暂无消息</div>';
    }

  } catch (error) {
    showNotification("清空消息失败:" + error, "warning");
    console.error('清空消息失败:', error);
    alert(`清空失败: ${error.message}`);
  }
}

/**
 * 修改 renderMessages 函数以支持消息操作按钮
 * @param {Array} messages - 消息列表
 */
export function renderMessages(messages) {
  const messagesContainer = document.getElementById("messages-container");
  if (!messagesContainer) return;

  messagesContainer.innerHTML = "";
  const hasMessages = messages && messages.length > 0;

  if (!hasMessages) {
    messagesContainer.innerHTML = '<div class="empty-message">暂无消息</div>';
    return;
  }

  messages.forEach((message, index) => {
    const messageElement = document.createElement("div");
    messageElement.className = `message ${message.role}`;
    messageElement.dataset.role = message.role;
    messageElement.dataset.originalContent = message.content; // 保存原始Markdown内容
    messageElement.dataset.messageId = message.id || index.toString();

    // 处理消息内容
    const content = formatMessageContent(message.content);
    messageElement.innerHTML = content;

    // 添加消息操作按钮
    const messageActions = document.createElement("div");
    messageActions.className = "message-actions";

    if (message.role === "user") {
      messageActions.innerHTML = `
        <button class="msg-action-btn copy-msg-btn" title="复制消息">复制</button>
        <button class="msg-action-btn delete-msg-btn" title="删除消息">删除</button>
      `;
    } else if (message.role === "ai" || message.role === "assistant") {
      messageActions.innerHTML = `
        <button class="msg-action-btn copy-msg-btn" title="复制消息">复制</button>
        <button class="msg-action-btn delete-msg-btn" title="删除消息">删除</button>
        <button class="msg-action-btn retry-msg-btn" title="重新生成">重试</button>
      `;
    }

    messageElement.appendChild(messageActions);
    messagesContainer.appendChild(messageElement);

    // 添加事件监听器
    const copyBtn = messageElement.querySelector('.copy-msg-btn');
    const deleteBtn = messageElement.querySelector('.delete-msg-btn');
    const retryBtn = messageElement.querySelector('.retry-msg-btn');

    if (copyBtn) {
      copyBtn.addEventListener('click', () => copyMessage(messageElement));
    }

    if (deleteBtn) {
      deleteBtn.addEventListener('click', () => deleteMessage(message.id || index.toString(), messageElement));
    }

    if (retryBtn) {
      retryBtn.addEventListener('click', () => retryMessage(message.id || index.toString()));
    }
  });

  // 隐藏非最后一条消息的重试按钮
  hideRetryButtonsExceptLast();

  // 滚动到底部
  messagesContainer.scrollTop = messagesContainer.scrollHeight;
}

/**
 * Format message content with markdown and code blocks
 * @param {string} content - Message content
 * @returns {string} Formatted HTML
 */
export function formatMessageContent(content) {
  if (!content) return "";

  // If marked library isn't loaded, use simple formatting
  if (typeof marked === "undefined") {
    // Fallback to simple code block replacement
    return formatMessageContentPlain(content);
  }

  // Configure marked options
  marked.setOptions({
    highlight: function (code, lang) {
      return code;
    },
  });

  // Remove <tool_use> tags from content to prevent them from being displayed
  let cleanContent = content.replace(/\<tool_use\>/g, "");

  // Convert content to HTML using marked
  const html = marked.parse(cleanContent);

  // Find all code blocks and add copy buttons
  const tempDiv = document.createElement("div");
  tempDiv.innerHTML = html;

  const codeBlocks = tempDiv.querySelectorAll("pre > code");
  codeBlocks.forEach((codeBlock) => {
    const pre = codeBlock.parentNode;
    pre.classList.add("code-block");

    const copyBtn = document.createElement("button");
    copyBtn.className = "copy-code-btn";
    copyBtn.textContent = "复制";
    copyBtn.setAttribute("onclick", "copyToClipboard(this)");

    pre.insertBefore(copyBtn, pre.firstChild);
  });

  return tempDiv.innerHTML;
}

/**
 * Simple format for message content (plain)
 * @param {string} content - Message content
 * @returns {string} Formatted HTML
 */
function formatMessageContentPlain(content) {
  // Simple formatting, handle markdown-like code blocks
  // Replace ```language\n...\n``` with <pre><code>...</code></pre>
  return content.replace(
    /```([\w-]*)\n([\s\S]*?)\n```/g,
    function (match, language, code) {
      return `<pre class="code-block ${language}"><code>${escapeHtml(
        code
      )}</code></pre>`;
    }
  );
}

// Define global copyToClipboard function if it doesn't exist
if (typeof window.copyToClipboard !== "function") {
  window.copyToClipboard = function (button) {
    const pre = button.closest("pre");
    const code = pre.querySelector("code").textContent;

    // Create temporary textarea
    const textArea = document.createElement("textarea");
    textArea.value = code;
    textArea.style.position = "fixed"; // Avoid scrolling to bottom
    textArea.style.left = "-999999px";
    textArea.style.top = "-999999px";
    document.body.appendChild(textArea);
    textArea.focus();
    textArea.select();

    try {
      const successful = document.execCommand("copy");
      if (successful) {
        button.textContent = "已复制!";
        setTimeout(function () {
          button.textContent = "复制";
        }, 2000);
      } else {
        button.textContent = "复制失败";
        setTimeout(function () {
          button.textContent = "复制";
        }, 2000);
      }
    } catch (err) {
      console.error("复制失败:", err);
      button.textContent = "复制失败";
      setTimeout(function () {
        button.textContent = "复制";
      }, 2000);
    }

    document.body.removeChild(textArea);
  };
}

/**
 * Session manager component
 * @module modules/sessions/sessionManager
 */
import { sessions } from "../../api/client.js";
import { showNotification } from "../../utils/dom.js";
import { renderMessages } from "./messaging.js";
import { updateContextFilesSelector } from "./contextManager.js";

/**
 * Load sessions for a project
 * @param {string} projectId - Project ID
 * @param {Function} setCurrentSession - Function to set current session
 */
export async function loadSessions(projectId, setCurrentSession) {
  try {
    const projectSessions = await sessions.getByProject(projectId);
    renderSessions(projectSessions, setCurrentSession);
  } catch (error) {
    console.error("加载会话失败:", error);
    showNotification("加载会话失败", "error");
  }
}

/**
 * Render sessions in the UI
 * @param {Array} sessions - Sessions to render
 * @param {Function} setCurrentSession - Function to set current session
 */
export function renderSessions(projectSessions, setCurrentSession) {
  const sessionsList = document.getElementById("sessions-list");
  if (!sessionsList) return;

  sessionsList.innerHTML = "";

  if (projectSessions.length === 0) {
    sessionsList.innerHTML = '<div class="empty-message">未找到会话</div>';
    return;
  }

  // 创建会话下拉容器，添加sessions-dropdown-container类名以区分
  const sessionsDropdown = document.createElement("div");
  sessionsDropdown.className = "sessions-dropdown sessions-dropdown-container";

  // 创建当前选择的会话显示区域
  const dropdownHeader = document.createElement("div");
  dropdownHeader.className = "sessions-dropdown-header sessions-header";

  const currentSessionName = document.createElement("span");
  currentSessionName.textContent = "选择会话";

  const dropdownIcon = document.createElement("span");
  dropdownIcon.textContent = "▼";

  dropdownHeader.appendChild(currentSessionName);
  dropdownHeader.appendChild(dropdownIcon);
  sessionsDropdown.appendChild(dropdownHeader);

  // 创建下拉内容区域，添加sessions-content类名以区分
  const dropdownContent = document.createElement("div");
  dropdownContent.className = "sessions-dropdown-content sessions-content";

  // 切换下拉显示/隐藏
  dropdownHeader.addEventListener("click", function () {
    dropdownContent.classList.toggle("active");
  });

  // 点击其他地方关闭下拉框 - 修改为只关闭会话下拉框
  document.addEventListener("click", function (event) {
    // 只有当点击的不是会话下拉框内的元素时，才关闭会话下拉框
    if (!sessionsDropdown.contains(event.target)) {
      dropdownContent.classList.remove("active");
    }
  });

  projectSessions.forEach((session) => {
    const sessionElement = document.createElement("div");
    sessionElement.className = "session-item";

    const sessionName = document.createElement("div");
    sessionName.className = "session-item-name";
    sessionName.textContent = session.name;
    sessionElement.appendChild(sessionName);

    // Create actions buttons container
    const sessionActions = document.createElement("div");
    sessionActions.className = "session-item-actions";

    // Edit button
    const editBtn = document.createElement("button");
    editBtn.className = "icon-btn";
    editBtn.innerHTML = "✎";
    editBtn.title = "编辑会话";
    editBtn.addEventListener("click", (e) => {
      e.stopPropagation();
      //   window.SessionsModule.showEditSessionModal(session.id);
      if (
        window.SessionsModule &&
        typeof window.SessionsModule.showEditSessionModal === "function"
      ) {
        window.SessionsModule.showEditSessionModal(session.id);
      } else {
        console.error("SessionsModule 或 showEditSessionModal 未定义！");
      }
    });
    sessionActions.appendChild(editBtn);

    // Delete button
    const deleteBtn = document.createElement("button");
    deleteBtn.className = "icon-btn";
    deleteBtn.innerHTML = "×";
    deleteBtn.title = "删除会话";
    deleteBtn.addEventListener("click", (e) => {
      e.stopPropagation();
      // Confirm delete
      if (confirm(`确定要删除会话 "${session.name}" 吗？`)) {
        deleteSession(session.id, setCurrentSession);
      }
    });

    sessionActions.appendChild(deleteBtn);
    sessionElement.appendChild(sessionActions);

    // Click on session item to select session
    sessionElement.addEventListener("click", () =>
      selectSession(session.id, setCurrentSession)
    );

    dropdownContent.appendChild(sessionElement);
  });

  sessionsDropdown.appendChild(dropdownContent);
  sessionsList.appendChild(sessionsDropdown);
}

/**
 * Delete a session
 * @param {string} sessionId - Session ID
 * @param {Function} setCurrentSession - Function to set current session
 */
export async function deleteSession(sessionId, setCurrentSession) {
  try {
    await sessions.delete(sessionId);
    showNotification("会话已删除", "success");

    // If deleted session is current session, clear current session
    const currentSession = window.SessionsModule.getCurrentSession();
    if (currentSession && currentSession.id === sessionId) {
      setCurrentSession(null);
      const messagesContainer = document.getElementById("messages-container");
      if (messagesContainer) {
        messagesContainer.innerHTML =
          '<div class="empty-message">暂无消息</div>';
      }

      const userMessageInput = document.getElementById("user-message");
      const sendMessageBtn = document.getElementById("send-message-btn");
      if (userMessageInput) userMessageInput.disabled = true;
      if (sendMessageBtn) sendMessageBtn.disabled = true;
    }

    // Reload sessions list
    const currentProject = window.ProjectsModule.getCurrentProject();
    if (currentProject) {
      loadSessions(currentProject.id, setCurrentSession);
    }
  } catch (error) {
    console.error("删除会话失败:", error);
    showNotification("删除会话失败", "error");
  }
}

/**
 * Select a session
 * @param {string} sessionId - Session ID
 * @param {Function} setCurrentSession - Function to set current session
 */
export async function selectSession(sessionId, setCurrentSession) {
  try {
    const session = await sessions.getById(sessionId);
    setCurrentSession(session);

    // 更新下拉框标题 - 修改为只更新会话下拉框
    const dropdownHeader = document.querySelector(
      ".sessions-header span:first-child"
    );
    if (dropdownHeader) {
      dropdownHeader.textContent = session.name;
    }

    // 隐藏下拉内容
    document
      .querySelector(".sessions-content")
      .classList.remove("active");

    // Render messages
    renderMessages(session.messages || []);

    // Update AI status
    document.getElementById("ai-status").textContent = "AI: 就绪";

    // Enable or disable file selector based on session mode
    const isManualMode = session.mode === "manual";
    const toggleContextFilesBtn = document.getElementById(
      "toggle-context-files"
    );
    if (toggleContextFilesBtn) {
      toggleContextFilesBtn.style.display = isManualMode
        ? "inline-block"
        : "none";
    }

    // Update context files selector
    updateContextFilesSelector();

    // Enable message input
    const userMessageInput = document.getElementById("user-message");
    const sendMessageBtn = document.getElementById("send-message-btn");
    if (userMessageInput) userMessageInput.disabled = false;
    if (sendMessageBtn) sendMessageBtn.disabled = false;
  } catch (error) {
    console.error("加载会话失败:", error);
    showNotification("加载会话失败", "error");
  }
}

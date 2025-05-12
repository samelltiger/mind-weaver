/**
 * Sessions module - Main entry point
 * @module modules/sessions
 */
import { loadSessions, renderSessions, selectSession, deleteSession } from './sessionManager.js';
import { toggleContextFilesSelector, updateContextFilesSelector } from './contextManager.js';
import { showNewSessionModal, hideNewSessionModal, createSession } from './sessionCreation.js';
import { createEditSessionModal, showEditSessionModal, hideEditSessionModal, saveSessionEdit } from './sessionEditing.js';
import { sendMessage, addMessageToUI, formatMessageContent, clearAllMessages } from './messaging.js';
import { closeHtmlPreview } from './messagingPreview.js';
import { hideFileSelector } from './fileSelector.js';
import { loadModels } from './models.js';

let currentSession = null;
let activeStream = null;
let selectedContextFiles = new Set();

let MSG_TYPE_NORMAL = "normal";
let MSG_TYPE_EXPLAIN = "explain";

/**
 * Initialize sessions module
 */
export function init() {
  // Attach event listeners
  const newSessionBtn = document.getElementById("new-session-btn");
  const newSessionModal = document.getElementById("new-session-modal");
  const newSessionForm = document.getElementById("new-session-form");
  const sendMessageBtn = document.getElementById("send-message-btn");
  const sendContinueMessageBtn = document.getElementById("send-continue-message-btn");
  const userMessageInput = document.getElementById("user-message");
  const toggleContextFilesBtn = document.getElementById("toggle-context-files");
  const explainCodeBtn = document.getElementById("explain-code");
  const clearMessagesBtn = document.getElementById("clear-messages-btn");

  // 绑定清空消息按钮事件
  if (clearMessagesBtn) {
    clearMessagesBtn.addEventListener("click", clearAllMessages);
  }

  if (newSessionBtn) {
    newSessionBtn.addEventListener("click", showNewSessionModal);
  }

  if (newSessionForm) {
    newSessionForm.addEventListener("submit", (event) => createSession(event, getCurrentSession));
  }

  if (newSessionModal) {
    const cancelBtn = newSessionModal.querySelector(".cancel-btn");
    if (cancelBtn) {
      cancelBtn.addEventListener("click", hideNewSessionModal);
    }
  }

  if (sendMessageBtn && userMessageInput) {
    sendMessageBtn.addEventListener("click", () => sendMessage(
      MSG_TYPE_NORMAL,
      userMessageInput.value,
      currentSession,
      selectedContextFiles,
      setActiveStream,
      getActiveStream
    ));

    let isComposing = false;
    userMessageInput.addEventListener('compositionstart', () => {
      isComposing = true;
    });

    userMessageInput.addEventListener('compositionend', () => {
      isComposing = false;
    });
    userMessageInput.addEventListener("keydown", function (event) {
      if (event.key === "Enter" && !event.shiftKey && !isComposing) {
        event.preventDefault();
        sendMessage(
          MSG_TYPE_NORMAL,
          userMessageInput.value,
          currentSession,
          selectedContextFiles,
          setActiveStream,
          getActiveStream
        );
      }
    });
  }

  // 添加全局事件监听器处理模态框关闭
  document.addEventListener('keydown', function (event) {
    if (event.key === 'Escape') {
      // 确保所有模态框都被正确关闭
      const newSessionModal = document.getElementById("new-session-modal");
      const fileSelectorModal = document.getElementById("file-selector-modal");
      const editSessionModal = document.getElementById("edit-session-modal");

      if (newSessionModal && newSessionModal.classList.contains('active')) {
        hideNewSessionModal();
      }

      if (fileSelectorModal) {
        hideFileSelector();
      }

      if (editSessionModal && editSessionModal.classList.contains('active')) {
        hideEditSessionModal();
      }
    }
  });

  // Add continue message button event listener
  if (sendContinueMessageBtn) {
    sendContinueMessageBtn.addEventListener("click", () => sendMessage(
      MSG_TYPE_NORMAL,
      // "continue", 
      "请继续",
      currentSession,
      selectedContextFiles,
      setActiveStream,
      getActiveStream
    ));
  }

  // Create edit session modal if it doesn't exist
  if (!document.getElementById('edit-session-modal')) {
    createEditSessionModal(getCurrentSession);
  }

  // Context files selector toggle button
  if (toggleContextFilesBtn) {
    toggleContextFilesBtn.addEventListener("click", toggleContextFilesSelector);
  }

  // 解释代码按钮
  if (explainCodeBtn && userMessageInput) {
    explainCodeBtn.addEventListener("click", function() {
      if (currentSession.mode!=="manual") {
        window.ProjectsModule.showNotification("请切换到手动模式", "warning");
        return;
      }
      sendMessage(
        MSG_TYPE_EXPLAIN,
        userMessageInput.value,
        currentSession,
        selectedContextFiles,
        setActiveStream,
        getActiveStream
      );
    })
  }

  // Load models list
  loadModels();

  // Add update selected code button event
  const updateSelectionBtn = document.getElementById("update-selection-btn");
  if (updateSelectionBtn) {
    updateSelectionBtn.addEventListener("click", function () {
      if (window.EditorModule && typeof window.EditorModule.updateSelectedCodeToContext === "function") {
        window.EditorModule.updateSelectedCodeToContext();
        window.ProjectsModule.showNotification("已更新选中代码到上下文", "info");
      } else {
        console.error("更新选中代码方法不可用");
      }
    });
  }
}

/**
 * Set the active stream
 * @param {Object} stream - Stream object
 */
function setActiveStream(stream) {
  activeStream = stream;
}

/**
 * Get the active stream
 * @returns {Object} Active stream
 */
function getActiveStream() {
  return activeStream;
}

/**
 * Load project sessions
 * @param {string} projectId - Project ID
 */
export async function loadProjectSessions(projectId) {
  await loadSessions(projectId, setCurrentSession);
}

/**
 * Set current session
 * @param {Object} session - Session object
 */
export function setCurrentSession(session) {
  currentSession = session;
}

/**
 * Get current session
 * @returns {Object} Current session
 */
export function getCurrentSession() {
  return currentSession;
}

/**
 * Add file to selected context files
 * @param {string} filePath - File path
 */
export function addContextFile(filePath) {
  selectedContextFiles.add(filePath);
}

/**
 * Remove file from selected context files
 * @param {string} filePath - File path
 */
export function removeContextFile(filePath) {
  selectedContextFiles.delete(filePath);
}

/**
 * Get selected context files
 * @returns {Set} Selected context files
 */
export function getSelectedContextFiles() {
  return selectedContextFiles;
}

// Export the public API
export default {
  init,
  loadProjectSessions,
  getCurrentSession,
  updateContextFilesSelector,
  addContextFile,
  removeContextFile,
  getSelectedContextFiles,
  showEditSessionModal,
  setCurrentSession,
  selectSession,
  renderSessions,
  deleteSession,
  hideEditSessionModal,
  saveSessionEdit,
  addMessageToUI,
  formatMessageContent,
  closeHtmlPreview,
  setActiveStream,
  getActiveStream
};
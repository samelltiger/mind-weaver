/**
 * Context manager component
 * @module modules/sessions/contextManager
 */
import { sessions } from '../../api/client.js';

/**
 * Toggle context files selector
 */
export function toggleContextFilesSelector() {
  const contextFilesSelector = document.getElementById("context-files-selector");
  const toggleContextFilesBtn = document.getElementById("toggle-context-files");

  if (contextFilesSelector) {
    contextFilesSelector.classList.toggle("visible");
    if (toggleContextFilesBtn) {
      toggleContextFilesBtn.textContent =
        contextFilesSelector.classList.contains("visible")
          ? "隐藏文件"
          : "选择文件";
    }
  }
}

/**
 * Update context files selector based on session mode
 */
export function updateContextFilesSelector() {
  const contextFilesSelector = document.getElementById("context-files-selector");
  if (!contextFilesSelector) return;

  const currentSession = window.SessionsModule.getCurrentSession();
  if (!currentSession) {
    const toggleContextFilesBtn = document.getElementById("toggle-context-files");
    if (toggleContextFilesBtn) toggleContextFilesBtn.style.display = "none";
    return;
  }

  // Show or hide file selector button based on mode
  const isManualMode = currentSession.mode === "manual";
  const toggleContextFilesBtn = document.getElementById("toggle-context-files");
  if (toggleContextFilesBtn) {
    toggleContextFilesBtn.style.display = isManualMode
      ? "inline-block"
      : "none";
  }

  // If not manual mode, don't show file selector
  if (!isManualMode) {
    contextFilesSelector.classList.remove("visible");
    return;
  }

  // Make sure FilesModule exists and getOpenFiles method is available
  if (!window.FilesModule || typeof window.FilesModule.getOpenFiles !== "function") {
    console.warn("FilesModule or getOpenFiles method not available");
    return;
  }

  const openFiles = window.FilesModule.getOpenFiles();
  const currentFilePath =
    window.EditorModule && window.EditorModule.getCurrentFilePath
      ? window.EditorModule.getCurrentFilePath()
      : null;

  // Clear selector
  contextFilesSelector.innerHTML = "";

  if (!openFiles || openFiles.length === 0) {
    contextFilesSelector.innerHTML =
      '<div class="empty-message">无打开的文件</div>';
    return;
  }

  // Reset selected files if current file path exists and no files are selected
  const selectedContextFiles = window.SessionsModule.getSelectedContextFiles();
  if (currentFilePath && !selectedContextFiles.size) {
    window.SessionsModule.addContextFile(currentFilePath);
  }

  // Create selection item for each open file
  openFiles.forEach((file) => {
    // Make sure file is not null and has path property
    if (!file || !file.path) return;

    const fileItem = document.createElement("div");
    fileItem.className = "context-file-item";

    const checkbox = document.createElement("input");
    checkbox.type = "checkbox";
    checkbox.id = `context-file-${file.path}`;
    checkbox.value = file.path;
    checkbox.checked = selectedContextFiles.has(file.path);

    checkbox.addEventListener("change", function () {
      if (this.checked) {
        window.SessionsModule.addContextFile(file.path);
      } else {
        window.SessionsModule.removeContextFile(file.path);
      }
    });

    const label = document.createElement("label");
    label.htmlFor = `context-file-${file.path}`;
    label.textContent = file.name || "未命名文件";

    // Highlight current active file
    if (file.path === currentFilePath) {
      label.style.fontWeight = "bold";
    }

    fileItem.appendChild(checkbox);
    fileItem.appendChild(label);
    contextFilesSelector.appendChild(fileItem);
  });
}

/**
 * Update selected code to session context
 * @param {string} selectedCode - Selected code
 * @param {string} currentFilePath - Current file path
 * @param {string} sessionId - Session ID
 */
export async function updateSelectedCodeToContext(selectedCode, currentFilePath, sessionId) {
  try {
    await sessions.updateContext(sessionId, {
      selected_code: selectedCode,
      current_file: currentFilePath
    });
    return true;
  } catch (error) {
    console.error('Failed to update selected code:', error);
    return false;
  }
}
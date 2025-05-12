/**

* Session editing component
* @module modules/sessions/sessionEditing
  */
import { sessions } from "../../api/client.js";
import { showNotification } from "../../utils/dom.js";
import { showFileSelector } from "./fileSelector.js";
import { createSessionModeOptions } from "./sessionCreation.js";

/**
 * Create edit session modal
 * @param {Function} getCurrentSession - Function to get current session
 */
export function createEditSessionModal(getCurrentSession) {
  const modal = document.createElement("div");
  modal.id = "edit-session-modal";
  modal.className = "modal";

  const modalContent = document.createElement("div");
  modalContent.className = "modal-content";

  const title = document.createElement("h3");
  title.textContent = "编辑会话";

  const form = document.createElement("form");
  form.id = "edit-session-form";

  // Hidden session ID field
  const sessionIdInput = document.createElement("input");
  sessionIdInput.type = "hidden";
  sessionIdInput.id = "edit-session-id";
  form.appendChild(sessionIdInput);

  // Session name field
  const nameFormGroup = document.createElement("div");
  nameFormGroup.className = "form-group";

  const nameLabel = document.createElement("label");
  nameLabel.htmlFor = "edit-session-name";
  nameLabel.textContent = "会话名称";

  const nameInput = document.createElement("input");
  nameInput.type = "text";
  nameInput.id = "edit-session-name";
  nameInput.required = true;

  nameFormGroup.appendChild(nameLabel);
  nameFormGroup.appendChild(nameInput);

  // Add session mode options
  const sessionModeGroup = createSessionModeOptions(
    "edit-session-mode",
    "edit-mode"
  );

  // File patterns container
  const filePatternContainer = createFilePatternContainer("edit");

  // Form actions
  const formActions = document.createElement("div");
  formActions.className = "form-actions";

  const cancelBtn = document.createElement("button");
  cancelBtn.type = "button";
  cancelBtn.className = "cancel-btn";
  cancelBtn.textContent = "取消";
  cancelBtn.addEventListener("click", hideEditSessionModal);

  const submitBtn = document.createElement("button");
  submitBtn.type = "submit";
  submitBtn.textContent = "保存";

  formActions.appendChild(cancelBtn);
  formActions.appendChild(submitBtn);

  // Assemble the form
  form.appendChild(nameFormGroup);
  form.appendChild(sessionModeGroup);
  form.appendChild(filePatternContainer);
  form.appendChild(formActions);

  // Attach form submit event
  form.addEventListener("submit", (event) =>
    saveSessionEdit(event, getCurrentSession)
  );

  // Assemble modal content
  modalContent.appendChild(title);
  modalContent.appendChild(form);
  modal.appendChild(modalContent);

  document.body.appendChild(modal);
}

/**
 * Create file pattern container
 * @param {string} prefix - Prefix for element IDs
 * @returns {HTMLElement} The file pattern container
 */
export function createFilePatternContainer(prefix) {
  const filePatternContainer = document.createElement("div");
  filePatternContainer.id = `${prefix}-file-pattern-container`;
  filePatternContainer.className = "file-pattern-container";

  const filePatternTitle = document.createElement("div");
  filePatternTitle.className = "patterns-title";
  filePatternTitle.textContent = "文件过滤设置";
  filePatternContainer.appendChild(filePatternTitle);

  // Include patterns section
  const includePatternsGroup = document.createElement("div");
  includePatternsGroup.className = "form-group patterns-group";

  const includeLabel = document.createElement("label");
  includeLabel.textContent = "包含的文件或文件夹";
  includePatternsGroup.appendChild(includeLabel);

  const includePatterns = document.createElement("div");
  includePatterns.className = "patterns-list";
  includePatterns.id = `${prefix}-include-patterns`;

  const addIncludeBtnContainer = document.createElement("div");
  addIncludeBtnContainer.className = "pattern-buttons";

  const addIncludeBtn = document.createElement("button");
  addIncludeBtn.type = "button";
  addIncludeBtn.className = "add-pattern-btn";
  addIncludeBtn.textContent = "+ 手动添加";
  addIncludeBtn.addEventListener("click", function () {
    addPatternInput(`${prefix}-include-patterns`, true);
  });

  const selectIncludeBtn = document.createElement("button");
  selectIncludeBtn.type = "button";
  selectIncludeBtn.className = "select-pattern-btn";
  selectIncludeBtn.textContent = "从文件树选择";
  selectIncludeBtn.addEventListener("click", function () {
    showFileSelector(`${prefix}-include`);
  });

  addIncludeBtnContainer.appendChild(addIncludeBtn);
  addIncludeBtnContainer.appendChild(selectIncludeBtn);

  includePatternsGroup.appendChild(includePatterns);
  includePatternsGroup.appendChild(addIncludeBtnContainer);
  filePatternContainer.appendChild(includePatternsGroup);

  // Exclude patterns section
  const excludePatternsGroup = document.createElement("div");
  excludePatternsGroup.className = "form-group patterns-group";

  const excludeLabel = document.createElement("label");
  excludeLabel.textContent = "排除的文件或文件夹";
  excludePatternsGroup.appendChild(excludeLabel);

  const excludePatterns = document.createElement("div");
  excludePatterns.className = "patterns-list";
  excludePatterns.id = `${prefix}-exclude-patterns`;

  const addExcludeBtnContainer = document.createElement("div");
  addExcludeBtnContainer.className = "pattern-buttons";

  const addExcludeBtn = document.createElement("button");
  addExcludeBtn.type = "button";
  addExcludeBtn.className = "add-pattern-btn";
  addExcludeBtn.textContent = "+ 手动添加";
  addExcludeBtn.addEventListener("click", function () {
    addPatternInput(`${prefix}-exclude-patterns`, false);
  });

  const selectExcludeBtn = document.createElement("button");
  selectExcludeBtn.type = "button";
  selectExcludeBtn.className = "select-pattern-btn";
  selectExcludeBtn.textContent = "从文件树选择";
  selectExcludeBtn.addEventListener("click", function () {
    showFileSelector(`${prefix}-exclude`);
  });

  addExcludeBtnContainer.appendChild(addExcludeBtn);
  addExcludeBtnContainer.appendChild(selectExcludeBtn);

  excludePatternsGroup.appendChild(excludePatterns);
  excludePatternsGroup.appendChild(addExcludeBtnContainer);
  filePatternContainer.appendChild(excludePatternsGroup);

  return filePatternContainer;
}

/**
 * Show edit session modal
 * @param {string} sessionId - Session ID
 */
export async function showEditSessionModal(sessionId) {
  try {
    const session = await sessions.getById(sessionId);
    if (!session) throw new Error("会话不存在");

    const modal = document.getElementById("edit-session-modal");
    if (!modal) return;

    // Populate form fields
    document.getElementById("edit-session-id").value = session.id;
    document.getElementById("edit-session-name").value = session.name;

    // Set session mode
    const modeRadio = document.querySelector(
      `input[name="edit-session-mode"][value="${session.mode || "manual"}"]`
    );
    if (modeRadio) modeRadio.checked = true;

    // Reset pattern containers
    document.getElementById("edit-include-patterns").innerHTML = "";
    document.getElementById("edit-exclude-patterns").innerHTML = "";

    // Add include patterns
    if (session.include_patterns && session.include_patterns.length > 0) {
      session.include_patterns.forEach((pattern) => {
        addPatternInputWithValue(
          "edit-include-patterns",
          pattern.path,
          pattern.is_dir
        );
      });
    }

    // Add exclude patterns
    if (session.exclude_patterns && session.exclude_patterns.length > 0) {
      session.exclude_patterns.forEach((pattern) => {
        addPatternInputWithValue(
          "edit-exclude-patterns",
          pattern.path,
          pattern.is_dir
        );
      });
    }

    modal.classList.add("active");
    if (typeof modal.showModal === "function") {
      modal.showModal();
    }
  } catch (error) {
    console.error("加载会话数据失败:", error);
    showNotification("加载会话数据失败", "error");
  }
}

/**

* Add a pattern input with value

* @param {string} containerId - Container ID

* @param {string} value - Pattern value

* @param {boolean} isDir - Whether it's a directory
*/
export function addPatternInputWithValue(containerId, value, isDir) {
  const container = document.getElementById(containerId);
  if (!container) return;

  const patternContainer = document.createElement("div");
  patternContainer.className = "pattern-input-container";

  const patternInput = document.createElement("input");
  patternInput.type = "text";
  patternInput.className = "pattern-input";
  patternInput.value = value;

  const isDirCheckbox = document.createElement("input");
  isDirCheckbox.type = "checkbox";
  isDirCheckbox.className = "is-dir-checkbox";
  isDirCheckbox.id = `is-dir-${containerId}-${container.children.length}`;
  isDirCheckbox.checked = isDir;

  const isDirLabel = document.createElement("label");
  isDirLabel.htmlFor = isDirCheckbox.id;
  isDirLabel.textContent = "文件夹";
  isDirLabel.className = "is-dir-label";

  const removeBtn = document.createElement("button");
  removeBtn.type = "button";
  removeBtn.className = "remove-pattern-btn";
  removeBtn.textContent = "×";
  removeBtn.addEventListener("click", function () {
    patternContainer.remove();
  });

  patternContainer.appendChild(patternInput);
  patternContainer.appendChild(isDirCheckbox);
  patternContainer.appendChild(isDirLabel);
  patternContainer.appendChild(removeBtn);

  container.appendChild(patternContainer);
}

/**
 * Add a pattern input
 * @param {string} containerId - Container ID
 * @param {boolean} isInclude - Whether it's an include pattern
 */
export function addPatternInput(containerId, isInclude) {
  const container = document.getElementById(containerId);
  if (!container) return;

  const patternContainer = document.createElement("div");
  patternContainer.className = "pattern-input-container";

  const patternInput = document.createElement("input");
  patternInput.type = "text";
  patternInput.className = "pattern-input";
  patternInput.placeholder = isInclude
    ? "输入要包含的文件或文件夹路径"
    : "输入要排除的文件或文件夹路径";

  const isDirCheckbox = document.createElement("input");
  isDirCheckbox.type = "checkbox";
  isDirCheckbox.className = "is-dir-checkbox";
  isDirCheckbox.id = `is-dir-${containerId}-${container.children.length}`;

  const isDirLabel = document.createElement("label");
  isDirLabel.htmlFor = isDirCheckbox.id;
  isDirLabel.textContent = "文件夹";
  isDirLabel.className = "is-dir-label";

  const removeBtn = document.createElement("button");
  removeBtn.type = "button";
  removeBtn.className = "remove-pattern-btn";
  removeBtn.textContent = "×";
  removeBtn.addEventListener("click", function () {
    patternContainer.remove();
  });

  patternContainer.appendChild(patternInput);
  patternContainer.appendChild(isDirCheckbox);
  patternContainer.appendChild(isDirLabel);
  patternContainer.appendChild(removeBtn);

  container.appendChild(patternContainer);
}

/**
 * Hide edit session modal
 */
export function hideEditSessionModal() {
  const modal = document.getElementById("edit-session-modal");
  if (modal) {
    modal.classList.remove("active");
    if (typeof modal.close === "function") {
      modal.close();
    }
  }
}

/**
 * Save session edit
 * @param {Event} event - Form submit event
 * @param {Function} getCurrentSession - Function to get current session
 */
export async function saveSessionEdit(event, getCurrentSession) {
  event.preventDefault();

  const id = document.getElementById("edit-session-id").value;
  const name = document.getElementById("edit-session-name").value;
  const modeElement = document.querySelector(
    'input[name="edit-session-mode"]:checked'
  );
  const mode = modeElement ? modeElement.value : "manual";

  // Collect include patterns
  const includePatterns = [];
  document
    .querySelectorAll("#edit-include-patterns .pattern-input-container")
    .forEach((container) => {
      const pathInput = container.querySelector(".pattern-input");
      const isDirCheckbox = container.querySelector(".is-dir-checkbox");

      if (pathInput && pathInput.value.trim()) {
        includePatterns.push({
          path: pathInput.value.trim(),
          is_dir: isDirCheckbox ? isDirCheckbox.checked : false,
        });
      }
    });

  // Collect exclude patterns
  const excludePatterns = [];
  document
    .querySelectorAll("#edit-exclude-patterns .pattern-input-container")
    .forEach((container) => {
      const pathInput = container.querySelector(".pattern-input");
      const isDirCheckbox = container.querySelector(".is-dir-checkbox");

      if (pathInput && pathInput.value.trim()) {
        excludePatterns.push({
          path: pathInput.value.trim(),
          is_dir: isDirCheckbox ? isDirCheckbox.checked : false,
        });
      }
    });

  try {
    const updates = {
      name,
      mode,
      include_patterns: includePatterns,
      exclude_patterns: excludePatterns,
    };

    await sessions.update(id, updates);
    hideEditSessionModal();

    // If current session was edited, update it
    const currentSession = getCurrentSession();
    if (currentSession && currentSession.id === id) {
      const updatedSession = await sessions.getById(id);
      window.SessionsModule.setCurrentSession(updatedSession);

      // Update session display in dropdown
      const dropdownHeader = document.querySelector(
        ".sessions-dropdown-header span:first-child"
      );
      if (dropdownHeader) {
        dropdownHeader.textContent = updatedSession.name;
      }

      // Update context files selector based on new mode
      if (window.SessionsModule.updateContextFilesSelector) {
        window.SessionsModule.updateContextFilesSelector();
      }
    }

    // Refresh sessions list
    const currentProject = window.ProjectsModule.getCurrentProject();
    if (currentProject && window.SessionsModule.loadProjectSessions) {
      await window.SessionsModule.loadProjectSessions(currentProject.id);
    }

    showNotification("会话更新成功", "success");
  } catch (error) {
    console.error("更新会话失败:", error);
    showNotification("更新会话失败", "error");
  }
}

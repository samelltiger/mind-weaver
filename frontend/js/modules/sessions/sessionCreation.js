/**

* Session creation component
* @module modules/sessions/sessionCreation
  */
import { sessions } from "../../api/client.js";
import { showNotification } from "../../utils/dom.js";
import { createFileSelector, showFileSelector } from "./fileSelector.js";

/**
 * Show new session modal
 */
export function showNewSessionModal() {
  const newSessionModal = document.getElementById("new-session-modal");
  if (!newSessionModal) return;

  // 获取表单和必要元素
  const newSessionForm = document.getElementById("new-session-form");

  // 保留名称输入和提交按钮区域
  const nameInput = newSessionForm.querySelector("#session-name");
  const formActions = newSessionForm.querySelector(".form-actions");

  // 清空表单
  newSessionForm.innerHTML = "";

  // 重新添加名称输入
  if (nameInput) {
    const nameFormGroup = document.createElement("div");
    nameFormGroup.className = "form-group";

    const nameLabel = document.createElement("label");
    nameLabel.htmlFor = "session-name";
    nameLabel.textContent = "会话名称";

    const newNameInput = document.createElement("input");
    newNameInput.type = "text";
    newNameInput.id = "session-name";
    newNameInput.required = true;

    nameFormGroup.appendChild(nameLabel);
    nameFormGroup.appendChild(newNameInput);
    newSessionForm.appendChild(nameFormGroup);
  }

  // 添加会话模式选项
  const sessionModeGroup = createSessionModeOptions("session-mode", "mode");
  newSessionForm.appendChild(sessionModeGroup);

  // 添加文件模式容器
  const filePatternContainer = createFilePatternContainer("");
  newSessionForm.appendChild(filePatternContainer);

  // 最后添加操作按钮
  if (formActions) {
    newSessionForm.appendChild(formActions);
  }

  // 确保文件选择器被创建
  createFileSelector();

  newSessionModal.classList.add("active");
  if (typeof newSessionModal.showModal === "function") {
    newSessionModal.showModal();
  }
}

/**
 * Add session mode options to modal
 * @param {HTMLElement} modal - Modal element
 */
function addSessionModeOptions(modal) {
  const newSessionForm = document.getElementById("new-session-form");
  if (!newSessionForm) return;

  // Create session mode options
  const sessionModeGroup = createSessionModeOptions("session-mode", "mode");

  // Find form actions
  const formActions = newSessionForm.querySelector(".form-actions");

  // Insert before form actions
  newSessionForm.insertBefore(sessionModeGroup, formActions);

  // Add file patterns container
  const filePatternContainer = createFilePatternContainer("");

  newSessionForm.insertBefore(filePatternContainer, formActions);

  // Create file selector modal
  createFileSelector();
}

/**
 * Create session mode options
 * @param {string} name - Name attribute for radio inputs
 * @param {string} idPrefix - Prefix for element IDs
 * @returns {HTMLElement} The session mode options container
 */
export function createSessionModeOptions(name, idPrefix) {
  const formGroup = document.createElement("div");
  formGroup.className = "form-group";
  formGroup.id = `${idPrefix}-container`;

  const label = document.createElement("label");
  label.textContent = "会话模式";
  formGroup.appendChild(label);

  const modesContainer = document.createElement("div");
  modesContainer.className = "session-modes";

  const modes = [
    {
      id: `${idPrefix}-manual`,
      value: "manual",
      label: "手动模式",
    },
    {
      id: `${idPrefix}-auto`,
      value: "auto",
      label: "智能模式 ",
    },
    {
      id: `${idPrefix}-single-html`,
      value: "single-html",
      label: "单HTML模式",
    },
    // {
    //   id: `${idPrefix}-product-design`,
    //   value: "product-design",
    //   label: "产品需求设计",
    // },
    // { id: `${idPrefix}-all`, value: "all", label: "全部文件" },
  ];

  modes.forEach((mode) => {
    const modeContainer = document.createElement("div");
    modeContainer.className = "mode-option";

    const radio = document.createElement("input");
    radio.type = "radio";
    radio.name = name;
    radio.id = mode.id;
    radio.value = mode.value;
    radio.checked = mode.value === "manual"; // Default to manual

    const modeLabel = document.createElement("label");
    modeLabel.htmlFor = mode.id;
    modeLabel.textContent = mode.label;

    modeContainer.appendChild(radio);
    modeContainer.appendChild(modeLabel);
    modesContainer.appendChild(modeContainer);
  });

  formGroup.appendChild(modesContainer);
  return formGroup;
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
  includePatterns.id = `${prefix}include-patterns`;

  const addIncludeBtnContainer = document.createElement("div");
  addIncludeBtnContainer.className = "pattern-buttons";

  const addIncludeBtn = document.createElement("button");
  addIncludeBtn.type = "button";
  addIncludeBtn.className = "add-pattern-btn";
  addIncludeBtn.textContent = "+ 手动添加";
  addIncludeBtn.addEventListener("click", function () {
    addPatternInput(`${prefix}include-patterns`, true);
  });

  const selectIncludeBtn = document.createElement("button");
  selectIncludeBtn.type = "button";
  selectIncludeBtn.className = "select-pattern-btn";
  selectIncludeBtn.textContent = "从文件树选择";
  selectIncludeBtn.addEventListener("click", function () {
    showFileSelector(`${prefix}include`);
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
  excludePatterns.id = `${prefix}exclude-patterns`;

  const addExcludeBtnContainer = document.createElement("div");
  addExcludeBtnContainer.className = "pattern-buttons";

  const addExcludeBtn = document.createElement("button");
  addExcludeBtn.type = "button";
  addExcludeBtn.className = "add-pattern-btn";
  addExcludeBtn.textContent = "+ 手动添加";
  addExcludeBtn.addEventListener("click", function () {
    addPatternInput(`${prefix}exclude-patterns`, false);
  });

  const selectExcludeBtn = document.createElement("button");
  selectExcludeBtn.type = "button";
  selectExcludeBtn.className = "select-pattern-btn";
  selectExcludeBtn.textContent = "从文件树选择";
  selectExcludeBtn.addEventListener("click", function () {
    showFileSelector(`${prefix}exclude`);
  });

  addExcludeBtnContainer.appendChild(addExcludeBtn);
  addExcludeBtnContainer.appendChild(selectExcludeBtn);

  excludePatternsGroup.appendChild(excludePatterns);
  excludePatternsGroup.appendChild(addExcludeBtnContainer);
  filePatternContainer.appendChild(excludePatternsGroup);

  return filePatternContainer;
}

/**
 * Hide new session modal
 */
export function hideNewSessionModal() {
  const newSessionModal = document.getElementById("new-session-modal");
  const newSessionForm = document.getElementById("new-session-form");

  // 同时关闭文件选择器模态框
  const fileSelectorModal = document.getElementById("file-selector-modal");
  if (fileSelectorModal) {
    if (typeof fileSelectorModal.close === "function") {
      fileSelectorModal.close();
    }
    fileSelectorModal.remove(); // 完全从DOM中移除
  }

  if (newSessionModal) {
    newSessionModal.classList.remove("active");
    if (typeof newSessionModal.close === "function") {
      newSessionModal.close();
    }
  }

  if (newSessionForm) {
    newSessionForm.reset();
  }
}

/**
 * Create a new session
 * @param {Event} event - Form submit event
 * @param {Function} getCurrentSession - Function to get current session
 */
export async function createSession(event, getCurrentSession) {
  event.preventDefault();

  const currentProject = window.ProjectsModule.getCurrentProject();
  if (!currentProject) {
    showNotification("未选择项目", "error");
    return;
  }

  const name = document.getElementById("session-name").value;
  const modeElement = document.querySelector(
    'input[name="session-mode"]:checked'
  );
  const mode = modeElement ? modeElement.value : "manual";

  try {
    // Collect include patterns for all modes
    const includePatterns = [];
    document
      .querySelectorAll("#include-patterns .pattern-input-container")
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

    // Collect exclude patterns for all modes
    const excludePatterns = [];
    document
      .querySelectorAll("#exclude-patterns .pattern-input-container")
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

    const session = await sessions.create(
      currentProject.id,
      name,
      mode,
      excludePatterns,
      includePatterns
    );

    // 确保关闭所有模态框
    hideNewSessionModal();

    // 显式检查并关闭文件选择器模态框
    const fileSelectorModal = document.getElementById("file-selector-modal");
    if (fileSelectorModal) {
      if (typeof fileSelectorModal.close === "function") {
        fileSelectorModal.close();
      }
      fileSelectorModal.remove();
    }

    // Reload project sessions and select the new one
    if (window.SessionsModule && window.SessionsModule.loadProjectSessions) {
      await window.SessionsModule.loadProjectSessions(currentProject.id);

      if (window.SessionsModule.selectSession) {
        window.SessionsModule.selectSession(
          session.id,
          window.SessionsModule.setCurrentSession
        );
      }
    }

    showNotification("会话创建成功");
  } catch (error) {
    console.error("创建会话失败:", error);
    showNotification("创建会话失败", "error");
  }
}

/**
 * Add a pattern input field
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
 * Add a pattern input field from a file
 * @param {HTMLElement} container - Container element
 * @param {Object} file - File object
 * @param {boolean} isInclude - Whether it's an include pattern
 */
export function addPatternInputFromFile(container, file, isInclude) {
  if (!container) return;

  const patternContainer = document.createElement("div");
  patternContainer.className = "pattern-input-container";

  const patternInput = document.createElement("input");
  patternInput.type = "text";
  patternInput.className = "pattern-input";
  patternInput.value = file.path;
  patternInput.readOnly = true;

  const isDirCheckbox = document.createElement("input");
  isDirCheckbox.type = "checkbox";
  isDirCheckbox.className = "is-dir-checkbox";
  isDirCheckbox.id = `is-dir-${container.id}-${container.children.length}`;
  isDirCheckbox.checked = file.is_dir;

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
 * Reset pattern inputs
 */
export function resetPatternInputs() {
  const includePatterns = document.getElementById("include-patterns");
  const excludePatterns = document.getElementById("exclude-patterns");

  if (includePatterns) includePatterns.innerHTML = "";
  if (excludePatterns) excludePatterns.innerHTML = "";
}

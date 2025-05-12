/**
 * File selector component
 * @module modules/sessions/fileSelector
 */
import { projects } from "../../api/client.js";
import { showNotification } from "../../utils/dom.js";
import { addPatternInputFromFile } from "./sessionCreation.js";

let currentSelectionType = null; // 'include', 'exclude', 'edit-include', or 'edit-exclude'

/**
 * Create file selector modal
 */
export function createFileSelector() {
  // 先移除已存在的选择器
  const existingSelector = document.getElementById("file-selector-modal");
  if (existingSelector) {
    existingSelector.remove();
  }

  const fileSelectorModal = document.createElement("dialog");
  fileSelectorModal.id = "file-selector-modal";
  fileSelectorModal.className = "file-selector-modal";

  const fileSelectorContent = document.createElement("div");
  fileSelectorContent.className = "file-selector-content";

  const fileSelectorHeader = document.createElement("div");
  fileSelectorHeader.className = "file-selector-header";

  const fileSelectorTitle = document.createElement("h4");
  fileSelectorTitle.id = "file-selector-title";
  fileSelectorTitle.textContent = "选择文件或文件夹";

  const closeBtn = document.createElement("button");
  closeBtn.type = "button";
  closeBtn.className = "file-selector-close";
  closeBtn.textContent = "×";
  closeBtn.addEventListener("click", hideFileSelector);

  fileSelectorHeader.appendChild(fileSelectorTitle);
  fileSelectorHeader.appendChild(closeBtn);
  fileSelectorContent.appendChild(fileSelectorHeader);

  const fileSelectorTree = document.createElement("div");
  fileSelectorTree.id = "file-selector-tree";
  fileSelectorTree.className = "file-selector-tree";
  fileSelectorContent.appendChild(fileSelectorTree);

  const fileSelectorActions = document.createElement("div");
  fileSelectorActions.className = "file-selector-actions";

  const cancelBtn = document.createElement("button");
  cancelBtn.type = "button";
  cancelBtn.className = "cancel-btn";
  cancelBtn.textContent = "取消";
  cancelBtn.addEventListener("click", hideFileSelector);

  const confirmBtn = document.createElement("button");
  confirmBtn.type = "button";
  confirmBtn.className = "confirm-btn";
  confirmBtn.textContent = "确认选择";
  confirmBtn.addEventListener("click", confirmFileSelection);

  fileSelectorActions.appendChild(cancelBtn);
  fileSelectorActions.appendChild(confirmBtn);
  fileSelectorContent.appendChild(fileSelectorActions);

  fileSelectorModal.appendChild(fileSelectorContent);

  // 确保模态框添加到正确的位置并设置正确的z-index
  document.body.appendChild(fileSelectorModal);
}


/**
 * Show file selector and populate with project files
 * @param {string} type - Selection type ('include', 'exclude', 'edit-include', or 'edit-exclude')
 */
export function showFileSelector(type) {
  // 创建或重新创建文件选择器
  createFileSelector();
  currentSelectionType = type;

  const modal = document.getElementById("file-selector-modal");
  const title = document.getElementById("file-selector-title");
  const tree = document.getElementById("file-selector-tree");

  if (!modal || !title || !tree) return;

  title.textContent =
    type === "include" || type === "edit-include"
      ? "选择要包含的文件或文件夹"
      : "选择要排除的文件或文件夹";
  tree.innerHTML = '<div class="loading-indicator">正在加载文件列表...</div>';

  // 使用showModal将其放在顶层
  if (typeof modal.showModal === "function") {
    modal.showModal();
  } else {
    modal.classList.add("active");
  }

  const currentProject = window.ProjectsModule.getCurrentProject();
  if (!currentProject) {
    tree.innerHTML = '<div class="empty-message">未选择项目</div>';
    return;
  }

  // 获取项目文件并渲染树
  projects
    .getFiles(currentProject.id)
    .then((files) => {
      tree.innerHTML = "";
      renderFileSelectorTree(files, tree);
    })
    .catch((error) => {
      console.error("加载项目文件失败:", error);
      tree.innerHTML = '<div class="empty-message">加载文件失败</div>';
    });
}

// 修改fileSelector.js中的hideFileSelector函数
export function hideFileSelector() {
  const modal = document.getElementById("file-selector-modal");
  if (modal) {
    try {
      if (typeof modal.close === "function") {
        modal.close();
      } else {
        modal.classList.remove("active");
      }
    } catch (e) {
      console.error("关闭文件选择器时出错:", e);
    } finally {
      // 无论如何都确保从DOM中移除
      modal.remove();
    }
  }
  currentSelectionType = null;
}

/**
 * Render file selector tree
 * @param {Object} fileData - File data
 * @param {HTMLElement} parentElement - Parent element
 */
export function renderFileSelectorTree(fileData, parentElement) {
  if (fileData.is_dir) {
    // Create directory element
    const dirContainer = document.createElement("div");
    dirContainer.className = "file-item-container";

    const dirElement = document.createElement("div");
    dirElement.className = "selector-file-item directory";

    const checkbox = document.createElement("input");
    checkbox.type = "checkbox";
    checkbox.className = "selector-file-checkbox";
    checkbox.dataset.path = fileData.path;
    checkbox.dataset.isDir = "true";
    checkbox.dataset.name = fileData.name;
    checkbox.id = `selector-file-${fileData.path}`;

    // Add event listener to handle checking/unchecking child files
    checkbox.addEventListener("change", function () {
      const childrenContainer = dirContainer.querySelector(
        ".selector-directory-children"
      );
      if (childrenContainer) {
        const childCheckboxes = childrenContainer.querySelectorAll(
          'input[type="checkbox"]'
        );
        childCheckboxes.forEach((childCheckbox) => {
          childCheckbox.checked = this.checked;
        });
      }
    });

    const dirLabel = document.createElement("label");
    dirLabel.htmlFor = checkbox.id;

    // Create a container for name and lines
    const nameContainer = document.createElement("span");
    nameContainer.className = "file-name";
    nameContainer.textContent = fileData.name;
    dirLabel.appendChild(nameContainer);

    // Add lines count with different styling
    const linesSpan = document.createElement("span");
    linesSpan.className = "file-lines";
    linesSpan.textContent = ` (${fileData.lines} lines)`;
    dirLabel.appendChild(linesSpan);

    // Create expand/collapse toggle
    const expandToggle = document.createElement("span");
    expandToggle.className = "expand-toggle";
    expandToggle.textContent = "▶";
    dirElement.appendChild(expandToggle);

    dirElement.appendChild(checkbox);
    dirElement.appendChild(dirLabel);

    dirContainer.appendChild(dirElement);

    // Create children container
    const childContainer = document.createElement("div");
    childContainer.className = "selector-directory-children";
    childContainer.style.display = "none"; // Initially collapsed
    dirContainer.appendChild(childContainer);

    parentElement.appendChild(dirContainer);

    // If there are children, render them
    if (fileData.children && fileData.children.length > 0) {
      fileData.children.forEach((child) => {
        renderFileSelectorTree(child, childContainer);
      });
    }

    // Add click event for expand/collapse
    let expanded = false;
    expandToggle.addEventListener("click", () => {
      expanded = !expanded;
      expandToggle.textContent = expanded ? "▼" : "▶";
      childContainer.style.display = expanded ? "block" : "none";
    });
  } else {
    // Create file element
    const fileElement = document.createElement("div");
    fileElement.className = "selector-file-item file";

    const checkbox = document.createElement("input");
    checkbox.type = "checkbox";
    checkbox.className = "selector-file-checkbox";
    checkbox.dataset.path = fileData.path;
    checkbox.dataset.isDir = "false";
    checkbox.dataset.name = fileData.name;
    checkbox.id = `selector-file-${fileData.path}`;

    const fileLabel = document.createElement("label");
    fileLabel.htmlFor = checkbox.id;

    // Create a container for name and lines
    const nameContainer = document.createElement("span");
    nameContainer.className = "file-name";
    nameContainer.textContent = fileData.name;
    fileLabel.appendChild(nameContainer);

    // Add lines count with different styling
    const linesSpan = document.createElement("span");
    linesSpan.className = "file-lines";
    linesSpan.textContent = ` (${fileData.lines} lines)`;
    fileLabel.appendChild(linesSpan);

    fileElement.appendChild(checkbox);
    fileElement.appendChild(fileLabel);

    parentElement.appendChild(fileElement);
  }
}

/**
 * Confirm file selection
 */
export function confirmFileSelection() {
  if (!currentSelectionType) return;

  const selectedFiles = [];
  const checkboxes = document.querySelectorAll(
    "#file-selector-tree .selector-file-checkbox:checked"
  );

  checkboxes.forEach((checkbox) => {
    selectedFiles.push({
      path: checkbox.dataset.path,
      is_dir: checkbox.dataset.isDir === "true",
      name: checkbox.dataset.name,
    });
  });

  // Determine target container based on selection type
  let targetContainer;
  let isInclude = true;

  if (currentSelectionType === "include") {
    targetContainer = document.getElementById("include-patterns");
  } else if (currentSelectionType === "exclude") {
    targetContainer = document.getElementById("exclude-patterns");
    isInclude = false;
  } else if (currentSelectionType === "edit-include") {
    targetContainer = document.getElementById("edit-include-patterns");
  } else if (currentSelectionType === "edit-exclude") {
    targetContainer = document.getElementById("edit-exclude-patterns");
    isInclude = false;
  }

  if (targetContainer) {
    selectedFiles.forEach((file) => {
      addPatternInputFromFile(targetContainer, file, isInclude);
    });
  }

  hideFileSelector();
}

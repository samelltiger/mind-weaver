/**

* Projects module
* @module modules/projects
  */
import { projects } from "../api/client.js";
import { showNotification, createElement } from "../utils/dom.js";

let currentProject = null;

/**

* Initialize projects module
*/
export function init() {
  const projectsList = document.getElementById("projects-list");
  const newProjectBtn = document.getElementById("new-project-btn");
  const newProjectModal = document.getElementById("new-project-modal");
  const newProjectForm = document.getElementById("new-project-form");
  const refreshProjectsBtn = document.getElementById("refresh-projects-btn");

  // Attach event listeners
  newProjectBtn.addEventListener("click", showNewProjectModal);
  newProjectForm.addEventListener("submit", createProject);
  newProjectModal
    .querySelector(".cancel-btn")
    .addEventListener("click", hideNewProjectModal);
  refreshProjectsBtn.addEventListener("click", loadProjects);

  // Load projects
  loadProjects();

  // Create edit project modal if it doesn't exist
  if (!document.getElementById("edit-project-modal")) {
    createEditProjectModal();
  }
}

/**

* Load all projects from the server
*/
export async function loadProjects() {
  try {
    showNotification("正在加载项目列表...", "info");
    const projectsData = await projects.getAll();
    renderProjects(projectsData);
    showNotification("项目列表已刷新", "success");
  } catch (error) {
    console.error("Failed to load projects:", error);
    showNotification("Failed to load projects", "error");
  }
}

/**
 * Render projects in the UI
 * @param {Array} projects - Projects to render
 */
function renderProjects(projectsData) {
  const projectsList = document.getElementById("projects-list");
  projectsList.innerHTML = "";

  if (projectsData.length === 0) {
    projectsList.innerHTML = '<div class="empty-message">没有找到项目</div>';
    return;
  }

  // 创建项目下拉容器，添加projects-dropdown类名以区分
  const projectsDropdown = createElement("div", { 
    className: "sessions-dropdown projects-dropdown" 
  }, []);

  // 创建当前选择的项目显示区域
  const dropdownHeader = createElement("div", { 
    className: "sessions-dropdown-header projects-dropdown-header" 
  }, [
    createElement("span", {}, currentProject ? currentProject.name : "选择项目"),
    createElement("span", {}, "▼")
  ]);

  projectsDropdown.appendChild(dropdownHeader);

  // 创建下拉内容区域，添加projects-dropdown-content类名以区分
  const dropdownContent = createElement("div", { 
    className: "sessions-dropdown-content projects-dropdown-content" 
  }, []);

  // 切换下拉显示/隐藏
  dropdownHeader.addEventListener("click", function() {
    dropdownContent.classList.toggle("active");
  });

  // 点击其他地方关闭下拉框 - 修改为只关闭项目下拉框
  document.addEventListener("click", function(event) {
    // 只有当点击的不是项目下拉框内的元素时，才关闭项目下拉框
    if (!projectsDropdown.contains(event.target)) {
      dropdownContent.classList.remove("active");
    }
  });

  projectsData.forEach((project) => {
    const projectElement = createElement(
      "div",
      {
        className: "project-item" + (currentProject && currentProject.id === project.id ? " active" : ""),
        dataset: { id: project.id },
        onClick: () => {
          selectProject(project);
          // 更新下拉框标题
          dropdownHeader.querySelector("span:first-child").textContent = project.name;
          // 隐藏下拉内容
          dropdownContent.classList.remove("active");
        }
      },
      [
        createElement("div", { className: "project-item-name" }, project.name),
        createElement("div", { className: "project-item-actions" }, [
          createElement("button", {
            className: "icon-btn",
            title: "编辑项目",
            innerHTML: "✎",
            onClick: (e) => {
              e.stopPropagation();
              showEditProjectModal(project);
            },
          }),
        ]),
      ]
    );

    dropdownContent.appendChild(projectElement);
  });

  projectsDropdown.appendChild(dropdownContent);
  projectsList.appendChild(projectsDropdown);
}

/**
 * Select a project
 * @param {Object} project - Project to select
 */
export async function selectProject(project) {
  currentProject = project;

  // 更新UI
  document.querySelectorAll(".project-item").forEach((el) => {
    el.classList.remove("active");
  });

  const projectElement = document.querySelector(
    `.project-item[data-id="${project.id}"]`
  );
  if (projectElement) {
    projectElement.classList.add("active");
  }

  // 更新下拉框标题 - 修改为只更新项目下拉框
  const dropdownHeader = document.querySelector(
    ".projects-dropdown-header span:first-child"
  );
  if (dropdownHeader) {
    dropdownHeader.textContent = project.name;
  }

  // 加载项目文件
  if (window.FilesModule && window.FilesModule.loadProjectFiles) {
    await window.FilesModule.loadProjectFiles(project.id);
  }

  // 加载项目会话
  if (window.SessionsModule && window.SessionsModule.loadProjectSessions) {
    await window.SessionsModule.loadProjectSessions(project.id);
  }

  // 更新状态栏
  document.getElementById("language-indicator").textContent = `语言: ${
    project.language || "未知"
  }`;

  // 发布事件
  window.dispatchEvent(
    new CustomEvent("project-selected", { detail: project })
  );
}

/**

* Show new project modal
*/
function showNewProjectModal() {
  const newProjectModal = document.getElementById("new-project-modal");
  newProjectModal.classList.add("active");
  newProjectModal.showModal();
}

/**

* Hide new project modal
*/
function hideNewProjectModal() {
  const newProjectModal = document.getElementById("new-project-modal");
  const newProjectForm = document.getElementById("new-project-form");
  newProjectModal.classList.remove("active");
  newProjectModal.close();
  newProjectForm.reset();
}

/**

* Create a new project

* @param {Event} event - Form submit event
*/
async function createProject(event) {
  event.preventDefault();

  const name = document.getElementById("project-name").value;
  const path = document.getElementById("project-path").value;
  const language = document.getElementById("project-language").value;

  try {
    const project = await projects.create({ name, path, language });
    hideNewProjectModal();
    await loadProjects();
    selectProject(project);
    showNotification("Project created successfully");
  } catch (error) {
    console.error("Failed to create project:", error);
    showNotification("Failed to create project", "error");
  }
}

/**

* Create edit project modal
*/
function createEditProjectModal() {
  // Continuing from previous code
  const modal = createElement(
    "dialog",
    {
      id: "edit-project-modal",
      className: "modal",
    },
    [
      createElement("div", { className: "modal-content" }, [
        createElement("h3", {}, "编辑项目"),
        createElement("form", { id: "edit-project-form" }, [
          createElement("input", { type: "hidden", id: "edit-project-id" }),
          createElement("div", { className: "form-group" }, [
            createElement(
              "label",
              { htmlFor: "edit-project-name" },
              "项目名称"
            ),
            createElement("input", {
              type: "text",
              id: "edit-project-name",
              required: true,
            }),
          ]),
          createElement("div", { className: "form-group" }, [
            createElement(
              "label",
              { htmlFor: "edit-project-path" },
              "项目路径"
            ),
            createElement("input", {
              type: "text",
              id: "edit-project-path",
              required: true,
            }),
          ]),
          createElement("div", { className: "form-group" }, [
            createElement(
              "label",
              { htmlFor: "edit-project-language" },
              "语言"
            ),
            createElement("select", { id: "edit-project-language" }, [
              createElement("option", { value: "go" }, "Go"),
              createElement("option", { value: "javascript" }, "JavaScript"),
              createElement("option", { value: "python" }, "Python"),
              createElement("option", { value: "java" }, "Java"),
              createElement("option", { value: "typescript" }, "TypeScript"),
            ]),
          ]),
          createElement("div", { className: "form-actions" }, [
            createElement(
              "button",
              { type: "button", className: "cancel-btn" },
              "取消"
            ),
            createElement("button", { type: "submit" }, "保存"),
          ]),
        ]),
      ]),
    ]
  );

  document.body.appendChild(modal);

  // Attach event listeners
  const form = document.getElementById("edit-project-form");
  form.addEventListener("submit", saveProjectEdit);
  modal
    .querySelector(".cancel-btn")
    .addEventListener("click", hideEditProjectModal);
}

/**
 * Show edit project modal
 * @param {Object} project - Project to edit
 */
function showEditProjectModal(project) {
  const modal = document.getElementById("edit-project-modal");
  if (!modal) return;

  document.getElementById("edit-project-id").value = project.id;
  document.getElementById("edit-project-name").value = project.name;
  document.getElementById("edit-project-path").value = project.path;
  document.getElementById("edit-project-language").value =
    project.language || "javascript";

  modal.classList.add("active");
  modal.showModal();
}

/**
 * Hide edit project modal
 */
function hideEditProjectModal() {
  const modal = document.getElementById("edit-project-modal");
  if (modal) {
    modal.classList.remove("active");
    modal.close();
  }
}

/**
 * Save project edit
 * @param {Event} event - Form submit event
 */
async function saveProjectEdit(event) {
  event.preventDefault();

  const id = document.getElementById("edit-project-id").value;
  const name = document.getElementById("edit-project-name").value;
  const path = document.getElementById("edit-project-path").value;
  const language = document.getElementById("edit-project-language").value;

  try {
    const updatedProject = await projects.update(id, {
      name,
      path,
      language,
    });
    hideEditProjectModal();
    await loadProjects();

    // If current project was edited, update the current project data
    if (currentProject && currentProject.id === id) {
      currentProject = updatedProject;
      document.getElementById("language-indicator").textContent = `Language: ${
        updatedProject.language || "Unknown"
      }`;
    }

    showNotification("项目更新成功", "success");
  } catch (error) {
    console.error("Failed to update project:", error);
    showNotification("更新项目失败", "error");
  }
}

/**
 * Get current project
 * @returns {Object|null} Current project
 */
export function getCurrentProject() {
  return currentProject;
}

// Export the module
export default {
  init,
  loadProjects,
  getCurrentProject,
  selectProject,
  showNotification,
};

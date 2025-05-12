/**
 * Main application entry point
 * @module app
 */
import ProjectsModule from './modules/projects.js';
import FilesModule from './modules/files.js';
import EditorModule from './modules/editor.js';
import SessionsModule from './modules/sessions/index.js';
import { showNotification } from './utils/dom.js';

// Make modules available globally for interoperability during refactoring
window.ProjectsModule = ProjectsModule;
window.FilesModule = FilesModule;
window.EditorModule = EditorModule;
window.SessionsModule = SessionsModule;

/**
 * Initialize application
 */
function initApp() {
  // Initialize modules
  ProjectsModule.init();
  FilesModule.init();
  SessionsModule.init();
  EditorModule.init();

  // Show welcome notification
  setTimeout(() => {
    showNotification('Welcome to MindWeaver AI');
  }, 1000);

  // Add global keyboard shortcuts
  document.addEventListener('keydown', function(e) {
    // Ctrl+Shift+S: Manually update selected code to context
    if (e.ctrlKey && e.shiftKey && e.key === 'S') {
      e.preventDefault();
      EditorModule.updateSelectedCodeToContext();
      showNotification('已更新选中代码到上下文', 'info');
    }
  });

  // 在 projects.js 的 init 函数中添加
// 全局点击事件，处理所有下拉框
document.addEventListener("click", function(event) {
  // 处理项目下拉框
  const projectsDropdown = document.querySelector(".projects-dropdown");
  if (projectsDropdown && !projectsDropdown.contains(event.target)) {
    document.querySelector(".projects-dropdown-content").classList.remove("active");
  }

  // 处理会话下拉框
  const sessionsDropdown = document.querySelector(".sessions-dropdown-container");
  if (sessionsDropdown && !sessionsDropdown.contains(event.target)) {
    document.querySelector(".sessions-content").classList.remove("active");
  }
});

// 然后在 renderProjects 和 renderSessions 中移除各自的全局点击事件监听器

  console.log('Application initialized');
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', initApp);
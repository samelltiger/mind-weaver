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

  // Setup UI enhancements
  setupRightPanelToggle();

  // Show welcome notification
  setTimeout(() => {
    showNotification('Welcome to MindWeaver AI');
  }, 1000);

  // Add global keyboard shortcuts
  document.addEventListener('keydown', function (e) {
    // Ctrl+Shift+S: Manually update selected code to context
    if (e.ctrlKey && e.shiftKey && e.key === 'S') {
      e.preventDefault();
      EditorModule.updateSelectedCodeToContext();
      showNotification('已更新选中代码到上下文', 'info');
    }
  });

  // 在 projects.js 的 init 函数中添加
  // 全局点击事件，处理所有下拉框
  document.addEventListener("click", function (event) {
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


/**
 * Initializes the toggle functionality for the right panel width.
 */
function setupRightPanelToggle() {
    const toggleBtn = document.getElementById('toggle-right-panel-width-btn');
    const rightPanel = document.getElementById('right-panel');

    if (!toggleBtn || !rightPanel) {
        console.warn('Right panel toggle button or right panel element not found.');
        return;
    }

    const originalWidth = '35%'; // 初始宽度，与 CSS 一致
    const expandedWidth = '65%';
    let isExpanded = false; // 初始状态为未展开 (宽度为 originalWidth)

    // 确保初始 title 与状态一致
    toggleBtn.title = "展开侧边栏"; 

    toggleBtn.addEventListener('click', () => {
        if (isExpanded) {
            rightPanel.style.width = originalWidth;
            toggleBtn.title = "展开侧边栏";
            // toggleBtn.innerHTML = '↔'; // 或者其他表示展开的图标/文字
        } else {
            rightPanel.style.width = expandedWidth;
            toggleBtn.title = "收起侧边栏";
            // toggleBtn.innerHTML = '↔'; // 或者其他表示收起的图标/文字
        }
        isExpanded = !isExpanded;

        // 触发 resize 事件，以便 Monaco 编辑器等组件可以自动调整布局
        // Monaco 编辑器的 automaticLayout: true 应该能处理，但显式触发更保险
        window.dispatchEvent(new Event('resize'));
    });
}

// Initialize app when DOM is ready
document.addEventListener('DOMContentLoaded', initApp);
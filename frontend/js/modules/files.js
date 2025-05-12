/**
 * Files module for handling file operations
 * @module modules/files
 */
import { projects, files } from '../api/client.js';
import { showNotification, createElement } from '../utils/dom.js';
import { getCurrentProject } from './projects.js';

let openFiles = [];

/**
 * Initialize files module
 */
export function init() {
  // Initialize file-related event listeners
  const refreshFileBtn = document.getElementById('refresh-file-btn');
  if (refreshFileBtn) {
    refreshFileBtn.addEventListener('click', refreshCurrentFile);
  }

  // Initialize refresh file list button
  const refreshFilesBtn = document.getElementById('refresh-files-btn');
  if (refreshFilesBtn) {
    refreshFilesBtn.addEventListener('click', refreshProjectFiles);
  }
}

/**
 * Refresh current project's file list
 */
export function refreshProjectFiles() {
  const currentProject = getCurrentProject();
  if (!currentProject) {
    showNotification('未选择项目', 'warning');
    return;
  }

  showNotification('正在刷新文件列表...', 'info');
  loadProjectFiles(currentProject.id)
    .then(() => {
      showNotification('文件列表已刷新', 'success');
    });
}

/**
 * Load project files
 * @param {string} projectId - Project ID
 */
export async function loadProjectFiles(projectId) {
  try {
    const fileData = await projects.getFiles(projectId);
    renderFileTree(fileData);
  } catch (error) {
    console.error('加载项目文件失败:', error);
    showNotification('加载项目文件失败', 'error');
  }
}

/**
 * Render file tree
 * @param {Object} fileData - File data
 * @param {HTMLElement} parentElement - Parent element to render into
 */
function renderFileTree(fileData, parentElement = document.getElementById('file-tree')) {
  // If rendering from root, clear current tree
  if (parentElement === document.getElementById('file-tree')) {
    parentElement.innerHTML = '';
  }

  if (fileData.is_dir) {
    // Create directory element
    const dirElement = createElement('div', {
      className: 'file-item directory',
      textContent: fileData.name,
      dataset: { path: fileData.path }
    });

    // Create expand/collapse functionality
    let expanded = false;
    dirElement.addEventListener('click', (event) => {
      event.stopPropagation();
      expanded = !expanded;
      const childContainer = dirElement.nextElementSibling;
      if (childContainer) {
        childContainer.style.display = expanded ? 'block' : 'none';
      }
    });

    parentElement.appendChild(dirElement);

    // Create child elements container
    const childContainer = createElement('div', {
      className: 'directory-children',
      style: 'display: none;' // Initially collapsed
    });
    parentElement.appendChild(childContainer);

    // If there are child elements, render them
    if (fileData.children && fileData.children.length > 0) {
      fileData.children.forEach(child => {
        renderFileTree(child, childContainer);
      });
    }
  } else {
    // Create file element
    const fileElement = createElement('div', {
      className: 'file-item file',
      textContent: fileData.name,
      dataset: { path: fileData.path },
      onClick: () => openFile(fileData.path)
    });

    parentElement.appendChild(fileElement);
  }
}

/**
 * Open a file
 * @param {string} filePath - File path
 */
export async function openFile(filePath) {
  try {
    const currentProject = getCurrentProject();
    if (!currentProject) {
      throw new Error('未选择项目');
    }

    // Check if file is already open
    const existingTabIndex = openFiles.findIndex(f => f && f.path === filePath);

    if (existingTabIndex !== -1) {
      // Just activate existing tab
      activateTab(existingTabIndex);
      return;
    }

    // Load file content
    const fileData = await files.read(currentProject.id, filePath);

    // Add to open files
    openFiles.push({
      path: filePath,
      name: filePath.split('/').pop(),
      content: fileData.content
    });

    // Create new tab and editor
    createTab(openFiles.length - 1);

    // Set file content in editor
    if (window.EditorModule && window.EditorModule.setContent) {
      window.EditorModule.setContent(fileData.content);
      window.EditorModule.setCurrentFilePath(filePath);
    }

    // Activate new tab
    activateTab(openFiles.length - 1);

    // Update session context to current file
    if (window.SessionsModule && window.SessionsModule.getCurrentSession) {
      const session = window.SessionsModule.getCurrentSession();
      if (session) {
        import('../api/client.js').then(({ sessions }) => {
          sessions.updateContext(session.id, {
            current_file: filePath
          });
        });
      }
    }

    // Update context files selector
    if (window.SessionsModule && window.SessionsModule.updateContextFilesSelector) {
      window.SessionsModule.updateContextFilesSelector();
    }
  } catch (error) {
    console.error('打开文件失败:', error);
    showNotification('打开文件失败', 'error');
  }
}

/**
 * Create a tab for file
 * @param {number} fileIndex - File index
 */
function createTab(fileIndex) {
  const file = openFiles[fileIndex];
  const tabsContainer = document.getElementById('editor-tabs');

  const tab = createElement('div', {
    className: 'tab',
    dataset: { index: fileIndex },
    onClick: () => activateTab(fileIndex)
  }, [
    createElement('span', { className: 'tab-title', textContent: file.name }),
    createElement('span', { 
      className: 'close-btn',
      onClick: (e) => {
        e.stopPropagation();
        closeFile(fileIndex);
      }
    })
  ]);

  tabsContainer.appendChild(tab);
}

/**
 * Activate a tab
 * @param {number} fileIndex - File index
 */
function activateTab(fileIndex) {
  // If file has been closed, ignore
  if (fileIndex >= openFiles.length || !openFiles[fileIndex]) {
    return;
  }

  // Deactivate all tabs
  document.querySelectorAll('.tab').forEach(tab => {
    tab.classList.remove('active');
  });

  // Activate selected tab
  const tab = document.querySelector(`.tab[data-index="${fileIndex}"]`);
  if (tab) {
    tab.classList.add('active');
  }

  // Update editor content
  const file = openFiles[fileIndex];
  if (window.EditorModule) {
    window.EditorModule.setContent(file.content);
    window.EditorModule.setCurrentFilePath(file.path);
  }

  // Update session context with current file
  if (window.SessionsModule && window.SessionsModule.getCurrentSession) {
    const session = window.SessionsModule.getCurrentSession();
    if (session) {
      import('../api/client.js').then(({ sessions }) => {
        sessions.updateContext(session.id, {
          current_file: file.path
        }).catch(console.error);
      });
    }
  }

  // Update context files selector
  if (window.SessionsModule && window.SessionsModule.updateContextFilesSelector) {
    window.SessionsModule.updateContextFilesSelector();
  }
}

/**
 * Close a file
 * @param {number} fileIndex - File index
 */
function closeFile(fileIndex) {
  // Ensure index is valid
  if (fileIndex >= openFiles.length || !openFiles[fileIndex]) {
    return;
  }

  // Get tab element
  const tab = document.querySelector(`.tab[data-index="${fileIndex}"]`);
  if (tab) {
    // Remove tab
    tab.remove();
  }

  // Mark file as closed (but keep index position to avoid renumbering complexity)
  openFiles[fileIndex] = null;

  // Check if it's the currently active tab
  const isActive = tab && tab.classList.contains('active');

  if (isActive) {
    // Find next available tab to activate
    let nextTabIndex = -1;

    // First try to find tab to the right
    for (let i = fileIndex + 1; i < openFiles.length; i++) {
      if (openFiles[i]) {
        nextTabIndex = i;
        break;
      }
    }

    // If no tab on the right, try to find one on the left
    if (nextTabIndex === -1) {
      for (let i = fileIndex - 1; i >= 0; i--) {
        if (openFiles[i]) {
          nextTabIndex = i;
          break;
        }
      }
    }

    // If we found a next tab, activate it
    if (nextTabIndex !== -1) {
      activateTab(nextTabIndex);
    } else {
      // No available tabs, clear editor
      if (window.EditorModule) {
        window.EditorModule.setContent('');
        window.EditorModule.setCurrentFilePath(null);
      }
    }
  }

  // Garbage collection: if all files are closed or null, reset array
  const hasOpenFiles = openFiles.some(file => file !== null);
  if (!hasOpenFiles) {
    openFiles = [];
  }

  // Update context files selector
  if (window.SessionsModule && window.SessionsModule.updateContextFilesSelector) {
    window.SessionsModule.updateContextFilesSelector();
  }
}

/**
 * Refresh current open file
 */
export async function refreshCurrentFile() {
  if (!window.EditorModule || !window.EditorModule.getCurrentFilePath) {
    return;
  }

  const currentFilePath = window.EditorModule.getCurrentFilePath();
  if (!currentFilePath) {
    showNotification('无打开的文件可刷新', 'warning');
    return;
  }

  const currentProject = getCurrentProject();
  if (!currentProject) {
    showNotification('未选择项目', 'error');
    return;
  }

  try {
    // Show refresh status
    showNotification('正在刷新文件...', 'info');

    // Get latest file content
    const fileData = await files.read(currentProject.id, currentFilePath);

    // Find current file's index
    const currentFileIndex = openFiles.findIndex(f => f && f.path === currentFilePath);
    if (currentFileIndex !== -1) {
      // Update file content
      openFiles[currentFileIndex].content = fileData.content;

      // Update editor content
      if (window.EditorModule && window.EditorModule.setContent) {
        window.EditorModule.setContent(fileData.content);
      }

      showNotification('文件已刷新', 'success');
    }
  } catch (error) {
    console.error('刷新文件失败:', error);
    showNotification('刷新文件失败', 'error');
  }
}

/**
 * Get all open files
 * @returns {Array} Open files
 */
export function getOpenFiles() {
  // Filter out null and undefined values
  return Array.isArray(openFiles) ? openFiles.filter(f => f !== null && f !== undefined) : [];
}

export default {
  init,
  loadProjectFiles,
  openFile,
  closeFile,
  refreshCurrentFile,
  refreshProjectFiles,
  getOpenFiles
};
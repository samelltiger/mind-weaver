/**
 * Monaco Editor module
 * @module modules/editor
 */
import { sessions } from '../api/client.js';

let editor = null;
let currentFilePath = null;
let selectionTimeout = null;
const SELECTION_DEBOUNCE_TIME = 300;

/**
 * Initialize editor
 */
export function init() {
  // Make sure loader is available
  if (typeof require === 'undefined') {
    console.error('Monaco Editor loader not available');
    setTimeout(init, 500); // Try to initialize later
    return;
  }

  // Configure Monaco loading path
  require.config({
    paths: { 'vs': 'js/third/monaco-editor@0.33.0/min/vs' }
  });

  // Load editor
  require(['vs/editor/editor.main'], function() {
    const container = document.getElementById('editor-container');
    if (!container) {
      console.error('Editor container not found');
      return;
    }

    // Create editor instance
    editor = monaco.editor.create(container, {
      value: '',
      language: 'plaintext',
      theme: 'vs',
      automaticLayout: true,
      minimap: { enabled: true },
      scrollBeyondLastLine: false,
      renderLineHighlight: 'all',
      fontFamily: 'Consolas, "Courier New", monospace',
      fontSize: 14,
      lineHeight: 20,
      lineNumbers: 'on',
      renderWhitespace: 'selection',
      tabSize: 2
    });

    // Listen for content changes
    editor.onDidChangeModelContent(function() {
      // Optional: Handle content change logic
    });

    // Listen for cursor position changes
    editor.onDidChangeCursorPosition(handleCursorPositionChange);

    // Listen for selection changes with debounce
    editor.onDidChangeCursorSelection(handleSelectionChange);

    console.log('Monaco Editor initialized successfully');
  });
}

/**
 * Handle cursor position change
 * @param {Object} e - Position change event
 */
function handleCursorPositionChange(e) {
  // Make sure editor and model exist
  if (!editor || !editor.getModel()) return;

  try {
    if (window.SessionsModule && window.SessionsModule.getCurrentSession) {
      const session = window.SessionsModule.getCurrentSession();
      if (session) {
        // Make sure getOffsetAt method exists
        if (typeof editor.getOffsetAt === 'function') {
          const offset = editor.getOffsetAt(e.position);
          sessions.updateContext(session.id, {
            cursor_position: offset,
            current_file: getCurrentFilePath()
          }).catch(err => console.error('Failed to update cursor position:', err));
        } else {
          // If getOffsetAt is not available, use alternative method
          const model = editor.getModel();
          if (model) {
            const offset = model.getOffsetAt(e.position);
            sessions.updateContext(session.id, {
              cursor_position: offset,
              current_file: getCurrentFilePath()
            }).catch(err => console.error('Failed to update cursor position:', err));
          }
        }
      }
    }
  } catch (error) {
    console.error('Error handling cursor position change:', error);
  }
}

/**
 * Handle selection change with debounce
 * @param {Object} e - Selection change event
 */
function handleSelectionChange(e) {
  // Clear previous timeout
  if (selectionTimeout) {
    clearTimeout(selectionTimeout);
    selectionTimeout = null;
  }

  // Set new timeout
  selectionTimeout = setTimeout(() => {
    try {
      // Make sure editor and model exist
      if (!editor || !editor.getModel()) return;

      // Make sure selection is not empty
      if (e.selection.isEmpty()) return;

      const model = editor.getModel();
      if (!model) return;

      const selectedText = model.getValueInRange(e.selection);
      if (!selectedText) return;

      if (window.SessionsModule && window.SessionsModule.getCurrentSession) {
        const session = window.SessionsModule.getCurrentSession();
        if (session) {
          sessions.updateContext(session.id, {
            selected_code: selectedText,
            current_file: getCurrentFilePath()
          }).catch(err => console.error('Failed to update selected code:', err));
        }
      }
    } catch (error) {
      console.error('Error handling selection change:', error);
    }
  }, SELECTION_DEBOUNCE_TIME);
}

/**
 * Check if editor is ready
 * @returns {boolean} Whether editor is ready
 */
export function isEditorReady() {
  return editor !== null && editor.getModel() !== null;
}

/**
 * Set editor content
 * @param {string} content - Content to set
 */
export function setContent(content) {
  if (!editor) {
    console.warn('Editor not initialized');
    return;
  }

  try {
    // If no content, show empty editor
    if (!content) {
      const emptyModel = monaco.editor.createModel('', 'plaintext');
      editor.setModel(emptyModel);
      return;
    }

    const langId = detectLanguage(currentFilePath);
    const model = monaco.editor.createModel(content, langId);
    editor.setModel(model);
  } catch (error) {
    console.error('Error setting editor content:', error);
    // Create empty model on error to avoid subsequent failures
    try {
      const emptyModel = monaco.editor.createModel('', 'plaintext');
      editor.setModel(emptyModel);
    } catch (e) {
      console.error('Failed to create empty model:', e);
    }
  }
}

/**
 * Detect language based on file extension
 * @param {string} filePath - File path
 * @returns {string} Language ID
 */
function detectLanguage(filePath) {
  if (!filePath) return 'plaintext';

  const extension = filePath.split('.').pop().toLowerCase();
  const langMap = {
    'js': 'javascript',
    'ts': 'typescript',
    'jsx': 'javascript',
    'tsx': 'typescript',
    'py': 'python',
    'go': 'go',
    'java': 'java',
    'html': 'html',
    'css': 'css',
    'json': 'json',
    'md': 'markdown',
    'rs': 'rust',
    'c': 'c',
    'cpp': 'cpp',
    'h': 'cpp',
    'php': 'php'
  };

  return langMap[extension] || 'plaintext';
}

/**
 * Get current editor content
 * @returns {string} Current content
 */
export function getCurrentContent() {
  if (!editor || !editor.getModel()) return '';
  return editor.getValue();
}

/**
 * Get current file path
 * @returns {string|null} Current file path
 */
export function getCurrentFilePath() {
  return currentFilePath;
}

/**
 * Set current file path
 * @param {string} path - File path
 */
export function setCurrentFilePath(path) {
  currentFilePath = path;
}

/**
 * Get cursor position
 * @returns {Object|null} Cursor position
 */
export function getCursorPosition() {
  if (!editor || !editor.getModel()) return null;

  try {
    const position = editor.getPosition();
    if (!position) return null;

    return {
      lineNumber: position.lineNumber,
      column: position.column,
      offset: editor.getModel().getOffsetAt(position)
    };
  } catch (error) {
    console.error('Failed to get cursor position:', error);
    return null;
  }
}

/**
 * Get selected code
 * @returns {string|null} Selected code
 */
export function getSelectedCode() {
  if (!editor || !editor.getModel()) return null;

  try {
    const selection = editor.getSelection();
    if (!selection || selection.isEmpty()) return null;

    return editor.getModel().getValueInRange(selection);
  } catch (error) {
    console.error('Failed to get selected code:', error);
    return null;
  }
}

/**
 * Update selected code to session context
 */
export function updateSelectedCodeToContext() {
  // Check if editor is ready
  if (!isEditorReady()) {
    console.warn('Editor not ready, cannot update selected code');
    return;
  }

  const selectedCode = getSelectedCode();
  if (!selectedCode) {
    console.warn('No selected code');
    return;
  }

  if (!window.SessionsModule || !window.SessionsModule.getCurrentSession) {
    console.warn('Sessions module not available');
    return;
  }

  const session = window.SessionsModule.getCurrentSession();
  if (!session) {
    console.warn('No active session');
    return;
  }

  try {
    sessions.updateContext(session.id, {
      selected_code: selectedCode,
      current_file: getCurrentFilePath()
    }).then(() => {
      console.log('Updated selected code to session context');
    }).catch(error => {
      console.error('Failed to update selected code:', error);
    });
  } catch (error) {
    console.error('Error processing selected code:', error);
  }
}

export default {
  init,
  isEditorReady,
  setContent,
  getCurrentContent,
  getCurrentFilePath,
  setCurrentFilePath,
  getCursorPosition,
  getSelectedCode,
  updateSelectedCodeToContext
};
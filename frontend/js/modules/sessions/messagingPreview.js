/**
 * Messaging component
 * @module modules/sessions/messaging
 */
import { SERVER_URL } from "../../api/client.js";
import {
  showNotification,
} from "../../utils/dom.js";

let editorPreviewContainer = null;
let editorPreviewIframe = null;
let closeEditorPreviewBtn = null;
let refreshEditorPreviewBtn = null;
let currentPreviewPath = null; // 用于存储当前预览的文件路径

export function getCurrentPreviewPath() {
  return currentPreviewPath;
}

/**
 * 显示或更新编辑器覆盖式 HTML 预览
 * @param {string} relativePath - AI 返回的相对文件路径
 */
export function showHtmlPreview(relativePath) {
  const currentProject = window.ProjectsModule.getCurrentProject();
  if (!currentProject || !relativePath) {
    console.warn("无法显示编辑器预览：缺少项目或路径信息。");
    // showNotification("无法显示编辑器预览：缺少项目或路径信息。", "warning");
    return;
  }

  // 首次或需要时获取 DOM 元素
  if (!editorPreviewContainer) {
    editorPreviewContainer = document.getElementById(
      "editor-preview-container"
    );
  }
  if (!editorPreviewIframe) {
    editorPreviewIframe = document.getElementById("editor-preview-iframe");
  }
  if (!closeEditorPreviewBtn) {
    closeEditorPreviewBtn = document.getElementById("close-editor-preview-btn");
    // 添加关闭事件监听器 (只添加一次)
    if (closeEditorPreviewBtn) {
      closeEditorPreviewBtn.addEventListener("click", closeHtmlPreview);
    }
  }
  if (!refreshEditorPreviewBtn) {
    refreshEditorPreviewBtn = document.getElementById(
      "refresh-editor-preview-btn"
    );
    // 添加刷新事件监听器 (只添加一次)
    if (refreshEditorPreviewBtn) {
      refreshEditorPreviewBtn.addEventListener("click", () => {
        if (currentPreviewPath) {
          console.log("手动刷新预览:", currentPreviewPath);
          showHtmlPreview(currentPreviewPath); // 使用存储的路径刷新
        } else {
          showNotification("没有可刷新的预览路径", "warning");
        }
      });
    }
  }

  // 检查元素是否存在
  if (
    !editorPreviewContainer ||
    !editorPreviewIframe ||
    !closeEditorPreviewBtn ||
    !refreshEditorPreviewBtn
  ) {
    console.error("编辑器预览所需的 DOM 元素未找到。");
    showNotification("无法初始化编辑器预览界面", "error");
    return;
  }

  // 存储当前路径以供刷新使用
  currentPreviewPath = relativePath;

  // 构造带有缓存清除参数的 URL
  const fullPath = currentProject.path + "/" + relativePath.replace(/^\//, "");
  const timestamp = Date.now(); // 使用时间戳作为缓存清除参数
  const previewUrl = `${SERVER_URL}/api/files/single-html?path=${encodeURIComponent(
    fullPath
  )}&t=${timestamp}`;

  console.log("准备在编辑器预览框中显示/刷新:", previewUrl);

  try {
    // 更新 iframe 的 src 来加载或刷新内容
    editorPreviewIframe.src = previewUrl;

    // 显示预览容器
    editorPreviewContainer.style.display = "flex"; // 使用 'flex'

    console.log("编辑器预览已更新并显示。");
  } catch (error) {
    console.error("更新编辑器预览时出错:", error);
    showNotification("更新预览时出错", "error");
    closeHtmlPreview(); // 出错时尝试关闭
  }
}

/**
 * 关闭编辑器覆盖式 HTML 预览
 */
export function closeHtmlPreview() {
  // 确保元素已获取
  if (!editorPreviewContainer) {
    editorPreviewContainer = document.getElementById(
      "editor-preview-container"
    );
  }
  if (!editorPreviewIframe) {
    editorPreviewIframe = document.getElementById("editor-preview-iframe");
  }

  if (editorPreviewContainer) {
    editorPreviewContainer.style.display = "none"; // 隐藏容器
    console.log("编辑器预览已关闭。");
  }
  if (editorPreviewIframe) {
    // 重置 iframe 内容，释放资源
    editorPreviewIframe.src = "about:blank";
  }
  // 重置当前预览路径
  currentPreviewPath = null;
  // 注意：不需要移除按钮的事件监听器，因为它们只添加了一次
}

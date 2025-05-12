/**
 * DOM utility functions
 * @module utils/dom
 */

/**
 * Create an element with attributes and children
 * @param {string} tag - Tag name
 * @param {Object} attrs - Attributes
 * @param {Array|string} children - Child elements or text
 * @returns {HTMLElement} Created element
 */
export function createElement(tag, attrs = {}, children = []) {
  const element = document.createElement(tag);

  // Set attributes
  Object.entries(attrs).forEach(([key, value]) => {
    if (key === "className") {
      element.className = value;
    } else if (key === "dataset") {
      Object.entries(value).forEach(([dataKey, dataValue]) => {
        element.dataset[dataKey] = dataValue;
      });
    } else if (key === "textContent") {
      element.textContent = value;
    } else if (key === "innerHTML") {
      element.innerHTML = value;
    } else if (key.startsWith("on") && typeof value === "function") {
      element.addEventListener(key.substring(2).toLowerCase(), value);
    } else {
      element.setAttribute(key, value);
    }
  });

  // Add children
  if (Array.isArray(children)) {
    children.forEach((child) => {
      if (typeof child === "string") {
        element.appendChild(document.createTextNode(child));
      } else if (child instanceof Node) {
        element.appendChild(child);
      }
    });
  } else if (typeof children === "string") {
    element.textContent = children;
  }

  return element;
}

/**
 * Show a notification
 * @param {string} message - Notification message
 * @param {string} type - Notification type ('info', 'success', 'error', 'warning')
 * @param {number} duration - Duration in milliseconds
 */
export function showNotification(message, type = "info", duration = 3000) {
  const notificationArea = document.getElementById("notification-area");
  if (!notificationArea) return;

  notificationArea.textContent = message;
  notificationArea.className = `notification ${type}`;

  setTimeout(() => {
    notificationArea.textContent = "";
    notificationArea.className = "";
  }, duration);
}

/**
 * Copy text to clipboard
 * @param {string} text - Text to copy
 * @returns {boolean} Success
 */
export function copyToClipboard(text) {
  // Create temporary textarea
  const textArea = document.createElement("textarea");
  textArea.value = text;
  textArea.style.position = "fixed";
  textArea.style.left = "-999999px";
  textArea.style.top = "-999999px";
  document.body.appendChild(textArea);
  textArea.focus();
  textArea.select();

  try {
    const successful = document.execCommand("copy");
    document.body.removeChild(textArea);
    return successful;
  } catch (err) {
    document.body.removeChild(textArea);
    return false;
  }
}

/**
 * Format markdown content to HTML with code blocks
 * @param {string} content - Markdown content
 * @returns {string} Formatted HTML
 */
export function formatMarkdown(content) {
  if (!content) return "";

  // If marked library isn't loaded, use simple formatting
  if (typeof marked === "undefined") {
    return simpleFormatMarkdown(content);
  }

  // Configure marked options
  marked.setOptions({
    highlight: function (code) {
      return code;
    },
  });

  // Convert content to HTML using marked
  const html = marked.parse(content);

  // Process all code blocks to add copy buttons
  const tempDiv = document.createElement("div");
  tempDiv.innerHTML = html;

  const codeBlocks = tempDiv.querySelectorAll("pre > code");
  codeBlocks.forEach((codeBlock) => {
    const pre = codeBlock.parentNode;
    pre.classList.add("code-block");

    const copyBtn = document.createElement("button");
    copyBtn.className = "copy-code-btn";
    copyBtn.textContent = "复制";
    copyBtn.setAttribute("onclick", "copyToClipboard(this)");

    pre.insertBefore(copyBtn, pre.firstChild);
  });

  return tempDiv.innerHTML;
}

/**
 * Simple markdown formatter (fallback)
 * @param {string} content - Markdown content
 * @returns {string} Formatted HTML
 */
function simpleFormatMarkdown(content) {
  return content.replace(
    /```([\w-]*)\n([\s\S]*?)\n```/g,
    function (match, language, code) {
      return `<pre class="code-block ${language}"><code>${escapeHtml(
        code
      )}</code></pre>`;
    }
  );
}

/**
 * Escape HTML special characters
 * @param {string} html - HTML to escape
 * @returns {string} Escaped HTML
 */
export function escapeHtml(html) {
  const div = document.createElement("div");
  div.textContent = html;
  return div.innerHTML;
}

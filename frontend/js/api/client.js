/**
 * MindWeaver API client
 * @module api/client
 */

// const SERVER_URL = "{{vars.API_URL}}";
export const SERVER_URL = "";

/**
 * Handle API response and extract data
 * @param {Response} response - Fetch API response
 * @returns {Promise<any>} Extracted data from response
 * @throws {Error} If API returns error code
 */
async function handleResponse(response) {
  const data = await response.json();
  if (data.code !== 0) {
    throw new Error(data.msg || "API request failed");
  }
  return data.data;
}

/**
 * Models API endpoints
 */
export const models = {
  /**
   * Get all available models
   * @returns {Promise<Array>} List of available models
   */
  getAll: async function () {
    const response = await fetch(`${SERVER_URL}/api/models`);
    return handleResponse(response);
  },
};

/**
 * Projects API endpoints
 */
export const projects = {
  /**
   * Get all projects
   * @returns {Promise<Array>} List of projects
   */
  getAll: async function () {
    const response = await fetch(`${SERVER_URL}/api/projects`);
    return handleResponse(response);
  },

  /**
   * Create a new project
   * @param {Object} project - Project data
   * @returns {Promise<Object>} Created project
   */
  create: async function (project) {
    const response = await fetch(`${SERVER_URL}/api/projects`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(project),
    });
    return handleResponse(response);
  },

  /**
   * Update existing project
   * @param {string} id - Project ID
   * @param {Object} project - Updated project data
   * @returns {Promise<Object>} Updated project
   */
  update: async function (id, project) {
    const response = await fetch(`${SERVER_URL}/api/projects/${id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(project),
    });
    return handleResponse(response);
  },

  /**
   * Get project by ID
   * @param {string} id - Project ID
   * @returns {Promise<Object>} Project data
   */
  getById: async function (id) {
    const response = await fetch(`${SERVER_URL}/api/projects/${id}`);
    return handleResponse(response);
  },

  /**
   * Get project files
   * @param {string} id - Project ID
   * @param {number} maxDepth - Maximum directory depth
   * @returns {Promise<Object>} Project files tree
   */
  getFiles: async function (id, maxDepth = 5) {
    const response = await fetch(
      `${SERVER_URL}/api/projects/${id}/files?maxDepth=${maxDepth}`
    );
    return handleResponse(response);
  },
};

/**
 * Files API endpoints
 */
export const files = {
  /**
   * Read file content
   * @param {string} projectId - Project ID
   * @param {string} filePath - File path
   * @returns {Promise<Object>} File content
   */
  read: async function (projectId, filePath) {
    const response = await fetch(`${SERVER_URL}/api/files/read`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ project_id: projectId, file_path: filePath }),
    });
    return handleResponse(response);
  },
};

/**
 * Sessions API endpoints
 */
export const sessions = {
  /**
   * Create a new session
   * @param {string} projectId - Project ID
   * @param {string} name - Session name
   * @param {string} mode - Session mode ('manual', 'auto', 'all')
   * @param {Array} excludePatterns - Patterns to exclude
   * @param {Array} includePatterns - Patterns to include
   * @returns {Promise<Object>} Created session
   */
  create: async function (
    projectId,
    name,
    mode = "manual",
    excludePatterns = [],
    includePatterns = []
  ) {
    const response = await fetch(`${SERVER_URL}/api/sessions`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        project_id: projectId,
        name,
        mode,
        exclude_patterns: excludePatterns,
        include_patterns: includePatterns,
      }),
    });
    return handleResponse(response);
  },

  /**
   * Update existing session
   * @param {string} id - Session ID
   * @param {Object} updates - Updates to apply
   * @returns {Promise<Object>} Updated session
   */
  update: async function (id, updates) {
    const response = await fetch(`${SERVER_URL}/api/sessions/${id}`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(updates),
    });
    return handleResponse(response);
  },

  /**
   * Get sessions by project
   * @param {string} projectId - Project ID
   * @returns {Promise<Array>} Project sessions
   */
  getByProject: async function (projectId) {
    const response = await fetch(
      `${SERVER_URL}/api/sessions/project/${projectId}`
    );
    return handleResponse(response);
  },

  /**
   * Get session by ID
   * @param {string} id - Session ID
   * @returns {Promise<Object>} Session data
   */
  getById: async function (id) {
    const response = await fetch(`${SERVER_URL}/api/sessions/${id}`);
    return handleResponse(response);
  },

  /**
   * Delete a session
   * @param {string} id - Session ID
   * @returns {Promise<Object>} Delete result
   */
  delete: async function (id) {
    const response = await fetch(`${SERVER_URL}/api/sessions/${id}`, {
      method: "DELETE",
    });
    return handleResponse(response);
  },

  /**
   * Send message to session
   * @param {string} sessionId - Session ID
   * @param {string} content - Message content
   * @param {string} projectPath - Project path
   * @param {Array} contextFiles - Context files
   * @param {string} model - Model to use
   * @returns {Promise<Object>} Message response
   */
  sendMessage: async function (
    sessionId,
    content,
    projectPath,
    contextFiles,
    model,
    tool = null, // Add tool parameter
  ) {
    const response = await fetch(
      `${SERVER_URL}/api/sessions/${sessionId}/message`,
      {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          content,
          project_path: projectPath,
          context_files: contextFiles,
          model: model,
          tool_use: tool,
        }),
      }
    );
    return handleResponse(response);
  },

  /**
   * OpenAI-compatible streaming completions
   * @param {string} sessionId - Session ID
   * @param {Object} request - OpenAI compatible request
   * @param {Function} onMessage - Message callback
   * @param {Function} onComplete - Completion callback
   * @param {Function} onError - Error callback
   * @returns {Object} Stream controller
   */
  streamCompletions: function (
    sessionId,
    request,
    onMessage,
    onComplete,
    onError
  ) {
    const encoder = new TextEncoder();
    const decoder = new TextDecoder();

    // Create a stream-compatible request
    const requestData = {
      ...request,
      stream: true,
    };

    // Use fetch with POST method but handle as EventSource
    const fetchController = new AbortController();
    const { signal } = fetchController;

    fetch(`${SERVER_URL}/api/sessions/${sessionId}/completions`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(requestData),
      signal,
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const reader = response.body.getReader();
        let buffer = "";

        function readStream() {
          reader
            .read()
            .then(({ done, value }) => {
              if (done) {
                if (buffer.length > 0) {
                  try {
                    const lastChunk = JSON.parse(buffer);
                    onMessage(lastChunk);
                  } catch (err) {
                    console.error("Error parsing last chunk:", err);
                  }
                }
                if (onComplete) onComplete();
                return;
              }

              // Decode the chunk and add to buffer
              const chunk = decoder.decode(value, { stream: true });
              buffer += chunk;

              // Process complete messages
              const lines = buffer.split("\n");
              buffer = lines.pop() || ""; // Keep the incomplete line

              for (const line of lines) {
                if (line.startsWith("data: ")) {
                  try {
                    const data = line.substring(6); // Remove 'data: ' prefix
                    // if (data === "[DONE]") {
                    if (data.startsWith("[DONE]")) {
                      if (onComplete) onComplete();
                      return;
                    }

                    const parsedData = JSON.parse(data);
                    onMessage(parsedData);
                  } catch (err) {
                    console.error("Error parsing stream data:", err);
                  }
                }
              }

              readStream();
            })
            .catch((error) => {
              if (error.name !== "AbortError") {
                console.error("Stream reading error:", error);
                if (onError) onError(error);
              }
            });
        }

        readStream();
      })
      .catch((error) => {
        if (error.name !== "AbortError") {
          console.error("Fetch error:", error);
          if (onError) onError(error);
        }
      });

    return {
      close: function () {
        fetchController.abort();
      },
    };
  },

  /**
   * Update session context
   * @param {string} sessionId - Session ID
   * @param {Object} context - Context data
   * @returns {Promise<Object>} Update result
   */
  updateContext: async function (sessionId, context) {
    const response = await fetch(
      `${SERVER_URL}/api/sessions/${sessionId}/context`,
      {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(context),
      }
    );
    return handleResponse(response);
  },

  /**
   * Get session context
   * @param {string} sessionId - Session ID
   * @returns {Promise<Object>} Session context
   */
  getContext: async function (sessionId) {
    const response = await fetch(
      `${SERVER_URL}/api/sessions/${sessionId}/context`
    );
    return handleResponse(response);
  },


  /**
   * Delete a message in a session
   * @param {string} sessionId - Session ID
   * @param {string} messageId - Message ID, use '0' to delete all messages
   * @returns {Promise<Object>} Delete result
   */
  deleteMessage: async function (sessionId, messageId) {
    const response = await fetch(
      `${SERVER_URL}/api/sessions/${sessionId}/messages/${messageId}`,
      {
        method: "DELETE",
      }
    );
    return handleResponse(response);
  },

  /**
   * Clear all messages in a session
   * @param {string} sessionId - Session ID
   * @returns {Promise<Object>} Clear result
   */
  clearAllMessages: async function (sessionId) {
    return this.deleteMessage(sessionId, "0");
  },
};

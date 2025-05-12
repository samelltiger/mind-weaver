/**
 * Models component
 * @module modules/sessions/models
 */
import { models } from '../../api/client.js';

/**
 * Load available models
 */
export async function loadModels() {
  try {
    // Make sure model selector exists
    const modelSelector = document.getElementById("model-selector");
    if (!modelSelector) return;

    // Clear existing options, just keep a loading option
    modelSelector.innerHTML = '<option value="">加载中...</option>';

    // Call API to get models list
    const modelsList = await models.getAll();

    // Clear loading option
    modelSelector.innerHTML = "";

    // Add model options
    if (modelsList && modelsList.length > 0) {
      modelsList.forEach((model) => {
        const option = document.createElement("option");
        option.value = model.name;
        option.textContent = model.name;
        // If GPT-4, select by default
        if (model.name.toLowerCase().includes("gpt-4")) {
          option.selected = true;
        }
        modelSelector.appendChild(option);
      });
    } else {
      // If no models returned, add default options
      const defaultModels = [
        { name: "gpt-3.5-turbo", display: "GPT-3.5" },
        { name: "gpt-4", display: "GPT-4" },
        { name: "claude-3", display: "Claude 3" },
      ];

      defaultModels.forEach((model) => {
        const option = document.createElement("option");
        option.value = model.name;
        option.textContent = model.display;
        if (model.name === "gpt-4") {
          option.selected = true;
        }
        modelSelector.appendChild(option);
      });
    }
  } catch (error) {
    console.error("加载模型列表失败:", error);
    // On failure, add default options
    const modelSelector = document.getElementById("model-selector");
    if (modelSelector) {
      modelSelector.innerHTML = `
        <option value="gpt-3.5-turbo">GPT-3.5</option>
        <option value="gpt-4" selected>GPT-4</option>
        <option value="claude-3">Claude 3</option>
      `;
    }
  }
}
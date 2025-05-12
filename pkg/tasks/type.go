package tasks

import "time"

// Task represents a task to be executed.
type Task struct {
	ID        string      `json:"id"`         // Unique task ID
	Name      string      `json:"name"`       // Task name (e.g., "send_email", "process_image")
	Status    string      `json:"status"`     // e.g., "pending", "running", "completed", "failed"
	Result    interface{} `json:"result"`     // Result of the task execution (if any)
	Payload   interface{} `json:"payload"`    // Task-specific data
	CreatedAt time.Time   `json:"created_at"` // Timestamp when the task was created
}

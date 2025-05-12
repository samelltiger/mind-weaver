package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// TaskHandler defines the function signature for handling tasks.
type TaskHandler func(ctx context.Context, task *Task) error

// RegisterHandlers registers task handlers for different task names.
var handlers = make(map[string]TaskHandler)

func RegisterHandler(taskName string, handler TaskHandler) {
	handlers[taskName] = handler
}

// CreateTask creates a new task with a unique ID and initial status.
func CreateTask(name string, payload interface{}) *Task {
	return &Task{
		ID:        uuid.NewString(),
		Name:      name,
		Status:    "pending",
		Payload:   payload,
		CreatedAt: time.Now(),
	}
}

// GetTask retrieves a task by ID from Redis.
func GetTask(ctx context.Context, rdb *redis.Client, taskID string) (*Task, error) {
	taskJSON, err := rdb.HGet(ctx, "tasks", taskID).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("task not found: %s", taskID)
		}
		return nil, err
	}

	var task Task
	err = json.Unmarshal([]byte(taskJSON), &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTaskStatus updates the status and result of a task in Redis.
func UpdateTaskStatus(ctx context.Context, rdb *redis.Client, taskID, status string, result interface{}) error {
	task, err := GetTask(ctx, rdb, taskID)
	if err != nil {
		return err
	}

	task.Status = status
	task.Result = result

	updatedTaskJSON, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return rdb.HSet(ctx, "tasks", taskID, updatedTaskJSON).Err()
}

// QueueLength returns the number of pending tasks in a specific queue.
func QueueLength(ctx context.Context, rdb *redis.Client, queueName string) (int64, error) {
	return rdb.LLen(ctx, queueName).Result()
}

// GetTaskQueuePosition returns the position of a task in a queue.
func GetTaskQueuePosition(ctx context.Context, rdb *redis.Client, queueName, taskID string) (int64, error) {
	tasks, err := rdb.LRange(ctx, queueName, 0, -1).Result()
	if err != nil {
		return -1, err
	}

	for i, taskJSON := range tasks {
		var task Task
		if err := json.Unmarshal([]byte(taskJSON), &task); err != nil {
			return -1, fmt.Errorf("error unmarshalling task in queue: %w", err) // Wrap the error
		}
		if task.ID == taskID {
			return int64(i), nil
		}
	}
	return -1, fmt.Errorf("task not found in queue")
}

// PublishTask publishes a task to the Redis queue.
func PublishTask(ctx context.Context, rdb *redis.Client, task *Task, queueName string) error {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return err
	}
	// Store the task in a hash for retrieval by ID
	err = rdb.HSet(ctx, "tasks", task.ID, taskJSON).Err()
	if err != nil {
		return err
	}

	return rdb.LPush(ctx, queueName, taskJSON).Err()

}

// ... (TaskHandler, RegisterHandler, and StartWorker remain largely the same)

func StartWorker(ctx context.Context, rdb *redis.Client, queueName string) { // Updated StartWorker
	for {
		select {
		case <-ctx.Done():
			return // Exit worker gracefully when context is canceled
		default:
			result, err := rdb.BRPop(ctx, 0, queueName).Result()
			if err != nil {
				if err == redis.Nil {
					continue // No tasks in the queue
				}
				log.Printf("Error retrieving task: %v", err)
				time.Sleep(time.Second)
				continue
			}

			var task Task
			if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
				log.Printf("Error unmarshalling task: %v", err)
				continue
			}

			handler, ok := handlers[task.Name]
			if !ok {
				log.Printf("No handler registered for task: %s", task.Name)
				continue
			}

			if err := UpdateTaskStatus(ctx, rdb, task.ID, "running", nil); err != nil {
				log.Printf("Error updating task status: %v", err)
			}

			if err := handler(ctx, &task); err != nil {
				log.Printf("Error handling task: %v", err)
				if err := UpdateTaskStatus(ctx, rdb, task.ID, "failed", err.Error()); err != nil { // Update status to "failed"
					log.Printf("Error updating task status: %v", err)
				}

				// Consider re-queueing the task or other error handling strategies
				continue // Skip to next task
			}

			if err := UpdateTaskStatus(ctx, rdb, task.ID, "completed", task.Result); err != nil {
				log.Printf("Error updating task status: %v", err)
			}

		}
	}
}

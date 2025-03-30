package storage

import "github.com/YuriLeonel/task-manager/pkg/models"

// copyTask creates a deep copy of a task.
func copyTask(sourceTask *models.Task) *models.Task {
	task := &models.Task{
		ID:          sourceTask.ID,
		Description: sourceTask.Description,
		Completed:   sourceTask.Completed,
		CreatedAt:   sourceTask.CreatedAt,
		UpdatedAt:   sourceTask.UpdatedAt,
		Priority:    sourceTask.Priority,
		Progress:    sourceTask.Progress,
		Tags:        make([]string, len(sourceTask.Tags)),
	}

	// Copy due date if exists
	if sourceTask.DueDate != nil {
		dueDate := *sourceTask.DueDate
		task.DueDate = &dueDate
	}

	// Copy tags
	copy(task.Tags, sourceTask.Tags)

	return task
}

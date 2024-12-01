package database

import (
	"database/sql"

	"github.com/sirupsen/logrus"
)

type Task struct {
	ID          int    `json:"id"`
	DagID       int    `json:"dag_id"`
	Status      string `json:"status"`
	ParentID    *int   `json:"parent_id"` // Nullable
	ExecutorID  int    `json:"executor_id"`
	Type        string `json:"type"`
	definition  string
	CreatedAt   string `json:"created_at"`
}

type TaskOperations struct {
    Logger *logrus.Logger
    DB     *sql.DB
}

func (to *TaskOperations) GetTasksByDagID(id int) ([]Task, error) {
	rows, err := to.DB.Query(`
		SELECT id, dag_id, status, parent_id, executor_id, type, definition, created_at
		FROM task
        WHERE id=$1
		ORDER BY created_at DESC
    `, id)
	if err != nil {
		to.Logger.Error("Error fetching tasks:", err)
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.DagID, &task.Status, &task.ParentID, &task.ExecutorID, &task.Type, &task.definition, &task.CreatedAt); err != nil {
			to.Logger.Error("Error scanning task:", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (to *TaskOperations) CreateTask(dagID int, executorID int, taskType string, definition string, parentID int) (int, *sql.Tx, error) {
	var id int
    txn, err := to.DB.Begin()
    if err != nil {
        return 0, nil, err
    }

	err = txn.QueryRow(`
		INSERT INTO task (dag_id, parent_id, executor_id, type, definition)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`, dagID, parentID, executorID, taskType, definition).Scan(&id)
	if err != nil {
		to.Logger.Error("Error creating task:", err)
		return 0, nil, err
	}

	return id, txn, nil
}

func (to *TaskOperations) DeleteTask(id int) error {
	_, err := to.DB.Exec(`DELETE FROM task WHERE id = $1`, id)
	if err != nil {
		to.Logger.Error("Error deleting task:", err)
		return err
	}

    return nil
}

func (to *TaskOperations) GetTaskByID(id int) (*Task, error) {
	var task Task
	err := to.DB.QueryRow(`
		SELECT id, dag_id, status, parent_id, executor_id, type, definition, created_at
		FROM task
		WHERE id = $1`, id).Scan(&task.ID, &task.DagID, &task.Status, &task.ParentID, &task.ExecutorID, &task.Type, &task.definition, &task.CreatedAt)

	if err != nil {
		to.Logger.Error("Error fetching task:", err)
		return nil, err
	}

	return &task, nil
}


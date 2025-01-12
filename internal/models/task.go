package models

import (
	"database/sql"

	"github.com/sirupsen/logrus"
    "github.com/chiragsoni81245/dagger/internal/types"
)

type TaskOperations struct {
    Logger *logrus.Logger
    DB     *sql.DB
}

func (to *TaskOperations) GetTasksByDagID(id int) ([]types.Task, error) {
	rows, err := to.DB.Query(`
		SELECT id, dag_id, name, status, parent_id, executor_id, type, definition, created_at
		FROM task
        WHERE dag_id=$1
		ORDER BY created_at DESC
    `, id)
	if err != nil {
		to.Logger.Error("Error fetching tasks:", err)
		return nil, err
	}
	defer rows.Close()

	var tasks []types.Task
	for rows.Next() {
		var task types.Task
		if err := rows.Scan(&task.ID, &task.DagID, &task.Name, &task.Status, &task.ParentID, &task.ExecutorID, &task.Type, &task.Definition, &task.CreatedAt); err != nil {
			to.Logger.Error("Error scanning task:", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (to *TaskOperations) CreateTask(dagID int, executorID int, name string, taskType string, definition string, parentID *int) (int, *sql.Tx, error) {
	var id int
    txn, err := to.DB.Begin()
    if err != nil {
        return 0, nil, err
    }

    // Validate this task should not be making any cycle

	err = txn.QueryRow(`
		INSERT INTO task (dag_id, name, parent_id, executor_id, type, definition)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`, dagID, name, parentID, executorID, taskType, definition).Scan(&id)
	if err != nil {
        txnErr := txn.Rollback()
        if txnErr != nil {
            to.Logger.Error("Error rolling back create task transaction: ", err)
            return 0, nil, txnErr
        }
		to.Logger.Error("Error creating task:", err)
		return 0, nil, err
	}

	return id, txn, nil
}

func (to *TaskOperations) DeleteTask(id int) (*sql.Tx, error) {
    txn, err := to.DB.Begin()
    if err != nil {
        return nil, err
    }
    result, err := txn.Exec(`
        DELETE FROM task
        USING dag
        WHERE 
            task.id = $1 and 
            dag.status = 'created'
    `, id)
	if err != nil {
        if txnErr := txn.Rollback(); txnErr != nil {
            to.Logger.Error("Error rolling back create task transaction: ", err)
            return nil, txnErr
        }
		to.Logger.Error("Error deleting task:", err)
		return nil, err
	}
    rows_affected, err := result.RowsAffected()
	if err != nil {
        if txnErr := txn.Rollback(); txnErr != nil {
            to.Logger.Error("Error rolling back create task transaction: ", err)
            return nil, txnErr
        }
		to.Logger.Error("Error deleting dag:", err)
		return nil, err
	}
    if rows_affected == 0 {
        return nil, NoRowsAffectedError
    }
    

    return txn, nil
}

func (to *TaskOperations) GetTaskByID(id int) (*types.Task, error) {
	var task types.Task
	err := to.DB.QueryRow(`
		SELECT id, dag_id, name, status, parent_id, executor_id, type, definition, created_at
		FROM task
		WHERE id = $1`, id).Scan(&task.ID, &task.Name, &task.DagID, &task.Status, &task.ParentID, &task.ExecutorID, &task.Type, &task.Definition, &task.CreatedAt)

	if err != nil {
		to.Logger.Error("Error fetching task:", err)
		return nil, err
	}

	return &task, nil
}


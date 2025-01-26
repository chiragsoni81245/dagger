package models

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/chiragsoni81245/dagger/internal/types"
	"github.com/sirupsen/logrus"
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
		WHERE id = $1`, id).Scan(&task.ID, &task.DagID, &task.Name, &task.Status, &task.ParentID, &task.ExecutorID, &task.Type, &task.Definition, &task.CreatedAt)

	if err != nil {
		to.Logger.Error("Error fetching task:", err)
		return nil, err
	}

	return &task, nil
}

func (to *TaskOperations) GetTaskLogs(id int) ([]*types.TaskLog, error) {
	var task types.Task
	err := to.DB.QueryRow(`
		SELECT id, dag_id, name, status, parent_id, executor_id, type, definition, created_at
		FROM task
		WHERE id = $1`, id).Scan(&task.ID, &task.DagID, &task.Name, &task.Status, &task.ParentID, &task.ExecutorID, &task.Type, &task.Definition, &task.CreatedAt)

	if err != nil {
		to.Logger.Error("Error fetching task logs:", err)
		return nil, err
	}

    if task.Status == "created" {
        return nil, TaskNotStarted
    }
    if task.Status == "running" {
        return nil, TaskIsStillRunning
    }

    taskLogs := []*types.TaskLog{}
    taskLogDir := fmt.Sprintf("logs/task-%d", id)
    entries, err := os.ReadDir(taskLogDir)
    if err != nil {
        if (os.IsNotExist(err)){
            return taskLogs, nil 
        }
        to.Logger.Error("Error in fetching log details: ", err)
        return nil, err
    }


    for _, entry := range entries {
        if entry.IsDir() { continue }
        fileNameParts := strings.Split(entry.Name(), ".")
        if fileNameParts[len(fileNameParts)-1] != "log" {
            // Skipping files which does not have extension as "log"
            continue
        }
        logFileName := strings.Join(fileNameParts[:len(fileNameParts)-1], ".")
        logNameWords := []string{}
        for _, word := range strings.Split(logFileName, "-") {
            logNameWords = append(logNameWords, strings.ToUpper(string(word[0])) + word[1:])
        }
        logName := strings.Join(logNameWords, " ")
        
        taskLogs = append(taskLogs, &types.TaskLog{
            Name: logName,
            URL: fmt.Sprintf("/api/v1/tasks/%d/logs/%s", id, logFileName),
        })
    }

	return taskLogs, nil
}

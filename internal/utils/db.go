package utils

import (
	"database/sql"
	"errors"

	"github.com/chiragsoni81245/dagger/internal/types"
)

func UpdateTaskStatus(db *sql.DB, eventCh chan types.Event, dagId int, taskId int, status string) error {
    result, err := db.Exec(`
    UPDATE task
    SET status=$1
    WHERE id=$2
    `, status, taskId)
    rowsAffected, err := result.RowsAffected()

    if err != nil {
        return err
    }

    if rowsAffected != 1 {
        return errors.New("error in updating task status")
    }
    eventCh <- types.Event{
        Resource: "task",
        ID: taskId,
        ParentResource: &types.EventResource{Resource: "dag", ID: dagId},
        Field: "status",
        NewValue: status,
    }

    return nil
}

func UpdateDagStatus(db *sql.DB, eventCh chan types.Event, dagId int, status string) error {
    result, err := db.Exec(`
    UPDATE dag
    SET status=$1
    WHERE id=$2
    `, status, dagId)
    rowsAffected, err := result.RowsAffected()

    if err != nil {
        return err
    }

    if rowsAffected != 1 {
        return errors.New("error in updating dag status")
    }

    eventCh <- types.Event{
        Resource: "dag",
        ID: dagId,
        Field: "status",
        NewValue: status,
    }

    return nil
}


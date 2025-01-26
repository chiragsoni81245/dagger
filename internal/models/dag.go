package models

import (
	"database/sql"

	"github.com/chiragsoni81245/dagger/internal/types"
	"github.com/chiragsoni81245/dagger/internal/controller"
	"github.com/sirupsen/logrus"
)

type DagOperations struct {
    Logger *logrus.Logger
    DB     *sql.DB
    EventCh chan types.Event
}

func (do *DagOperations) GetDags(page int, perPage int) ([]types.Dag, int, error) {
    var total_dags int;
    total_query_row := do.DB.QueryRow(`
        SELECT count(*)
        FROM dag
    `)
    total_query_row.Scan(&total_dags)

	rows, err := do.DB.Query(`
    WITH task_counts AS (
    	SELECT 
    		dag_id,
    		status,
    		COUNT(*) as count 
    	FROM task
    	GROUP BY status, dag_id
    )
    SELECT 
    		d.id, 
    		d.name, 
    		d.status, 
    		d.created_at, 
    		coalesce(tc_total.count, 0) - coalesce(tc_completed.count, 0) as pending_tasks, 
    		coalesce(tc_running.count, 0) as running_tasks, 
    		coalesce(tc_completed.count, 0) as completed_tasks 
    FROM dag as d
    LEFT JOIN (
    	SELECT dag_id, count(*)
    	FROM task
    	GROUP BY dag_id
    ) as tc_total ON tc_total.dag_id=d.id
    LEFT JOIN task_counts as tc_running ON tc_running.dag_id=d.id and tc_running.status='running'
    LEFT JOIN task_counts as tc_completed ON tc_completed.dag_id=d.id and tc_completed.status='completed'
    ORDER BY d.created_at DESC 
    LIMIT $1 OFFSET $2`, perPage, (page-1)*perPage)
	if err != nil {
        do.Logger.Error(err)
		return nil, 0, err
	}
	defer rows.Close()

	var dags []types.Dag
	for rows.Next() {
		var dag types.Dag
		if err := rows.Scan(&dag.ID, &dag.Name, &dag.Status, &dag.CreatedAt, &dag.PendingTasks, &dag.RunningTasks, &dag.CompletedTasks); err != nil {
			do.Logger.Println("Error scanning dag:", err)
			return nil, 0, err
		}
		dags = append(dags, dag)
	}

	return dags, total_dags, nil
}

func (do *DagOperations) GetDagByID(id int) (*types.DagWithTasks, error) {
	row := do.DB.QueryRow(`
		SELECT id, name, status, created_at 
		FROM dag 
		WHERE id=$1`, id)

	var dag types.DagWithTasks
    if err := row.Scan(&dag.ID, &dag.Name, &dag.Status, &dag.CreatedAt); err != nil {
        do.Logger.Println("Error scanning dag:", err)
        return nil, err
    }
    to := TaskOperations{Logger: do.Logger, DB: do.DB}
    tasks, err := to.GetTasksByDagID(id) 
	if err != nil {
		return nil, err
	}
    dag.Tasks = tasks
	return &dag, nil
}

func (do *DagOperations) CreateDag(name string) (int, error) {
	var id int
	err := do.DB.QueryRow(`
		INSERT INTO dag (name) 
		VALUES ($1) 
		RETURNING id`, name).Scan(&id)
	if err != nil {
		do.Logger.Error("Error creating dag:", err)
		return 0, err
	}

	return id, nil
}

func (do *DagOperations) DeleteDag(id int) error {
	result, err := do.DB.Exec(`DELETE FROM dag WHERE id = $1 and status = 'created'`, id)
    rows_affected, err := result.RowsAffected()
	if err != nil {
		do.Logger.Error("Error deleting dag:", err)
		return err
	}
    if rows_affected == 0 {
        return NoRowsAffectedError
    }
	return err
}

func (do *DagOperations) RunDag(id int) error {
	row := do.DB.QueryRow(`
		SELECT id, name, status, created_at 
		FROM dag 
		WHERE id=$1`, id)

	var dag types.Dag
    if err := row.Scan(&dag.ID, &dag.Name, &dag.Status, &dag.CreatedAt); err != nil {
        return err
    }
    
    if dag.Status != "created" {
        return AlreadyInRunningState
    }

    err := controller.RunDag(do.Logger, do.EventCh, id)
    if err != nil {
        return err
    }
    return nil
}


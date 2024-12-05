package database

import (
	"database/sql"

	"github.com/sirupsen/logrus"
)

type Dag struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
    Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
    PendingTasks int `json:"pending_tasks"`
    CompletedTasks int `json:"completed_tasks"`
    ProcessingTasks int `json:"processing_tasks"`
}

type DagWithTasks struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
    Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
    Tasks     []Task `json:"tasks"`
}

type DagOperations struct {
    Logger *logrus.Logger
    DB     *sql.DB
}

func (do *DagOperations) GetDags(page int, perPage int) ([]Dag, int, error) {
    var total_dags int;
    total_query_row := do.DB.QueryRow(`
        SELECT count(*)
        FROM dag
    `)
    total_query_row.Scan(&total_dags)

	rows, err := do.DB.Query(`
		SELECT d.id, d.name, d.status, d.created_at, coalesce(t.pending_tasks, 0) 
		FROM dag as d
        LEFT JOIN (
            SELECT 
                dag_id,
                COUNT(*) OVER(partition by status) as pending_tasks 
            FROM task
            GROUP BY dag_id, status
        ) as t ON t.dag_id=d.id
		ORDER BY d.created_at DESC 
		LIMIT $1 OFFSET $2`, perPage, (page-1)*perPage)
	if err != nil {
        do.Logger.Error(err)
		return nil, 0, err
	}
	defer rows.Close()

	var dags []Dag
	for rows.Next() {
		var dag Dag
		if err := rows.Scan(&dag.ID, &dag.Name, &dag.Status, &dag.CreatedAt, &dag.PendingTasks); err != nil {
			do.Logger.Println("Error scanning dag:", err)
			return nil, 0, err
		}
		dags = append(dags, dag)
	}

	return dags, total_dags, nil
}

func (do *DagOperations) GetDagByID(id int) (*DagWithTasks, error) {
	row := do.DB.QueryRow(`
		SELECT id, name, status, created_at 
		FROM dag 
		WHERE id=$1`, id)

	var dag DagWithTasks
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

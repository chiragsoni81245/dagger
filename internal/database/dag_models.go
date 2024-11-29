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
}

type DagOperations struct {
    Logger *logrus.Logger
    DB     *sql.DB
}

func (do *DagOperations) GetDags(page int, perPage int) ([]Dag, error) {
	rows, err := do.DB.Query(`
		SELECT id, name, status, created_at 
		FROM dag 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`, perPage, (page-1)*perPage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dags []Dag
	for rows.Next() {
		var dag Dag
		if err := rows.Scan(&dag.ID, &dag.Name, &dag.Status, &dag.CreatedAt); err != nil {
			do.Logger.Println("Error scanning dag:", err)
			return nil, err
		}
		dags = append(dags, dag)
	}

	return dags, nil
}

func (do *DagOperations) GetDagByID(id int) (*Dag, error) {
	rows, err := do.DB.Query(`
		SELECT id, name, status, created_at 
		FROM dag 
		WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dag Dag
    if err := rows.Scan(&dag.ID, &dag.Name, &dag.Status, &dag.CreatedAt); err != nil {
        do.Logger.Println("Error scanning dag:", err)
        return nil, err
    }

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
	_, err := do.DB.Exec(`DELETE FROM dag WHERE id = $1`, id)
	return err
}

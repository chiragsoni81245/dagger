package database

import (
	"database/sql"

	"github.com/sirupsen/logrus"
)

type Executor struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Config    string `json:"config"` // Storing as string for JSONB
	CreatedAt string `json:"created_at"`
}

type ExecutorOperations struct {
    Logger *logrus.Logger
    DB     *sql.DB
}

func (eo *ExecutorOperations) GetExecutors(page int, perPage int) ([]Executor, error) {
	rows, err := eo.DB.Query(`
		SELECT id, type, config, created_at
		FROM executor
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, perPage, (page-1)*perPage)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var executors []Executor
	for rows.Next() {
		var executor Executor
		if err := rows.Scan(&executor.ID, &executor.Type, &executor.Config, &executor.CreatedAt); err != nil {
			eo.Logger.Error("Error scanning executor:", err)
			return nil, err
		}
		executors = append(executors, executor)
	}

	return executors, nil
}

func (eo *ExecutorOperations) CreateExecutor(name string, executorType string, config string) (int, error) {
	var id int
	err := eo.DB.QueryRow(`
		INSERT INTO executor (name, type, config)
		VALUES ($1, $2, $3)
		RETURNING id`, name, executorType, config).Scan(&id)
	if err != nil {
		eo.Logger.Error("Error creating executor:", err)
		return 0, err
	}

	return id, nil
}

func (eo *ExecutorOperations) DeleteExecutor(id int) error {
	_, err := eo.DB.Exec(`DELETE FROM executor WHERE id = $1`, id)
	return err
}

func (eo *ExecutorOperations) GetExecutorByID(id int) (*Executor, error) {
	var executor Executor
	err := eo.DB.QueryRow(`
		SELECT id, type, config, created_at
		FROM executor
		WHERE id = $1`, id).Scan(&executor.ID, &executor.Type, &executor.Config, &executor.CreatedAt)

	if err != nil {
		eo.Logger.Error("Error fetching executor by ID:", err)
		return nil, err
	}

	return &executor, nil
}


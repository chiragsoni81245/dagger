package executor

import (
	"database/sql"

	"github.com/chiragsoni81245/dagger/internal/types"
	"github.com/sirupsen/logrus"
)

type Executor interface {
    RunTask(int) <-chan struct{}
}

func ExecuteTask(logger *logrus.Logger, db *sql.DB, executorId int, taskId int) (<-chan struct{}, error){
    row := db.QueryRow(`
        SELECT name, type, config
        FROM executor
        WHERE id=$1
    `, executorId)

    var executor types.Executor
    executor.ID = executorId
    if err := row.Scan(&executor.Name, &executor.Type, &executor.Config); err != nil {
        return nil, err
    }

    var e Executor

    switch (executor.Type) {
    case "docker":
        e = DockerExecutor{
            DB: db,
            Logger: logger,
        }
    }

    c := e.RunTask(taskId)

    return c, nil 
}


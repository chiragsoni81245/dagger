package executor

import (
	"database/sql"
	"time"

	"github.com/chiragsoni81245/dagger/internal/types"
	"github.com/sirupsen/logrus"
)

func ExecuteTask(logger *logrus.Logger, db *sql.DB, executorId int, taskId int) (chan struct{}, error){
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

    c := make(chan struct{})

    go func(){
        time.Sleep(time.Second * 2)
        c <- struct{}{}  
    }()

    return c, nil 
}

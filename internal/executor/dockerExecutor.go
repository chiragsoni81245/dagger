package executor

import (
	"time"
	"database/sql"

	"github.com/sirupsen/logrus"
)

type DockerExecutor struct {
    DB *sql.DB
    Logger *logrus.Logger
}

func (de DockerExecutor) RunTask(taskId int) <-chan struct{} {
    c := make(chan struct{})

    go func(){
        time.Sleep(time.Second * 2)
        c <- struct{}{}  
    }()
    
    return c
}

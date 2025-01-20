package executor

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/chiragsoni81245/dagger/internal/types"
	"github.com/chiragsoni81245/dagger/internal/utils"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

type DockerExecutor struct {
    DB *sql.DB
    Logger *logrus.Logger
}

func (de DockerExecutor) runTask(c chan struct{}, taskId int) {
    var err error
    defer func() {
        if err == nil { return }
        de.Logger.Errorf("[Task %d] %s", taskId, err)
        err := utils.UpdateTaskStatus(de.DB, taskId, "error")
        if err != nil {
            de.Logger.Errorf("[Task %d]: %s", taskId, err)
            return
        }
        return
    }()

    row := de.DB.QueryRow(`
    SELECT id, dag_id, name, status, parent_id, executor_id, type, definition, created_at
    FROM task
    WHERE id=$1
    `, taskId)
    var task types.Task
    if err = row.Scan(&task.ID, &task.DagID, &task.Name, &task.Status, &task.ParentID, &task.ExecutorID, &task.Type, &task.Definition, &task.CreatedAt); err != nil {
        err = errors.New(fmt.Sprintf("Error in fetching task from db: %s\n", err))
        return
    }


    err = utils.UpdateTaskStatus(de.DB, taskId, "running")
    if err != nil {
        return
    }

    // Create a Docker client
	_, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        return
    }
    
    // Unzip submitted code
    sourceCodeZip := fmt.Sprintf("storage/code-files-zip/%d/code.zip", taskId)
    sourceCodePath := fmt.Sprintf("storage/codes/%d/", taskId)
    err = utils.Unzip(sourceCodeZip, sourceCodePath)
    if err != nil {
        return
    }

    var taskDefination struct {
        Dockerfile string `json:"dockerfile"`
    }
    err = json.Unmarshal([]byte(task.Definition), &taskDefination)
    if err != nil {
        return
    }
    dockerfileLocation := fmt.Sprintf("storage/code-files-zip/%d/code/%s", taskId, taskDefination.Dockerfile) 
    de.Logger.Println("==> dockerfile: ", dockerfileLocation)

    c <- struct{}{}  
    close(c)
}

func (de DockerExecutor) RunTask(taskId int) <-chan struct{} {
    c := make(chan struct{})

    go de.runTask(c, taskId)
    
    return c
}

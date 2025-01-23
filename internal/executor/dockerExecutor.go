package executor

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/chiragsoni81245/dagger/internal/types"
	"github.com/chiragsoni81245/dagger/internal/utils"
	"github.com/sirupsen/logrus"

	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type DockerExecutor struct {
    DB *sql.DB
    Logger *logrus.Logger
}

func (de DockerExecutor) runTask(c chan struct{}, taskId int) {
    ctx := context.Background()
    var err error
    var buildResponse dockerTypes.ImageBuildResponse 
    defer func() {
        if err == nil {
            buildResponse.Body.Close()
            return
        }
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
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        return
    }
    
    // Unzip submitted code
    sourceCodeZip := fmt.Sprintf("storage/code-files-zip/%d/code.zip", taskId)

    var taskDefination struct {
        Dockerfile string `json:"dockerfile"`
    }
    err = json.Unmarshal([]byte(task.Definition), &taskDefination)
    if err != nil {
        return
    }

    // Create a tarball of the Dockerfile
    tarBuf := new(bytes.Buffer)
	if err = utils.CreateTarFromZip(sourceCodeZip, tarBuf, map[string]string{
        taskDefination.Dockerfile: "Dockerfile", 
    }); err != nil {
		err = fmt.Errorf("Error creating tarball from zip: %v", err)
        return
	}

    // Build the Docker image
	imageName := fmt.Sprintf("dagger-task-%d", taskId)
	buildResponse, err = cli.ImageBuild(ctx, tarBuf, dockerTypes.ImageBuildOptions{
		Tags: []string{imageName},
	})
	if err != nil {
        return
	}

    logDir := fmt.Sprintf("logs/task-%d", taskId)
    err = os.MkdirAll(logDir, 0755)
    if err != nil {
        return
    }

    imageBuildLogFilePath := fmt.Sprintf("%s/image-build.log", logDir)
	imageBuildLogFile, err := os.Create(imageBuildLogFilePath)
	if err != nil {
		err = fmt.Errorf("Error creating %s file: %s\n", imageBuildLogFilePath, err)
		return
	}

    if _, err = io.Copy(imageBuildLogFile, buildResponse.Body); err != nil {
        return
	}
    imageBuildLogFile.Close()
    de.Logger.Infof("[Task %d] Docker image '%s' built successfully", taskId, imageName)

    c <- struct{}{}  
    close(c)
}

func (de DockerExecutor) RunTask(taskId int) <-chan struct{} {
    c := make(chan struct{})

    go de.runTask(c, taskId)
    
    return c
}

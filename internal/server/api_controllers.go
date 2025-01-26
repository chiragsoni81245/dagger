package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/chiragsoni81245/dagger/internal/models"
	"github.com/chiragsoni81245/dagger/internal/types"
	"github.com/gin-gonic/gin"
)


type APIControllers struct {
    Server *types.Server
}

// --------------------------------------------------------------------------------------------------
// --------------------------- For Dags -------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
func (apiC *APIControllers) GetDags(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    do := models.DagOperations{Logger: logger, DB: db}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	dags, total_dags, err := do.GetDags(page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dags"})
		return
	}

    c.JSON(http.StatusOK, gin.H{"dags": dags, "total_dags": total_dags})
}

func (apiC *APIControllers) GetDagByID(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    do := models.DagOperations{Logger: logger, DB: db}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	dag, err := do.GetDagByID(id)
	if err != nil {
        if err == sql.ErrNoRows {
		    c.JSON(http.StatusNotFound, gin.H{"error": "Invalid ID"})
		    return
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dag": *dag})
}

func (apiC *APIControllers) CreateDag(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    do := models.DagOperations{Logger: logger, DB: db}

	var input struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	id, err := do.CreateDag(input.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dag"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (apiC *APIControllers) DeleteDag(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    do := models.DagOperations{Logger: logger, DB: db}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := do.DeleteDag(id); err != nil {
        if err == models.NoRowsAffectedError {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't delete already running dag"})
            return
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete dag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dag deleted"})
}
func (apiC *APIControllers) RunDag(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    do := models.DagOperations{Logger: logger, DB: db, EventCh: apiC.Server.EventCh}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := do.RunDag(id); err != nil {
        if err == sql.ErrNoRows {
		    c.JSON(http.StatusNotFound, gin.H{"error": "Invalid ID"})
		    return
        }
        if err == models.AlreadyInRunningState {
		    c.JSON(http.StatusNotFound, gin.H{"error": "Dag is already in running state"})
		    return
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run dag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dag started running"})
}



// --------------------------------------------------------------------------------------------------
// --------------------------- For Tasks ------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
func (apiC *APIControllers) GetTasksByDagID(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    to := models.TaskOperations{Logger: logger, DB: db}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	tasks, err := to.GetTasksByDagID(id)
	if err != nil {
        if err == sql.ErrNoRows {
		    c.JSON(http.StatusNotFound, gin.H{"error": "Invalid ID"})
		    return
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (apiC *APIControllers) GetTaskByID(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    to := models.TaskOperations{Logger: logger, DB: db}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	task, err := to.GetTaskByID(id)
	if err != nil {
        if err == sql.ErrNoRows {
		    c.JSON(http.StatusNotFound, gin.H{"error": "Invalid ID"})
		    return
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": *task})
}

func (apiC *APIControllers) CreateTask(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    to := models.TaskOperations{Logger: logger, DB: db}

	var input struct {
        DagId int `json:"dag_id"`
        ExecutorID int `json:"executor_id"`
        ParentID *int `json:"parent_id"`
        Name string `json:"name"`
        Type string `json:"type"`
        DockerfilePath string `json:"dockerfile-path"`
	}

    var err error;
	input.DagId, err = strconv.Atoi(c.PostForm("dag_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dag id"})
		return
	}
	input.ExecutorID, err = strconv.Atoi(c.PostForm("executor_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid executor id"})
		return
	}
    input.Type = c.PostForm("type")
    input.Name = c.PostForm("name")
    input.DockerfilePath = c.PostForm("dockerfile-path")
    parentId := c.PostForm("parent_id")
    if parentId != "null" {
        parentIdInt, err := strconv.Atoi(parentId)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parent id"})
            return
        }
        input.ParentID = &parentIdInt
    }

    codeFilesZip, err := c.FormFile("code_files_zip")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input, code files zip missing"})
		return
	}
    
    var definition string = "{}";
    if input.DockerfilePath != "" {
        definition = fmt.Sprintf("{\"dockerfile\": \"%s\"}", input.DockerfilePath)
    } 

	id, txn, err := to.CreateTask(input.DagId, input.ExecutorID, input.Name, input.Type, definition, input.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

    // Validate File Type
    fileExt := filepath.Ext(codeFilesZip.Filename)
    if fileExt != ".zip" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid code files extension, we only support zip"})
		return
    }

    // Store Code Zip file in that task's specific directory
    taskDir := fmt.Sprintf("storage/code-files-zip/%d", id)
    err = os.MkdirAll(taskDir, 0755)
    if err != nil {
        err = txn.Rollback()
        if err != nil {
            logger.Error(err)
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
    }
    err = c.SaveUploadedFile(codeFilesZip, taskDir+"/code.zip")
    if err != nil {
        if txnErr := txn.Rollback(); txnErr != nil {
            logger.Error(txnErr)
            return
        }
        logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
    }

    err = txn.Commit()
    if err != nil {
        logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
    }
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (apiC *APIControllers) DeleteTask(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    to := models.TaskOperations{Logger: logger, DB: db}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

    txn, err := to.DeleteTask(id)
	if err != nil {
        if err == models.NoRowsAffectedError {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't delete already running dag or task"})
            return
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

    taskDir := fmt.Sprintf("storage/code-files-zip/%d", id)
    if err = os.RemoveAll(taskDir); err != nil {
        if txnErr := txn.Rollback(); txnErr != nil {
            logger.Error(err)
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
    }

    err = txn.Commit()
    if err != nil {
        logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
    }
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}

func (apiC *APIControllers) GetTaskLogs(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    to := models.TaskOperations{Logger: logger, DB: db}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

    taskLogs, err := to.GetTaskLogs(id)
	if err != nil {
        if err == models.TaskIsStillRunning || err == models.TaskNotStarted {
		    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task logs"})
		return
	}

    c.JSON(http.StatusOK, gin.H{"logs": taskLogs})
}

func (apiC *APIControllers) GetTaskLogByName(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
    logName := c.Param("name")

    logFilePath := fmt.Sprintf("logs/task-%d/%s.log", id, logName)
    c.File(logFilePath)
}


// --------------------------------------------------------------------------------------------------
// --------------------------- For Executors --------------------------------------------------------
// --------------------------------------------------------------------------------------------------
func (apiC *APIControllers) GetExecutors(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    eo := models.ExecutorOperations{Logger: logger, DB: db}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	executors, total_executors, err := eo.GetExecutors(page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get executors"})
		return
	}

    c.JSON(http.StatusOK, gin.H{"executors": executors, "total_executors": total_executors})
}

func (apiC *APIControllers) GetExecutorByID(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    to := models.ExecutorOperations{Logger: logger, DB: db}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
        if err == sql.ErrNoRows {
		    c.JSON(http.StatusNotFound, gin.H{"error": "Invalid ID"})
		    return
        }
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	executor, err := to.GetExecutorByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"executor": *executor})
}

func (apiC *APIControllers) CreateExecutor(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    to := models.ExecutorOperations{Logger: logger, DB: db}

	var input struct {
		Name string `json:"name"`
        Type string `json:"type"`
        Config string `json:"-"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

    if input.Type == "docker" {
        input.Config = "{}"
    } else {
        input.Config = "{}"
        // To-Do here in feature we can add logic for other fields which can be added into this config to setup those executors
    }

	id, err := to.CreateExecutor(input.Name, input.Type, input.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dag"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (apiC *APIControllers) DeleteExecutor(c *gin.Context) {
    logger := apiC.Server.Logger
    db := apiC.Server.DB
    to := models.ExecutorOperations{Logger: logger, DB: db}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := to.DeleteExecutor(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete dag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Executor deleted"})
}

package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/chiragsoni81245/dagger/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)


// --------------------------------------------------------------------------------------------------
// --------------------------- For Dags -------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
func GetDags(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    db := c.MustGet("db").(*sql.DB)
    do := database.DagOperations{Logger: logger, DB: db}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	dags, err := do.GetDags(page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dags": dags})
}

func GetDagByID(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    db := c.MustGet("db").(*sql.DB)
    do := database.DagOperations{Logger: logger, DB: db}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	dag, err := do.GetDagByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dag": *dag})
}

func CreateDag(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    db := c.MustGet("db").(*sql.DB)
    do := database.DagOperations{Logger: logger, DB: db}

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

func DeleteDag(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    db := c.MustGet("db").(*sql.DB)
    do := database.DagOperations{Logger: logger, DB: db}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := do.DeleteDag(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete dag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Dag deleted"})
}


// --------------------------------------------------------------------------------------------------
// --------------------------- For Tasks ------------------------------------------------------------
// --------------------------------------------------------------------------------------------------
func GetTasks(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    db := c.MustGet("db").(*sql.DB)
    to := database.TaskOperations{Logger: logger, DB: db}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	dags, err := to.GetTasks(page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": dags})
}

func GetTaskByID(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    db := c.MustGet("db").(*sql.DB)
    to := database.TaskOperations{Logger: logger, DB: db}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	dag, err := to.GetTaskByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": *dag})
}

func CreateTask(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    db := c.MustGet("db").(*sql.DB)
    to := database.TaskOperations{Logger: logger, DB: db}

	var input struct {
        DagId int `json:"dag_id"`
        ExecutorID int `json:"executor_id"`
        Type string `json:"type"`
        Definition string `json:"definition"`
        ParentID int `json:"parent_id"`
	}

    // Validate Definition
    var goTaskDefinition struct {
        Version string `json:"version"`
        Command string `json:"command"`
    }
    var pythonTaskDefinition struct {
        Version string `json:"version"`
        Command string `json:"command"`
    }
    var dockerTaskDefinition struct {
        Command string `json:"command"`
    }

    if err := json.Unmarshal([]byte(input.Definition), &goTaskDefinition); err != nil {
        if err := json.Unmarshal([]byte(input.Definition), &pythonTaskDefinition); err != nil {
            if err := json.Unmarshal([]byte(input.Definition), &dockerTaskDefinition); err != nil {
		        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task definition"})
                return 
            }
        }
    }

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	id, txn, err := to.CreateTask(input.DagId, input.ExecutorID, input.Type, input.Definition, input.ParentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

    codeZipFile, _ := c.FormFile("code_zip_file")

    // Validate File Type
    fileExt := filepath.Ext(codeZipFile.Filename)
    if fileExt != "zip" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid code files extension, we only support zip"})
		return
    }

    // Store Code Zip file in that task's specific directory
    taskDir := fmt.Sprintf("storage/code-zip-files/%d", id)
    err = os.MkdirAll(taskDir, 0755)
    if err != nil {
        err = txn.Rollback()
        if err != nil {
            logger.Error(err)
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
    }
    err = c.SaveUploadedFile(codeZipFile, taskDir+"/code.tar")
    if err != nil {
        err = txn.Rollback()
        if err != nil {
            logger.Error(err)
        }
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
    }

    txn.Commit()
    if err != nil {
        logger.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
    }
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func DeleteTask(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    db := c.MustGet("db").(*sql.DB)
    to := database.TaskOperations{Logger: logger, DB: db}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
    taskDir := fmt.Sprintf("storage/code-zip-files/%d", id)
    if err = os.RemoveAll(taskDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
    }

	if err = to.DeleteTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}


// --------------------------------------------------------------------------------------------------
// --------------------------- For Executors --------------------------------------------------------
// --------------------------------------------------------------------------------------------------
func GetExecutors(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    db := c.MustGet("db").(*sql.DB)
    to := database.ExecutorOperations{Logger: logger, DB: db}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	dags, err := to.GetExecutors(page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dags": dags})
}

func GetExecutorByID(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    db := c.MustGet("db").(*sql.DB)
    to := database.ExecutorOperations{Logger: logger, DB: db}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	dag, err := to.GetExecutorByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dag"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"dag": *dag})
}

func CreateExecutor(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    db := c.MustGet("db").(*sql.DB)
    to := database.ExecutorOperations{Logger: logger, DB: db}

	var input struct {
		Name string `json:"name"`
        Type string `json:"type"`
        Config string `json:"config"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	id, err := to.CreateExecutor(input.Name, input.Type, input.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create dag"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func DeleteExecutor(c *gin.Context) {
    logger := c.MustGet("logger").(*logrus.Logger)
    db := c.MustGet("db").(*sql.DB)
    to := database.ExecutorOperations{Logger: logger, DB: db}

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

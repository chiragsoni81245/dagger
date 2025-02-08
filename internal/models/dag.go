package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/chiragsoni81245/dagger/internal/controller"
	"github.com/chiragsoni81245/dagger/internal/queue"
	"github.com/chiragsoni81245/dagger/internal/types"
	"github.com/sirupsen/logrus"
)

type DagOperations struct {
    Logger *logrus.Logger
    DB     *sql.DB
    EventCh chan types.Event
}

func (do *DagOperations) GetDags(page int, perPage int) ([]types.Dag, int, error) {
    var total_dags int;
    total_query_row := do.DB.QueryRow(`
        SELECT count(*)
        FROM dag
    `)
    total_query_row.Scan(&total_dags)

	rows, err := do.DB.Query(`
    WITH task_counts AS (
    	SELECT 
    		dag_id,
    		status,
    		COUNT(*) as count 
    	FROM task
    	GROUP BY status, dag_id
    )
    SELECT 
    		d.id, 
    		d.name, 
    		d.status, 
    		d.created_at, 
    		coalesce(tc_total.count, 0) - coalesce(tc_completed.count, 0) as pending_tasks, 
    		coalesce(tc_running.count, 0) as running_tasks, 
    		coalesce(tc_completed.count, 0) as completed_tasks 
    FROM dag as d
    LEFT JOIN (
    	SELECT dag_id, count(*)
    	FROM task
    	GROUP BY dag_id
    ) as tc_total ON tc_total.dag_id=d.id
    LEFT JOIN task_counts as tc_running ON tc_running.dag_id=d.id and tc_running.status='running'
    LEFT JOIN task_counts as tc_completed ON tc_completed.dag_id=d.id and tc_completed.status='completed'
    ORDER BY d.created_at DESC 
    LIMIT $1 OFFSET $2`, perPage, (page-1)*perPage)
	if err != nil {
        do.Logger.Error(err)
		return nil, 0, err
	}
	defer rows.Close()

	var dags []types.Dag
	for rows.Next() {
		var dag types.Dag
		if err := rows.Scan(&dag.ID, &dag.Name, &dag.Status, &dag.CreatedAt, &dag.PendingTasks, &dag.RunningTasks, &dag.CompletedTasks); err != nil {
			do.Logger.Println("Error scanning dag: ", err)
			return nil, 0, err
		}
		dags = append(dags, dag)
	}

	return dags, total_dags, nil
}

func (do *DagOperations) GetDagByID(id int) (*types.DagWithTasks, error) {
	row := do.DB.QueryRow(`
		SELECT id, name, status, created_at 
		FROM dag 
		WHERE id=$1`, id)

	var dag types.DagWithTasks
    if err := row.Scan(&dag.ID, &dag.Name, &dag.Status, &dag.CreatedAt); err != nil {
        do.Logger.Println("Error scanning dag: ", err)
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
		do.Logger.Error("Error creating dag: ", err)
		return 0, err
	}

	return id, nil
}

func (do *DagOperations) DeleteDag(id int) error {
	result, err := do.DB.Exec(`DELETE FROM dag WHERE id = $1 and status in ('created','completed')`, id)
    rows_affected, err := result.RowsAffected()
	if err != nil {
		do.Logger.Error("Error deleting dag: ", err)
		return err
	}
    if rows_affected == 0 {
        return NoRowsAffectedError
    }
	return err
}

func (do *DagOperations) RunDag(id int) error {
	row := do.DB.QueryRow(`
		SELECT id, name, status, created_at 
		FROM dag 
		WHERE id=$1`, id)

	var dag types.Dag
    if err := row.Scan(&dag.ID, &dag.Name, &dag.Status, &dag.CreatedAt); err != nil {
        return err
    }
    
    if dag.Status != "created" {
        return AlreadyInRunningState
    }

    err := controller.RunDag(do.Logger, do.EventCh, id)
    if err != nil {
        return err
    }
    return nil
}

func (do *DagOperations) CreateDagWithYAML(dag *types.DagNode, files map[string]*multipart.FileHeader) (dagId int, err error) {
    // Parse YAML file content
    txn, err := do.DB.Begin()
    if err != nil {
        return -1, err
    }

    defer func () {
        if err != nil {
            do.Logger.Error(err)
            secondaryErr := txn.Rollback()
            if secondaryErr != nil {
                do.Logger.Error(secondaryErr)
                err = secondaryErr
            }
        }
    }()

    eo := ExecutorOperations{Logger: do.Logger, DB: do.DB}
    executors, _, err := eo.GetExecutors(1, -1) // We want to fetch all executors so here page =1, and perPage = -1 which is considered infinity
    if err != nil {
        do.Logger.Errorf("Error fetching executor details from db: %v", err)
        return -1, err
    }
    executorIds := make(map[string]int)
    for _, executor := range executors {
        executorIds[executor.Name] = executor.ID
    }

    // Creating Dag
    err = txn.QueryRow(`
    INSERT INTO dag (name) 
    VALUES ($1) 
    RETURNING id`, dag.Name).Scan(&dagId)
    if err != nil {
        err = fmt.Errorf("Error creating dag: %v", err)
        return -1, err
    }

    if len(dag.Tasks) == 0 {
        err = txn.Commit()
        if err != nil {
            return -1, err
        }
        return dagId, nil
    }

    // Creating Tasks
    type qElement struct {
        Node types.TaskNode
        ParentId *int
    }

    q := &queue.Queue[qElement]{}
    q.Enqueue(qElement{Node: dag.Tasks[0], ParentId: nil})

    for !q.IsEmpty() {
        ele, _ := q.Dequeue()
        task := ele.Node
        parentId := ele.ParentId
        var taskId int

        taskDefinitionJson, err := json.Marshal(task.Definition)
        if err != nil {
            err = fmt.Errorf("Error in task definition json parsing: %v", err)
            return -1, err
        }

        // Validate File Type
        fileExt := filepath.Ext(task.CodeZipFileName)
        if fileExt != ".zip" {
            err = fmt.Errorf("Invalid file extension %s for task %s", fileExt, task.Name)
            return -1, err
        }

        // Insert Task into DB
        err = txn.QueryRow(`
        INSERT INTO task (dag_id, name, type, executor_id, definition, parent_id) 
        VALUES ($1, $2, $3, $4, $5, $6) 
        RETURNING id`, dagId, task.Name, task.Type, executorIds[task.ExecutorName], taskDefinitionJson, parentId).Scan(&taskId)
        if err != nil {
            err = fmt.Errorf("Error creating task: %v", err)
            return -1, err
        }

        // Store Code Zip file in that task's specific directory
        taskDir := fmt.Sprintf("storage/code-files-zip/%d", taskId)
        err = os.MkdirAll(taskDir, 0755)
        if err != nil {
            err = fmt.Errorf("Error in creating code file directory for task: %v", err)
            return -1, err
        }

        codeZipFile, err := files[task.CodeZipFileName].Open()
        if err != nil {
            return -1, err
        }

        // Read file content
	    zipFile, err := os.OpenFile(taskDir+"/code.zip", os.O_RDWR|os.O_CREATE, 0644)
        if err != nil {
            err = fmt.Errorf("Error creating new zip file for task %s: %v", task.Name, err)
            return -1, err
        }
        _, err = io.Copy(zipFile, codeZipFile)
        if err != nil {
            err = fmt.Errorf("Error saving task code zip file: %v", err)
            return -1, err
        }

        for _, child := range task.Childs {
            q.Enqueue(qElement{Node: child, ParentId: &taskId})
        }
    }

    err = txn.Commit()
    if err != nil {
        return -1, err
    }

    return dagId, nil
}

func (do *DagOperations) ValidateDagYAML(dag *types.DagNode) types.DagYAMLValidationResponse {
    // Parse YAML file content
    var err error
    internalServerError := types.DagYAMLValidationResponse{
        IsValid: false,
        Error: "Something went wrong",
        RequiredFiles: []string{},
    }

    if len(dag.Name) == 0 {
        return types.DagYAMLValidationResponse{
            IsValid: false,
            Error: "Invalid dag name",
            RequiredFiles: []string{},
        }

    }

    eo := ExecutorOperations{Logger: do.Logger, DB: do.DB}
    executors, _, err := eo.GetExecutors(1, -1) // We want to fetch all executors so here page =1, and perPage = -1 which is considered infinity
    if err != nil {
        do.Logger.Errorf("Error fetching executor details from db: %v", err)
        return internalServerError
    }
    executorIds := make(map[string]int)
    for _, executor := range executors {
        executorIds[executor.Name] = executor.ID
    }

    var existingDagId int
    err = do.DB.QueryRow(`
        SELECT id FROM dag WHERE name=$1
    `, dag.Name).Scan(&existingDagId)
    if err == nil {
        return types.DagYAMLValidationResponse{
            IsValid: false,
            Error: fmt.Sprintf("Dag with name %s already exists", dag.Name),
            RequiredFiles: []string{},
        }
    } else if err == sql.ErrNoRows {
        err = nil
    } else {
        return internalServerError
    }

    requiredFiles := make(map[string]struct{})
    if len(dag.Tasks) == 0 {
        keys := make([]string, 0, len(requiredFiles))
        for k := range requiredFiles {
            keys = append(keys, k)
        }
        
        return types.DagYAMLValidationResponse{
            IsValid: true,
            RequiredFiles: keys,
            Error: "",
        }
    }

    // Creating Tasks
    type qElement struct {
        Node types.TaskNode
        ParentId int
    }

    q := &queue.Queue[qElement]{}
    q.Enqueue(qElement{Node: dag.Tasks[0]})

    for !q.IsEmpty() {
        ele, _ := q.Dequeue()
        task := ele.Node
        requiredFiles[task.CodeZipFileName] = struct{}{}
        if task.Type == "docker" {
            _, ok := task.Definition["dockerfile"];
            if !ok {
                err = fmt.Errorf("Missing dockerfile in task %s definition", task.Name) 
                break
            }
            if _, ok := executorIds[task.ExecutorName]; !ok {
                err = fmt.Errorf("Invalid executor name %s in task %s", task.ExecutorName, task.Name) 
                break
            }
        }

        for _, child := range task.Childs {
            q.Enqueue(qElement{Node: child})
        }
    }

    if err != nil {
        return types.DagYAMLValidationResponse{
            IsValid: false,
            RequiredFiles: []string{},
            Error: err.Error(),
        }
    }

    keys := make([]string, 0, len(requiredFiles))
    for k := range requiredFiles {
        keys = append(keys, k)
    }

    return types.DagYAMLValidationResponse{
        IsValid: true,
        RequiredFiles: keys,
        Error: "",
    }
}


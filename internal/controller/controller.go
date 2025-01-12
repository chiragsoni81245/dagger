package controller

import (
	"fmt"
	"time"

	"github.com/chiragsoni81245/dagger/internal/config"
	"github.com/chiragsoni81245/dagger/internal/database"
	"github.com/chiragsoni81245/dagger/internal/executor"
	"github.com/chiragsoni81245/dagger/internal/queue"
	"github.com/chiragsoni81245/dagger/internal/types"
	"github.com/sirupsen/logrus"
)


func RunDag(logger *logrus.Logger, id int) error{
    var err error

    // Generate application configuration
    config, err := config.GetConfig()
    if err != nil {
        logger.Error(err)
        return err
    }

    db, err := database.GetDB(config) 
    if err != nil {
        logger.Error(err)
        return err
    }

    // Move dag into running state
    _, err = db.Exec(`
        UPDATE dag
        SET status='running'
        WHERE id=$1
    `, id)
    if err != nil {
        logger.Error(err)
        return err
    }

    // Execute the actual running logic in background
    go func() {
        var tasksIdMap map[int]types.Task = make(map[int]types.Task)
        var tasksGraph map[int][]int = make(map[int][]int)
        var rootTask types.Task

        rows, err := db.Query(`
            SELECT id, dag_id, name, status, parent_id, executor_id, type, definition, created_at
            FROM task
            WHERE dag_id=$1
            ORDER BY parent_id ASC
        `, id)
        if err != nil {
            logger.Error(fmt.Sprintf("[Dag %d] Error fetching tasks: %s", id, err))
            return
        }
        defer rows.Close()

        for rows.Next() {
            var task types.Task
            if err := rows.Scan(&task.ID, &task.DagID, &task.Name, &task.Status, &task.ParentID, &task.ExecutorID, &task.Type, &task.Definition, &task.CreatedAt); err != nil {
                logger.Error("Error scanning task:", err)
                return
            }
            tasksIdMap[task.ID] = task
            if task.ParentID == nil {
                rootTask = task
                if _, ok := tasksGraph[task.ID]; !ok {
                    tasksGraph[task.ID] = []int{}
                }
            } else {
                if _, ok := tasksGraph[*task.ParentID]; ok {
                    tasksGraph[*task.ParentID] = append(tasksGraph[*task.ParentID], task.ID)
                } else {
                    tasksGraph[*task.ParentID] = []int{task.ID}
                }
            }
        }


        fmt.Println(tasksGraph)

        // BFS Traversal in task graph
        runningTasks := make(map[int]chan struct{})
        q := &queue.Queue[int]{}
        q.Enqueue(rootTask.ID)
        
        for ; !q.IsEmpty() || len(runningTasks) != 0 ; {
            currentTaskId, ok := q.Dequeue()

            if ok {
                // Start Executing this task, in its specified executor
                task := tasksIdMap[currentTaskId]
                c, err := executor.ExecuteTask(logger, db, task.ExecutorID, currentTaskId)
                if err != nil {
                    logger.Error(err)
                    return
                }
                runningTasks[currentTaskId] = c 
                fmt.Printf("Started execution for task: %d\n", currentTaskId)
            } else {
                for {
                    completedTaskId := -1 
                    for taskId, c := range runningTasks {
                        select {
                        case <-c:
                            completedTaskId = taskId
                            break
                        default:
                        }
                    }

                    if completedTaskId != -1 {
                        // find all the child tasks of this task, and get them running
                        for _, id := range tasksGraph[completedTaskId] {
                            q.Enqueue(id)
                        }
                        break
                    }

                    time.Sleep(time.Microsecond * 10)
                }
            }
        }

    }()

    return nil
}

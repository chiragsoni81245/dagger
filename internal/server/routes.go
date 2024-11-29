package server

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
    // UI Routes
    {
        r.GET("/", Dashboard)
        r.GET("/dags", Dags)
        r.GET("/executors", Executors)
    }

	// Group routes under /api/v1
	v1 := r.Group("/api/v1")
	{
		// DAG routes
		v1.GET("/dags", GetDags)
		v1.POST("/dags", CreateDag)
		v1.GET("/dags/:id", GetDagByID)
		v1.DELETE("/dags/:id", DeleteDag)

        // Task routes
        v1.GET("/tasks", GetTasks)
        v1.POST("/tasks", CreateTask)
        v1.GET("/tasks/:id", GetTaskByID)
        v1.DELETE("/tasks/:id", DeleteTask)

        // Executor routes
        v1.GET("/executor", GetExecutors)
        v1.POST("/executor", CreateExecutor)
        v1.GET("/executor/:id", GetExecutorByID)
        v1.DELETE("/executor/:id", DeleteExecutor)
	}
}


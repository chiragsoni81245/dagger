package server

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
    // UI Routes
    ui := UIControllers{}
    {
        r.GET("/", ui.Dashboard)
        r.GET("/dags", ui.Dags)
        r.GET("/dags/:id", ui.Dag)
        r.GET("/executors", ui.Executors)
    }

	// Group routes under /api/v1
	v1 := r.Group("/api/v1")
    api := APIControllers{}
	{
		// DAG routes
		v1.GET("/dags", api.GetDags)
		v1.POST("/dags", api.CreateDag)
		v1.GET("/dags/:id", api.GetDagByID)
        v1.GET("/dags/:id/tasks", api.GetTasksByDagID)
		v1.DELETE("/dags/:id", api.DeleteDag)

        // Task routes
        v1.POST("/tasks", api.CreateTask)
        v1.GET("/tasks/:id", api.GetTaskByID)
        v1.DELETE("/tasks/:id", api.DeleteTask)

        // Executor routes
        v1.GET("/executor", api.GetExecutors)
        v1.POST("/executor", api.CreateExecutor)
        v1.GET("/executor/:id", api.GetExecutorByID)
        v1.DELETE("/executor/:id", api.DeleteExecutor)
	}
}


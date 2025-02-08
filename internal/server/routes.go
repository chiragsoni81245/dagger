package server

import (
	"github.com/chiragsoni81245/dagger/internal/types"
)

func SetupRoutes(server *types.Server) {
	// UI Routes
	ui := UIControllers{Server: server}
	{
		server.Router.GET("/", ui.Dashboard)
		server.Router.GET("/dags", ui.Dags)
		server.Router.GET("/dags/:id", ui.Dag)
		server.Router.GET("/executors", ui.Executors)
	}

	// Web Socket Route
	ws := WebSocketControllers{Server: server}
	{
		server.Router.GET("/ws", ws.HandleWebSocket)
	}

	// Group routes under /api/v1
	v1 := server.Router.Group("/api/v1")
	api := APIControllers{Server: server}
	{
		// DAG routes
		v1.GET("/dags", api.GetDags)
		v1.POST("/dags", api.CreateDag)
		v1.POST("/dags/yaml", api.CreateDagWithYAML)
		v1.POST("/dags/yaml/validate", api.ValidateDagYAML)
		v1.GET("/dags/:id", api.GetDagByID)
		v1.GET("/dags/:id/tasks", api.GetTasksByDagID)
		v1.DELETE("/dags/:id", api.DeleteDag)
		v1.POST("/dags/:id/run", api.RunDag)

		// Task routes
		v1.POST("/tasks", api.CreateTask)
		v1.GET("/tasks/:id", api.GetTaskByID)
		v1.DELETE("/tasks/:id", api.DeleteTask)
		v1.GET("/tasks/:id/logs", api.GetTaskLogs)
        v1.GET("/tasks/:id/logs/:name", api.GetTaskLogByName)

		// Executor routes
		v1.GET("/executors", api.GetExecutors)
		v1.POST("/executors", api.CreateExecutor)
		v1.GET("/executors/:id", api.GetExecutorByID)
		v1.DELETE("/executors/:id", api.DeleteExecutor)
	}
}

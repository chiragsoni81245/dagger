package types

type Task struct {
	ID          int    `json:"id"`
	DagID       int    `json:"dag_id"`
    Name        string `json:"name"`
	Status      string `json:"status"`
	ParentID    *int   `json:"parent_id"` // Nullable
	ExecutorID  int    `json:"executor_id"`
	Type        string `json:"type"`
	Definition  string
	CreatedAt   string `json:"created_at"`
}

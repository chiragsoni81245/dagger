package types

type Dag struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	CreatedAt      string `json:"created_at"`
	PendingTasks   int    `json:"pending_tasks"`
	CompletedTasks int    `json:"completed_tasks"`
	RunningTasks   int    `json:"running_tasks"`
}

type DagWithTasks struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	Tasks     []Task `json:"tasks"`
}

type DagYAMLValidationResponse struct {
	IsValid      bool     `json:"isValid"`
	RequiredFiles []string `json:"requiredFiles"`
	Error        string   `json:"error"`
}

type DagNode struct {
	Name  string     `yaml:"name"`
	Tasks []TaskNode `yaml:"tasks"`
}

type DagNode_ struct {
	Name  string     `yaml:"name"`
	Tasks []TaskNode `yaml:"tasks"`
}

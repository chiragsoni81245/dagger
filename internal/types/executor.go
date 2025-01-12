package types

type Executor struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
    Name      string `json:"name"`
	Config    string
	CreatedAt string `json:"created_at"`
}

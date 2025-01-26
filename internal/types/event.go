package types

type EventDescriptor struct {
    Resource string
    ID       int
}

type EventResource struct {
    Resource string
    ID       int
}

type Event struct {
    Resource       string `json:"resource"`
    ID             int    `json:"id"` 
    ParentResource *EventResource `json:"-"`
    Field          string `json:"field"`
    NewValue       string `json:"newValue"`
}

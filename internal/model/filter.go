package model

type TaskFilter struct {
	Status    string
	ContextID *int
	Overdue   bool
	Limit     int
	Cursor    *int
}

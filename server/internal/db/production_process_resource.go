package db

// ProductionProcessResource represents a resource input/output for a process.
type ProductionProcessResource struct {
	ProcessID  int64
	ResourceID int64
	Direction  string
	Quantity   int64
}

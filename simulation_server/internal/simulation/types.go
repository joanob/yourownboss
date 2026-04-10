package simulation

type SimulationResourceRange struct {
	ResourceID int64   `json:"resource_id"`
	MinPrice   float64 `json:"min_price"`
	MaxPrice   float64 `json:"max_price"`
	Step       float64 `json:"step"`
	IsOutput   bool    `json:"is_output,omitempty"`
}

type SimulationRequest struct {
	ProcessID      int64                     `json:"process_id"`
	TimeMinMs      int                       `json:"time_min_ms"`
	TimeMaxMs      int                       `json:"time_max_ms"`
	TimeStepMs     int                       `json:"time_step_ms"`
	ResourceRanges []SimulationResourceRange `json:"resource_ranges"`
}

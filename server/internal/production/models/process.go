package models

// ProductionProcessDBO represents a production process as stored in the database
type ProductionProcessDBO struct {
	ID                   string `db:"id"`
	MasterID             string `db:"master_id"`
	ProductionBuildingID string `db:"production_building_id"`
	Name                 string `db:"name"`
	CycleTimeSec         int64  `db:"cycle_time_s"`
	WindowStartHour      *int64 `db:"window_start_hour"` // nullable - segundos desde medianoche UTC
	WindowEndHour        *int64 `db:"window_end_hour"`   // nullable - segundos desde medianoche UTC
}

// ProductionProcessResourceDBO represents a resource in a production process
type ProductionProcessResourceDBO struct {
	ProcessID  string `db:"process_id"`
	ResourceID string `db:"resource_id"`
	IsOutput   int64  `db:"is_output"` // 1 = output, 0 = input
	Quantity   int64  `db:"quantity"`
}

// ProductionProcess represents a production process in the domain
type ProductionProcess struct {
	ID                   string                     `json:"id"`
	MasterID             string                     `json:"master_id"`
	ProductionBuildingID string                     `json:"production_building_id"`
	Name                 string                     `json:"name"`
	CycleTimeSec         int64                      `json:"cycle_time_s"`
	WindowStartHour      *int64                     `json:"window_start_hour"`
	WindowEndHour        *int64                     `json:"window_end_hour"`
	InputResources       []*ProductionProcessInput  `json:"input_resources"`
	OutputResources      []*ProductionProcessOutput `json:"output_resources"`
}

// ProductionProcessInput represents input resources for a process
type ProductionProcessInput struct {
	ResourceID string `json:"resource_id"`
	Quantity   int64  `json:"quantity"`
}

// ProductionProcessOutput represents output resources for a process
type ProductionProcessOutput struct {
	ResourceID string `json:"resource_id"`
	Quantity   int64  `json:"quantity"`
}

// NewProductionProcess creates a new production process from a DBO
func NewProductionProcess(dbo *ProductionProcessDBO) *ProductionProcess {
	if dbo == nil {
		return nil
	}
	return &ProductionProcess{
		ID:                   dbo.ID,
		MasterID:             dbo.MasterID,
		ProductionBuildingID: dbo.ProductionBuildingID,
		Name:                 dbo.Name,
		CycleTimeSec:         dbo.CycleTimeSec,
		WindowStartHour:      dbo.WindowStartHour,
		WindowEndHour:        dbo.WindowEndHour,
		InputResources:       make([]*ProductionProcessInput, 0),
		OutputResources:      make([]*ProductionProcessOutput, 0),
	}
}

// ToProductionProcessDBO converts a ProductionProcess to a DBO
func (pp *ProductionProcess) ToProductionProcessDBO() *ProductionProcessDBO {
	return &ProductionProcessDBO{
		ID:                   pp.ID,
		MasterID:             pp.MasterID,
		ProductionBuildingID: pp.ProductionBuildingID,
		Name:                 pp.Name,
		CycleTimeSec:         pp.CycleTimeSec,
		WindowStartHour:      pp.WindowStartHour,
		WindowEndHour:        pp.WindowEndHour,
	}
}

// AddInputResource adds an input resource to the process
func (pp *ProductionProcess) AddInputResource(resourceID string, quantity int64) {
	if pp.InputResources == nil {
		pp.InputResources = make([]*ProductionProcessInput, 0)
	}
	pp.InputResources = append(pp.InputResources, &ProductionProcessInput{
		ResourceID: resourceID,
		Quantity:   quantity,
	})
}

// AddOutputResource adds an output resource to the process
func (pp *ProductionProcess) AddOutputResource(resourceID string, quantity int64) {
	if pp.OutputResources == nil {
		pp.OutputResources = make([]*ProductionProcessOutput, 0)
	}
	pp.OutputResources = append(pp.OutputResources, &ProductionProcessOutput{
		ResourceID: resourceID,
		Quantity:   quantity,
	})
}

// IsAvailableNow checks if the process is available right now (for window constraints)
// Returns true if no window is defined or if current UTC time is within the window
func (pp *ProductionProcess) IsAvailableNow(nowUTCSeconds int64) bool {
	// If no window is defined, process is always available
	if pp.WindowStartHour == nil || pp.WindowEndHour == nil {
		return true
	}

	// Get current UTC time within a day (0-86400 seconds)
	secondsInDay := int64(86400)
	currentTimeOfDay := nowUTCSeconds % secondsInDay

	// Check if within window
	return currentTimeOfDay >= *pp.WindowStartHour && currentTimeOfDay <= *pp.WindowEndHour
}

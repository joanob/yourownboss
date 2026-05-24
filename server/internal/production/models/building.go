package models

// ProductionBuildingDBO represents a production building as stored in the database
type ProductionBuildingDBO struct {
	ID                  string `db:"id"`
	MasterID            string `db:"master_id"`
	Name                string `db:"name"`
	ConstructionCost    int64  `db:"construction_cost"`
	ConstructionTimeSec int64  `db:"construction_time_s"`
}

// ProductionBuilding represents a production building in the domain
type ProductionBuilding struct {
	ID                  string               `json:"id"`
	MasterID            string               `json:"master_id"`
	Name                string               `json:"name"`
	ConstructionCost    int64                `json:"construction_cost"`
	ConstructionTimeSec int64                `json:"construction_time_s"`
	Processes           []*ProductionProcess `json:"processes"`
}

// NewProductionBuilding creates a new production building from a DBO
func NewProductionBuilding(dbo *ProductionBuildingDBO) *ProductionBuilding {
	if dbo == nil {
		return nil
	}
	return &ProductionBuilding{
		ID:                  dbo.ID,
		MasterID:            dbo.MasterID,
		Name:                dbo.Name,
		ConstructionCost:    dbo.ConstructionCost,
		ConstructionTimeSec: dbo.ConstructionTimeSec,
		Processes:           make([]*ProductionProcess, 0),
	}
}

// ToProductionBuildingDBO converts a ProductionBuilding to a DBO
func (pb *ProductionBuilding) ToProductionBuildingDBO() *ProductionBuildingDBO {
	return &ProductionBuildingDBO{
		ID:                  pb.ID,
		MasterID:            pb.MasterID,
		Name:                pb.Name,
		ConstructionCost:    pb.ConstructionCost,
		ConstructionTimeSec: pb.ConstructionTimeSec,
	}
}

// AddProcess adds a process to the building
func (pb *ProductionBuilding) AddProcess(process *ProductionProcess) {
	if pb.Processes == nil {
		pb.Processes = make([]*ProductionProcess, 0)
	}
	pb.Processes = append(pb.Processes, process)
}

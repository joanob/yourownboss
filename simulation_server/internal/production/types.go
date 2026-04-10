package production

// DTOs for production buildings and processes
type ProcessResourceDTO struct {
	ID         *int64 `json:"id,omitempty"`
	ResourceID int64  `json:"resource_id"`
	IsOutput   bool   `json:"is_output"`
}

type ProcessDTO struct {
	ID        *int64               `json:"id,omitempty"`
	Name      string               `json:"name"`
	StartHour *int                 `json:"start_hour,omitempty"`
	EndHour   *int                 `json:"end_hour,omitempty"`
	Resources []ProcessResourceDTO `json:"resources"`
}

type BuildingDTO struct {
	ID        int64        `json:"id"`
	Name      string       `json:"name"`
	Processes []ProcessDTO `json:"processes"`
}

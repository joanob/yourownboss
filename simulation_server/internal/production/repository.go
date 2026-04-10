package production

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) BuildingExists(ctx context.Context, tx *sql.Tx, id int64) (bool, error) {
	var cnt int
	row := tx.QueryRowContext(ctx, "SELECT COUNT(1) FROM production_buildings WHERE id = ?", id)
	if err := row.Scan(&cnt); err != nil {
		return false, fmt.Errorf("building exists scan: %w", err)
	}
	return cnt > 0, nil
}

func (r *Repository) InsertBuilding(ctx context.Context, tx *sql.Tx, id int64, name string) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO production_buildings (id, name) VALUES (?, ?)", id, name)
	return err
}

func (r *Repository) UpdateBuilding(ctx context.Context, tx *sql.Tx, id int64, name string) error {
	_, err := tx.ExecContext(ctx, "UPDATE production_buildings SET name = ? WHERE id = ?", name, id)
	return err
}

// Processes
func (r *Repository) InsertProcess(ctx context.Context, tx *sql.Tx, id int64, buildingID int64, name string, startHour *int, endHour *int) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO production_processes (id, building_id, name, start_hour, end_hour) VALUES (?, ?, ?, ?, ?)", id, buildingID, name, startHour, endHour)
	return err
}

func (r *Repository) UpdateProcess(ctx context.Context, tx *sql.Tx, id int64, name string, startHour *int, endHour *int) error {
	_, err := tx.ExecContext(ctx, "UPDATE production_processes SET name = ?, start_hour = ?, end_hour = ? WHERE id = ?", name, startHour, endHour, id)
	return err
}

func (r *Repository) DeleteProcessesNotIn(ctx context.Context, tx *sql.Tx, buildingID int64, keepIDs []int64) error {
	// If keepIDs empty, delete all processes for building
	if len(keepIDs) == 0 {
		_, err := tx.ExecContext(ctx, "DELETE FROM production_processes WHERE building_id = ?", buildingID)
		return err
	}
	// build query
	query := "DELETE FROM production_processes WHERE building_id = ? AND id NOT IN ("
	args := make([]interface{}, 0, len(keepIDs)+1)
	args = append(args, buildingID)
	for i, id := range keepIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, id)
	}
	query += ")"
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

// Process resources: we'll delete existing for a process and re-insert
func (r *Repository) DeleteProcessResources(ctx context.Context, tx *sql.Tx, processID int64) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM process_resources WHERE process_id = ?", processID)
	return err
}

func (r *Repository) InsertProcessResource(ctx context.Context, tx *sql.Tx, processID int64, resourceID int64, isOutput bool) error {
	out := 0
	if isOutput {
		out = 1
	}
	_, err := tx.ExecContext(ctx, "INSERT INTO process_resources (process_id, resource_id, is_output) VALUES (?, ?, ?)", processID, resourceID, out)
	return err
}

// Detailed read types
type ResourceDetail struct {
	ResourceID int64  `json:"resource_id"`
	IsOutput   bool   `json:"is_output"`
	Name       string `json:"name"`
}

type ProcessDetail struct {
	ID        *int64           `json:"id,omitempty"`
	Name      *string          `json:"name,omitempty"`
	StartHour *int             `json:"start_hour,omitempty"`
	EndHour   *int             `json:"end_hour,omitempty"`
	Resources []ResourceDetail `json:"resources"`
}

type BuildingDetail struct {
	ID        int64           `json:"id"`
	Name      string          `json:"name"`
	Processes []ProcessDetail `json:"processes"`
}

// GetAllBuildingsDetailed returns buildings with nested processes and resource data
func (r *Repository) GetAllBuildingsDetailed(ctx context.Context) ([]BuildingDetail, error) {
	q := `SELECT b.id AS b_id, b.name AS b_name,
				   p.id AS p_id, p.name AS p_name, p.start_hour AS p_start, p.end_hour AS p_end,
				   pr.resource_id AS res_id, pr.is_output AS res_is_output,
				   rr.name AS res_name
			FROM production_buildings b
			LEFT JOIN production_processes p ON p.building_id = b.id
			LEFT JOIN process_resources pr ON pr.process_id = p.id
			LEFT JOIN resources rr ON rr.id = pr.resource_id
			ORDER BY b.id, p.id;`

	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query buildings: %w", err)
	}
	defer rows.Close()

	buildings := make([]BuildingDetail, 0)
	var curB *BuildingDetail
	var curP *ProcessDetail

	for rows.Next() {
		var bID sql.NullInt64
		var bName sql.NullString
		var pID sql.NullInt64
		var pName sql.NullString
		var pStart sql.NullInt64
		var pEnd sql.NullInt64
		var resID sql.NullInt64
		var resIsOutput sql.NullInt64
		var resName sql.NullString

		if err := rows.Scan(&bID, &bName, &pID, &pName, &pStart, &pEnd, &resID, &resIsOutput, &resName); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if !bID.Valid {
			continue
		}

		if curB == nil || curB.ID != bID.Int64 {
			// new building
			curB = &BuildingDetail{ID: bID.Int64, Name: bName.String, Processes: []ProcessDetail{}}
			buildings = append(buildings, *curB)
			// point to newly appended
			curB = &buildings[len(buildings)-1]
			curP = nil
		}

		if pID.Valid {
			if curP == nil || (curP.ID != nil && *curP.ID != pID.Int64) {
				// new process
				pname := pName.String
				var pidPtr *int64
				var namePtr *string
				pidVal := pID.Int64
				pidPtr = &pidVal
				namePtr = &pname
				var startPtr *int
				var endPtr *int
				if pStart.Valid {
					v := int(pStart.Int64)
					startPtr = &v
				}
				if pEnd.Valid {
					v := int(pEnd.Int64)
					endPtr = &v
				}
				proc := ProcessDetail{ID: pidPtr, Name: namePtr, StartHour: startPtr, EndHour: endPtr, Resources: []ResourceDetail{}}
				curB.Processes = append(curB.Processes, proc)
				curP = &curB.Processes[len(curB.Processes)-1]
			}

			if resID.Valid {
				rd := ResourceDetail{ResourceID: resID.Int64, IsOutput: resIsOutput.Int64 == 1, Name: resName.String}

				curP.Resources = append(curP.Resources, rd)
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return buildings, nil
}

// GetBuildingDetailed returns one building with nested processes and resources
func (r *Repository) GetBuildingDetailed(ctx context.Context, bid int64) (*BuildingDetail, error) {
	q := `SELECT b.id AS b_id, b.name AS b_name,
				   p.id AS p_id, p.name AS p_name, p.start_hour AS p_start, p.end_hour AS p_end,
				   pr.resource_id AS res_id, pr.is_output AS res_is_output,
				   rr.name AS res_name
			FROM production_buildings b
			LEFT JOIN production_processes p ON p.building_id = b.id
			LEFT JOIN process_resources pr ON pr.process_id = p.id
			LEFT JOIN resources rr ON rr.id = pr.resource_id
			WHERE b.id = ?
			ORDER BY p.id;`

	rows, err := r.db.QueryContext(ctx, q, bid)
	if err != nil {
		return nil, fmt.Errorf("query building: %w", err)
	}
	defer rows.Close()

	var building *BuildingDetail
	var curP *ProcessDetail

	for rows.Next() {
		var bID sql.NullInt64
		var bName sql.NullString
		var pID sql.NullInt64
		var pName sql.NullString
		var pStart sql.NullInt64
		var pEnd sql.NullInt64
		var resID sql.NullInt64
		var resIsOutput sql.NullInt64
		var resName sql.NullString

		if err := rows.Scan(&bID, &bName, &pID, &pName, &pStart, &pEnd, &resID, &resIsOutput, &resName); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if !bID.Valid {
			continue
		}

		if building == nil {
			building = &BuildingDetail{ID: bID.Int64, Name: bName.String, Processes: []ProcessDetail{}}
			curP = nil
		}

		if pID.Valid {
			if curP == nil || (curP.ID != nil && *curP.ID != pID.Int64) {
				pname := pName.String
				var pidPtr *int64
				var namePtr *string
				pidVal := pID.Int64
				pidPtr = &pidVal
				namePtr = &pname
				var startPtr *int
				var endPtr *int
				if pStart.Valid {
					v := int(pStart.Int64)
					startPtr = &v
				}
				if pEnd.Valid {
					v := int(pEnd.Int64)
					endPtr = &v
				}
				proc := ProcessDetail{ID: pidPtr, Name: namePtr, StartHour: startPtr, EndHour: endPtr, Resources: []ResourceDetail{}}
				building.Processes = append(building.Processes, proc)
				curP = &building.Processes[len(building.Processes)-1]
			}

			if resID.Valid {
				rd := ResourceDetail{ResourceID: resID.Int64, IsOutput: resIsOutput.Int64 == 1, Name: resName.String}
				curP.Resources = append(curP.Resources, rd)
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	if building == nil {
		return nil, nil
	}
	return building, nil
}

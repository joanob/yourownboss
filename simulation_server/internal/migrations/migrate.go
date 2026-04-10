package migrations

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"
)

// Migrate reads SQL files from dir and applies any not-present migrations.
// Files must follow the convention NNN_description.sql (e.g. 001_init.sql).
func Migrate(db *sql.DB, dir string) error {
	// ensure migrations table exists
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS migrations (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        checksum TEXT,
        applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
    );`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			// no migrations folder — nothing to do
			return nil
		}
		return fmt.Errorf("read migrations dir: %w", err)
	}

	re := regexp.MustCompile(`^(\d+)_.*\.sql$`)
	type mf struct {
		name  string
		path  string
		order int
	}
	var files []mf
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !re.MatchString(name) {
			continue
		}
		m := re.FindStringSubmatch(name)
		ord, _ := strconv.Atoi(m[1])
		files = append(files, mf{name: name, path: filepath.Join(dir, name), order: ord})
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].order == files[j].order {
			return files[i].name < files[j].name
		}
		return files[i].order < files[j].order
	})

	for _, f := range files {
		var cnt int
		// sqlite uses ? param placeholder — this works for the default driver.
		if err := db.QueryRow("SELECT COUNT(1) FROM migrations WHERE name = ?", f.name).Scan(&cnt); err != nil {
			return fmt.Errorf("check migration %s: %w", f.name, err)
		}
		if cnt > 0 {
			continue
		}

		b, err := os.ReadFile(f.path)
		if err != nil {
			return fmt.Errorf("read file %s: %w", f.path, err)
		}
		sum := sha256.Sum256(b)
		checksum := hex.EncodeToString(sum[:])

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", f.name, err)
		}

		if _, err := tx.Exec(string(b)); err != nil {
			tx.Rollback()
			return fmt.Errorf("exec migration %s: %w", f.name, err)
		}

		if _, err := tx.Exec("INSERT INTO migrations (name, checksum, applied_at) VALUES (?, ?, ?)", f.name, checksum, time.Now().UTC()); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", f.name, err)
		}

		if err := tx.Commit(); err != nil {
			tx.Rollback()
			return fmt.Errorf("commit migration %s: %w", f.name, err)
		}
	}

	return nil
}

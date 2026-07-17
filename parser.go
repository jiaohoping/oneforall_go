package oneforall

import (
	"fmt"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// FromDB populates the Result from the OneForAll SQLite database at dbPath
// for a single target domain.
//
// The table name is derived by replacing "." with "_" in target. It first
// tries <target>_now_result (the table OneForAll writes final results to),
// then falls back to the plain <target> name.
func (r *Result) FromDB(dbPath, target string) error {
	return r.FromDBMulti(dbPath, []string{target})
}

// FromDBMulti populates the Result from the OneForAll SQLite database at
// dbPath for multiple target domains. Results from all targets are merged
// into r.Subdomains.
func (r *Result) FromDBMulti(dbPath string, targets []string) error {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	defer sqlDB.Close()

	var all []Subdomain
	for _, target := range targets {
		base := strings.ReplaceAll(target, ".", "_")
		table := resolveTable(db, base)
		if table == "" {
			continue
		}
		var records []Subdomain
		if err := db.Table(table).Find(&records).Error; err != nil {
			return fmt.Errorf("failed to query table %q: %w", table, err)
		}
		all = append(all, records...)
	}

	r.Subdomains = all
	return nil
}

// resolveTable returns the best-matching SQLite table name for baseName.
// It prefers <baseName>_now_result (standard OneForAll output), then falls
// back to <baseName>. Returns "" when neither table exists.
func resolveTable(db *gorm.DB, baseName string) string {
	if preferred := baseName + "_now_result"; db.Migrator().HasTable(preferred) {
		return preferred
	}
	if db.Migrator().HasTable(baseName) {
		return baseName
	}
	return ""
}

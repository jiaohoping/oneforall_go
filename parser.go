package oneforall

import (
	"fmt"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func (r *Result) FromDB(dbPath, tableName string) error {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %v", err)
	}

	defer sqlDB.Close()

	var dbRecords []Subdomain
	if err := db.Table(strings.ReplaceAll(tableName, ".", "_")).Find(&dbRecords).Error; err != nil {
		return fmt.Errorf("failed to query database: %v", err)
	}

	r.Subdomains = dbRecords

	return nil
}

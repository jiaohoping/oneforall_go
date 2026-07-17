package oneforall_test

import (
	"os"
	"path/filepath"
	"testing"

	oneforall "github.com/jiaohoping/oneforall_go"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// createTestDB creates a temporary SQLite database with the given table name
// and inserts the provided subdomain records. Returns the database file path.
func createTestDB(t *testing.T, tableName string, records []oneforall.Subdomain) string {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "result.sqlite3")

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	if err := db.Table(tableName).AutoMigrate(&oneforall.Subdomain{}); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}
	for _, r := range records {
		if err := db.Table(tableName).Create(&r).Error; err != nil {
			t.Fatalf("Create record failed: %v", err)
		}
	}

	sqlDB, _ := db.DB()
	sqlDB.Close()

	return dbPath
}

func TestFromDB_PlainTableName(t *testing.T) {
	records := []oneforall.Subdomain{
		{Subdomain: "www.example.com", IP: "1.2.3.4", Alive: 1},
		{Subdomain: "mail.example.com", IP: "5.6.7.8", Alive: 0},
	}
	dbPath := createTestDB(t, "example_com", records)

	r := &oneforall.Result{}
	if err := r.FromDB(dbPath, "example.com"); err != nil {
		t.Fatalf("FromDB error: %v", err)
	}
	if len(r.Subdomains) != 2 {
		t.Errorf("got %d subdomains, want 2", len(r.Subdomains))
	}
}

func TestFromDB_PrefersNowResultTable(t *testing.T) {
	plain := []oneforall.Subdomain{
		{Subdomain: "old.example.com", IP: "1.1.1.1"},
	}
	now := []oneforall.Subdomain{
		{Subdomain: "new.example.com", IP: "2.2.2.2"},
		{Subdomain: "new2.example.com", IP: "3.3.3.3"},
	}

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "result.sqlite3")

	db, _ := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: logger.Discard})
	db.Table("example_com").AutoMigrate(&oneforall.Subdomain{})
	for _, r := range plain {
		db.Table("example_com").Create(&r)
	}
	db.Table("example_com_now_result").AutoMigrate(&oneforall.Subdomain{})
	for _, r := range now {
		db.Table("example_com_now_result").Create(&r)
	}
	sqlDB, _ := db.DB()
	sqlDB.Close()

	result := &oneforall.Result{}
	if err := result.FromDB(dbPath, "example.com"); err != nil {
		t.Fatalf("FromDB error: %v", err)
	}
	// Should have read _now_result (2 rows), not the plain table (1 row).
	if len(result.Subdomains) != 2 {
		t.Errorf("got %d subdomains, want 2 (from _now_result table)", len(result.Subdomains))
	}
	for _, s := range result.Subdomains {
		if s.Subdomain == "old.example.com" {
			t.Error("plain table record leaked into result; _now_result should have been used")
		}
	}
}

func TestFromDBMulti_MergesTargets(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "result.sqlite3")

	db, _ := gorm.Open(sqlite.Open(dbPath), &gorm.Config{Logger: logger.Discard})
	db.Table("a_com").AutoMigrate(&oneforall.Subdomain{})
	db.Table("a_com").Create(&oneforall.Subdomain{Subdomain: "www.a.com"})
	db.Table("b_com").AutoMigrate(&oneforall.Subdomain{})
	db.Table("b_com").Create(&oneforall.Subdomain{Subdomain: "www.b.com"})
	db.Table("b_com").Create(&oneforall.Subdomain{Subdomain: "mail.b.com"})
	sqlDB, _ := db.DB()
	sqlDB.Close()

	result := &oneforall.Result{}
	if err := result.FromDBMulti(dbPath, []string{"a.com", "b.com"}); err != nil {
		t.Fatalf("FromDBMulti error: %v", err)
	}
	if len(result.Subdomains) != 3 {
		t.Errorf("got %d subdomains, want 3", len(result.Subdomains))
	}
}

func TestFromDB_MissingTable_ReturnsEmpty(t *testing.T) {
	records := []oneforall.Subdomain{
		{Subdomain: "www.other.com"},
	}
	dbPath := createTestDB(t, "other_com", records)

	r := &oneforall.Result{}
	if err := r.FromDB(dbPath, "example.com"); err != nil {
		t.Fatalf("FromDB should not error when table is missing, got: %v", err)
	}
	if len(r.Subdomains) != 0 {
		t.Errorf("expected 0 subdomains for missing table, got %d", len(r.Subdomains))
	}
}

func TestFromDB_MissingDatabase(t *testing.T) {
	r := &oneforall.Result{}
	err := r.FromDB("/nonexistent/path/result.sqlite3", "example.com")
	if err == nil {
		t.Error("expected error for non-existent database, got nil")
	}
}

func TestFromDB_SubdomainFields(t *testing.T) {
	records := []oneforall.Subdomain{
		{
			Subdomain: "api.example.com",
			IP:        "10.0.0.1,10.0.0.2",
			Alive:     1,
			CDN:       0,
			Status:    200,
			Title:     "API",
			Module:    "dns",
		},
	}
	dbPath := createTestDB(t, "example_com", records)

	r := &oneforall.Result{}
	if err := r.FromDB(dbPath, "example.com"); err != nil {
		t.Fatalf("FromDB error: %v", err)
	}
	if len(r.Subdomains) != 1 {
		t.Fatalf("got %d subdomains, want 1", len(r.Subdomains))
	}
	sub := r.Subdomains[0]
	if sub.Subdomain != "api.example.com" {
		t.Errorf("Subdomain = %q, want api.example.com", sub.Subdomain)
	}
	ips := sub.IPs()
	if len(ips) != 2 || ips[0] != "10.0.0.1" || ips[1] != "10.0.0.2" {
		t.Errorf("IPs() = %v, want [10.0.0.1 10.0.0.2]", ips)
	}
	if !sub.IsAlive() {
		t.Error("IsAlive() should be true")
	}
	if sub.Status != 200 {
		t.Errorf("Status = %d, want 200", sub.Status)
	}
	if sub.Module != "dns" {
		t.Errorf("Module = %q, want dns", sub.Module)
	}

	// Ensure parser_test.go imports are all used (sqlite, os)
	_ = os.DevNull
}

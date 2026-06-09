package query

import (
	"context"
	"testing"

	"dealscanner/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCheckIfNew(t *testing.T) {
	gdb, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}
	err = gdb.AutoMigrate(&models.TtlItem{})
	if err != nil {
		t.Fatalf("Failed to migrate test DB: %v", err)
	}
	defer func() {
		sqlDB, _ := gdb.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	dbClient := Use(gdb)
	ctx := context.Background()

	itemId := "test-item-123"
	scanner := ScannerEbay

	isNew1, err := dbClient.CheckIfNew(ctx, itemId, scanner)
	if err != nil {
		t.Fatalf("CheckIfNew failed: %v", err)
	}
	if !isNew1 {
		t.Errorf("Expected item to be new")
	}

	isNew2, err := dbClient.CheckIfNew(ctx, itemId, scanner)
	if err != nil {
		t.Fatalf("CheckIfNew failed: %v", err)
	}
	if isNew2 {
		t.Errorf("Expected item to NOT be new")
	}

	isNewCs1, err := dbClient.CheckIfNewCsItem(ctx, "AK-47 | Redline", 0.15, "CS_TRADE")
	if err != nil {
		t.Fatalf("CheckIfNewCsItem failed: %v", err)
	}
	if !isNewCs1 {
		t.Errorf("Expected CS item to be new")
	}

	isNewCs2, err := dbClient.CheckIfNewCsItem(ctx, "AK-47 | Redline", 0.15, "CS_TRADE")
	if err != nil {
		t.Fatalf("CheckIfNewCsItem failed: %v", err)
	}
	if isNewCs2 {
		t.Errorf("Expected CS item to NOT be new")
	}
}

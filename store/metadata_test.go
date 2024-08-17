package store

import (
	"os"
	"testing"
	"time"
)

func TestMetadataStore(t *testing.T) {
	dbPath := "test_metadata.db"
	defer os.Remove(dbPath)

	ms, err := NewMetadataStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create MetadataStore: %v", err)
	}
	defer ms.db.Close()

	t.Run("CreateMetadata", testCreateMetadata(ms))
	t.Run("GetMetadata", testGetMetadata(ms))
	t.Run("UpdateMetadata", testUpdateMetadata(ms))
	t.Run("DeleteMetadata", testDeleteMetadata(ms))
}

func testCreateMetadata(ms *MetadataStore) func(*testing.T) {
	return func(t *testing.T) {
		metadata := &Metadata{
			ObjectID:  "test1",
			LocalPath: "path/to/test1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := ms.Create(metadata)
		if err != nil {
			t.Errorf("Failed to create metadata: %v", err)
		}

		// Verify creation
		stored, err := ms.Get("test1")
		if err != nil {
			t.Errorf("Failed to get created metadata: %v", err)
		}
		if stored.ObjectID != metadata.ObjectID || stored.LocalPath != metadata.LocalPath {
			t.Errorf("Stored metadata does not match created metadata")
		}
	}
}

func testGetMetadata(ms *MetadataStore) func(*testing.T) {
	return func(t *testing.T) {
		metadata, err := ms.Get("test1")
		if err != nil {
			t.Errorf("Failed to get metadata: %v", err)
		}
		if metadata.ObjectID != "test1" || metadata.LocalPath != "path/to/test1" {
			t.Errorf("Retrieved metadata does not match expected values")
		}
	}
}

func testUpdateMetadata(ms *MetadataStore) func(*testing.T) {
	return func(t *testing.T) {
		metadata := &Metadata{
			ObjectID:  "test1",
			LocalPath: "updated/path/to/test1",
			UpdatedAt: time.Now(),
		}

		err := ms.Update(metadata)
		if err != nil {
			t.Errorf("Failed to update metadata: %v", err)
		}

		// Verify update
		updated, err := ms.Get("test1")
		if err != nil {
			t.Errorf("Failed to get updated metadata: %v", err)
		}
		if updated.LocalPath != "updated/path/to/test1" {
			t.Errorf("Updated metadata does not reflect changes")
		}
	}
}

func testDeleteMetadata(ms *MetadataStore) func(*testing.T) {
	return func(t *testing.T) {
		err := ms.Delete("test1")
		if err != nil {
			t.Errorf("Failed to delete metadata: %v", err)
		}

		// Verify deletion
		_, err = ms.Get("test1")
		if err == nil {
			t.Errorf("Metadata still exists after deletion")
		}
	}
}

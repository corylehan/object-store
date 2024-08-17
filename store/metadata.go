package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Metadata struct {
	ObjectID   string
	LocalPath  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type MetadataStore struct {
	db *sql.DB
}

func NewMetadataStore(dbPath string) (*MetadataStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create table if it doesn't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS metadata (
			object_id TEXT PRIMARY KEY,
			local_path TEXT NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata table: %w", err)
	}

	return &MetadataStore{db: db}, nil
}

func (ms *MetadataStore) Create(metadata *Metadata) error {
	_, err := ms.db.Exec(`
		INSERT INTO metadata (object_id, local_path, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`, metadata.ObjectID, metadata.LocalPath, metadata.CreatedAt, metadata.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create metadata: %w", err)
	}
	return nil
}

func (ms *MetadataStore) Get(objectID string) (*Metadata, error) {
	var metadata Metadata
	err := ms.db.QueryRow(`
		SELECT object_id, local_path, created_at, updated_at
		FROM metadata
		WHERE object_id = ?
	`, objectID).Scan(&metadata.ObjectID, &metadata.LocalPath, &metadata.CreatedAt, &metadata.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("metadata not found for object_id: %s", objectID)
		}
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}
	return &metadata, nil
}

func (ms *MetadataStore) Update(metadata *Metadata) error {
	_, err := ms.db.Exec(`
		UPDATE metadata
		SET local_path = ?, updated_at = ?
		WHERE object_id = ?
	`, metadata.LocalPath, metadata.UpdatedAt, metadata.ObjectID)

	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}
	return nil
}

func (ms *MetadataStore) Delete(objectID string) error {
	_, err := ms.db.Exec(`
		DELETE FROM metadata
		WHERE object_id = ?
	`, objectID)

	if err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}
	return nil
}

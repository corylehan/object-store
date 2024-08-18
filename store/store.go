package store

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"time"
)

type Store struct {
	FileStorage   *FileStorage
	MetadataStore *MetadataStore
}

func NewStore(configFile, dbPath string) (*Store, error) {
	fs, err := NewFileStorage(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create FileStorage: %w", err)
	}

	ms, err := NewMetadataStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create MetadataStore: %w", err)
	}

	return &Store{
		FileStorage:   fs,
		MetadataStore: ms,
	}, nil
}

func (s *Store) CreateObject(objectPath string, data []byte) (string, error) {
	objectID := generateObjectID(data)
	localPath := filepath.Join(s.FileStorage.config.StorageDirectory, objectID)

	// Check if the object already exists
	_, err := s.MetadataStore.Get(objectID)
	if err == nil {
		return "", fmt.Errorf("object with ID %s already exists", objectID)
	}

	err = s.FileStorage.Create(objectID, data)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	metadata := &Metadata{
		ObjectID:   objectID,
		ObjectPath: objectPath,
		LocalPath:  localPath,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = s.MetadataStore.Create(metadata)
	if err != nil {
		// If metadata creation fails, rollback file creation
		s.FileStorage.Delete(objectID)
		return "", fmt.Errorf("failed to create metadata: %w", err)
	}

	return objectID, nil
}

func (s *Store) ReadObject(objectIDOrPath string) ([]byte, error) {
	metadata, err := s.getMetadata(objectIDOrPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	data, err := s.FileStorage.Read(metadata.ObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

func (s *Store) UpdateObject(objectIDOrPath string, data []byte) error {
	metadata, err := s.getMetadata(objectIDOrPath)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	err = s.FileStorage.Update(metadata.ObjectID, data)
	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	metadata.UpdatedAt = time.Now()
	err = s.MetadataStore.Update(metadata)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}

	return nil
}

func (s *Store) DeleteObject(objectIDOrPath string) error {
	metadata, err := s.getMetadata(objectIDOrPath)
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}

	err = s.FileStorage.Delete(metadata.ObjectID)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	err = s.MetadataStore.Delete(metadata.ObjectID)
	if err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}

func (s *Store) getMetadata(objectIDOrPath string) (*Metadata, error) {
	metadata, err := s.MetadataStore.Get(objectIDOrPath)
	if err == nil {
		return metadata, nil
	}

	// If not found by ObjectID, try by ObjectPath
	metadata, err = s.MetadataStore.GetByObjectPath(objectIDOrPath)
	if err != nil {
		return nil, fmt.Errorf("object not found: %s", objectIDOrPath)
	}

	return metadata, nil
}

func generateObjectID(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

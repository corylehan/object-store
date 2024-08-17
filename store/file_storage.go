package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the configuration for the FileStorage.
type Config struct {
	StorageDirectory string `json:"storage_directory"`
}

// FileStorage represents a simple object storage system.
type FileStorage struct {
	config Config
}

// NewFileStorage creates a new FileStorage instance using the provided configuration file.
func NewFileStorage(configFile string) (*FileStorage, error) {
	config := &Config{}
	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config: %w", err)
	}

	if err := os.MkdirAll(config.StorageDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &FileStorage{
		config: *config,
	}, nil
}

// Create stores a new object with the given name and data.
func (s *FileStorage) Create(name string, data []byte) error {
	filePath := filepath.Join(s.config.StorageDirectory, name)
	_, err := os.Stat(filePath)
	if err == nil {
		return fmt.Errorf("object with name %s already exists", name)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// Read retrieves the object with the given name.
func (s *FileStorage) Read(name string) ([]byte, error) {
	filePath := filepath.Join(s.config.StorageDirectory, name)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read object %s: %w", name, err)
	}
	return data, nil
}

// Update modifies the content of an existing object.
func (s *FileStorage) Update(name string, data []byte) error {
	filePath := filepath.Join(s.config.StorageDirectory, name)
	_, err := os.Stat(filePath)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("object %s does not exist", name)
	}
	if err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	}

	return os.WriteFile(filePath, data, 0644)
}

// Delete removes the object with the given name.
func (s *FileStorage) Delete(name string) error {
	filePath := filepath.Join(s.config.StorageDirectory, name)
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", name, err)
	}
	return nil
}

// List returns a slice of all object names in the store.
func (s *FileStorage) List() ([]string, error) {
	entries, err := os.ReadDir(s.config.StorageDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() {
			names = append(names, entry.Name())
		}
	}
	return names, nil
}

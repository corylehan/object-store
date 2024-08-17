package store

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestStore(t *testing.T) {
    tempDir, err := os.MkdirTemp("", "store-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	configFile := filepath.Join(tempDir, "config.json")
	storageDir := filepath.Join(tempDir, "storage")
	dbPath := filepath.Join(tempDir, "metadata.db")
	config := Config{StorageDirectory: storageDir}
	configData, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
    if err := os.WriteFile(configFile, configData, 0644); err != nil {
		t.Fatal(err)
	}

	s, err := NewStore(configFile, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer s.MetadataStore.db.Close()

    t.Run("CreateReadUpdateDeleteByObjectID", func(t *testing.T) {
        testStoreOperations(t, s, "object-id", "test.txt", []byte("Hello, ObjectID test!"))
    })

    t.Run("CreateReadUpdateDeleteByObjectPath", func(t *testing.T) {
        testStoreOperations(t, s, "object-path", "documents/report.docx", []byte("Hello, ObjectPath test!"))
    })
}

func testStoreOperations(t *testing.T, s *Store, testCase, objectPath string, data []byte) {
	updatedData := []byte("Hello, updated world!")

	// Test CreateObject
	objectID, err := s.CreateObject(objectPath, data)
	if err != nil {
		t.Fatalf("%s: Failed to create object: %v", testCase, err)
	}

	// Test ReadObject by ObjectID
	readData, err := s.ReadObject(objectID)
	if err != nil {
		t.Fatalf("%s: Failed to read object by ObjectID: %v", testCase, err)
	}
	if !bytes.Equal(data, readData) {
		t.Errorf("%s: Read data by ObjectID doesn't match", testCase)
	}

	// Test ReadObject by ObjectPath
	readData, err = s.ReadObject(objectPath)
	if err != nil {
		t.Fatalf("%s: Failed to read object by ObjectPath: %v", testCase, err)
	}
	if !bytes.Equal(data, readData) {
		t.Errorf("%s: Read data by ObjectPath doesn't match", testCase)
	}

	// Test UpdateObject
	err = s.UpdateObject(objectID, updatedData)
	if err != nil {
		t.Fatalf("%s: Failed to update object by ObjectID: %v", testCase, err)
	}

	readData, err = s.ReadObject(objectPath)
	if err != nil {
		t.Fatalf("%s: Failed to read updated object: %v", testCase, err)
	}
	if !bytes.Equal(updatedData, readData) {
		t.Errorf("%s: Updated data doesn't match", testCase)
	}

	// Test DeleteObject
	err = s.DeleteObject(objectPath)
	if err != nil {
		t.Fatalf("%s: Failed to delete object: %v", testCase, err)
	}

	_, err = s.ReadObject(objectID)
	if err == nil {
		t.Errorf("%s: Object still exists after deletion", testCase)
	}

	_, err = s.ReadObject(objectPath)
	if err == nil {
		t.Errorf("%s: Object still exists after deletion", testCase)
	}
}

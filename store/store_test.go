package store

import (
    "bytes"
    "os"
    "path/filepath"
    "testing"
    "encoding/json"
)

func resetObjectStore(t *testing.T, tempDir string) *ObjectStore {
	storageDir := filepath.Join(tempDir, "storage")
	os.RemoveAll(storageDir)
	config := Config{StorageDirectory: storageDir}
	configFile := filepath.Join(tempDir, "config.json")

	configData, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configFile, configData, 0644); err != nil {
		t.Fatal(err)
	}

	objectStore, err := NewObjectStore(configFile)
	if err != nil {
		t.Fatal(err)
	}

	return objectStore
}

func TestObjectStore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "object-store-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testCases := []struct {
		name     string
		testFunc func(t *testing.T, os *ObjectStore)
	}{
		{"TestCreate", testCreate},
		{"TestReadNonExistent", testReadNonExistent},
		{"TestUpdate", testUpdate},
		{"TestDelete", testDelete},
		{"TestList", testList},
		{"TestCreateDuplicateName", testCreateDuplicateName},
		{"TestUpdateNonExistent", testUpdateNonExistent},
		{"TestDeleteNonExistent", testDeleteNonExistent},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objectStore := resetObjectStore(t, tempDir)
			tc.testFunc(t, objectStore)
		})
	}
}

func testCreate(t *testing.T, os *ObjectStore) {
    data := []byte("test data")
    err := os.Create("testfile", data)
    if err != nil {
        t.Errorf("Create failed: %v", err)
    }

    readData, err := os.Read("testfile")
    if err != nil {
        t.Errorf("Read after Create failed: %v", err)
    }
    if !bytes.Equal(data, readData) {
        t.Errorf("Read data doesn't match: expected %v, got %v", data, readData)
    }
}

func testReadNonExistent(t *testing.T, os *ObjectStore) {
    _, err := os.Read("nonexistent")
    if err == nil {
        t.Error("Expected error when reading non-existent file")
    }
}

func testUpdate(t *testing.T, os *ObjectStore) {
    initialData := []byte("initial data")
    updatedData := []byte("updated data")
    
    err := os.Create("updatefile", initialData)
    if err != nil {
        t.Fatalf("Create failed: %v", err)
    }

    err = os.Update("updatefile", updatedData)
    if err != nil {
        t.Errorf("Update failed: %v", err)
    }

    readData, err := os.Read("updatefile")
    if err != nil {
        t.Errorf("Read after Update failed: %v", err)
    }
    if !bytes.Equal(updatedData, readData) {
        t.Errorf("Updated data doesn't match: expected %v, got %v", updatedData, readData)
    }
}

func testDelete(t *testing.T, os *ObjectStore) {
    data := []byte("data to delete")
    
    err := os.Create("deletefile", data)
    if err != nil {
        t.Fatalf("Create failed: %v", err)
    }

    err = os.Delete("deletefile")
    if err != nil {
        t.Errorf("Delete failed: %v", err)
    }

    _, err = os.Read("deletefile")
    if err == nil {
        t.Error("Expected error when reading deleted file")
    }
}

func testList(t *testing.T, os *ObjectStore) {
    files := []string{"file1", "file2", "file3"}
    data := []byte("data")

    for _, file := range files {
        err := os.Create(file, data)
        if err != nil {
            t.Fatalf("Create failed: %v", err)
        }
    }

    listed, err := os.List()
    if err != nil {
        t.Errorf("List failed: %v", err)
    }

    if len(listed) != len(files) {
        t.Errorf("List returned wrong number of files: expected %d, got %d", len(files), len(listed))
    }

    for _, file := range files {
        found := false
        for _, listedFile := range listed {
            if file == listedFile {
                found = true
                break
            }
        }
        if !found {
            t.Errorf("File %s not found in list", file)
        }
    }
}

func testCreateDuplicateName(t *testing.T, os *ObjectStore) {
    data := []byte("data")
    err := os.Create("duplicate", data)
    if err != nil {
        t.Fatalf("First Create failed: %v", err)
    }

    err = os.Create("duplicate", data)
    if err == nil {
        t.Error("Expected error when creating file with duplicate name")
    }
}

func testUpdateNonExistent(t *testing.T, os *ObjectStore) {
    data := []byte("data")
    err := os.Update("nonexistent", data)
    if err == nil {
        t.Error("Expected error when updating non-existent file")
    }
}

func testDeleteNonExistent(t *testing.T, os *ObjectStore) {
    err := os.Delete("nonexistent")
    if err == nil {
        t.Error("Expected error when deleting non-existent file")
    }
}

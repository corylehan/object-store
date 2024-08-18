package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/corylehan/object-store/api"
	"github.com/corylehan/object-store/store"
)

const (
	testPort       = 8081
	testConfigFile = "test_config.json"
	testDBPath     = "test_metadata.db"
)

func setupTestEnvironment(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "objectstore")
	if err != nil {
		t.Fatal(err)
	}

	configContent := fmt.Sprintf(`{"storage_directory":"%s"}`, filepath.Join(tempDir, "storage"))
	os.WriteFile(testConfigFile, []byte(configContent), 0644)
}

func teardownTestEnvironment() {
	os.Remove(testConfigFile)
	os.Remove(testDBPath)
}

func startTestServer(t *testing.T) {
	setupTestEnvironment(t)

	s, err := store.NewStore(testConfigFile, testDBPath)
	if err != nil {
		t.Fatal(err)
	}

	server := api.NewServer(testPort, s)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			t.Logf("Server exited with error: %v", err)
		}
	}()

	time.Sleep(time.Second) // Wait for server to start
}

func TestMain(m *testing.M) {
	startTestServer(&testing.T{})
	defer teardownTestEnvironment()

	os.Exit(m.Run())
}

func baseURL() string {
	return fmt.Sprintf("http://localhost:%d", testPort)
}

func TestObjectLifecycle(t *testing.T) {
	objectPath := "test/lifecycle.txt"
	encodedPath := url.PathEscape(objectPath)

	content := []byte("Lifecycle at " + time.Now().String())
	updatedContent := []byte("Updated lifecycle at " + time.Now().String())

	// Create object
	createResp, err := http.Post(baseURL()+"/objects?path="+encodedPath, "application/octet-stream", bytes.NewBuffer(content))
	if err != nil || createResp.StatusCode != http.StatusCreated {
		t.Fatalf("Failed to create object: %v, status: %d", err, createResp.StatusCode)
	}

	// Read object
	readResp, err := http.Get(baseURL() + "/objects/" + encodedPath)
	if err != nil || readResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to read object: %v, status: %d", err, readResp.StatusCode)
	}

	readData, _ := io.ReadAll(readResp.Body)
	if !bytes.Equal(readData, content) {
		t.Error("Retrieved data doesn't match stored data")
	}

	// Update object
	req, _ := http.NewRequest(http.MethodPut, baseURL()+"/objects/"+encodedPath, bytes.NewBuffer(updatedContent))
	updateResp, err := http.DefaultClient.Do(req)
	if err != nil || updateResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to update object: %v", err)
	}

	// Read updated object
	readUpdatedResp, _ := http.Get(baseURL() + "/objects/" + encodedPath)
	readUpdatedData, _ := io.ReadAll(readUpdatedResp.Body)
	if !bytes.Equal(readUpdatedData, updatedContent) {
		t.Error("Updated data doesn't match")
	}

	// Delete object
	req, _ = http.NewRequest(http.MethodDelete, baseURL()+"/objects/"+encodedPath, nil)
	deleteResp, err := http.DefaultClient.Do(req)
	if err != nil || deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("Failed to delete object: %v", err)
	}

	// Verify object is deleted
	notFoundResp, _ := http.Get(baseURL() + "/objects/" + encodedPath)
	if notFoundResp.StatusCode != http.StatusNotFound {
		t.Error("Object should be deleted")
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("CreateEmptyObject", func(t *testing.T) {
		resp, err := http.Post(baseURL()+"/objects?path=empty.txt", "application/octet-stream", bytes.NewBuffer([]byte{}))
		if err != nil || resp.StatusCode != http.StatusCreated {
			t.Errorf("Failed on empty object: %v", err)
		}
	})

	t.Run("CreateWithoutPath", func(t *testing.T) {
		resp, err := http.Post(baseURL()+"/objects", "application/octet-stream", bytes.NewBuffer([]byte("test")))
		if err != nil || resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Should fail without path: %v", err)
		}
	})

	t.Run("DeleteNonExistentObject", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, baseURL()+"/objects/nonexistent.txt", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Delete should fail: %v", err)
		}
	})

	t.Run("InvalidObjectNames", func(t *testing.T) {
		invalidNames := []string{"../../etc/passwd", "com1", "nul", " ", "object\n.txt"}
		for _, name := range invalidNames {
			encodedName := url.PathEscape(name)
			_, err := http.Post(baseURL()+"/objects?path="+encodedName, "application/octet-stream", bytes.NewBuffer([]byte("test")))
			if err == nil {
				t.Errorf("Should refuse invalid object name: %s", name)
			}
		}
	})
}

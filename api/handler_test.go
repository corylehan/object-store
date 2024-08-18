// api/handler_test.go
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/corylehan/object-store/store"
)

func setupTestServer(t *testing.T) (*httptest.Server, *store.Store) {
    tempDir, err := os.MkdirTemp("", "api-test")
    if err != nil {
        t.Fatal(err)
    }

    configFile := filepath.Join(tempDir, "config.json")
    storageDir := filepath.Join(tempDir, "storage")
    dbPath := filepath.Join(tempDir, "metadata.db")
    config := store.Config{StorageDirectory: storageDir}
    configData, _ := json.Marshal(config)
    os.WriteFile(configFile, configData, 0644)

    s, err := store.NewStore(configFile, dbPath)
    if err != nil {
        t.Fatal(err)
    }

    h := NewHandler(s)
    mux := http.NewServeMux()
    mux.HandleFunc("/objects", h.handleObjects)
    mux.HandleFunc("/objects/", h.handleObject)
    server := httptest.NewServer(mux)

    return server, s
}

func TestCreateAndGetObject(t *testing.T) {
    server, _ := setupTestServer(t)
    defer server.Close()

    objectPath := "test/object.txt"
    encodedPath := url.QueryEscape(objectPath)
    content := []byte("Hello, world!")

    // Create object
    resp, err := http.Post(fmt.Sprintf("%s/objects?path=%s", server.URL, encodedPath), "application/octet-stream", bytes.NewReader(content))
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        t.Errorf("Expected status %d, got %d", http.StatusCreated, resp.StatusCode)
    }

    // Retrieve object
    resp, err = http.Get(fmt.Sprintf("%s/objects/%s", server.URL, encodedPath))
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        t.Fatal(err)
    }

    if !bytes.Equal(content, body) {
        t.Errorf("Expected content %q, got %q", content, body)
    }

    // Retrieve non-existent object
    resp, err = http.Get(fmt.Sprintf("%s/objects/non-existent.txt", server.URL))
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusNotFound {
        t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
    }
}

func TestUpdateAndDeleteObject(t *testing.T) {
    server, _ := setupTestServer(t)
    defer server.Close()

    objectPath := "test/object.txt"
    encodedPath := url.QueryEscape(objectPath)
    content := []byte("Initial content")
    updatedContent := []byte("Updated content")

    // Create object
    http.Post(fmt.Sprintf("%s/objects?path=%s", server.URL, encodedPath), "application/octet-stream", bytes.NewReader(content))

    // Update object
    req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/objects/%s", server.URL, encodedPath), bytes.NewReader(updatedContent))
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
    }

    // Verify updated content
    resp, _ = http.Get(fmt.Sprintf("%s/objects/%s", server.URL, encodedPath))
    body, _ := io.ReadAll(resp.Body)
    resp.Body.Close()

    if !bytes.Equal(updatedContent, body) {
        t.Errorf("Expected updated content %q, got %q", updatedContent, body)
    }

    // Delete object
    req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/objects/%s", server.URL, encodedPath), nil)
    resp, err = http.DefaultClient.Do(req)
    if err != nil {
        t.Fatal(err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
    }

    // Verify object is deleted
    resp, _ = http.Get(fmt.Sprintf("%s/objects/%s", server.URL, encodedPath))
    if resp.StatusCode != http.StatusNotFound {
        t.Errorf("Expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
    }
}

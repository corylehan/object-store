package api

import (
    "fmt"
    "io"
    "net/http"
    "strings"

    "github.com/corylehan/object-store/store"
)

type Handler struct {
    store *store.Store
}

func NewHandler(s *store.Store) *Handler {
    return &Handler{store: s}
}

func (h *Handler) handleObjects(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodPost:
        h.createObject(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func (h *Handler) handleObject(w http.ResponseWriter, r *http.Request) {
    // Extract everything after /objects/
    objectPath := strings.TrimPrefix(r.URL.Path, "/objects/")
    switch r.Method {
    case http.MethodGet:
        h.getObject(w, r, objectPath)
    case http.MethodPut:
        h.updateObject(w, r, objectPath)
    case http.MethodDelete:
        h.deleteObject(w, r, objectPath)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func (h *Handler) createObject(w http.ResponseWriter, r *http.Request) {
    data, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    objectPath := r.URL.Query().Get("path")
    if objectPath == "" {
        http.Error(w, "Missing 'path' query parameter", http.StatusBadRequest)
        return
    }

    objectID, err := h.store.CreateObject(objectPath, data)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    fmt.Fprintf(w, "Created object %s", objectID)
}

func (h *Handler) getObject(w http.ResponseWriter, r *http.Request, objectPath string) {
    data, err := h.store.ReadObject(objectPath)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write(data)
}

func (h *Handler) updateObject(w http.ResponseWriter, r *http.Request, objectPath string) {
    data, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if err := h.store.UpdateObject(objectPath, data); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Updated object %s", objectPath)
}

func (h *Handler) deleteObject(w http.ResponseWriter, r *http.Request, objectPath string) {
    if err := h.store.DeleteObject(objectPath); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "Deleted object %s", objectPath)
}


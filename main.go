// main.go

package main

import (
	"log"

	"github.com/corylehan/object-store/api"
	"github.com/corylehan/object-store/store"
)

func main() {
	configFile := "./config.json"
	dbPath := "./metadata.db"

	s, err := store.NewStore(configFile, dbPath)
	if err != nil {
		log.Fatalf("Failed to create Store: %v", err)
	}
	defer s.MetadataStore.db.Close()

	api.StartServer(s)
}

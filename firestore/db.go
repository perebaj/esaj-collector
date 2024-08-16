// Package firestore gather all function that deal with the firestore database
package firestore

import (
	"cloud.google.com/go/firestore"
)

// Storage is a struct that holds the firestore client and the projectID and database name
type Storage struct {
	client    *firestore.Client
	projectID string
	database  string
}

// NewStorage creates a new Storage struct
func NewStorage(client *firestore.Client, projectID, database string) *Storage {
	return &Storage{
		client:    client,
		projectID: projectID,
		database:  database,
	}
}

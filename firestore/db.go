// Package firestore gather all function that deal with the firestore database
package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/perebaj/esaj/esaj"
)

// Storage is a struct that holds the firestore client and the projectID and database name
type Storage struct {
	client    *firestore.Client
	projectID string
}

// NewStorage creates a new Storage struct
func NewStorage(client *firestore.Client, projectID string) *Storage {
	return &Storage{
		client:    client,
		projectID: projectID,
	}
}

// SaveProcessSeeds saves the process seeds in the firestore database
func (s *Storage) SaveProcessSeeds(ctx context.Context, ps []esaj.ProcessSeed) error {
	collection := s.client.Collection("process_seeds")
	bulkWriter := s.client.BulkWriter(ctx)
	for _, seed := range ps {
		docRef := collection.NewDoc()

		_, err := bulkWriter.Set(docRef, seed)
		if err != nil {
			return err
		}
	}

	bulkWriter.Flush()
	return nil
}

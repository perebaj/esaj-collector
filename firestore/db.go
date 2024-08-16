package firestore

import (
	"cloud.google.com/go/firestore"
)

type Storage struct {
	client    *firestore.Client
	projectID string
	database  string
}

func NewStorage(client *firestore.Client, projectID, database string) *Storage {
	return &Storage{
		client:    client,
		projectID: projectID,
		database:  database,
	}
}

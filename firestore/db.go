// Package firestore gather all function that deal with the firestore database
package firestore

import (
	"context"
	"log/slog"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/perebaj/esaj/esaj"
	"github.com/perebaj/esaj/tracing"
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
	traceID := tracing.GetTraceIDFromContext(ctx)
	collection := s.client.Collection("process_seeds")
	bulkWriter := s.client.BulkWriter(ctx)
	for _, seed := range ps {
		docRef := collection.Doc(seed.ProcessID)
		m := make(map[string]interface{})
		m["process_id"] = seed.ProcessID
		m["oab"] = seed.OAB
		m["url"] = seed.URL
		m["trace_id"] = traceID
		_, err := bulkWriter.Set(docRef, m)
		if err != nil {
			return err
		}
	}

	bulkWriter.Flush()
	return nil
}

// ProcessSeed is the struct that represents the process seed in the firestore database
type ProcessSeed struct {
	ID        string
	ProcessID string
	OAB       string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GetSeedsByOAB returns all the process seeds given an OAB identifier
func (s *Storage) GetSeedsByOAB(ctx context.Context, oab string) ([]ProcessSeed, error) {
	collection := s.client.Collection("process_seeds")
	iter := collection.Where("oab", "==", oab).Documents(ctx)

	doc, err := iter.GetAll()
	if err != nil {
		return nil, err
	}

	var seeds []ProcessSeed
	for _, d := range doc {
		seed := ProcessSeed{
			ID:        d.Ref.ID,
			ProcessID: d.Data()["process_id"].(string),
			OAB:       d.Data()["oab"].(string),
			URL:       d.Data()["url"].(string),
			CreatedAt: d.CreateTime,
			UpdatedAt: d.UpdateTime,
		}

		seeds = append(seeds, seed)
	}

	return seeds, nil
}

// SaveProcessBasicInfo saves the process basic information in the firestore database
func (s *Storage) SaveProcessBasicInfo(ctx context.Context, pBasicInfo esaj.ProcessBasicInfo) error {
	traceID := tracing.GetTraceIDFromContext(ctx)
	slog.Info("saving process basic info", "process_id", pBasicInfo.ProcessID, "trace_id", traceID)

	collection := s.client.Collection("process_basic_info")
	docRef := collection.Doc(pBasicInfo.ProcessID)

	m := make(map[string]interface{})
	m["process_id"] = pBasicInfo.ProcessID
	m["foro_code"] = pBasicInfo.ProcessForo
	m["foro_name"] = pBasicInfo.ForoName
	m["process_code"] = pBasicInfo.ProcessCode
	m["judge"] = pBasicInfo.Judge
	m["class"] = pBasicInfo.Class
	m["claimant"] = pBasicInfo.Claimant
	m["defendant"] = pBasicInfo.Defendant
	m["vara"] = pBasicInfo.Vara
	m["trace_id"] = traceID
	m["url"] = pBasicInfo.URL

	_, err := docRef.Set(ctx, m)
	return err
}

// ProcessBasicInfo is the struct that represents the process basic information in the firestore database
type ProcessBasicInfo struct {
	ProcessID   string    `firestore:"process_id"`
	ProcessForo string    `firestore:"foro_code"`
	ForoName    string    `firestore:"foro_name"`
	ProcessCode string    `firestore:"process_code"`
	Judge       string    `firestore:"judge"`
	Class       string    `firestore:"class"`
	Claimant    string    `firestore:"claimant"`
	Defendant   string    `firestore:"defendant"`
	Vara        string    `firestore:"vara"`
	URL         string    `firestore:"url"`
	CreatedAt   time.Time `firestore:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at"`
	TraceID     string    `firestore:"trace_id"`
}

// ProcessBasicInfoByProcessID returns the process basic information given a processID
func (s *Storage) ProcessBasicInfoByProcessID(ctx context.Context, processID string) (*ProcessBasicInfo, error) {
	collection := s.client.Collection("process_basic_info")
	doc, err := collection.Doc(processID).Get(ctx)
	if err != nil {
		return nil, err
	}

	p := ProcessBasicInfo{
		ProcessID:   doc.Data()["process_id"].(string),
		ProcessForo: doc.Data()["foro_code"].(string),
		ForoName:    doc.Data()["foro_name"].(string),
		ProcessCode: doc.Data()["process_code"].(string),
		Judge:       doc.Data()["judge"].(string),
		Class:       doc.Data()["class"].(string),
		Claimant:    doc.Data()["claimant"].(string),
		Defendant:   doc.Data()["defendant"].(string),
		Vara:        doc.Data()["vara"].(string),
		URL:         doc.Data()["url"].(string),
		CreatedAt:   doc.CreateTime,
		UpdatedAt:   doc.UpdateTime,
		TraceID:     doc.Data()["trace_id"].(string),
	}

	return &p, nil
}

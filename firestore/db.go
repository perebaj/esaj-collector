// Package firestore gather all function that deal with the firestore database
package firestore

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/perebaj/esaj/esaj"
	"github.com/perebaj/esaj/tracing"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	logger := slog.With("traceID", traceID)
	logger.Info("saving process basic info", "process_id", pBasicInfo.ProcessID)

	collection := s.client.Collection("process_basic_info")
	docRef := collection.Doc(pBasicInfo.ProcessID)

	doc, err := docRef.Get(ctx)
	if err != nil && status.Code(err) != codes.NotFound {
		return fmt.Errorf("error getting document: %w", err)
	}
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

	if status.Code(err) == codes.NotFound {
		m["oabs"] = firestore.ArrayUnion(pBasicInfo.OAB)
	} else {
		existingOABs := doc.Data()["oabs"].([]interface{})
		existingOABs = append(existingOABs, pBasicInfo.OAB)
		m["oabs"] = firestore.ArrayUnion(existingOABs...)
	}
	_, err = docRef.Set(ctx, m)

	return err
}

// ProcessBasicInfoByOAB returns all process that has the same OAB identifier
func (s *Storage) ProcessBasicInfoByOAB(ctx context.Context, oab string) ([]esaj.ProcessBasicInfo, error) {
	collection := s.client.Collection("process_basic_info")
	iter := collection.Where("oabs", "array-contains", oab).Documents(ctx)

	doc, err := iter.GetAll()
	if err != nil {
		return nil, err
	}

	var processBasicInfo []esaj.ProcessBasicInfo
	for _, d := range doc {
		p := esaj.ProcessBasicInfo{
			ProcessID:   d.Data()["process_id"].(string),
			ProcessForo: d.Data()["foro_code"].(string),
			ForoName:    d.Data()["foro_name"].(string),
			ProcessCode: d.Data()["process_code"].(string),
			Judge:       d.Data()["judge"].(string),
			Class:       d.Data()["class"].(string),
			Claimant:    d.Data()["claimant"].(string),
			Defendant:   d.Data()["defendant"].(string),
			Vara:        d.Data()["vara"].(string),
			URL:         d.Data()["url"].(string),
		}

		processBasicInfo = append(processBasicInfo, p)
	}

	return processBasicInfo, nil
}

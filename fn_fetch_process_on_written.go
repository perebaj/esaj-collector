// Package collector Triggered when a new object is created or updated in the Firestore database(process_seeds). projectID := "blup-432616"
// It will receive the object, parse it and trigger a new search in the esaj website to scrape basic information about the process.
package collector

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	fs "cloud.google.com/go/firestore"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/googleapis/google-cloudevents-go/cloud/firestoredata"
	"github.com/perebaj/esaj/esaj"
	"github.com/perebaj/esaj/firestore"
	"github.com/perebaj/esaj/logger"
	"github.com/perebaj/esaj/tracing"
	"google.golang.org/protobuf/proto"
)

func init() {
	logger, err := logger.NewLoggerSlog(logger.ConfigLogger{
		Level:  logger.LevelInfo,
		Format: logger.FormatGCP,
	})
	if err != nil {
		slog.Error("error initializing logger", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(logger)

	functions.CloudEvent("fn-fetch-process-on-written", Parser)
}

// Parser ...
// TODO(@perebaj): improve the desing of functions that are triggered by cloud events like im doing in the http functions
func Parser(_ context.Context, event event.Event) error {
	var data firestoredata.DocumentEventData
	err := proto.Unmarshal(event.Data(), &data)
	if err != nil {
		return fmt.Errorf("error unmarshalling data. Data received: %v. error: %w", data.Value, err)
	}

	doc := data.GetValue().GetFields()

	processID := doc["process_id"].GetStringValue()
	traceID := doc["trace_id"].GetStringValue()
	if processID == "" || traceID == "" {
		return fmt.Errorf("process_id or trace_id is empty")
	}

	logger := slog.With("trace_id", traceID)

	ctx := context.Background()
	ctx = tracing.SetTraceIDInContext(ctx, traceID)

	u := doc["url"].GetStringValue()

	projectID := "blup-432616"
	databaseName := "blup-db"

	fsClient, err := fs.NewClientWithDatabase(ctx, projectID, databaseName)
	if err != nil {
		logger.Error("error creating firestore client", "error", err)
		return fmt.Errorf("error creating firestore client. error: %w", err)
	}
	storage := firestore.NewStorage(fsClient, projectID)

	// This request doesn't require cookies to access information
	esajClient := esaj.New(esaj.Config{}, &http.Client{
		Timeout: 90 * time.Second,
	})

	pBasicInfo, err := esajClient.FetchBasicProcessInfo(ctx, u, processID)
	if err != nil {
		logger.Error("error fetching basic process info", "error", err)
		return fmt.Errorf("error fetching basic process info. error: %w", err)
	}

	err = storage.SaveProcessBasicInfo(ctx, *pBasicInfo)
	if err != nil {
		logger.Error("error saving basic process info", "error", err)
		return fmt.Errorf("error saving basic process info. error: %w", err)
	}

	return nil
}

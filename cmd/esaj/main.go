// // Package main
// package main

// import (
// 	"fmt"
// 	"log/slog"
// 	"net/http"
// 	"os"
// 	"time"

// 	"github.com/perebaj/esaj"
// 	"github.com/perebaj/esaj/api"
// 	"github.com/perebaj/esaj/postgres"
// )

// func main() {
// 	logger, err := esaj.NewLoggerSlog(esaj.ConfigLogger{
// 		Level:  esaj.LevelInfo,
// 		Format: esaj.FormatLogFmt,
// 	})
// 	if err != nil {
// 		slog.Info("error initializing logger: %v", "error", err)
// 		os.Exit(1)
// 	}

// 	slog.SetDefault(logger)

// 	postgresCfg := postgres.Config{
// 		URL:             os.Getenv("POSTGRES_URL"),
// 		MaxOpenConns:    10,
// 		MaxIdleConns:    10,
// 		ConnMaxIdleTime: 10,
// 	}

// 	db, err := postgres.OpenDB(postgresCfg)
// 	if err != nil {
// 		slog.Error("error opening database", "error", err)
// 		os.Exit(1)
// 	}
// 	storage := postgres.NewStorage(db)

// 	esaj := esaj.New(esaj.Config{}, &http.Client{
// 		Timeout: 30 * time.Second,
// 	})

// 	mux := api.NewServerMux(storage, esaj)

// 	slog.Info("server running on port 8080")

// 	svc := &http.Server{
// 		Addr:         fmt.Sprintf(":%d", 8080),
// 		Handler:      mux,
// 		ReadTimeout:  5 * time.Second,
// 		WriteTimeout: 30 * time.Second,
// 	}

//		if err := svc.ListenAndServe(); err != nil {
//			slog.Error("error starting server", "error", err)
//			os.Exit(1)
//		}
//	}
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!")
}

func esajApiHandler(w http.ResponseWriter, r *http.Request) {
	message := "This HTTP triggered function executed successfully. Pass a name in the query string for a personalized response.\n"
	name := r.URL.Query().Get("name")
	if name != "" {
		message = fmt.Sprintf("Hello, %s. This HTTP triggered function executed successfully.\n", name)
	}
	fmt.Fprint(w, message)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/hello", helloHandler)
	mux.HandleFunc("GET /api/esaj-api", esajApiHandler)
	mux.HandleFunc("GET /api/jojo", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Temporario")
	})

	listenAddr := ":8080"
	if val, ok := os.LookupEnv("FUNCTIONS_CUSTOMHANDLER_PORT"); ok {
		listenAddr = ":" + val
	}

	svr := &http.Server{
		Addr:         listenAddr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	slog.Info(fmt.Sprintf("Listening on %s", listenAddr))
	if err := svr.ListenAndServe(); err != nil {
		slog.Error("Error starting server", "error", err)
		os.Exit(1)
	}
}

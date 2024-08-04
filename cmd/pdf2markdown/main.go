// Package main Main entry point for the pdf2markdown command line tool.
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/perebaj/esaj/llamaparser"
)

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func main() {
	cfg := llamaparser.Config{
		APIKey: getEnvWithDefault("LLAMA_CLOUD_API_KEY", ""),
	}

	if cfg.APIKey == "" {
		slog.Error("LLAMA_CLOUD_API_KEY not set")
		os.Exit(1)
	}

	ll := llamaparser.NewLlamaParser(cfg, &http.Client{
		Timeout: 10 * time.Second,
	})

	// read all files inside the /tmp directory and send a buffer.Bytes to the llama cloud
	entries, err := os.ReadDir("tmp")
	if err != nil {
		slog.Error("Error reading directory", "error", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		fmt.Println(entry.Name())

		//read all the bytes from the file
		buf := new(bytes.Buffer)
		// read all the bytes from the file
		writer := multipart.NewWriter(buf)
		fw, err := writer.CreateFormFile("file", "tmp/"+entry.Name())
		if err != nil {
			log.Fatal(err)
		}
		fd, err := os.Open("tmp/" + entry.Name())
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			_ = fd.Close()
		}()

		_, err = io.Copy(fw, fd)
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			_ = writer.Close()
		}()

		markdown, err := ll.PDFToMarkdown(buf, writer.FormDataContentType())
		if err != nil {
			slog.Error("Error converting pdf to markdown", "error", err)
			os.Exit(1)
		}
		// save the markdown to a file
		err = os.WriteFile("tmp/markdown/"+entry.Name()+".md", []byte(markdown.Markdown), 0644)
		if err != nil {
			slog.Error("Error writing markdown to file", "error", err)
			os.Exit(1)
		}
	}
	// ll.PDFToMarkdown()

}

// Package main Main entry point for the pdf2markdown command line tool.
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/perebaj/esaj"
	"github.com/perebaj/esaj/llamaparser"
)

func main() {
	cfg := llamaparser.Config{
		APIKey: esaj.GetEnvWithDefault("LLAMA_CLOUD_API_KEY", ""),
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

	ch := make(chan llamaparser.Entry)

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		wg.Done()
		for e := range ch {
			markdown, err := ll.PDFToMarkdown(e)
			if err != nil {
				slog.Error("Error converting pdf to markdown", "error", err)
			}

			// save the markdown file
			err = os.WriteFile("tmp/markdown/"+e.EntryPath+".md", []byte(markdown.Markdown), 0644)
			if err != nil {
				slog.Error("Error saving markdown file", "error", err)
			}
		}
	}()
	generateEntries(entries, ch, &wg)

	fmt.Println("Waiting for goroutines to finish")
	close(ch)
	wg.Wait()
}

func generateEntries(entries []fs.DirEntry, ch chan llamaparser.Entry, wg *sync.WaitGroup) {
	wg.Done()
	for _, entry := range entries {
		fmt.Println(entry.Name())

		buf := new(bytes.Buffer)

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

		err = writer.Close()
		if err != nil {
			log.Fatal(err)
		}

		ch <- llamaparser.Entry{
			Buf:         buf,
			ContentType: writer.FormDataContentType(),
			EntryPath:   entry.Name(),
		}
	}
}

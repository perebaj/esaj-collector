// Package llamaparser Package llamaparser provides a client to interact with the Llama Cloud API.
// The Llama Cloud API allows you to parse PDF files and extract the text from them.
package llamaparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	llamaURL = "https://api.cloud.llamaindex.ai"
)

// Config struct to hold the configuration of the llama parser
type Config struct {
	APIKey string
}

// LlamaParser struct to hold the llama parser instance
type LlamaParser struct {
	// URL is the base URL of the llama cloud.
	URL    string
	config Config
	client *http.Client
}

// Entry struct to hold the parsed pdf file
type Entry struct {
	Buf         *bytes.Buffer
	ContentType string
	EntryPath   string
}

// NewLlamaParser create a new instance of LlamaParser
func NewLlamaParser(config Config, client *http.Client) *LlamaParser {
	return &LlamaParser{
		config: config,
		client: client,
		URL:    llamaURL,
	}
}

// PDFToMarkdown convert parsed pdf file in *bytes.Buffer to markdown
func (ll LlamaParser) PDFToMarkdown(e Entry) (*MarkdownResponse, error) {
	slog.Info("uploading the pdf to llama cloud")

	uploadResponse, err := ll.uploadPDF(e.Buf, e.ContentType)
	if err != nil {
		return nil, fmt.Errorf("error while uploading the pdf to llama cloud. error: %v", err)
	}

	// before retrive the parsed markdown, we need to wait the job to be processed, this can take some time
	// so we need to check the job status until it is completed
	ok, err := ll.poolJobStatus(uploadResponse.ID)
	if err != nil || !ok {
		return nil, fmt.Errorf("error while pooling the job status. error: %v", err)
	}

	slog.Info(fmt.Sprintf("job %s is completed. Retriving the markdown", uploadResponse.ID), "job_id", uploadResponse.ID)

	markdownResp, err := ll.getMarkdown(uploadResponse.ID)
	if err != nil {
		return nil, err
	}

	slog.Info("markdown retrived successfully", "job_id", uploadResponse.ID)

	return markdownResp, nil
}

func (ll LlamaParser) uploadPDF(r *bytes.Buffer, contentType string) (*UploadResponse, error) {
	req, err := http.NewRequest("POST", ll.URL+"/api/parsing/upload", r)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+ll.config.APIKey)

	resp, err := ll.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("receive a non 200 status code. status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var uploadResponse UploadResponse
	err = json.Unmarshal(body, &uploadResponse)
	if err != nil {
		return nil, err
	}

	return &uploadResponse, nil
}

func (ll LlamaParser) getMarkdown(jobID string) (*MarkdownResponse, error) {
	req, err := http.NewRequest("GET", ll.URL+"/api/parsing/job/"+jobID+"/result/markdown", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+ll.config.APIKey)

	resp, err := ll.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var markdownResponse MarkdownResponse
	err = json.Unmarshal(body, &markdownResponse)
	if err != nil {
		return nil, err
	}

	return &markdownResponse, nil
}

// poolJobStatus validate the job status until it is completed
// this function will return a boolean that indicates if the pool was successful or not
func (ll LlamaParser) poolJobStatus(jobID string) (bool, error) {
	for {
		req, err := http.NewRequest("GET", ll.URL+"/api/parsing/job/"+jobID, nil)
		if err != nil {
			return false, err
		}

		req.Header.Set("accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+ll.config.APIKey)

		resp, err := ll.client.Do(req)
		if err != nil {
			return false, err
		}

		if resp.StatusCode != http.StatusOK {
			return false, err
		}

		defer func() {
			_ = resp.Body.Close()
		}()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		var uploadResponse UploadResponse
		err = json.Unmarshal(body, &uploadResponse)
		if err != nil {
			return false, err
		}

		if uploadResponse.Status == "SUCCESS" {
			break
		}

		t := time.Second * 1
		slog.Info(fmt.Sprintf("job %s is still processing, waiting %s", uploadResponse.ID, t), "job_id", uploadResponse.ID)
		time.Sleep(time.Second * 1)
	}
	return true, nil
}

// UploadResponse struct to hold the response of the upload request
type UploadResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// MarkdownResponse struct to hold the response of the markdown request
type MarkdownResponse struct {
	Markdown string      `json:"markdown"`
	Metadata JobMetadata `json:"metadata"`
}

// JobMetadata struct to hold the metadata of the job
type JobMetadata struct {
	CreditsUsed     float64 `json:"credits_used"`
	CreditsMax      int     `json:"credits_max"`
	JobCreditsUsage int     `json:"job_credits_usage"`
	JobPages        int     `json:"job_pages"`
	JobIsCacheHit   bool    `json:"job_is_cache_hit"`
}

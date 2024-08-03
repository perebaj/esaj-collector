package llamaparser

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// create a test for poolJobStatus mocking the output of the api/parsing/job endpoint
func TestLlamaParser_poolJobStatus(t *testing.T) {
	// create a new LlamaParser
	ll := NewLlamaParser(
		Config{
			APIKey: "fake-api-key",
		},
		&http.Client{
			Timeout: time.Second * 2,
		},
	)

	// create a new UploadResponse
	uploadResponse := UploadResponse{
		ID: "123",
	}

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check the request method
		if r.Method != http.MethodGet {
			t.Errorf("expected %s, got %s", http.MethodGet, r.Method)
		}

		// check the request URL
		if r.URL.Path != "/api/parsing/job/123" {
			t.Errorf("expected %s, got %s", "/api/parsing/job/123", r.URL.Path)
		}

		callCount++
		if callCount == 1 {
			// write the response
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": "123", "status": "PENDING"}`))
		} else if callCount == 2 {
			// write the response
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id": "123", "status": "SUCCESS"}`))
		}
	}))

	// close the server when the test is done
	defer server.Close()

	ll.URL = server.URL

	ok, err := ll.poolJobStatus(uploadResponse.ID)
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	if !ok {
		t.Errorf("expected true, got false")
	}
}

func TestLlamaParser_getMarkdown(t *testing.T) {
	// create a new LlamaParser
	ll := NewLlamaParser(
		Config{APIKey: "fake-api-key"},
		&http.Client{Timeout: time.Second * 2},
	)

	// create a new MarkdownResponse
	markdownResponse := MarkdownResponse{
		Markdown: "# Hello, World!",
	}

	// create a new http server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected %s, got %s", http.MethodGet, r.Method)
		}

		if r.URL.Path != "/api/parsing/job/123/result/markdown" {
			t.Errorf("expected %s, got %s", "/api/parsing/job/123", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"markdown": "# Hello, World!"}`))
	}))

	defer server.Close()

	ll.URL = server.URL

	markdownResp, err := ll.getMarkdown("123")
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}

	if markdownResp.Markdown != markdownResponse.Markdown {
		t.Errorf("expected %s, got %s", markdownResponse.Markdown, markdownResp.Markdown)
	}
}

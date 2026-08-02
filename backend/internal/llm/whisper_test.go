package llm

import (
	"context"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWhisperClientTranscribeSendsRequestAndParsesResponse(t *testing.T) {
	var gotAuth, gotModel, gotResponseFormat, gotFileContent, gotFileName string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")

		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
			t.Fatalf("expected multipart content type, got %q (err: %v)", mediaType, err)
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		form, err := mr.ReadForm(1 << 20)
		if err != nil {
			t.Fatalf("read multipart form: %v", err)
		}

		gotModel = form.Value["model"][0]
		gotResponseFormat = form.Value["response_format"][0]

		fileHeaders := form.File["file"]
		if len(fileHeaders) != 1 {
			t.Fatalf("expected 1 file part, got %d", len(fileHeaders))
		}
		gotFileName = fileHeaders[0].Filename

		f, err := fileHeaders[0].Open()
		if err != nil {
			t.Fatalf("open file part: %v", err)
		}
		defer f.Close()
		buf := make([]byte, 1024)
		n, _ := f.Read(buf)
		gotFileContent = string(buf[:n])

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"text": "hive three looks good",
			"language": "english",
			"segments": [
				{"text": "hive three looks good", "avg_logprob": -0.25, "no_speech_prob": 0.01}
			]
		}`))
	}))
	defer server.Close()

	client := NewWhisperClient("test-api-key", WithWhisperBaseURL(server.URL))

	result, err := client.Transcribe(context.Background(), strings.NewReader("fake audio bytes"), "recording.m4a")
	if err != nil {
		t.Fatalf("Transcribe returned error: %v", err)
	}

	if gotAuth != "Bearer test-api-key" {
		t.Errorf("expected Authorization header %q, got %q", "Bearer test-api-key", gotAuth)
	}
	if gotModel != "openai/whisper-1" {
		t.Errorf("expected model %q, got %q", "openai/whisper-1", gotModel)
	}
	if gotResponseFormat != "verbose_json" {
		t.Errorf("expected response_format %q, got %q", "verbose_json", gotResponseFormat)
	}
	if gotFileName != "recording.m4a" {
		t.Errorf("expected filename %q, got %q", "recording.m4a", gotFileName)
	}
	if gotFileContent != "fake audio bytes" {
		t.Errorf("expected file content %q, got %q", "fake audio bytes", gotFileContent)
	}

	if result.Text != "hive three looks good" {
		t.Errorf("expected text %q, got %q", "hive three looks good", result.Text)
	}
	if result.Language != "english" {
		t.Errorf("expected language %q, got %q", "english", result.Language)
	}
	if len(result.Segments) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(result.Segments))
	}
	seg := result.Segments[0]
	if seg.Text != "hive three looks good" {
		t.Errorf("expected segment text %q, got %q", "hive three looks good", seg.Text)
	}
	if seg.AvgLogprob != -0.25 {
		t.Errorf("expected avg_logprob %v, got %v", -0.25, seg.AvgLogprob)
	}
	if seg.NoSpeechProb != 0.01 {
		t.Errorf("expected no_speech_prob %v, got %v", 0.01, seg.NoSpeechProb)
	}
}

func TestWhisperClientTranscribeWithCustomModelSendsOverriddenModel(t *testing.T) {
	var gotModel string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || !strings.HasPrefix(mediaType, "multipart/") {
			t.Fatalf("expected multipart content type, got %q (err: %v)", mediaType, err)
		}

		mr := multipart.NewReader(r.Body, params["boundary"])
		form, err := mr.ReadForm(1 << 20)
		if err != nil {
			t.Fatalf("read multipart form: %v", err)
		}
		gotModel = form.Value["model"][0]

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"text": "", "language": "english", "segments": []}`))
	}))
	defer server.Close()

	client := NewWhisperClient("test-api-key", WithWhisperBaseURL(server.URL), WithWhisperModel("openai/whisper-large-v3"))

	_, err := client.Transcribe(context.Background(), strings.NewReader("fake audio bytes"), "recording.m4a")
	if err != nil {
		t.Fatalf("Transcribe returned error: %v", err)
	}

	if gotModel != "openai/whisper-large-v3" {
		t.Errorf("expected model %q, got %q", "openai/whisper-large-v3", gotModel)
	}
}

func TestWhisperClientTranscribeNonOKResponseReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "invalid api key"}`))
	}))
	defer server.Close()

	client := NewWhisperClient("bad-api-key", WithWhisperBaseURL(server.URL))

	_, err := client.Transcribe(context.Background(), strings.NewReader("fake audio bytes"), "recording.m4a")
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected error to mention status code 401, got: %v", err)
	}
	if !strings.Contains(err.Error(), "invalid api key") {
		t.Errorf("expected error to include response body, got: %v", err)
	}
}

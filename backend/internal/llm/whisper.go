package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const (
	defaultWhisperBaseURL = "https://openrouter.ai/api/v1"
	defaultWhisperModel   = "openai/whisper-1"
)

type WhisperAPIError struct {
	StatusCode int
	Body       string
}

func (e *WhisperAPIError) Error() string {
	return fmt.Sprintf("whisper API returned status %d: %s", e.StatusCode, e.Body)
}

type WhisperClient struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

type WhisperOption func(*WhisperClient)

func WithWhisperBaseURL(url string) WhisperOption {
	return func(c *WhisperClient) { c.baseURL = url }
}

func WithWhisperModel(model string) WhisperOption {
	return func(c *WhisperClient) { c.model = model }
}

func NewWhisperClient(apiKey string, opts ...WhisperOption) *WhisperClient {
	c := &WhisperClient{
		apiKey:     apiKey,
		baseURL:    defaultWhisperBaseURL,
		model:      defaultWhisperModel,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type TranscriptionSegment struct {
	Text         string  `json:"text"`
	AvgLogprob   float64 `json:"avg_logprob"`
	NoSpeechProb float64 `json:"no_speech_prob"`
}

type TranscriptionResult struct {
	Text     string                 `json:"text"`
	Language string                 `json:"language"`
	Segments []TranscriptionSegment `json:"segments"`
}

func (c *WhisperClient) Transcribe(ctx context.Context, audio io.Reader, filename string) (*TranscriptionResult, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, audio); err != nil {
		return nil, fmt.Errorf("write audio to form: %w", err)
	}
	if err := writer.WriteField("model", c.model); err != nil {
		return nil, fmt.Errorf("write model field: %w", err)
	}
	if err := writer.WriteField("response_format", "verbose_json"); err != nil {
		return nil, fmt.Errorf("write response_format field: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/audio/transcriptions", &body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &WhisperAPIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	var result TranscriptionResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &result, nil
}

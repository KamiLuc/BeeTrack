package llm

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client wraps the Anthropic SDK client, giving both callers one shared
// construction point instead of each wiring up their own API key/options.
type Client struct {
	anthropic.Client
}

// NewClient builds a Client authenticated with the given API key.
func NewClient(apiKey string, opts ...option.RequestOption) *Client {
	opts = append([]option.RequestOption{option.WithAPIKey(apiKey)}, opts...)
	return &Client{Client: anthropic.NewClient(opts...)}
}

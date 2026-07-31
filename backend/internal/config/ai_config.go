package config

import "fmt"

type AIConfig struct {
	AnthropicAPIKey string
	OpenAIAPIKey    string
}

func LoadAIConfig() (AIConfig, error) {
	cfg := AIConfig{
		AnthropicAPIKey: getEnv("ANTHROPIC_API_KEY", ""),
		OpenAIAPIKey:    getEnv("OPENAI_API_KEY", ""),
	}

	if err := cfg.validate(); err != nil {
		return AIConfig{}, err
	}
	return cfg, nil
}

// Deliberately bypasses AIConfig's validate(), which also hard-requires
// OPENAI_API_KEY for voice logging (not yet built) — an unconfigured key
// here just means the assistant route isn't registered, not that the
// whole API fails to start.
func AnthropicAPIKey() string {
	return getEnv("ANTHROPIC_API_KEY", "")
}

func OpenAIAPIKey() string {
	return getEnv("OPENAI_API_KEY", "")
}

// AnthropicModel is the model the assistant agent loop uses. Set ANTHROPIC_MODEL to override, e.g.:
//
//	ANTHROPIC_MODEL=claude-haiku-4-5         # default — cheapest/fastest, weaker at multi-step tool use
//	ANTHROPIC_MODEL=claude-sonnet-5          # better multi-step tool use, higher cost
//	ANTHROPIC_MODEL=claude-opus-5            # highest quality, highest cost
func AnthropicModel() string {
	return getEnv("ANTHROPIC_MODEL", "claude-haiku-4-5")
}

func (c AIConfig) validate() error {
	if c.AnthropicAPIKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY is required")
	}
	if c.OpenAIAPIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required")
	}
	return nil
}

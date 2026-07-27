package config

import "fmt"

// AIConfig holds settings for the voice-logging and assistant features.
type AIConfig struct {
	AnthropicAPIKey  string
	OpenAIAPIKey     string
	AudioStoragePath string
}

// LoadAIConfig reads AI feature settings from the environment and validates them.
func LoadAIConfig() (AIConfig, error) {
	cfg := AIConfig{
		AnthropicAPIKey:  getEnv("ANTHROPIC_API_KEY", ""),
		OpenAIAPIKey:     getEnv("OPENAI_API_KEY", ""),
		AudioStoragePath: getEnv("AUDIO_STORAGE_PATH", "/data/audio"),
	}

	if err := cfg.validate(); err != nil {
		return AIConfig{}, err
	}
	return cfg, nil
}

// Deliberately bypasses AIConfig's validate(), which also hard-requires
// OPENAI_API_KEY/AUDIO_STORAGE_PATH for voice logging (not yet built) — an
// unconfigured key here just means the assistant route isn't registered,
// not that the whole API fails to start.
func AnthropicAPIKey() string {
	return getEnv("ANTHROPIC_API_KEY", "")
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

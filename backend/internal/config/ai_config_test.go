package config

import "testing"

func TestAIConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     AIConfig
		wantErr bool
	}{
		{
			name: "valid",
			cfg: AIConfig{
				AnthropicAPIKey: "sk-ant-test",
				OpenAIAPIKey:    "sk-openai-test",
			},
			wantErr: false,
		},
		{
			name: "missing anthropic key",
			cfg: AIConfig{
				AnthropicAPIKey: "",
				OpenAIAPIKey:    "sk-openai-test",
			},
			wantErr: true,
		},
		{
			name: "missing openai key",
			cfg: AIConfig{
				AnthropicAPIKey: "sk-ant-test",
				OpenAIAPIKey:    "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOpenRouterAPIKeyDefaultsToEmpty(t *testing.T) {
	if got := OpenRouterAPIKey(); got != "" {
		t.Errorf("expected empty default, got %q", got)
	}

	t.Setenv("OPENROUTER_API_KEY", "sk-or-test")
	if got := OpenRouterAPIKey(); got != "sk-or-test" {
		t.Errorf("expected %q, got %q", "sk-or-test", got)
	}
}

func TestOpenRouterModelDefaultsToClaudeHaiku(t *testing.T) {
	if got := OpenRouterModel(); got != "anthropic/claude-haiku-4.5" {
		t.Errorf("expected default %q, got %q", "anthropic/claude-haiku-4.5", got)
	}

	t.Setenv("OPENROUTER_MODEL", "anthropic/claude-sonnet-5")
	if got := OpenRouterModel(); got != "anthropic/claude-sonnet-5" {
		t.Errorf("expected %q, got %q", "anthropic/claude-sonnet-5", got)
	}
}

func TestOpenRouterWhisperModelDefaultsToWhisper1(t *testing.T) {
	if got := OpenRouterWhisperModel(); got != "openai/whisper-1" {
		t.Errorf("expected default %q, got %q", "openai/whisper-1", got)
	}

	t.Setenv("OPENROUTER_WHISPER_MODEL", "openai/gpt-4o-mini-transcribe")
	if got := OpenRouterWhisperModel(); got != "openai/gpt-4o-mini-transcribe" {
		t.Errorf("expected %q, got %q", "openai/gpt-4o-mini-transcribe", got)
	}
}

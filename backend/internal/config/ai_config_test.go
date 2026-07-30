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

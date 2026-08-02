package config

// OpenRouterAPIKey authenticates both the assistant agent loop
// (internal/service/assistant.go) and the voice worker's chat/transcription
// calls (internal/worker, internal/llm/whisper.go) against OpenRouter.
func OpenRouterAPIKey() string {
	return getEnv("OPENROUTER_API_KEY", "")
}

// OpenRouterModel is the OpenRouter model slug the assistant agent loop and
// voice worker use. Set OPENROUTER_MODEL to override, e.g.:
//
//	OPENROUTER_MODEL=anthropic/claude-haiku-4.5   # default — cheapest/fastest, weaker at multi-step tool use
//	OPENROUTER_MODEL=anthropic/claude-sonnet-5    # better multi-step tool use, higher cost
//	OPENROUTER_MODEL=anthropic/claude-opus-5      # highest quality, highest cost
func OpenRouterModel() string {
	return getEnv("OPENROUTER_MODEL", "anthropic/claude-haiku-4.5")
}

// OpenRouterWhisperModel is the OpenRouter transcription model slug the voice
// worker uses. Set OPENROUTER_WHISPER_MODEL to override.
func OpenRouterWhisperModel() string {
	return getEnv("OPENROUTER_WHISPER_MODEL", "openai/whisper-1")
}

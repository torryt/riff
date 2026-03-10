package internal

import (
	"testing"
)

func TestHasLLM_DisabledByEnv(t *testing.T) {
	t.Setenv("RIFF_NO_AI", "1")

	if HasLLM() {
		t.Error("HasLLM() = true, want false when RIFF_NO_AI=1")
	}
}

func TestLLMProvider_DisabledByEnv(t *testing.T) {
	t.Setenv("RIFF_NO_AI", "1")

	if provider := LLMProvider(); provider != "" {
		t.Errorf("LLMProvider() = %q, want empty when RIFF_NO_AI=1", provider)
	}
}

func TestGenerateDescription_DisabledByEnv(t *testing.T) {
	t.Setenv("RIFF_NO_AI", "1")

	_, err := GenerateDescription(t.TempDir())
	if err != ErrNoLLM {
		t.Errorf("GenerateDescription error = %v, want ErrNoLLM", err)
	}
}

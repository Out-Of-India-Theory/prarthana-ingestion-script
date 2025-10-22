package ai_platform

type TranslateText struct {
	Index       int    `json:"index"`
	Translation string `json:"translation"`
}

type TranslateTextResponse struct {
	Translations []TranslateText `json:"translations"`
}

type BatchTranslateRequest struct {
	Texts           []string `json:"texts"`
	TargetLanguage  string   `json:"target_language"`
	PromptId        *int     `json:"prompt_id"`
}

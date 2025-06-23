package ai_platform

type TranslateText struct {
	Index       int    `json:"index"`
	Translation string `json:"translation"`
}

type TranslateTextResponse struct {
	TranslatedTexts []TranslateText `json:"translated_texts"`
}

type BatchTranslateRequest struct {
	Texts           []string `json:"texts"`
	TargetLanguage  string   `json:"target_language"`
	IsTransliterate bool     `json:"is_transliterate"`
}

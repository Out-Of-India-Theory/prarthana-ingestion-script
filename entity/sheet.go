package entity

type IngestionRequest struct {
	StartID int `json:"start_id" binding:"required,min=1"`
	EndID   int `json:"end_id" binding:"required,max=100000"`
}

type ShlokaSheetResponse struct {
	Records []map[string]interface{} `json:"records"`
}

type ShlokaRecord struct {
	TranslationAssamese  string `json:"translation_assamese"`
	TranslationPunjabi   string `json:"translation_punjabi"`
	TextSanskrit         string `json:"text_sanskrit"`
	TranslationHindi     string `json:"translation_hindi"`
	TranslationBengali   string `json:"translation_bengali"`
	TextKannada          string `json:"text_kannada"`
	TranslationOdiya     string `json:"translation_odiya"`
	TextGujarati         string `json:"text_gujarati"`
	TranslationMalayalam string `json:"translation_malayalam"`
	NameOptional         string `json:"Name (Optional)"`
	RowIndex             int    `json:"row_index"`
	ID                   int    `json:"ID"`
	TranslationMarathi   string `json:"translation_marathi"`
	TranslationTamil     string `json:"translation_tamil"`
	TextMalayalam        string `json:"text_malayalam"`
	TranslationKannada   string `json:"translation_kannada"`
	TranslationTelugu    string `json:"translation_telugu"`
	TextTelugu           string `json:"text_telugu"`
	TranslationEnglish   string `json:"translation_english"`
	TextBengali          string `json:"text_bengali"`
	TextTamil            string `json:"text_tamil"`
	TextOdiya            string `json:"text_odiya"`
	TranslationGujarati  string `json:"translation_gujarati"`
	TextAssamese         string `json:"text_assamese"`
	TextPunjabi          string `json:"text_punjabi"`
	TextHindi            string `json:"text_hindi"`
	TextMarathi          string `json:"text_marathi"`
	TextEnglish          string `json:"text_english"`
}

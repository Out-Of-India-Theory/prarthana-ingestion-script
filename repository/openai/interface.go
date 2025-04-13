package openai

type ClientRepository interface {
	TranslateText(text string, lang string, isTranslation bool) (string, error)
}

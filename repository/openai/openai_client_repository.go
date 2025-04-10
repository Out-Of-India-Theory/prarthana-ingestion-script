package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/configuration"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type OpenAIClientRepository struct {
	logger        *zap.Logger
	configuration configuration.Configuration
}

func InitOpenAIClientRepository(ctx context.Context, configuration configuration.Configuration) *OpenAIClientRepository {
	return &OpenAIClientRepository{
		logger:        logging.WithContext(ctx),
		configuration: configuration,
	}
}

func (r *OpenAIClientRepository) TranslateText(text string, lang string, isTranslation bool) (string, error) {
	err := limiter.Wait(context.Background())
	if err != nil {
		return "", fmt.Errorf("rate limit exceeded: %v", err)
	}

	var taskPrompt string
	if isTranslation {
		taskPrompt = fmt.Sprintf(
			"Translate this Sanskrit Shloka to everyday formal vernacular language in %s. Provide only the translated text without any explanations. Translated text can only contain %s text and nothing else.",
			lang, lang,
		)
	} else {
		taskPrompt = fmt.Sprintf(
			"Transliterate this Sanskrit Shloka to %s. Provide only the transliterated text without any explanations. Transliterated text can only contain %s text and nothing else.",
			lang, lang,
		)
	}

	messages := []RequestMessage{
		{Role: "system", Content: taskPrompt},
		{Role: "user", Content: text},
	}

	requestBody := OpenAIRequest{
		Model:       "gpt-4-turbo",
		Messages:    messages,
		Temperature: 0.3,
	}

	payloadBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	url := "https://api.openai.com/v1/chat/completions"
	maxRetries := 5
	backoff := time.Second

	for i := 0; i < maxRetries; i++ {
		if err := limiter.Wait(context.Background()); err != nil {
			log.Printf("Rate limit wait error: %v", err)
			continue
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
		if err != nil {
			return "", fmt.Errorf("failed to create HTTP request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+r.configuration.OpenAIConfig.Key)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Request failed: %v", err)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %v", err)
		}

		if resp.StatusCode == 429 {
			log.Printf("Rate limited. Retrying in %v...", backoff)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
		}

		var openAIResponse OpenAIResponse
		if err := json.Unmarshal(body, &openAIResponse); err != nil {
			return "", fmt.Errorf("failed to parse response body: %v", err)
		}

		if len(openAIResponse.Choices) > 0 {
			return strings.TrimSpace(openAIResponse.Choices[0].Message.Content), nil
		}

		return "", fmt.Errorf("no translation found in response")
	}

	return "", fmt.Errorf("max retries reached, request failed")
}

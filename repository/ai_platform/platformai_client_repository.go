package ai_platform

import (
	// "bytes"
	"context"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/client/http_client"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"

	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/configuration"
	"go.uber.org/zap"
	"net/http"
)

type PlatformAIClientRepository struct {
	logger       *zap.Logger
	httpClient   *http_client.HttpBaseClient
	clientConfig configuration.HttpClientConfig
}

func InitPlatformAIClientRepository(ctx context.Context, configuration configuration.HttpClientConfig) *PlatformAIClientRepository {
	httpClient := http_client.NewHttpClient("", 0, configuration.Timeout) // Initialize with default values, can be customized later
	return &PlatformAIClientRepository{
		logger:       logging.WithContext(ctx),
		httpClient:   httpClient,
		clientConfig: configuration,
	}
}

const translateURI = "ai-platform/v1//translate/batch"

func (r *PlatformAIClientRepository) BatchTranslateText(ctx context.Context, texts []string, lang string, isTransliterate bool) ([]TranslateText, error) {

	url := fmt.Sprintf("%s/%s", r.clientConfig.Address, translateURI)

	//prepare request body payload
	requestBody := BatchTranslateRequest{
		Texts:           texts,
		TargetLanguage:  lang,
		IsTransliterate: isTransliterate,
	}
	header := http.Header{}
	header.Set("Content-Type", "application/json")
	var response TranslateTextResponse
	err := r.httpClient.Call(ctx, url, requestBody, header, http.MethodPost, &response)
	if err != nil {
		r.logger.Error(fmt.Sprintf("error while Translating Text form external service: %v", err))
		return nil, err
	}
	return response.TranslatedTexts, nil
}

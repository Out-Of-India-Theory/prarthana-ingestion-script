package ai_platform

import (
	"context"
	"fmt"
	"time"

	"github.com/Out-Of-India-Theory/oit-go-commons/client/http_client"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"

	"net/http"

	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/configuration"
	"go.uber.org/zap"
)

type PlatformAIClientRepository struct {
	logger       *zap.Logger
	httpClient   *http_client.HttpBaseClient
	clientConfig configuration.AIPlatformConfig
}

func InitPlatformAIClientRepository(ctx context.Context, config configuration.Configuration) *PlatformAIClientRepository {

	httpClient := http_client.NewHttpClient("", 0, 30*time.Second)
	return &PlatformAIClientRepository{
		logger:       logging.WithContext(ctx),
		httpClient:   httpClient,
		clientConfig: config.AIPlatformClientConfig,
	}
}

const translateURI = "ai-platform/v1/translate/batch"

func (r *PlatformAIClientRepository) BatchTranslateText(ctx context.Context, texts []string, lang string, promptId int) ([]TranslateText, error) {
	url := fmt.Sprintf("%s/%s", r.clientConfig.Address, translateURI)
	r.logger.Info("Final translation URL", zap.String("url", url))

	requestBody := BatchTranslateRequest{
		Texts:           texts,
		TargetLanguage:  lang,
		PromptId:        &promptId,
	}
	header := http.Header{}
	header.Set("Content-Type", "application/json")

	var response TranslateTextResponse
	err := r.httpClient.Call(ctx, url, requestBody, header, http.MethodPost, &response)
	if err != nil {
		r.logger.Error(fmt.Sprintf("error while Translating Text from external service: %v", err))
		return nil, err
	}

	return response.Translations, nil
}

package prarthana

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	config "github.com/Out-Of-India-Theory/prarthana-ingestion-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type PrarthanaESRepository struct {
	logger *zap.Logger
	config config.ESConfig
}

func InitPrarthanaESRepository(ctx context.Context, config config.ESConfig) *PrarthanaESRepository {
	return &PrarthanaESRepository{
		logger: logging.WithContext(ctx),
		config: config,
	}
}

func (r *PrarthanaESRepository) InsertDeitySearchDocument(doc entity.DeitySearchData) error {
	url := fmt.Sprintf("%s/%s/_doc", r.config.Host, r.config.DeityIndex)
	jsonData, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	payload := strings.NewReader(string(jsonData))
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", r.config.Auth)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("Elasticsearch returned non-2xx status: %d, body: %s", res.StatusCode, string(body))
	}

	return nil
}

func (r *PrarthanaESRepository) InsertPrarthanaSearchDocument(doc entity.PrarthanaSearchData) error {
	url := fmt.Sprintf("%s/%s/_doc", r.config.Host, r.config.PrarthanaIndex)
	jsonData, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Add("Authorization", r.config.Auth)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer res.Body.Close()

	_, err = io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("Elasticsearch returned non-success status: %s", res.Status)
	}
	return nil
}

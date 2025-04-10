package zoho

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/util"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ZohoService struct {
	logger        *zap.Logger
	configuration *configuration.Configuration
	httpClient    *http.Client
	tokenManager  TokenManager
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type TokenManager struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

func InitZohoService(ctx context.Context,
	configuration *configuration.Configuration,
	httpClient *http.Client,
) *ZohoService {
	return &ZohoService{
		logger:        logging.WithContext(ctx),
		configuration: configuration,
		httpClient:    httpClient,
	}
}

func (s *ZohoService) RefreshAccessToken() (string, error) {
	if s.configuration.ZohoConfig.RefreshToken == "" {
		return "", fmt.Errorf("refresh token not set")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", s.configuration.ZohoConfig.RefreshToken)
	data.Set("client_id", s.configuration.ZohoConfig.ClientId)
	data.Set("client_secret", s.configuration.ZohoConfig.ClientSecret)

	resp, err := http.PostForm(s.configuration.ZohoConfig.TokenUrl, data)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to refresh token, response: %s", string(body))
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", err
	}
	return tokenResp.AccessToken, nil
}

func (s *ZohoService) GetSheetData(ctx context.Context, sheetName string, response interface{}) error {
	accessToken := util.GetZohoAccessTokenFromContext(ctx)
	url1 := fmt.Sprintf("https://sheet.zoho.in/api/v2/%s", s.configuration.ZohoConfig.SheetId)
	data := url.Values{}
	data.Set("method", "worksheet.records.fetch")
	data.Set("worksheet_name", sheetName)
	data.Set("header_row", "1")

	// Create a new HTTP request with POST method
	req, err := http.NewRequest(http.MethodPost, url1, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set the required headers
	req.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("error response from server: %s", string(body))
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	err = json.Unmarshal(bytes, &response)
	if err != nil {
		return fmt.Errorf("failed to parse response body: %w", err)
	}
	return nil
}

func (s *ZohoService) SetSheetData(ctx context.Context, sheetName string, translatedRecords entity.ShlokaSheetResponse) error {
	accessToken := util.GetZohoAccessTokenFromContext(ctx)
	url1 := fmt.Sprintf("https://sheet.zoho.com/api/v2/%s", s.configuration.ZohoConfig.SheetId)

	// Marshal the records to JSON
	jsonData, err := json.Marshal(translatedRecords.Records)
	if err != nil {
		return fmt.Errorf("failed to marshal records: %w", err)
	}

	data := url.Values{}
	data.Set("method", "worksheet.records.add")
	data.Set("worksheet_name", sheetName)
	data.Set("header_row", "1")
	data.Set("json_data", string(jsonData))

	// Create POST request
	req, err := http.NewRequest(http.MethodPost, url1, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error response from server: %s", string(body))
	}

	return nil
}

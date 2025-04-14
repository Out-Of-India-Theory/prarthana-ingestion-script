package zoho

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/util"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
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

func (s *ZohoService) SetSheetData(ctx context.Context, sheetName string, row int, columnIndexes []int, dataArray []interface{}) error {
	accessToken := util.GetZohoAccessTokenFromContext(ctx)
	resourceID := s.configuration.ZohoConfig.SheetId
	apiURL := fmt.Sprintf("https://sheet.zoho.in/api/v2/%s", resourceID)
	payload := url.Values{}
	payload.Set("method", "row.content.set")
	payload.Set("worksheet_name", sheetName)
	payload.Set("row", fmt.Sprintf("%d", row))
	colArrayJSON, err := json.Marshal(columnIndexes)
	if err != nil {
		return fmt.Errorf("failed to marshal column array: %w", err)
	}
	payload.Set("column_array", string(colArrayJSON))
	dataArrayJSON, err := json.Marshal(dataArray)
	if err != nil {
		return fmt.Errorf("failed to marshal data array: %w", err)
	}
	payload.Set("data_array", string(dataArrayJSON))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(payload.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Zoho API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Zoho API error (%d): %s", resp.StatusCode, string(body))
	}
	//log.Printf("✅ Zoho row update successful for row %d. Response: %s", row, string(body))
	return nil
}

func (s *ZohoService) AddUUIDToSheet(ctx context.Context, sheetName string, uuid string, row int) error {
	accessToken := util.GetZohoAccessTokenFromContext(ctx)
	resourceID := s.configuration.ZohoConfig.SheetId
	url1 := fmt.Sprintf("https://sheet.zoho.in/api/v2/%s", resourceID)

	const uuidColumnIndex = 1

	postData := url.Values{}
	postData.Set("method", "cell.content.set")
	postData.Set("worksheet_name", sheetName)
	postData.Set("row", strconv.Itoa(row))
	postData.Set("column", strconv.Itoa(uuidColumnIndex))
	postData.Set("content", uuid)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url1, strings.NewReader(postData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Zoho API request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Zoho API error: %s", string(body))
	}

	return nil
}

package zoho

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-ingestion-script/entity"
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

func (s *ZohoService) SetSheetData(ctx context.Context, sheetName string, data entity.ShlokaSheetResponse, startId int) error {
	accessToken := util.GetZohoAccessTokenFromContext(ctx)
	resourceID := s.configuration.ZohoConfig.SheetId
	url1 := fmt.Sprintf("https://sheet.zoho.in/api/v2/%s", resourceID)

	columnMap := map[string]int{
		"ID": 1, "Name (Optional)": 2, "text_sanskrit": 3, "text_english": 4, "translation_english": 5,
		"text_kannada": 6, "translation_kannada": 7, "text_hindi": 8, "translation_hindi": 9,
		"text_telugu": 10, "translation_telugu": 11, "text_bengali": 12, "translation_bengali": 13,
		"text_marathi": 14, "translation_marathi": 15, "text_tamil": 16, "translation_tamil": 17,
		"text_gujarati": 18, "translation_gujarati": 19, "text_odiya": 20, "translation_odiya": 21,
		"text_malayalam": 22, "translation_malayalam": 23, "text_assamese": 24, "translation_assamese": 25,
		"text_punjabi": 26, "translation_punjabi": 27,
	}

	for i, record := range data.Records {
		for key, val := range record {
			colIndex, ok := columnMap[key]
			if !ok {
				continue
			}
			content := fmt.Sprintf("%v", val)
			if strings.TrimSpace(content) == "" {
				continue
			}
			row := startId + i + 1
			postData := url.Values{}
			postData.Set("method", "cell.content.set")
			postData.Set("worksheet_name", sheetName)
			postData.Set("row", strconv.Itoa(row))
			postData.Set("column", strconv.Itoa(colIndex))
			postData.Set("content", content)
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
		}
	}
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

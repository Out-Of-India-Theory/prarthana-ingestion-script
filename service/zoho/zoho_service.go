package zoho

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/configuration"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/entity"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/util"
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

func (s *ZohoService) GetAuthorizationURL(state string) string {
	u, _ := url.Parse(s.configuration.ZohoConfig.AuthUrl)
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", s.configuration.ZohoConfig.ClientId)
	q.Set("scope", "ZohoSheet.dataAPI.READ,ZohoSheet.dataAPI.UPDATE")
	q.Set("redirect_uri", s.configuration.ZohoConfig.RedirectUrl)
	q.Set("access_type", "offline")
	q.Set("prompt", "consent")
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *ZohoService) ExchangeCodeForTokens(ctx context.Context, code string) error {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", s.configuration.ZohoConfig.ClientId)
	data.Set("client_secret", s.configuration.ZohoConfig.ClientSecret)
	data.Set("redirect_uri", s.configuration.ZohoConfig.RedirectUrl)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.configuration.ZohoConfig.TokenUrl, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to exchange code, response: %s", string(body))
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %v", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	s.tokenManager.AccessToken = tokenResp.AccessToken
	s.tokenManager.RefreshToken = tokenResp.RefreshToken
	s.tokenManager.ExpiresAt = expiresAt
	return nil
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

func (s *ZohoService) GetSheetData(ctx context.Context, sheetName string) (*entity.SheetResponse, error) {
	accessToken := util.GetZohoAccessTokenFromContext(ctx)
	url1 := fmt.Sprintf("https://sheet.zoho.in/api/v2/%s", s.configuration.ZohoConfig.SheetId)
	data := url.Values{}
	data.Set("method", "worksheet.records.fetch")
	data.Set("worksheet_name", sheetName)

	// Create a new HTTP request with POST method
	req, err := http.NewRequest(http.MethodPost, url1, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the required headers
	req.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("error response from server: %s", string(body))
	}
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var sheetRecords entity.SheetResponse
	err = json.Unmarshal(bytes, &sheetRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response body: %w", err)
	}
	return &sheetRecords, nil
}

func (s *ZohoService) IsTokenExpired() bool {
	return time.Now().After(s.tokenManager.ExpiresAt)
}

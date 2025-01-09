package zoho

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Out-Of-India-Theory/oit-go-commons/logging"
	"github.com/Out-Of-India-Theory/prarthana-automated-script/configuration"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ZohoAuthService struct {
	logger        *zap.Logger
	configuration *configuration.Configuration
	httpClient    *http.Client
	tokenManager  TokenManager
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type TokenManager struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

func InitZohoService(ctx context.Context,
	configuration *configuration.Configuration,
	httpClient *http.Client,
) *ZohoAuthService {
	return &ZohoAuthService{
		logger:        logging.WithContext(ctx),
		configuration: configuration,
		httpClient:    httpClient,
	}
}

func (s *ZohoAuthService) GetAuthorizationURL(state string) string {
	u, _ := url.Parse(s.configuration.ZohoAuthConfig.AuthUrl)
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", s.configuration.ZohoAuthConfig.ClientId)
	q.Set("scope", "ZohoSheet.dataAPI.READ,ZohoSheet.dataAPI.UPDATE")
	q.Set("redirect_uri", s.configuration.ZohoAuthConfig.RedirectUrl)
	q.Set("access_type", "offline")
	q.Set("prompt", "consent")
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *ZohoAuthService) ExchangeCodeForTokens(ctx context.Context, code string) error {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", s.configuration.ZohoAuthConfig.ClientId)
	data.Set("client_secret", s.configuration.ZohoAuthConfig.ClientSecret)
	data.Set("redirect_uri", s.configuration.ZohoAuthConfig.RedirectUrl)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.configuration.ZohoAuthConfig.TokenUrl, bytes.NewBufferString(data.Encode()))
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

func (s *ZohoAuthService) RefreshAccessToken() error {
	if s.tokenManager.RefreshToken == "" {
		return fmt.Errorf("refresh token not set")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", s.tokenManager.RefreshToken)
	data.Set("client_id", s.configuration.ZohoAuthConfig.ClientId)
	data.Set("client_secret", s.configuration.ZohoAuthConfig.ClientSecret)

	resp, err := http.PostForm(s.configuration.ZohoAuthConfig.TokenUrl, data)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %v", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to refresh token, response: %s", string(body))
	}
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return err
	}
	s.tokenManager.AccessToken = tokenResp.AccessToken
	s.tokenManager.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return nil
}

func (s *ZohoAuthService) GetSheetData(sheetId string, sheetName string) ([]byte, error) {
	if time.Now().After(s.tokenManager.ExpiresAt) {
		if err := s.RefreshAccessToken(); err != nil {
			return nil, err
		}
	}

	url1 := fmt.Sprintf("https://sheet.zoho.com/api/v2/%s", sheetId)
	data := url.Values{}
	data.Set("method", "worksheet.records.fetch")
	data.Set("worksheet_name", sheetName)

	// Create a new HTTP request with POST method
	req, err := http.NewRequest(http.MethodPost, url1, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the required headers
	req.Header.Set("Authorization", "Zoho-oauthtoken "+s.tokenManager.AccessToken)
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
	return ioutil.ReadAll(resp.Body)
}

func (s *ZohoAuthService) IsTokenExpired() bool {
	return time.Now().After(s.tokenManager.ExpiresAt)
}

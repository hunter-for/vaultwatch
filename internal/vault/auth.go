package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// TokenInfo holds metadata about the current Vault token.
type TokenInfo struct {
	LeaseID    string
	Accessor   string
	Policies   []string
	TTL        time.Duration
	Renewable  bool
	ExpireTime time.Time
}

type tokenSelfResponse struct {
	Data struct {
		Accessor   string   `json:"accessor"`
		Policies   []string `json:"policies"`
		TTL        int      `json:"ttl"`
		Renewable  bool     `json:"renewable"`
		ExpireTime string   `json:"expire_time"`
	} `json:"data"`
}

// GetTokenInfo returns metadata about the token currently configured on the client.
func (c *Client) GetTokenInfo() (*TokenInfo, error) {
	req, err := http.NewRequest(http.MethodGet, c.address+"/v1/auth/token/lookup-self", nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from token lookup", resp.StatusCode)
	}

	var result tokenSelfResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	info := &TokenInfo{
		Accessor:  result.Data.Accessor,
		Policies:  result.Data.Policies,
		TTL:       time.Duration(result.Data.TTL) * time.Second,
		Renewable: result.Data.Renewable,
	}

	if et := strings.TrimSpace(result.Data.ExpireTime); et != "" {
		parsed, err := time.Parse(time.RFC3339, et)
		if err == nil {
			info.ExpireTime = parsed
		}
	}

	return info, nil
}

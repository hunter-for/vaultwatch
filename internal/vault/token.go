package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TokenInfo holds metadata about the current Vault token.
type TokenInfo struct {
	Accessor   string
	DisplayName string
	Policies   []string
	TTL        time.Duration
	Renewable  bool
	ExpireTime time.Time
}

type tokenLookupResponse struct {
	Data struct {
		Accessor    string   `json:"accessor"`
		DisplayName string   `json:"display_name"`
		Policies    []string `json:"policies"`
		TTL         int      `json:"ttl"`
		Renewable   bool     `json:"renewable"`
		ExpireTime  string   `json:"expire_time"`
	} `json:"data"`
}

// LookupSelfToken retrieves metadata about the token currently in use.
func (c *Client) LookupSelfToken() (*TokenInfo, error) {
	req, err := http.NewRequest(http.MethodGet, c.address+"/v1/auth/token/lookup-self", nil)
	if err != nil {
		return nil, fmt.Errorf("building token lookup request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token lookup request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token lookup returned status %d", resp.StatusCode)
	}

	var result tokenLookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding token lookup response: %w", err)
	}

	info := &TokenInfo{
		Accessor:    result.Data.Accessor,
		DisplayName: result.Data.DisplayName,
		Policies:    result.Data.Policies,
		TTL:         time.Duration(result.Data.TTL) * time.Second,
		Renewable:   result.Data.Renewable,
	}

	if result.Data.ExpireTime != "" {
		t, err := time.Parse(time.RFC3339, result.Data.ExpireTime)
		if err == nil {
			info.ExpireTime = t
		}
	}

	return info, nil
}

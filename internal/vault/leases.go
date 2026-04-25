package vault

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LeaseEntry represents a single lease returned by the Vault list endpoint.
type LeaseEntry struct {
	LeaseID   string        `json:"lease_id"`
	Renewable bool          `json:"renewable"`
	TTL       time.Duration `json:"-"`
	RawTTL    int           `json:"lease_duration"`
}

// listLeasesResponse mirrors the Vault API response for listing leases.
type listLeasesResponse struct {
	Data struct {
		Keys []string `json:"keys"`
	} `json:"data"`
}

// lookupLeaseResponse mirrors the Vault API response for a lease lookup.
type lookupLeaseResponse struct {
	Data struct {
		ID        string `json:"id"`
		Renewable bool   `json:"renewable"`
		TTL       int    `json:"ttl"`
	} `json:"data"`
}

// ListLeases returns all lease IDs under the given prefix.
func (c *Client) ListLeases(prefix string) ([]string, error) {
	url := fmt.Sprintf("%s/v1/sys/leases/lookup/%s", c.address, prefix)
	req, err := http.NewRequest("LIST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("building list request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing list request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list leases returned status %d", resp.StatusCode)
	}

	var result listLeasesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding list response: %w", err)
	}
	return result.Data.Keys, nil
}

// GetLease fetches details for a single lease by ID.
func (c *Client) GetLease(leaseID string) (*LeaseEntry, error) {
	if leaseID == "" {
		return nil, fmt.Errorf("leaseID must not be empty")
	}

	body, err := jsonBody(map[string]string{"lease_id": leaseID})
	if err != nil {
		return nil, fmt.Errorf("encoding request body: %w", err)
	}

	url := fmt.Sprintf("%s/v1/sys/leases/lookup", c.address)
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, fmt.Errorf("building lookup request: %w", err)
	}
	req.Header.Set("X-Vault-Token", c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing lookup request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusBadRequest {
		return nil, fmt.Errorf("lease not found: %s", leaseID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lease lookup returned status %d", resp.StatusCode)
	}

	var result lookupLeaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding lookup response: %w", err)
	}

	return &LeaseEntry{
		LeaseID:   result.Data.ID,
		Renewable: result.Data.Renewable,
		TTL:       time.Duration(result.Data.TTL) * time.Second,
		RawTTL:    result.Data.TTL,
	}, nil
}

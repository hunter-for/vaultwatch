package vault

import (
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with lease-monitoring helpers.
type Client struct {
	api     *vaultapi.Client
	address string
}

// Lease represents a Vault secret lease with its metadata.
type Lease struct {
	LeaseID       string
	Path          string
	TTL           time.Duration
	Renewable     bool
	ExpireTime    time.Time
}

// NewClient creates and configures a new Vault client.
func NewClient(address, token string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	api, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}

	if token != "" {
		api.SetToken(token)
	}

	return &Client{
		api:     api,
		address: address,
	}, nil
}

// LookupLease retrieves metadata for a given lease ID.
func (c *Client) LookupLease(leaseID string) (*Lease, error) {
	secret, err := c.api.Sys().Lookup(leaseID)
	if err != nil {
		return nil, fmt.Errorf("looking up lease %q: %w", leaseID, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("lease %q not found", leaseID)
	}

	ttlRaw, ok := secret.Data["ttl"]
	if !ok {
		return nil, fmt.Errorf("lease %q missing ttl field", leaseID)
	}

	ttlSeconds, ok := ttlRaw.(float64)
	if !ok {
		return nil, fmt.Errorf("lease %q ttl has unexpected type", leaseID)
	}

	ttl := time.Duration(ttlSeconds) * time.Second
	renewable, _ := secret.Data["renewable"].(bool)

	return &Lease{
		LeaseID:    leaseID,
		TTL:        ttl,
		Renewable:  renewable,
		ExpireTime: time.Now().Add(ttl),
	}, nil
}

// IsHealthy checks whether the Vault server is reachable and unsealed.
func (c *Client) IsHealthy() error {
	health, err := c.api.Sys().Health()
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}
	if health.Sealed {
		return fmt.Errorf("vault is sealed")
	}
	return nil
}

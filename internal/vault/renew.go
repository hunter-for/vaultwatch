package vault

import (
	"context"
	"fmt"
	"net/http"
)

// renewLeaseRequest is the JSON body sent to Vault's lease renew endpoint.
type renewLeaseRequest struct {
	LeaseID   string `json:"lease_id"`
	Increment int    `json:"increment"`
}

// renewLeaseResponse captures the relevant fields from Vault's renew response.
type renewLeaseResponse struct {
	LeaseID       string `json:"lease_id"`
	LeaseDuration int    `json:"lease_duration"`
	Renewable     bool   `json:"renewable"`
}

// RenewLease calls the Vault sys/leases/renew endpoint for the given leaseID.
// increment is the requested TTL extension in seconds (0 uses Vault's default).
func (c *Client) RenewLease(ctx context.Context, leaseID string, increment int) error {
	if leaseID == "" {
		return fmt.Errorf("RenewLease: leaseID must not be empty")
	}

	payload := renewLeaseRequest{
		LeaseID:   leaseID,
		Increment: increment,
	}

	var result renewLeaseResponse
	resp, err := c.doJSON(ctx, http.MethodPut, "/v1/sys/leases/renew", payload, &result)
	if err != nil {
		return fmt.Errorf("RenewLease %s: %w", leaseID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("RenewLease %s: unexpected status %d", leaseID, resp.StatusCode)
	}
	return nil
}

// Package vault provides a thin wrapper around the HashiCorp Vault API client
// tailored for vaultwatch's lease-monitoring use case.
//
// Usage:
//
//	client, err := vault.NewClient(cfg.VaultAddress, cfg.VaultToken)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	if err := client.IsHealthy(); err != nil {
//		log.Fatalf("vault unreachable: %v", err)
//	}
//
//	lease, err := client.LookupLease("database/creds/my-role/abc123")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Lease expires in: %v\n", lease.TTL)
//
// The Client type is the primary entry point. It exposes:
//   - NewClient: construct a configured client
//   - IsHealthy: verify Vault is reachable and unsealed
//   - LookupLease: fetch TTL and renewability for a lease ID
package vault

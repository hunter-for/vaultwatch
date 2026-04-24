# vaultwatch

A CLI tool that monitors HashiCorp Vault secret leases and sends alerts before expiration.

---

## Installation

```bash
go install github.com/yourusername/vaultwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vaultwatch.git
cd vaultwatch && go build -o vaultwatch .
```

---

## Usage

Set your Vault address and token, then run `vaultwatch` with a warning threshold:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.xxxxxxxxxxxxxxxx"

# Alert on leases expiring within the next 24 hours
vaultwatch watch --threshold 24h --alert slack
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--threshold` | `24h` | Warn when lease expires within this duration |
| `--interval` | `5m` | How often to poll Vault for lease status |
| `--alert` | `stdout` | Alert method: `stdout`, `slack`, or `pagerduty` |
| `--config` | `~/.vaultwatch.yaml` | Path to config file |

### Config File Example

```yaml
vault_addr: https://vault.example.com
threshold: 12h
interval: 10m
alert: slack
slack_webhook: https://hooks.slack.com/services/xxx/yyy/zzz
```

---

## Requirements

- Go 1.21+
- HashiCorp Vault 1.12+

---

## License

[MIT](LICENSE) © 2024 Your Name
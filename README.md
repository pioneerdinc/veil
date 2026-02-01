# Veil
```
██╗   ██╗███████╗██╗██╗     
██║   ██║██╔════╝██║██║     
██║   ██║█████╗  ██║██║     
╚██╗ ██╔╝██╔══╝  ██║██║     
 ╚████╔╝ ███████╗██║███████╗
  ╚═══╝  ╚══════╝╚═╝╚══════╝
```
A lightweight, secure CLI secrets manager for developers. Store passwords, API keys, JWT secrets, and database credentials locally with AES-256-GCM encryption.

## Features

- AES-256-GCM encryption with random nonces
- Single binary, no servers, no dependencies
- Group secrets by project/environment
- Global Search - Find secrets across all vaults
- Export to .env - Generate and export secrets directly to environment files
- Secret Generation - Generate strong passwords, API keys, and JWT secrets
- Direct .env Integration - Generate secrets straight into your .env files
## Installation

```bash
# Quick install (Linux/macOS)
curl -fsSL https://raw.githubusercontent.com/pioneerdinc/veil/master/install.sh | bash

# With Go
go install github.com/pioneerdinc/veil/cmd/veil@latest

# Or build from source
git clone https://github.com/pioneerdinc/veil.git
cd veil
go build -o veil ./cmd/veil
```

## Quick Start

```bash
# 1. Generate your master encryption key
veil init
# Save this key: export MASTER_KEY=<your-key>

# 2. Store a secret
veil set production DATABASE_URL "postgresql://user:pass@localhost/db"

# 3. Retrieve it
veil get production DATABASE_URL

# 4. Generate a strong password
veil generate production DB_PASSWORD --length 32

# 5. Export all secrets to .env
veil export production --to .env
```

## Commands

### Basic Operations

```bash
# Store a secret
veil set <vault> <name> <value>

# Retrieve a secret
veil get <vault> <name>

# Delete a secret
veil delete <vault> <name>

# List secrets in a vault
veil list <vault>

# List all vaults
veil vaults
```

### Search

```bash
# Search for secrets across all vaults (case-insensitive, supports wildcards)
veil search API_KEY
veil search "DB_*"
veil search "*SECRET*"
```

### Export

```bash
# Export vault secrets to .env file
veil export production --to .env

# Append to existing file
veil export staging --to .env --append

# Force overwrite
veil export production --to .env --force

# Preview without writing
veil export production --to .env --dry-run
```

### Generate Secrets

```bash
# Generate a strong password (default: 32 chars with symbols)
veil generate <vault> <name>

# Custom length, no symbols
veil generate production API_KEY --length 48 --no-symbols

# Generate API key (uuid, hex, or base64)
veil generate stripe-api STRIPE_KEY --type apikey --format base64 --prefix "sk_live_"

# Generate JWT secret (256 or 512 bits)
veil generate auth JWT_SECRET --type jwt --bits 256
```

### Generate Directly to .env

```bash
# Generate and append to .env in one command
veil generate myapp API_KEY --type apikey --to-env .env

# Overwrite existing key
veil generate myapp API_KEY --to-env .env --force
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MASTER_KEY` | Your 64-character hex encryption key | **Required** |
| `VEIL_DB_PATH` | Path to the SQLite database | `~/.veil.db` |
| `VEIL_STORE_TYPE` | Storage backend | `sqlite` |

## Security

- **AES-256-GCM** encryption with cryptographically random nonces
- **No server** - all data stays on your machine
- **Encrypted at rest** - database file is useless without the master key
- **File permissions** - database and .env files get 0600 permissions
- **No logging** - secrets never appear in logs or stdout (except during generation)

## Workflow Examples

### Development Setup

```bash
# Generate local development credentials
veil generate dev DB_PASSWORD --to-env .env.local
veil generate dev JWT_SECRET --type jwt --to-env .env.local
veil generate dev STRIPE_KEY --type apikey --prefix "sk_test_" --to-env .env.local
```

### Production Deployment

```bash
# Export production secrets for deployment
veil export production --to .env.production
# Deploy with .env.production
```

### Team Onboarding

```bash
# New team member searches for what they need
veil search "STRIPE*"
# Found 1 match:
#   production/STRIPE_SECRET_KEY

# Get the key
veil get production STRIPE_SECRET_KEY
```

### Database Credential Rotation

```bash
# Generate new database password
veil generate production DB_PASSWORD --length 32 --to-env .env --force
# Update database, restart services
```

## Best Practices

1. **Keep your MASTER_KEY safe** - If you lose it, your secrets are gone forever
2. **Use vaults for environments** - `dev`, `staging`, `production`
3. **Never commit .env files** - Add them to .gitignore
4. **Use descriptive names** - `DB_PASSWORD` not `PASS`
5. **Rotate secrets regularly** - Use `veil generate` with `--force` to update

## Technical Details

- **Storage**: SQLite with AES-256-GCM encrypted values
- **Key Derivation**: Master key must be 32 bytes (64 hex characters)
- **Format**: Encrypted values stored as `nonce || ciphertext` (hex encoded)
- **Search**: Case-insensitive SQL LIKE queries on unencrypted vault/name fields

## License

MIT

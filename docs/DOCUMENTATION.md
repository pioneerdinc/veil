# Veil Documentation

> A lightweight, secure CLI secrets manager for developers. Store passwords, API keys, JWT secrets, and database credentials locally with AES-256-GCM encryption.

---

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
- [Commands Reference](#commands-reference)
  - [init](#init)
  - [set](#set)
  - [get](#get)
  - [delete](#delete)
  - [list](#list)
  - [vaults](#vaults)
  - [search](#search)
  - [generate](#generate)
  - [export](#export)
  - [quick](#quick)
  - [reset](#reset)
  - [version](#version)
- [Environment Variables](#environment-variables)
- [Workflow Examples](#workflow-examples)
- [Security](#security)
- [FAQ](#faq)

---

## Installation

### Quick Install (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/pioneerdinc/veil/master/install.sh | bash
```

### With Go

```bash
go install github.com/pioneerdinc/veil/cmd/veil@latest
```

### Build from Source

```bash
git clone https://github.com/pioneerdinc/veil.git
cd veil
go build -o veil ./cmd/veil
```

---

## Quick Start

```bash
# 1. Generate your master encryption key
veil init

# 2. Export the key to your environment
export MASTER_KEY=<your-64-character-hex-key>

# 3. Store a secret
veil set production DATABASE_URL "postgresql://user:pass@localhost/db"

# 4. Retrieve it
veil get production DATABASE_URL

# 5. Generate a strong password
veil generate production DB_PASSWORD --length 32

# 6. Export all secrets to .env
veil export production --to .env
```

---

## Core Concepts

### Master Key

Your `MASTER_KEY` is a 64-character hexadecimal string (32 bytes) used for AES-256-GCM encryption. Every secret you store is encrypted with this key before being written to the database.

**Critical**: If you lose your master key, your secrets are gone forever. There is no recovery mechanism by design.

```bash
# Generate a new key
veil init

# Set it in your environment
export MASTER_KEY=abc123...  # 64 hex characters
```

### Vaults

Vaults are namespaces for organizing secrets. Think of them as folders or environments. Common vault names:

- `dev` - Local development secrets
- `staging` - Staging environment
- `production` - Production secrets
- `myapp` - Project-specific secrets

Vaults are created implicitly when you store a secret in them.

### Secrets

Secrets are key-value pairs stored within vaults. The key (name) is stored in plaintext for searchability, but the value is encrypted at rest.

```
vault/name = encrypted_value
production/DATABASE_URL = <encrypted>
```

---

## Commands Reference

### init

Generate a new master encryption key.

```bash
veil init
```

**Output:**
```
Your new MASTER_KEY is:

a1b2c3d4e5f6...  (64 hex characters)

SAVE THIS KEY! If you lose it, your secrets are gone forever.
Export it to your environment:
export MASTER_KEY=a1b2c3d4e5f6...
```

**Notes:**
- Does not require `MASTER_KEY` to be set
- Will warn if a database already exists (generating a new key makes existing secrets unreadable)
- Run only once per setup

---

### set

Store a secret in a vault.

```bash
veil set <vault> <name> <value>
```

**Arguments:**
| Argument | Description |
|----------|-------------|
| `vault` | The vault name (e.g., `production`, `dev`) |
| `name` | The secret name (e.g., `DATABASE_URL`, `API_KEY`) |
| `value` | The secret value |

**Examples:**
```bash
# Store a database URL
veil set production DATABASE_URL "postgresql://user:pass@localhost/db"

# Store an API key
veil set stripe STRIPE_SECRET_KEY "sk_live_abc123..."

# Store a password
veil set dev DB_PASSWORD "supersecret"
```

**Notes:**
- Overwrites existing secret with the same vault/name
- No output on success (silent success)
- Creates the vault if it doesn't exist

---

### get

Retrieve a secret from a vault.

```bash
veil get <vault> <name>
```

**Arguments:**
| Argument | Description |
|----------|-------------|
| `vault` | The vault name |
| `name` | The secret name |

**Examples:**
```bash
# Get a secret
veil get production DATABASE_URL
# Output: postgresql://user:pass@localhost/db

# Use in a script
DB_URL=$(veil get production DATABASE_URL)

# Pipe to clipboard (macOS)
veil get production API_KEY | pbcopy
```

**Notes:**
- Prints only the secret value (no formatting)
- Exits with error if secret not found
- Exits with error if decryption fails (wrong master key)

---

### delete

Remove a secret from a vault.

```bash
veil delete <vault> <name>
```

**Arguments:**
| Argument | Description |
|----------|-------------|
| `vault` | The vault name |
| `name` | The secret name |

**Examples:**
```bash
# Delete a secret
veil delete production OLD_API_KEY

# Delete and confirm it's gone
veil delete dev TEMP_PASSWORD && echo "Deleted"
```

**Notes:**
- Silent success (no output)
- Permanent deletion - no undo

---

### list

List all secret names in a vault.

```bash
veil list <vault>
```

**Arguments:**
| Argument | Description |
|----------|-------------|
| `vault` | The vault name |

**Examples:**
```bash
# List secrets in production vault
veil list production
# Output:
# DATABASE_URL
# API_KEY
# JWT_SECRET
```

**Notes:**
- Shows only names, never values
- One name per line
- Empty output if vault has no secrets

---

### vaults

List all vaults.

```bash
veil vaults
```

**Examples:**
```bash
veil vaults
# Output:
# dev
# production
# staging
```

**Notes:**
- Shows only vault names
- One vault per line

---

### search

Search for secrets across all vaults.

```bash
veil search <pattern>
```

**Arguments:**
| Argument | Description |
|----------|-------------|
| `pattern` | Search pattern (supports `*` wildcard) |

**Examples:**
```bash
# Find all secrets containing "API"
veil search API
# Output:
# Found 2 matches:
#   production/API_KEY
#   staging/API_KEY

# Wildcard: all secrets starting with DB_
veil search "DB_*"
# Output:
# Found 3 matches:
#   production/DB_PASSWORD
#   production/DB_HOST
#   dev/DB_PASSWORD

# Wildcard: all secrets ending with _SECRET
veil search "*_SECRET"

# Wildcard: contains "STRIPE"
veil search "*STRIPE*"
```

**Notes:**
- Case-insensitive matching
- Shows vault/name pairs, never values
- Supports `*` wildcard at start, end, or both

---

### generate

Generate a cryptographically secure secret and store it in a vault.

```bash
veil generate <vault> <name> [options]
```

**Arguments:**
| Argument | Description |
|----------|-------------|
| `vault` | The vault name |
| `name` | The secret name |

**Options:**

| Option | Description | Default |
|--------|-------------|---------|
| `--type <type>` | Secret type: `password`, `apikey`, `jwt` | `password` |
| `--length <n>` | Password length (8-128) | `32` |
| `--no-symbols` | Alphanumeric only (no special characters) | `false` |
| `--format <fmt>` | API key format: `uuid`, `hex`, `base64` | `base64` |
| `--prefix <str>` | Prefix for API keys (e.g., `sk_live_`) | none |
| `--bits <n>` | JWT secret bits: 128-512 | `256` |
| `--to-env <path>` | Also append to .env file | none |
| `--force` | Overwrite existing key in .env | `false` |

#### Password Generation

```bash
# Default: 32-character password with symbols
veil generate production DB_PASSWORD
# Output:
# Generated secret: aB3$xY9!mN2@pQ5^...
# Stored in production/DB_PASSWORD

# Custom length
veil generate production API_TOKEN --length 48

# Alphanumeric only (no symbols)
veil generate production SIMPLE_KEY --no-symbols
```

#### API Key Generation

```bash
# Base64 format (default)
veil generate stripe STRIPE_KEY --type apikey
# Output: dGhpcyBpcyBhIHRlc3Q=

# UUID format
veil generate myapp CLIENT_ID --type apikey --format uuid
# Output: 550e8400-e29b-41d4-a716-446655440000

# Hex format
veil generate myapp HEX_KEY --type apikey --format hex
# Output: a1b2c3d4e5f6...

# With prefix (e.g., Stripe-style)
veil generate stripe STRIPE_SECRET --type apikey --format base64 --prefix "sk_live_"
# Output: sk_live_dGhpcyBpcyBhIHRlc3Q=
```

#### JWT Secret Generation

```bash
# Default: 256-bit secret
veil generate auth JWT_SECRET --type jwt
# Output: 64-character hex string (256 bits)

# 512-bit for HS512
veil generate auth JWT_SECRET_512 --type jwt --bits 512
# Output: 128-character hex string (512 bits)
```

#### Direct .env Integration

```bash
# Generate and append to .env file
veil generate myapp API_KEY --type apikey --to-env .env
# Stores in vault AND appends to .env

# Overwrite existing key in .env
veil generate myapp API_KEY --type apikey --to-env .env --force
```

---

### export

Export vault secrets to a file.

```bash
veil export <vault> [options]
```

**Arguments:**
| Argument | Description |
|----------|-------------|
| `vault` | The vault name |

**Options:**

| Option | Description | Default |
|--------|-------------|---------|
| `--to <path>` | Output file path | `.env` |
| `--force` | Overwrite existing file | `false` |
| `--append` | Append to existing file | `false` |
| `--dry-run` | Preview without writing | `false` |
| `--backup` | Create backup before overwriting | `false` |
| `--include <pattern>` | Only export matching keys | all |
| `--exclude <pattern>` | Skip matching keys | none |

**Examples:**

```bash
# Export to .env (default)
veil export production --to .env

# Export to custom file
veil export production --to .env.production

# Append to existing file
veil export staging --to .env --append

# Preview changes without writing
veil export production --to .env --dry-run
# Output:
# DRY RUN - No files will be modified
# Would write to .env:
#   + DATABASE_URL
#   + API_KEY
# Summary: 2 new, 0 updates, 0 skipped

# Force overwrite
veil export production --to .env --force

# Create backup before overwriting
veil export production --to .env --force --backup

# Filter exports
veil export production --to .env --include "DB_*"
veil export production --to .env --exclude "*_SECRET"
```

**Output format:**
```env
# Added by veil on 2024-01-15T10:30:00Z
DATABASE_URL=postgresql://user:pass@localhost/db
API_KEY=sk_live_abc123
JWT_SECRET=deadbeef...
```

---

### quick

Generate ephemeral secrets without storing them in the vault. Perfect for one-off needs.

```bash
veil quick [type] [options]
```

**Arguments:**
| Argument | Description |
|----------|-------------|
| `type` | Secret type: `password`, `apikey`, `jwt`, `hex`, `base64`, `uuid` | `password` |

**Options:**

| Option | Description | Default |
|--------|-------------|---------|
| `--length <n>` | Password length (8-128) | `32` |
| `--no-symbols` | Alphanumeric only | `false` |
| `--format <fmt>` | API key format: `uuid`, `hex`, `base64` | `base64` |
| `--prefix <str>` | Prefix for generated value | none |
| `--bits <n>` | JWT secret bits | `256` |
| `--count <n>` | Generate multiple secrets | `1` |
| `--to <path>` | Append to .env file | stdout |
| `--name <KEY>` | Env variable name (required with `--to`) | none |
| `--force` | Overwrite existing key in .env | `false` |
| `--template <str>` | Custom output format | none |
| `--batch <file>` | Generate from JSON config | none |

**Key differences from `generate`:**
- No vault required
- No master key required
- Secrets are NOT stored in the database
- Displayed once and forgotten

**Examples:**

```bash
# Quick password (default)
veil quick
# Output: Generated password: aB3$xY9!mN2@pQ5^...

# Quick password, alphanumeric
veil quick password --no-symbols

# Quick UUID
veil quick uuid
# Output: Generated apikey: 550e8400-e29b-41d4-a716-446655440000

# Quick hex string
veil quick hex --length 16
# Output: Generated apikey: a1b2c3d4e5f6a1b2...

# Quick base64
veil quick base64

# Quick JWT secret
veil quick jwt --bits 512

# Generate multiple at once
veil quick password --count 5
# Output:
# Generated 5 passwords:
# 1. aB3$xY9!mN2@pQ5^...
# 2. xK7#mL4&nP8*qR2!...
# ...

# Write directly to .env file
veil quick apikey --to .env --name API_KEY

# Custom template
veil quick password --template "Password: {value} (generated at {timestamp})"
```

#### Batch Generation

Create multiple secrets at once using a JSON configuration file.

**batch.json:**
```json
{
  "secrets": [
    {
      "name": "DB_PASSWORD",
      "type": "password",
      "length": 32
    },
    {
      "name": "API_KEY",
      "type": "apikey",
      "format": "base64",
      "prefix": "sk_"
    },
    {
      "name": "JWT_SECRET",
      "type": "jwt",
      "bits": 256
    }
  ]
}
```

```bash
# Generate all secrets from config
veil quick --batch batch.json
# Output:
# Batch: batch.json
# Generated 3 secrets:
#   DB_PASSWORD: aB3$xY9!mN2@...
#   API_KEY: sk_dGhpcyBpcyBh...
#   JWT_SECRET: deadbeef...

# Generate and append all to .env
veil quick --batch batch.json --to .env
```

---

### reset

Delete all secrets and start fresh. Use when you've lost your master key.

```bash
veil reset
```

**Behavior:**
1. Displays a warning
2. Requires typing `yes` to confirm
3. Wipes all data from the database

```bash
veil reset
# Output:
# ⚠️  WARNING: This will permanently DELETE ALL SECRETS in the database.
# Are you sure? (type 'yes' to confirm): yes
# Database wiped successfully. You can now run 'veil init' to start over.
```

**Notes:**
- Does not require master key (since you may have lost it)
- Irreversible
- After reset, run `veil init` to generate a new key

---

### version

Show version information.

```bash
veil version
# Output: veil version 1.0.0
```

Also works with flags:
```bash
veil --version
veil -v
```

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MASTER_KEY` | Your 64-character hex encryption key | **Required** for most commands |
| `VEIL_DB_PATH` | Path to the SQLite database | `~/.veil.db` |
| `VEIL_STORE_TYPE` | Storage backend type | `sqlite` |

**Example .bashrc / .zshrc:**
```bash
export MASTER_KEY="your-64-character-hex-key-here"
export VEIL_DB_PATH="$HOME/.config/veil/secrets.db"
```

---

## Workflow Examples

### Development Setup

Bootstrap a new project with generated secrets:

```bash
# Generate all dev secrets directly to .env.local
veil generate dev DB_PASSWORD --to-env .env.local
veil generate dev JWT_SECRET --type jwt --to-env .env.local
veil generate dev STRIPE_KEY --type apikey --prefix "sk_test_" --to-env .env.local
```

### Production Deployment

Export production secrets for deployment:

```bash
# Export to production env file
veil export production --to .env.production

# Or inject directly into environment
export $(veil export production --to /dev/stdout | xargs)
```

### CI/CD Integration

```bash
# In your CI script
export MASTER_KEY="${MASTER_KEY_SECRET}"  # From CI secrets
veil export production --to .env
```

### Team Onboarding

New team member searches for what they need:

```bash
# Find Stripe-related secrets
veil search "STRIPE*"
# Output:
# Found 2 matches:
#   production/STRIPE_SECRET_KEY
#   staging/STRIPE_SECRET_KEY

# Get the one they need
veil get staging STRIPE_SECRET_KEY
```

### Secret Rotation

Rotate a compromised credential:

```bash
# Generate new password and update .env
veil generate production DB_PASSWORD --length 32 --to-env .env --force

# Update database with new password
# Restart services
```

### Backup Existing Secrets

```bash
# Export with backup
veil export production --to .env.production --force --backup
# Creates .env.production.backup.20240115-103000
```

---

## Security

### Encryption

- **Algorithm**: AES-256-GCM (Galois/Counter Mode)
- **Key size**: 256 bits (32 bytes, 64 hex characters)
- **Nonce**: Random 12-byte nonce per encryption
- **Storage format**: `nonce || ciphertext` (hex encoded)

### What's Encrypted

| Data | Encrypted? |
|------|------------|
| Secret values | Yes |
| Vault names | No (for searchability) |
| Secret names | No (for searchability) |
| Database file | Partially (values only) |

### File Permissions

- Database: `0600` (owner read/write only)
- Exported .env files: `0600`

### What Veil Does NOT Do

- Sync to cloud (all data stays local)
- Store or transmit your master key
- Log secret values
- Phone home or collect analytics

---

## FAQ

### What if I lose my MASTER_KEY?

Your secrets are gone forever. There is no recovery mechanism. This is by design - if we could recover your secrets, so could an attacker.

**Recommendation**: Store your master key in a password manager like 1Password or Bitwarden.

### Can I change my MASTER_KEY?

Not directly. You would need to:
1. Export all secrets to a file
2. Run `veil reset`
3. Run `veil init` to generate a new key
4. Re-import all secrets

### Where is my data stored?

By default: `~/.veil.db` (SQLite database)

You can change this with: `export VEIL_DB_PATH=/path/to/secrets.db`

### Can I sync across devices?

Not built-in. However, you can:
1. Store your `~/.veil.db` in a synced folder (Dropbox, iCloud, etc.)
2. Set `VEIL_DB_PATH` to that location
3. Use the same `MASTER_KEY` on all devices

**Warning**: This means your encrypted database is on a cloud service. While values are encrypted, vault and secret names are visible.

### Is there a GUI?

No. Veil is CLI-only by design. This keeps it simple, scriptable, and auditable.

### How does this compare to 1Password/Bitwarden?

Those are designed for human users (browser extensions, autofill, etc.). Veil is designed for developers and machines (CLI, .env files, scripts).

### Can I use this in production?

Yes, but understand the trade-offs:
- No built-in backup/sync
- Single point of failure (the master key)
- No access controls or audit logs

For production, consider also:
- Backing up your database file
- Storing your master key securely (HSM, vault, etc.)
- Using environment-specific vaults

### What happens if my database is stolen?

An attacker would have:
- Your vault names (plaintext)
- Your secret names (plaintext)
- Your secret values (encrypted)

Without your master key, the values are useless. AES-256-GCM is considered unbreakable with current technology.

---

## Getting Help

```bash
# Show all commands
veil --help

# Show version
veil --version
```

**Issues & Contributions**: [github.com/pioneerdinc/veil](https://github.com/pioneerdinc/veil)

---

*Veil - Your secrets, encrypted. Your machine, not ours.*

# DQ Vault Deployment Guide

This guide provides instructions for deploying DQ Vault using Docker and Docker Compose with a simplified setup process.

## Prerequisites

- Docker and Docker Compose installed on your system
- Access to the server where you want to deploy DQ Vault

## Quick Start

### 1. Prepare the Environment

Create an empty `.env` file in the project root:
```bash
touch .env
```

### 2. Build and Start the Vault

Build the Docker image and start the container:
```bash
docker build -t dq .
docker-compose up -d
```

### 3. Initialize Vault (First Time Setup)

Get the container ID:
```bash
docker ps
```

Enter the container:
```bash
docker exec -it <container-id> sh
```

Inside the container, initialize the vault:
```bash
# Clean up existing data
rm -rf /var/lib/vault/data
mkdir -p /var/lib/vault/data
chmod -R 777 /var/lib/vault/data

# Initialize vault
vault operator init
```

**Important:** Save the unseal keys and root token from the initialization output in a secure location.

### 4. Unseal the Vault

Run the unseal command three times with different keys:
```bash
vault operator unseal
# Enter unseal key 1

vault operator unseal
# Enter unseal key 2

vault operator unseal
# Enter unseal key 3
```

Verify the vault is unsealed:
```bash
vault status
```

### 5. Configure the Plugin

Export the root token:
```bash
export VAULT_TOKEN=<your-root-token>
```

Set up the DQ Vault plugin:
```bash
# Disable existing plugin (if any)
vault secrets disable /dq

# Calculate plugin hash
export SHA256=$(sha256sum "/vault/plugins/dq-vault" | cut -d' ' -f1)

# Register plugin
vault write sys/plugins/catalog/dq sha_256="${SHA256}" command="dq-vault"

# Enable plugin
vault secrets enable -path="dq" -plugin-name="dq" plugin

# Create policy
echo 'path "dq/*" { capabilities = ["read", "update"] }' > /tmp/dq-policy.hcl
vault policy write dq-policy /tmp/dq-policy.hcl

# Create application token
vault token create -policy=dq-policy -renewable=false -no-default-policy
```

**Important:** Save the token from the last command - this is your application token for API access.

### 6. Install curl (for testing)

Inside the container:
```bash
apk add curl
```

### 7. Register a User

Replace `<register-token>` and `<mnemonic>` with your actual values:
```bash
curl -k --location 'http://127.0.0.1:8200/v1/dq/register' \
  --header 'X-Vault-Token: <register-token>' \
  --header 'Content-Type: application/json' \
  --data '{"username": "root","mnemonic": "<mnemonic>"}'
```

Save the returned UUID for future operations.

## Configuration Files

### vault.hcl
The main Vault configuration file:
```hcl
disable_mlock = "true"
plugin_directory = "/vault/plugins"

listener "tcp" {
  address = "0.0.0.0:8200"
  tls_disable = 1
}

storage "file" {
  path = "/var/lib/vault/data"
}
log_level="Debug"
ui = false
api_addr = "http://127.0.0.1:8200"
```

### docker-compose.yml
Docker Compose configuration:
```yaml
services:
  vault:
    image: dq
    ports:
      - 8200:8200
    volumes:
      - /var/vault:/var/lib/vault
      - /var/log:/var/log
      - ./vault.hcl:/etc/vault.d/vault.hcl
    command:
      - vault
      - server
      - -config=/etc/vault.d/vault.hcl
      - -log-level=info
    cap_add:
      - IPC_LOCK
    env_file: ['.env']
```

## Usage

### Generate Address
```bash
vault write dq/address uuid="<uuid>" path="<path>" coinType=<coin-type>
```

Example for Solana:
```bash
vault write dq/address uuid="cql4aua0negc60hrrshg" path="m/44'/501'/0'" coinType=501
```

### Sign Transaction
```bash
vault write dq/signature uuid="<uuid>" path="<path>" payload="<payload>" coinType=<coin-type>
```

Example for Solana:
```bash
vault write dq/signature uuid="cqo63c4u54ms4l5mffpg" path="m/44'/501'/0'" \
  payload='{"rawTxHex": "010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000508e9589e3c33514a47f3f8e4655fda1908915b4938274253256a63c000a77761ee8759142feccf57082eb2032543e7f68d29a666fc25992c024899ad234cce0164d40ec3100887c28cc85aceb030d588343f287696f469accf82f208e018d06bc300000000000000000000000000000000000000000000000000000000000000001b576941a405b0a8e641a08eb009ac674459306d5dc22f202d992768eb968d273b442cb3912157f13a933d0134282d032b5ffecd01a2dbf1b7790608df002ea78c97258f4e2489f1bb3d1029148e0d830b5a1399daff1084048e7bd8dbe9f85906ddf6e1d765a193d9cbe146ceeb79ac1cb485ed5f5b37913a8cf5857eff00a9b2c97b742167eeea0460cf0faa2388ff7fc692e70ece259dac03c436b04d907202060600010405030700070302010009031027000000000000", "tokenAddress":"4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU"}' \
  coinType=501
```

## Restarting Vault

For subsequent starts (not first-time setup):
```bash
docker-compose up -d
docker exec -it <container-id> sh

# Set permissions
chmod -R 777 /var/lib/vault/data

# Unseal with the same 3 keys
vault operator unseal
vault operator unseal
vault operator unseal
```

## Troubleshooting

### View Logs
```bash
docker-compose logs -f
```

### Check Vault Status
```bash
docker exec -it <container-id> vault status
```

### Container Access
```bash
docker exec -it <container-id> sh
```

## Security Notes

- Store unseal keys and root token securely
- Use environment variables for sensitive data
- Consider using TLS in production
- Regularly backup vault data
- Use proper access policies for production use

## Support

For detailed API documentation and usage examples, refer to the [DQ Vault plugin usage documentation](https://deqode.github.io/dq-vault/docs/guides/plugin-usage/). 
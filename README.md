# DQ Vault - Hashicorp Vault Cryptocurrency Plugin

<p align="center"><img src="https://deqode.github.io/dq-vault/assets/images/vault-dq-192x192-202df720d6d8d239d0fbf4cdc208c1c8.png"></p>

![GitHub](https://img.shields.io/github/license/deqode/dq-vault)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/deqode/dq-vault)](https://pkg.go.dev/github.com/github.com/deqode/dq-vault)
[![Go Report Card](https://goreportcard.com/badge/github.com/deqode/dq-vault)](https://goreportcard.com/report/github.com/deqode/dq-vault)
![GitHub last commit](https://img.shields.io/github/last-commit/deqode/codeanalyser)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/deqode/dq-vault)

## Introduction

DQ Vault is a plugin for [Hashicorp Vault](https://vaultproject.io/) that allows it to generate cryptocurrency addresses and sign transactions. This vault plugin stores a user's mnemonic inside vault in an encrypted manner. The plugin uses this stored mnemonic to derive a private key based on an HD wallet path provided by the user and signs a raw transaction given as input using that private key.

All this process happens inside the vault and the user never knows the mnemonic (unless he has provided it manually) or the private key derived. All he needs to do is give a raw transaction as input and the vault returns a signed transaction. A particular user is identified in the vault using a UUID generated when the user is initially registered in the vault.

## Supported Currencies

- Bitcoin (BTC)
- Ethereum (ETH)
- Litecoin (LTC)
- Dogecoin (DOGE)
- Ripple (XRP)
- Stellar (XLM)
- Solana (SOL)
- Bitshares (BTS)
- Tron (TRX)

## Quick Start

### Prerequisites

- Docker and Docker Compose installed on your system

### 1. Simple Deployment

Run the deployment script:
```bash
./deploy.sh
```

### 2. Manual Deployment

Alternatively, you can deploy manually:

```bash
# Create .env file
touch .env

# Build and start
docker build -t dq .
docker-compose up -d

# Get container ID
docker ps

# Initialize vault (first time only)
docker exec -it <container-id> sh
vault operator init  # Save the keys and root token!
vault operator unseal  # Run 3 times with different keys
```

### 3. Complete Setup Guide

For detailed setup instructions, see [DEPLOYMENT.md](DEPLOYMENT.md)

## Configuration Files

- `vault.hcl` - Vault server configuration
- `docker-compose.yml` - Docker deployment configuration
- `Dockerfile` - Container build configuration

## API Usage

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
  payload='{"rawTxHex": "...", "tokenAddress": "..."}' \
  coinType=501
```

For detailed API documentation and usage examples, see the [plugin usage guide](https://deqode.github.io/dq-vault/docs/guides/plugin-usage/)

## Documentation

- [Deployment Guide](DEPLOYMENT.md) - Comprehensive deployment instructions
- [Official Documentation](https://deqode.github.io/dq-vault/) - Full API and usage documentation

## Architecture

The deployment uses a multi-stage Docker build:
1. **Builder stage**: Compiles the Go application using Go 1.21
2. **Runtime stage**: HashiCorp Vault 1.15.6 with the DQ plugin installed

## Troubleshooting

### View Logs
```bash
docker-compose logs -f
```

### Check Vault Status
```bash
docker exec -it <container-id> vault status
```

### Access Container
```bash
docker exec -it <container-id> sh
```

## Roles

There are two roles communicating with vault:

1. **Admin**: The one who sets up the vault.
2. **Application Server**: The one who uses vault to read and update data.

The application server can communicate with a vault server using API requests/calls. Both CLI commands and API call methods have been included in this guide.

## License

```
Copyright 2021, DeqodeLabs (https://deqode.com/)

Licensed under the MIT License(the "License");
```

<p align="center"><img src="https://deqode.com/wp-content/uploads/presskit-logo.png" width="400"></p>
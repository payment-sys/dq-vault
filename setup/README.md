# DQ Vault Postman Collection

This directory contains a comprehensive Postman collection and environment for testing the DQ Vault API endpoints.

## Files

- `vault_postman_collection.json` - Complete Postman collection with all DQ Vault endpoints
- `vault_postman_environment.json` - Environment variables for the collection
- `README.md` - This documentation file

## Quick Start

### 1. Import into Postman

1. Open Postman
2. Click **Import** button
3. Import both files:
   - `vault_postman_collection.json`
   - `vault_postman_environment.json`

### 2. Configure Environment

1. Select "DQ Vault Environment" from the environment dropdown
2. Set the required variables:
   - `vault_url`: Your Vault server URL (default: `http://localhost:8200`)
   - `vault_token`: Your Vault authentication token

### 3. Get Your Vault Token

You can get a Vault token in several ways:

#### Using Root Token (Development)
```bash
# Use the root token from vault init
export VAULT_TOKEN=<your-root-token>
```

#### Using DQ Policy Token
```bash
# Create a token with DQ policy
vault token create -policy=dq-policy -renewable=false -no-default-policy
```

#### From Kubernetes Deployment
```bash
# Get from the staging deployment credentials
kubectl port-forward svc/dq-vault-staging 8200:8200 -n dq-vault-staging
# Use the DQ_TOKEN from the deployment credentials
```

## API Endpoints Overview

### 1. Authentication & Info
- **GET** `/v1/dq/info` - Get plugin information and help

### 2. User Management
- **POST** `/v1/dq/register` - Register user with mnemonic
  - With provided mnemonic
  - Auto-generate mnemonic (leave mnemonic field empty)

### 3. Address Generation
Generate addresses for supported cryptocurrencies:
- **POST** `/v1/dq/address` - Generate address for any supported coin

| Cryptocurrency | Coin Type | Derivation Path | Example |
|---------------|-----------|-----------------|---------|
| Bitcoin (BTC) | 0 | `m/44'/0'/0'/0/0` | Generate Bitcoin Address |
| Litecoin (LTC) | 2 | `m/44'/2'/0'/0/0` | Generate Litecoin Address |
| Dogecoin (DOGE) | 3 | `m/44'/3'/0'/0/0` | Generate Dogecoin Address |
| Ethereum (ETH) | 60 | `m/44'/60'/0'/0/0` | Generate Ethereum Address |
| Ripple (XRP) | 144 | `m/44'/144'/0'/0/0` | Generate Ripple Address |
| Stellar (XLM) | 148 | `m/44'/148'/0'/0/0` | Generate Stellar Address |
| Tron (TRX) | 195 | `m/44'/195'/0'/0/0` | Generate Tron Address |
| Solana (SOL) | 501 | `m/44'/501'/0'` | Generate Solana Address |
| Bitshares (BTS) | 69 | `m/44'/69'/0'/0/0` | Generate Bitshares Address |

### 4. Transaction Signing
- **POST** `/v1/dq/signature` - Sign transactions for any supported coin

### 5. Development Mode
- Test with testnets by setting `isDev: true` in requests

## Workflow Example

### Complete User Workflow

1. **Register User**
   ```json
   POST /v1/dq/register
   {
     "username": "test-user",
     "mnemonic": "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
     "passphrase": "test-passphrase"
   }
   ```
   
   Response: `{"data": {"uuid": "user-uuid-123"}}`

2. **Generate Addresses**
   ```json
   POST /v1/dq/address
   {
     "uuid": "user-uuid-123",
     "path": "m/44'/60'/0'/0/0",
     "coinType": 60
   }
   ```
   
   Response: `{"data": {"address": "0x..."}}`

3. **Sign Transaction**
   ```json
   POST /v1/dq/signature
   {
     "uuid": "user-uuid-123",
     "path": "m/44'/60'/0'/0/0",
     "coinType": 60,
     "payload": "{\"nonce\":42,\"value\":1000000000000000000,\"gasLimit\":21000,\"gasPrice\":20000000000,\"to\":\"0x742d35Cc6634C0532925a3b8D359A5C5119e32C8\",\"data\":\"0x\",\"chainId\":1}"
   }
   ```
   
   Response: `{"data": {"signature": "0x..."}}`

## Request Parameters

### Common Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `uuid` | string | Yes (for address/sign) | User UUID from registration |
| `path` | string | Yes (for address/sign) | BIP-44 derivation path |
| `coinType` | integer | Yes (for address/sign) | SLIP-44 coin type |
| `isDev` | boolean | No | Enable development/testnet mode |

### Registration Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `username` | string | No | Username for the user |
| `mnemonic` | string | No | BIP-39 mnemonic (auto-generated if empty) |
| `passphrase` | string | No | Optional passphrase for additional security |

### Signing Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `payload` | string | Yes | JSON string with transaction data |

## Payload Examples

### Ethereum Transaction
```json
{
  "nonce": 42,
  "value": 1000000000000000000,
  "gasLimit": 21000,
  "gasPrice": 20000000000,
  "to": "0x742d35Cc6634C0532925a3b8D359A5C5119e32C8",
  "data": "0x",
  "chainId": 1
}
```

### Solana Transaction
```json
{
  "rawTxHex": "010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000508e9589e3c33514a47f3f8e4655fda1908915b4938274253256a63c000a77761ee...",
  "tokenAddress": "4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU"
}
```

### Bitcoin Transaction
```json
{
  "inputs": [
    {
      "txid": "abc123",
      "vout": 0,
      "amount": 100000
    }
  ],
  "outputs": [
    {
      "address": "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa",
      "amount": 50000
    }
  ],
  "fee": 1000
}
```

## Testing Features

### Automated Tests
The collection includes automated tests that:
- Validate response status codes
- Check response structure
- Verify data types
- Auto-populate environment variables (like `user_uuid`)

### Global Scripts
- **Pre-request**: Validates environment variables
- **Tests**: Checks response time and content type

## Environment Variables

The environment includes pre-configured variables for:
- **Server Configuration**: URLs, tokens
- **User Data**: usernames, mnemonics, passphrases
- **Derivation Paths**: For all supported cryptocurrencies
- **Coin Types**: SLIP-44 standard coin type numbers
- **Chain IDs**: For Ethereum networks (mainnet, testnets)
- **Sample Data**: Test addresses and token addresses

## Security Notes

⚠️ **Important Security Considerations:**

1. **Test Mnemonic**: The default mnemonic is for testing only - NEVER use in production
2. **Token Security**: Never commit real Vault tokens to version control
3. **Production Use**: Always use unique, securely generated mnemonics for production
4. **Environment Isolation**: Use separate environments for development, staging, and production

## Troubleshooting

### Common Issues

1. **Authentication Errors**
   - Verify `vault_token` is set correctly
   - Check token permissions (needs access to `dq/*` paths)

2. **UUID Not Found**
   - Ensure user is registered first
   - Verify `user_uuid` environment variable is set

3. **Invalid Derivation Path**
   - Check path format matches BIP-44 standard
   - Solana uses 4-part paths (`m/44'/501'/0'`)

4. **Plugin Not Found**
   - Verify DQ plugin is installed and enabled
   - Check `vault secrets list` for `dq/` mount

### Debug Tips

1. Check Vault logs for detailed error messages
2. Use the "Get Plugin Info" request to verify plugin status
3. Enable Postman console for request/response debugging
4. Verify environment variables are properly set

## Support

For additional help:
- Check the [DQ Vault documentation](https://deqode.github.io/dq-vault/)
- Review the API code in `/api/` directory
- Run the test suite: `make test`

## Version Information

- **Collection Version**: 1.0.0
- **API Version**: v1
- **Supported Vault Version**: 1.15.6+ 
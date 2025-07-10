# Vault Initialization Guide

This guide explains how to initialize and unseal HashiCorp Vault after deployment.

## Why This is Needed

HashiCorp Vault follows a **security-first design**:
1. **Deployed** → Vault container starts but is "uninitialized"
2. **Initialized** → Generate master keys and root token (one-time)
3. **Unsealed** → Provide 3 of 5 keys to unlock Vault (after every restart)
4. **Ready** → Vault is operational and can serve requests

## Post-Deployment Steps

### Step 1: Check Pod Status
```bash
# Verify pod is running and ready
kubectl get pods -n dq-vault-staging -l app.kubernetes.io/name=dq-vault

# Should show: 1/1 Running (not 0/1)
```

### Step 2: Access Vault
```bash
# Port forward to access Vault
kubectl port-forward svc/dq-vault-staging 8200:8200 -n dq-vault-staging
```

### Step 3: Check Initialization Status
```bash
# Check if Vault is initialized
curl -s http://localhost:8200/v1/sys/init | jq

# Response: {"initialized": false} = Needs initialization
# Response: {"initialized": true} = Already initialized
```

### Step 4: Initialize Vault (First Time Only)
```bash
# Initialize with 5 keys, requiring 3 to unseal
curl -X POST http://localhost:8200/v1/sys/init \
  -H "Content-Type: application/json" \
  -d '{"secret_shares": 5, "secret_threshold": 3}' | jq

# SAVE THE OUTPUT! You'll get:
# - 5 unseal keys (keys_base64)
# - 1 root token (root_token)
```

**⚠️ CRITICAL: Save the keys and root token securely!**

### Step 5: Unseal Vault
```bash
# Unseal with any 3 of the 5 keys
curl -X POST http://localhost:8200/v1/sys/unseal \
  -H "Content-Type: application/json" \
  -d '{"key": "UNSEAL_KEY_1"}'

curl -X POST http://localhost:8200/v1/sys/unseal \
  -H "Content-Type: application/json" \
  -d '{"key": "UNSEAL_KEY_2"}'

curl -X POST http://localhost:8200/v1/sys/unseal \
  -H "Content-Type: application/json" \
  -d '{"key": "UNSEAL_KEY_3"}'

# Check if unsealed
curl -s http://localhost:8200/v1/sys/health | jq
```

### Step 6: Configure DQ Plugin
```bash
# Set the root token
export VAULT_TOKEN="your-root-token"

# Disable existing plugin (if any)
curl -X DELETE http://localhost:8200/v1/sys/mounts/dq \
  -H "X-Vault-Token: $VAULT_TOKEN"

# Calculate plugin hash
export SHA256=$(kubectl exec -n dq-vault-staging deployment/dq-vault-staging -- sha256sum /vault/plugins/dq-vault | cut -d' ' -f1)

# Register plugin
curl -X POST http://localhost:8200/v1/sys/plugins/catalog/secret/dq \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"sha_256\": \"$SHA256\", \"command\": \"dq-vault\"}"

# Enable plugin
curl -X POST http://localhost:8200/v1/sys/mounts/dq \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"type": "dq"}'

# Create policy
curl -X POST http://localhost:8200/v1/sys/policies/acl/dq-policy \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"policy": "path \"dq/*\" { capabilities = [\"read\", \"update\"] }"}'

# Create application token
curl -X POST http://localhost:8200/v1/auth/token/create \
  -H "X-Vault-Token: $VAULT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"policies": ["dq-policy"], "renewable": false, "no_default_policy": true}' | jq
```

### Step 7: Test DQ Plugin
```bash
# Register a user (save the UUID!)
curl -X POST http://localhost:8200/v1/dq/register \
  -H "X-Vault-Token: your-app-token" \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "mnemonic": "your twelve word mnemonic phrase here for testing purposes only"}' | jq

# Generate an address
curl -X POST http://localhost:8200/v1/dq/address \
  -H "X-Vault-Token: your-app-token" \
  -H "Content-Type: application/json" \
  -d '{"uuid": "your-uuid", "path": "m/44'\'''/501'\'''/0'\''", "coinType": 501}' | jq
```

## Troubleshooting

### Pod Not Ready
```bash
# Check pod logs
kubectl logs -n dq-vault-staging deployment/dq-vault-staging

# Should see: "core: security barrier not initialized" (normal)
```

### Health Check Failing
```bash
# Test health endpoint directly
curl -s "http://localhost:8200/v1/sys/health?standbyok=true&uninitcode=204&sealedcode=204"

# Should return HTTP 204 even when uninitialized
```

### Vault Sealed After Restart
```bash
# This is normal! Vault seals on restart for security
# Repeat Step 5 (Unseal) with the same 3 keys
```

## Important Security Notes

1. **Save unseal keys and root token securely** - You cannot recover them
2. **Vault seals on every restart** - This is intentional security behavior
3. **Use separate tokens for applications** - Don't use the root token for apps
4. **Rotate tokens regularly** - Especially in production environments
5. **Store keys in a secure vault/safe** - Not in plain text files

## Automation Options

For production environments, consider:
- **Vault Auto-unseal** with cloud KMS
- **Kubernetes secrets** for storing tokens (encrypted at rest)
- **Init containers** for automated plugin setup
- **Backup strategies** for Vault data

---

**Next Steps**: Once initialized, your DQ Vault is ready to generate cryptocurrency addresses and sign transactions! 
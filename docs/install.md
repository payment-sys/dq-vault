# How to Run DQ Vault

## Step 1:
Connect to the relevant server via SSH.

## Step 2:
Make the following changes to the specified files:

### `/etc/vault.d/vault.hcl`:
```
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

### `./docker-compose.yml`:
```yaml
services:
  vault:
    image: dq
    ports:
      - 8200:8200
    volumes:
      - /var/vault:/var/lib/vault
      - /var/log:/var/log
      - /etc/vault.d:/etc/vault.d
    command:
      - vault
      - server
      - -config=/etc/vault.d/vault.hcl
      - -log-level=info
    cap_add:
      - IPC_LOCK
    env_file: ['.env']
```

## Step 3:
Run the following commands in order:
```sh
touch .env
docker build -t dq .
docker-compose up
```
Once executed, you should see the message `==> Vault server started! Log data will stream in below`.

## Step 4:
Open a new terminal tab and run `docker ps`. Find the container named `dq` and copy its `CONTAINER ID`. Then, run the following command with the copied `CONTAINER ID`:
```sh
docker exec -it <container-id> sh
```
In the new shell, execute the necessary commands to use the Vault.

## Step 5:
### Case 1: If setting up the Vault for the first time, run the following commands in order:
```sh
rm -R /var/lib/vault/data
mkdir /var/lib/vault/data
chmod -R 777 /var/lib/vault/data
vault operator init
```
The output of the last command will be as follows; save it somewhere secure:
```
Unseal Key 1: <unseal-key-1>
Unseal Key 2: <unseal-key-2>
Unseal Key 3: <unseal-key-3>
Unseal Key 4: <unseal-key-4>
Unseal Key 5: <unseal-key-5>

Initial Root Token: <initial-root-token>
```

### Case 2: If the Vault was previously set up, run only the following command:
```sh
chmod -R 777 /var/lib/vault/data
```

## Step 6:
To continue the setup, unseal the Vault by running the following command three times:
```sh
vault operator unseal
```
Each time, you will be prompted to enter one of the `unseal keys` obtained during the init process. After three runs, execute:
```sh
vault status
```
Ensure the `Sealed` parameter is `false`. The output should look like:
```
Key             Value
---             -----
Seal Type       shamir
Initialized     true
Sealed          false
Total Shares    5
Threshold       3
Version         1.15.6
Build Date      2024-02-28T17:07:34Z
Storage Type    file
Cluster Name    vault-cluster-e5f4c07f
Cluster ID      28aee9a5-3926-0046-c293-42294bec147c
HA Enabled      false
```

## Step 7:
Export the `initial root token` obtained during the init process:
```sh
export VAULT_TOKEN=<initial-root-token>
```

## Step 8:
Run the following commands in order:
```sh
vault secrets disable /dq

export SHA256=$(sha256sum "/vault/plugins/dq-vault" | cut -d' ' -f1)

vault write sys/plugins/catalog/dq sha_256="${SHA256}" command="dq-vault"

vault secrets enable -path="dq" -plugin-name="dq" plugin

echo 'path "dq/*"  { capabilities = ["read", "update"] }' > ./dq-policy.hcl

vault policy write dq-policy ./dq-policy.hcl

vault token create -policy=dq-policy -renewable=false -no-default-policy
```
If any errors occur, you may need to restart the Vault by repeating Step 5 (Case 1) and continue from there.

The output of the last command will be similar to:
```
Key                  Value
---                  -----
token                <register-token>
token_accessor       ...
token_duration       768h
token_renewable      false
token_policies       ["dq-policy"]
identity_policies    []
policies             ["dq-policy"]
```
Save the `<register-token>` value securely.

## Step 9:
In the container shell, install curl:
```sh
apk add curl
```

## Step 10:
Replace `<register-token>` and `<mnemonic>` with the appropriate values and run the following command:
```sh
curl -k --location 'http://127.0.0.1:8200/v1/dq/register' --header 'X-Vault-Token: <register-token>' --header 'Content-Type: application/json' --data '{"username": "root","mnemonic": "<mnemonic>"}'
```
The output will be similar to:
```json
{"request_id":"...","lease_id":"","renewable":false,"lease_duration":0,"data":{"uuid":"<uuid>"},"wrap_info":null,"warnings":null,"auth":null}
```
Save the `<uuid>` value securely.

Now the Vault is ready for use, and you can use it to create addresses or sign transactions.

# How to Use DQ Vault

## Official Documentation

First, check the [DQ Vault plugin usage documentation](https://deqode.github.io/dq-vault/docs/guides/plugin-usage/).

Here, the method of interacting with Vault in the CLI is explained, and for further details and using cURL, please refer to the main documentation.

## Deriving an Address:
Replace `<uuid>`, `<path>`, and `<coin-type>` with the appropriate values and run:
```sh
vault write dq/address uuid="<uuid>" path="<path>" coinType=<coin-type>
```
For Solana, the `path` should be in the four-part format (e.g., "m/44'/501'/0'").

The `coin-type` corresponds to the network ID as per the slip44 standard; for example, Ethereum is 60, and Solana is 501.

### Example:
```sh
vault write dq/address uuid="cql4aua0negc60hrrshg" path="m/44'/501'/0'" coinType=501
```
The output will be:
```
Key          Value
---          -----
address      <address>
publicKey    <public-key>
```
For Solana, the public key and address are the same.

## Signing a Transaction:
Replace `<uuid>`, `<path>`, `<payload>`, and `<coin-type>` with the appropriate values and run:
```sh
vault write dq/signature uuid="<uuid>" path="<path>" payload="<payload>" coinType=<coin-type>
```
For Solana, the `path` should be in the four-part format (e.g., "m/44'/501'/0'").

The `<payload>` should match the payload interface for each network as defined in `.lib/payload.go`.

The `<coin-type>` corresponds to the network ID as per the slip44 standard; for example, Ethereum is 60, and Solana is 501.

### Example:
For the `SolanaRawTx` struct in `.lib/payload.go`:
```
type SolanaRawTx struct {
  RawTxHex     string `json:"rawTxHex"`
  TokenAddress string `json:"tokenAddress"`
  IRawTx
}
```
Command:
```sh
vault write dq/signature uuid="cqo63c4u54ms4l5mffpg" path="m/44'/501'/0'" payload="{\"rawTxHex\": \"010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001000508e9589e3c33514a47f3f8e4655fda1908915b4938274253256a63c000a77761ee8759142feccf57082eb2032543e7f68d29a666fc25992c024899ad234cce0164d40ec3100887c28cc85aceb030d588343f287696f469accf82f208e018d06bc300000000000000000000000000000000000000000000000000000000000000001b576941a405b0a8e641a08eb009ac674459306d5dc22f202d992768eb968d273b442cb3912157f13a933d0134282d032b5ffecd01a2dbf1b7790608df002ea78c97258f4e2489f1bb3d1029148e0d830b5a1399daff1084048e7bd8dbe9f85906ddf6e1d765a193d9cbe146ceeb79ac1cb485ed5f5b37913a8cf5857eff00a9b2c97b742167eeea0460cf0faa2388ff7fc692e70ece259dac03c436b04d907202060600010405030700070302010009031027000000000000\", \"tokenAddress\":\"4zMMC9srt5Ri5X14GAgXhaHii3GnPAEERYPJgZJDncDU\"}" coinType=501
```
Ensure that hexadecimal inputs are sent without the `0x` prefix.

The output will be:
```
Key          Value
---          -----
signature    <signed-tx>
```
The `<signed-tx>` contains the entire transaction content along with the signature and is ready to be broadcast to the network. You might need to convert the format (e.g., to base58) depending on the network's requirements.

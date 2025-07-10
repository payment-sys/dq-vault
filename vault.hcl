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
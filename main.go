package main

import (
	"log"
	"os"

	hasApi "github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/plugin"
	api "github.com/payment-system/dq-vault/api"
)

func main() {
	apiClientMeta := &hasApi.PluginAPIClientMeta{}
	flags := apiClientMeta.FlagSet()
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	tlsConfig := apiClientMeta.GetTLSConfig()
	tlsProviderFunc := hasApi.VaultPluginTLSProvider(tlsConfig)

	err := plugin.Serve(&plugin.ServeOpts{
		BackendFactoryFunc: api.Factory,
		TLSProviderFunc:    tlsProviderFunc,
	})

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

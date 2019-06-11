package main

import (
	"github.com/hashicorp/terraform/plugin"
	"windns"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: windns.Provider,
	})
}

package windns

import (
    "github.com/hashicorp/terraform/helper/schema"
    "github.com/hashicorp/terraform/terraform"

    "fmt"
)

// Provider allows making changes to Windows DNS server
// Utilizes Powershell to connect to domain controller
func Provider() terraform.ResourceProvider {
    return &schema.Provider{
        Schema: map[string]*schema.Schema{
            "domain_controller": &schema.Schema{
                Type:        schema.TypeString,
                Optional:    true,
                DefaultFunc: schema.EnvDefaultFunc("DOMAIN_CONTROLLER", nil),
                Description: "The AD Domain controller to apply changes to. ",
            },
        },
        ResourcesMap: map[string]*schema.Resource{
            "windns": resourceWinDNSRecord(),
        },

        ConfigureFunc: providerConfigure,
    }
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
    domain_controller := d.Get("domain_controller").(string)
    if domain_controller == "" {
        return nil, fmt.Errorf("The 'domain_controller' property was not specified.")
    }

    client := DNSClient {
        domain_controller:  domain_controller,
    }

    return &client, nil
}

type DNSClient struct {
    domain_controller   string
}

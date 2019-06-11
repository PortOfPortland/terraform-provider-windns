package windns

import (
    "github.com/hashicorp/terraform/helper/schema"
    "runpwsh"
    "errors"
    "strings"
)

func resourceWinDNSRecord() *schema.Resource {
    return &schema.Resource{
        Create: resourceWinDNSRecordCreate,
        Read:   resourceWinDNSRecordRead,
        Delete: resourceWinDNSRecordDelete,

        Schema: map[string]*schema.Schema{
            "zone_name": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
                ForceNew: true,
            },
            "record_name": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
                ForceNew: true,
            },
            "record_type": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
                ForceNew: true,
            },
            "ipv4address": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                ForceNew: true,
            },
            "hostnamealias": &schema.Schema{
                Type:     schema.TypeString,
                Optional: true,
                ForceNew: true,
            },
        },
    }
}

func resourceWinDNSRecordCreate(d *schema.ResourceData, m interface{}) error {
    //convert the interface so we can use the variables like username, etc
    client := m.(*DNSClient)

    domain_controller := client.domain_controller
    zone_name := d.Get("zone_name").(string)
    record_type := d.Get("record_type").(string)
    record_name := d.Get("record_name").(string)
    ipv4address := d.Get("ipv4address").(string)
    hostnamealias := d.Get("hostnamealias").(string)

    var id string = zone_name + "_" + record_name + "_" + record_type

    var psCommand string

    switch record_type {
        case "A":
            if ipv4address == "" {
                return errors.New("Must provide ipv4address if record_type is 'A'")
            }
            psCommand = "Add-DNSServerResourceRecord -ZoneName " + zone_name + " -" + record_type + " -Name " + record_name + " -IPv4Address " + ipv4address + " -ComputerName " + domain_controller
        case "CNAME":
            if hostnamealias == "" {
                return errors.New("Must provide hostnamealias if record_type is 'CNAME'")
            }
            psCommand = "Add-DNSServerResourceRecord -ZoneName " + zone_name + " -" + record_type + " -Name " + record_name + " -HostNameAlias " + hostnamealias + " -ComputerName " + domain_controller
        default:
            return errors.New("Unknown record type. This provider currently only supports 'A' and 'CNAME' records.")
    }

        _, err := runpwsh.RunPowershellCommand(psCommand)
    if err != nil {
        //something bad happened
        return err
    }

    d.SetId(id)

    return nil
}

func resourceWinDNSRecordRead(d *schema.ResourceData, m interface{}) error {
    //convert the interface so we can use the variables like username, etc
    client := m.(*DNSClient)

    domain_controller := client.domain_controller
    zone_name := d.Get("zone_name").(string)
    record_type := d.Get("record_type").(string)
    record_name := d.Get("record_name").(string)

    //Get-DnsServerResourceRecord -ZoneName "contoso.com" -Name "Host03" -RRType "A"
    var psCommand string = "try { $record = Get-DnsServerResourceRecord -ZoneName " + zone_name + " -RRType " + record_type + " -Name " + record_name + " -ComputerName " + domain_controller + " -ErrorAction Stop } catch { $record = '''' }; if ($record) { write-host 'RECORD_FOUND' }"
    _, err := runpwsh.RunPowershellCommand(psCommand)
    if err != nil {
        if !strings.Contains(err.Error(), "ObjectNotFound") {
            //something bad happened
            return err
        } else {
            //not able to find the record - this is an error but ok
            d.SetId("")
            return nil
        }
    }

    var id string = zone_name + "_" + record_name + "_" + record_type
    d.SetId(id)

    return nil
}

func resourceWinDNSRecordDelete(d *schema.ResourceData, m interface{}) error {
    //convert the interface so we can use the variables like username, etc
    client := m.(*DNSClient)

    domain_controller := client.domain_controller
    zone_name := d.Get("zone_name").(string)
    record_type := d.Get("record_type").(string)
    record_name := d.Get("record_name").(string)

    //Remove-DnsServerResourceRecord -ZoneName "contoso.com" -RRType "A" -Name "Host01"
    var psCommand string = "Remove-DNSServerResourceRecord -ZoneName " + zone_name + " -RRType " + record_type + " -Name " + record_name + " -ComputerName " + domain_controller + " -Confirm:$false -Force"

    _, err := runpwsh.RunPowershellCommand(psCommand)
    if err != nil {
        //something bad happened
        return err
    }

    // d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
    d.SetId("")

    return nil
}

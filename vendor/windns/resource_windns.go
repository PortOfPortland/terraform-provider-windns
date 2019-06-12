package windns

import (
    "github.com/hashicorp/terraform/helper/schema"
    "runpwsh"
    "bytes"
    "errors"
    "strings"
    "text/template"
)

type DNSRecord struct {
    Id string
    ZoneName string
    RecordType string
    RecordName string
    IPv4Address string
    HostnameAlias string
    DomainController string
} 

var createTemplate = `
try { 
    $newRecord = $record = Get-DnsServerResourceRecord -ZoneName '{{.ZoneName}}' -RRType '{{.RecordType}}' -Name '{{.RecordName}}' -ComputerName '{{.DomainController}}' -ErrorAction Stop 
} catch { $record = $null }; 
if ($record) { 
    Write-Host 'Existing Record Found, Modifying record.'
    Switch ('{{.RecordType}}')
    {
        'A'     { $newRecord.RecordData = '{{.IPv4Address }}' }
        'CNAME' { $newRecord.RecordData = '{{.HostnameAlias}}' }
    }
    $newRecord.RecordType = '{{.RecordType}}'
    $newRecord.HostName = '{{.RecordName}}'
    Set-DnsServerResourceRecord -ZoneName '{{.ZoneName}}' -OldObject $record -NewObject $newRecord -PassThru -ComputerName '{{.DomainController}}'
}
else {
    Write-Host 'Creating record.'
    Switch ('{{.RecordType}}')
    {
        'A'     { Add-DnsServerResourceRecord -ZoneName '{{.ZoneName}}' -RRType '{{.RecordType}}' -Name '{{.RecordName}}' -ComputerName '{{.DomainController}}' -IPv4Address '{{.IPv4Address}}' }
        'CNAME' { Add-DnsServerResourceRecord -ZoneName '{{.ZoneName}}' -RRType '{{.RecordType}}' -Name '{{.RecordName}}' -ComputerName '{{.DomainController}}' -HostNameAlias '{{.HostnameAlias}}' }
    }
}`

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

    record := DNSRecord {
        Id: d.Get("zone_name").(string) + "_" + d.Get("record_name").(string) + "_" + d.Get("record_type").(string),
        ZoneName: d.Get("zone_name").(string),
        RecordName: d.Get("record_name").(string),
        RecordType: d.Get("record_type").(string),
        IPv4Address: d.Get("ipv4address").(string),
        HostnameAlias: d.Get("hostnamealias").(string),
        DomainController: client.domain_controller,
    }

    t := template.New("CreateTemplate")
    t, err := t.Parse(createTemplate)
    if err != nil {
        return err
    }

    var createComandBuffer bytes.Buffer
    if err := t.Execute(&createComandBuffer, record); err != nil {
        return err
    }

    createCommand := createComandBuffer.String()

    switch record.RecordType {
        case "A":
            if record.IPv4Address == "" {
                return errors.New("Must provide ipv4address if record_type is 'A'")
            }
        case "CNAME":
            if record.HostnameAlias == "" {
                return errors.New("Must provide hostnamealias if record_type is 'CNAME'")
            }
        default:
            return errors.New("Unknown record type. This provider currently only supports 'A' and 'CNAME' records.")
    }

    _, err = runpwsh.RunPowershellCommand(createCommand)
    if err != nil {
        //something bad happened
        return err
    }

    d.SetId(record.Id)

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

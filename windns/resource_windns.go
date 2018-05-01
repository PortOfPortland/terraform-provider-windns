package windns

import (
	"github.com/hashicorp/terraform/helper/schema"

	ps "github.com/gorillalabs/go-powershell"
	"github.com/gorillalabs/go-powershell/backend"

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
			psCommand = "Add-DNSServerResourceRecord -ZoneName " + zone_name + " -" + record_type + " -Name " + record_name + " -IPv4Address " + ipv4address
		case "CNAME":
			if hostnamealias == "" {
				return errors.New("Must provide hostnamealias if record_type is 'CNAME'")
			}
			psCommand = "Add-DNSServerResourceRecord -ZoneName " + zone_name + " -" + record_type + " -Name " + record_name + " -HostNameAlias " + hostnamealias
		default:
			return errors.New("Unknown record type. This provider currently only supports 'A' and 'CNAME' records.")
	}

        _, err := runWinRMCommand(client.username, client.password, client.server, psCommand, client.usessl)
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

	zone_name := d.Get("zone_name").(string)
	record_type := d.Get("record_type").(string)
	record_name := d.Get("record_name").(string)

	//Get-DnsServerResourceRecord -ZoneName "contoso.com" -Name "Host03" -RRType "A"
	var psCommand string = "try { $record = Get-DnsServerResourceRecord -ZoneName " + zone_name + " -RRType " + record_type + " -Name " + record_name + "} catch { $record = '' }; if ($record) { write-host 'RECORD_FOUND' }"
	_, err := runWinRMCommand(client.username, client.password, client.server, psCommand, client.usessl)
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
	d.Set("address", id)

	return nil
}

func resourceWinDNSRecordDelete(d *schema.ResourceData, m interface{}) error {
	//convert the interface so we can use the variables like username, etc
	client := m.(*DNSClient)

	zone_name := d.Get("zone_name").(string)
	record_type := d.Get("record_type").(string)
	record_name := d.Get("record_name").(string)

	//Remove-DnsServerResourceRecord -ZoneName "contoso.com" -RRType "A" -Name "Host01"
	var psCommand string = "Remove-DNSServerResourceRecord -ZoneName " + zone_name + " -RRType " + record_type + " -Name " + record_name + " -Confirm:$false -Force"

        _, err := runWinRMCommand(client.username, client.password, client.server, psCommand, client.usessl)
	if err != nil {
		//something bad happened
		return err
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
	d.SetId("")

	return nil
}

func runWinRMCommand(username string, password string, server string, command string, usessl string) (string, error) {
	// choose a backend
	back := &backend.Local{}

	// start a local powershell process
	shell, err := ps.New(back)
	if err != nil {
		//something bad happened - return an error
		return "", err
	}
	defer shell.Exit()

	// ... and interact with it
	var winRMPre string = "$SecurePassword = '" + password + "' | ConvertTo-SecureString -AsPlainText -Force; $cred = New-Object System.Management.Automation.PSCredential -ArgumentList '" + username + "', $SecurePassword; $s = New-PSSession -ComputerName " + server + " -Credential $cred"
        var winRMPost string = "; Invoke-Command -Session $s -Scriptblock { " + command + " }; Remove-PSSession $s"

	// use SSL if requested
	var winRMCommand string
	if (usessl == "1") {
		winRMCommand = winRMPre + " -UseSSL" + winRMPost
	} else {
		winRMCommand = winRMPre + winRMPost
	}
	stdout, _, err := shell.Execute(winRMCommand)
	
	return stdout, err
}
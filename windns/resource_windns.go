package windns

import (
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/pawelwiecaszek/goPSRemoting"

	"errors"
	"strings"

	"math/rand"
	"os"
	"time"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func waitForLock(client *DNSClient) bool {

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	time.Sleep(time.Duration(r.Intn(100)) * time.Millisecond)

	locked := fileExists(client.lockfile)

	for locked == true {
		time.Sleep(100 * time.Millisecond)
		locked = fileExists(client.lockfile)
	}

	time.Sleep(1000 * time.Millisecond)
	return true
}

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
			"ptrdomainname": &schema.Schema{
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
	ptrdomainname := d.Get("ptrdomainname").(string)

	var id string = zone_name + "_" + record_name + "_" + record_type

	var psCommand string
	var psServerName string

	waitForLock(client)

	file, err := os.Create(client.lockfile)
	if err != nil {
		return err
	}

	if client.usejumphost == "1" {
		psServerName = " -ComputerName " + zone_name
	} else {
		psServerName = ""
	}

	switch record_type {
	case "A":
		if ipv4address == "" {
			return errors.New("Must provide ipv4address if record_type is 'A'")
		}
		psCommand = "Add-DNSServerResourceRecord -ZoneName " + zone_name + " -" + record_type + " -Name " + record_name + " -IPv4Address " + ipv4address + psServerName
	case "CNAME":
		if hostnamealias == "" {
			return errors.New("Must provide hostnamealias if record_type is 'CNAME'")
		}
		psCommand = "Add-DNSServerResourceRecord -ZoneName " + zone_name + " -" + record_type + " -Name " + record_name + " -HostNameAlias " + hostnamealias + psServerName
	case "PTR":
		if ptrdomainname == "" {
			return errors.New("Must provide ptrdomainname if record_type is 'PTR'")
		}
		psCommand = "Add-DNSServerResourceRecord -ZoneName " + zone_name + " -" + record_type + " -Name " + record_name + " -PtrDomainName " + ptrdomainname + psServerName
	default:
		return errors.New("Unknown record type. This provider currently only supports 'A', 'CNAME', and 'PTR' records.")
	}

	_, err = goPSRemoting.RunPowershellCommand(client.username, client.password, client.server, psCommand, client.usessl, client.usessh)

	if err != nil {
		//something bad happened
		return err
	}

	d.SetId(id)

	file.Close()
	os.Remove(client.lockfile)

	return nil
}

func resourceWinDNSRecordRead(d *schema.ResourceData, m interface{}) error {
	//convert the interface so we can use the variables like username, etc
	client := m.(*DNSClient)

	zone_name := d.Get("zone_name").(string)
	record_type := d.Get("record_type").(string)
	record_name := d.Get("record_name").(string)

	var psServerName string

	if client.usejumphost == "1" {
		psServerName = " -ComputerName " + zone_name
	} else {
		psServerName = ""
	}

	//Get-DnsServerResourceRecord -ZoneName "contoso.com" -Name "Host03" -RRType "A"
	var psCommand string = "try { $record = Get-DnsServerResourceRecord -ZoneName " + zone_name + " -RRType " + record_type + " -Name " + record_name + psServerName + " -ErrorAction Stop } catch { $record = '''' }; if ($record) { write-host 'RECORD_FOUND' }"
	_, err := goPSRemoting.RunPowershellCommand(client.username, client.password, client.server, psCommand, client.usessl, client.usessh)
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

	var psServerName string

	waitForLock(client)

	file, err := os.Create(client.lockfile)
	if err != nil {
		return err
	}

	zone_name := d.Get("zone_name").(string)
	record_type := d.Get("record_type").(string)
	record_name := d.Get("record_name").(string)

	if client.usejumphost == "1" {
		psServerName = " -ComputerName " + zone_name
	} else {
		psServerName = ""
	}

	//Remove-DnsServerResourceRecord -ZoneName "contoso.com" -RRType "A" -Name "Host01"
	var psCommand string = "Remove-DNSServerResourceRecord -ZoneName " + zone_name + " -RRType " + record_type + " -Name " + record_name + " -Confirm:$false -Force" + psServerName

	if client.usejumphost == "1" {
		psCommand = "Get-DnsServerResourceRecord -ZoneName " + zone_name + " -RRType " + record_type + " -Name " + record_name + psServerName + " | " + "Remove-DNSServerResourceRecord -ZoneName " + zone_name + " -Confirm:$false -Force" + psServerName
	}

	_, err = goPSRemoting.RunPowershellCommand(client.username, client.password, client.server, psCommand, client.usessl, client.usessh)
	if err != nil {
		//something bad happeneds
		return err
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but it is added here for explicitness.
	d.SetId("")

	file.Close()
	os.Remove(client.lockfile)

	return nil
}

# Terraform Windows DNS Provider

This is the repository for a Terraform Windows DNS Provider, which you can use to create DNS records in Microsoft Windows DNS.

The provider uses the [github.com/gorillalabs/go-powershell/backend](github.com/gorillalabs/go-powershell/backend) package to "shell out" to PowerShell, fire up a WinRM session, and perform the actual DNS work. I made this decision because the Go WinRM packages I was able to find only supported WinRM in Basic/Unencrypted mode, which is not doable in our environment. Shelling out to PowerShell is admittedly ugly, but it allows the use of domain accounts, HTTPS, etc.

# Using the Provider

### Example

```hcl
# configure the provider
# username + password - used to build a powershell credential
# server - the server we'll create a WinRM session into to perform the DNS operations
# usessl - whether or not to use HTTPS for our WinRM session (by default port TCP/5986)
variable "username" {
  type = "string"
}

variable "password" {
  type = "string"
}

provider "windns" {
  server = "mydc.mydomain.com"
  username = "${var.username}"
  password = "${var.password}"
  usessl = true
}

#create an a record
resource "windns" "dns" {
  record_name = "testentry1"
  record_type = "A"
  zone_name = "mydomain.com"
  ipv4address = "192.168.1.5"
}

#create a cname record
resource "windns" "dnscname" {
  record_name = "testcname1"
  record_type = "CNAME"
  zone_name = "mydomain.com"
  hostnamealias = "myhost1.mydomain.com"
}
```

# Building
0. Make sure you have $GOPATH set ($env:GOPATH='c:\wip\go' on Windows, etc)
1. git clone https://github.com/PortOfPortland/terraform-provider-windns
2. cd github.com\portofportland\terraform-provider-windns
3. switch to a feature branch
```
git checkout -b myfeature
```
4. get the dependencies
```
go get
```
5. prune any unnecessary dependencies
```
go mod tidy
```
6. vendor our dependencies
```
go mod vendor
```
7. build the module
```
go build

#cross-compile for windows
GOOS=windows GOARCH=386 go build -o terraform-provider-windns.exe
```

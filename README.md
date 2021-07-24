# azure-dyndns

Various clients for updating Azure DNS records with current public IP for use with Dynamic DNS scenarios

All clients makes use of the following environment variables if found:

- AZURE_TENANT_ID
- AZURE_CLIENT_ID
- AZURE_CLIENT_SECRET

These can be overriden using command line arguments as well - but remember that these can show up in history, process listings etc.

All clients can also make use of a configuration like this:

```json
{
 "subscriptionId": "",
 "resourceGroup": "",
 "zoneName": "",
 "recordName": "",
 "clientId": "",
 "clientSecret": "",
 "tenantId": ""
}
```

## azure-dyndns-client

Golang implementation.

Install using `go get -u github.com/frodehus/azure-dyndns/azure-dyndns-client` or compile from source

```text
$ azure-dyndns-client --help

Usage of azure-dyndns-client:
  -client-id string
        Client ID of the service principal used to login (or set AZURE_CLIENT_ID)
  -client-secret string
        Client secret used to authenticate (or set AZURE_CLIENT_SECRET)
  -config string
        Path of the configuration file to use
  -record string
        Name of the DNS record to update
  -resource-group string
        Name of the resource group where the Azure DNS zone is located
  -subscription-id string
        ID of the subscription where the Azure DNS zone is located
  -tenant string
        Azure tenant where the Azure DNS is located (or set AZURE_TENANT_ID)
  -zone string
        Name of the Azure DNS zone
```


## python

Python 3 implementation.

```
$ python3 azure-dyndns.py --help  

usage: azure-dyndns.py [-h] --subscription-id SUBSCRIPTION_ID --resource-group
                       RESOURCE_GROUP --zone ZONE --record RECORD [--tenant-id TENANT_ID]
                       [--client-id CLIENT_ID] [--client-secret CLIENT_SECRET]

Update Azure DNS record based on current public IP

optional arguments:
  -h, --help            show this help message and exit
  --config CONFIG       Path to configuration file
  --subscription-id SUBSCRIPTION_ID
                        Azure subscription ID
  --resource-group RESOURCE_GROUP
                        Azure resource group name
  --zone ZONE           Azure DNS zone name
  --record RECORD       DNS record name to create/update
  --tenant-id TENANT_ID
                        Azure tenant ID (or set AZURE_TENANT_ID)
  --client-id CLIENT_ID
                        Azure service principal client id (or set AZURE_CLIENT_ID)
  --client-secret CLIENT_SECRET
                        Service principal client secret (or set AZURE_CLIENT_SECRET)
```

## net

.NET 5.0 implementation

### Build standalone executable

`dotnet publish -c Release -r <RID>` where [RID](https://docs.microsoft.com/en-us/dotnet/core/rid-catalog) is the Runtime Identifier for example `linux-x64`

```
$ dotnet run -- --help 

Azure-DynDns 1.0.0
Copyright (C) 2021 Azure-DynDns

  -g, --resource-group     Required. Azure resource group where Azure DNS is located

  -z, --zone               Required. Azure DNS zone name

  -r, --record             Required. DNS record name to be created/updated

  -s, --subscription-id    Required. Azure subscription ID

  -t, --tenant-id          Azure tenant ID (or set AZURE_TENANT_ID)

  -c, --client-id          Azure service principal client ID (or set AZURE_CLIENT_ID)

  -x, --client-secret      Azure service principal client secret (or set AZURE_CLIENT_SECRET)

  --help                   Display this help screen.

  --version                Display version information.

```

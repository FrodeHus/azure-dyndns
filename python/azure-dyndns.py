from azure.mgmt.dns import DnsManagementClient
from azure.identity import ClientSecretCredential, DefaultAzureCredential
from datetime import datetime
import argparse
import json
import os

parser = argparse.ArgumentParser(
    description="Update Azure DNS record based on current public IP"
)
parser.add_argument("--config", help="Path to configuration file")
parser.add_argument("--subscription-id", help="Azure subscription ID")
parser.add_argument("--resource-group", help="Azure resource group name")
parser.add_argument("--zone", help="Azure DNS zone name")
parser.add_argument("--record", help="DNS record name to create/update")
parser.add_argument("--tenant-id", help="Azure tenant ID (or set AZURE_TENANT_ID)")
parser.add_argument(
    "--client-id", help="Azure service principal client id (or set AZURE_CLIENT_ID)"
)
parser.add_argument(
    "--client-secret",
    help="Service principal client secret (or set AZURE_CLIENT_SECRET)",
)
args = parser.parse_args()

if args.config:
    with open(args.config, "r") as config_file:
        config = json.load(config_file)
else:
    config = {
        "subscriptionId": args.subscription_id,
        "tenantId": args.tenant_id,
        "clientId": args.client_id,
        "clientSecret": args.client_secret,
        "resourceGroup": args.resource_group,
        "zoneName": args.zone,
        "recordName": args.record,
    }

if (
    os.getenv("AZURE_TENANT_ID")
    and os.getenv("AZURE_CLIENT_ID")
    and os.getenv("AZURE_CLIENT_SECRET")
):
    credentials = DefaultAzureCredential()
else:
    credentials = ClientSecretCredential(
        config["tenantId"], config["clientId"], config["clientSecret"]
    )


def update_dns(ip: str):
    dns_client = DnsManagementClient(
        credentials, subscription_id=config["subscriptionId"]
    )
    record_set = dns_client.record_sets.create_or_update(
        config["resourceGroup"],
        config["zoneName"],
        config["recordName"],
        "A",
        {
            "ttl": 300,
            "arecords": [{"ipv4_address": ip}],
            "metadata": {
                "createdBy": "azure-dyndns (python)",
                "updated": datetime.now().strftime("%m/%d/%Y, %H:%M:%S"),
            },
        },
    )
    print(f"{record_set.fqdn} - {ip} - {record_set.provisioning_state}")


def get_external_ip():
    import urllib3

    client = urllib3.connection_from_url("https://ifconfig.me")
    response = client.request("get", "/")
    return response.data.decode("utf-8")


if __name__ == "__main__":
    ip = get_external_ip()
    update_dns(ip)

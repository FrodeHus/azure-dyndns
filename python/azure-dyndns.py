from azure.mgmt.dns import DnsManagementClient
from azure.identity import ClientSecretCredential, DefaultAzureCredential
import argparse
import os

parser = argparse.ArgumentParser(
    description="Update Azure DNS record based on current public IP"
)
parser.add_argument("--subscription-id", help="Azure subscription ID", required=True)
parser.add_argument("--resource-group", help="Azure resource group name", required=True)
parser.add_argument("--zone", help="Azure DNS zone name", required=True)
parser.add_argument("--record", help="DNS record name to create/update", required=True)
parser.add_argument("--tenant-id", help="Azure tenant ID (or set AZURE_TENANT_ID)")
parser.add_argument("--client-id", help="Azure service principal client id (or set AZURE_CLIENT_ID)")
parser.add_argument("--client-secret", help="Service principal client secret (or set AZURE_CLIENT_SECRET)")
args = parser.parse_args()

subscription_id = args.subscription_id
tenant_id = args.tenant_id
client_id = args.client_id
client_secret = args.client_secret

if os.getenv("AZURE_TENANT_ID") and os.getenv("AZURE_CLIENT_ID") and os.getenv("AZURE_CLIENT_SECRET"):
    credentials = DefaultAzureCredential()
else:
    credentials = ClientSecretCredential(tenant_id, client_id, client_secret)

resource_group = args.resource_group
dns_zone = args.zone
record = args.record


def update_dns(ip: str):
    import json

    dns_client = DnsManagementClient(credentials, subscription_id=subscription_id)
    record_set = dns_client.record_sets.create_or_update(
        resource_group,
        dns_zone,
        record,
        "A",
        {"ttl": 300, "arecords": [{"ipv4_address": ip}]},
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

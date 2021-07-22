from azure.mgmt.dns import DnsManagementClient
from azure.identity import ClientSecretCredential

subscription_id = ""
tenant_id = ""
client_id = ""
client_secret = ""
credentials = ClientSecretCredential(
    tenant_id, client_id, client_secret
)

resource_group = ""
dns_zone = ""
record = ""


def update_dns(ip: str):
    import json

    dns_client = DnsManagementClient(
        credentials, subscription_id=subscription_id
    )
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

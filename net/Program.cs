using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Text.Json;
using System.Threading.Tasks;
using Azure.Identity;
using CommandLine;
using Microsoft.Azure.Management.Dns;
using Microsoft.Azure.Management.Dns.Models;
using Microsoft.Rest;
using Microsoft.Rest.Azure.Authentication;

namespace AzureDynDns
{
    class Program
    {
        public class Options
        {
            [Option('r', "resource-group", Required = true, HelpText = "Azure resource group where Azure DNS is located")]
            public string ResourceGroup { get; set; }
            [Option('z', "zone", Required = true, HelpText = "Azure DNS zone name")]
            public string Zone { get; set; }
            [Option('n', "record", Required = true, HelpText = "DNS record name to be created/updated")]
            public string Record { get; set; }
            [Option('s', "subscription-id", Required = true, HelpText = "Azure subscription ID")]
            public string SubscriptionId { get; set; }
            [Option('t', "tenant-id", HelpText = "Azure tenant ID (or set AZURE_TENANT_ID)")]
            public string TenantId { get; set; }
            [Option('c', "client-id", HelpText = "Azure service principal client ID (or set AZURE_CLIENT_ID)")]
            public string ClientId { get; set; }
            [Option('x', "client-secret", HelpText = "Azure service principal client secret (or set AZURE_CLIENT_SECRET)")]
            public string ClientSecret { get; set; }
        }
        public static async Task Main(string[] args)
        {
            await Parser.Default.ParseArguments<Options>(args).WithParsedAsync(async (o) => await UpdateDNS(o));
        }

        public static async Task UpdateDNS(Options options)
        {
            var (tenantId, clientId, clientSecret) = GetCredentialInfo(options);
            var creds = await ApplicationTokenProvider.LoginSilentAsync(tenantId, clientId, clientSecret);
            var dnsClient = new DnsManagementClient(creds);
            dnsClient.SubscriptionId = options.SubscriptionId;

            var ip = await GetPublicIP();
            var recordSet = new RecordSet();
            recordSet.TTL = 3600;
            recordSet.ARecords = new List<ARecord>();
            recordSet.ARecords.Add(new ARecord(ip));
            recordSet.Metadata = new Dictionary<string, string>();
            recordSet.Metadata.Add("createdBy", "Azure-DynDns");
            recordSet.Metadata.Add("updated", DateTime.Now.ToString());
            var result = await dnsClient.RecordSets.CreateOrUpdateAsync(options.ResourceGroup, options.Zone, options.Record, RecordType.A, recordSet);

            Console.WriteLine(JsonSerializer.Serialize(result));
        }

        public static async Task<string> GetPublicIP()
        {
            var client = new HttpClient();
            return await client.GetStringAsync("https://ifconfig.me");
        }

        public static (string, string, string) GetCredentialInfo(Options options)
        {
            var tenantId = options.TenantId ?? Environment.GetEnvironmentVariable("AZURE_TENANT_ID");
            var clientId = options.ClientId ?? Environment.GetEnvironmentVariable("AZURE_CLIENT_ID");
            var clientSecret = options.ClientSecret ?? Environment.GetEnvironmentVariable("AZURE_CLIENT_SECRET");
            return (tenantId, clientId, clientSecret);
        }
    }
}

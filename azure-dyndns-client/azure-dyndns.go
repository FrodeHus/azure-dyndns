package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/dns/mgmt/dns"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
)

type DynDnsConfig struct {
	subscriptionId string
	resourceGroup  string
	zoneName       string
	recordName     string
	clientId       string
	clientSecret   string
	tenantId       string
}

func main() {
	subscription := flag.String("subscription-id", "", "ID of the subscription where the Azure DNS zone is located")
	resourceGroup := flag.String("resource-group", "", "Name of the resource group where the Azure DNS zone is located")
	zoneName := flag.String("zone", "", "Name of the Azure DNS zone")
	recordName := flag.String("record", "", "Name of the DNS record to update")
	clientId := flag.String("client-id", "", "Client ID of the service principal used to login (or set AZURE_CLIENT_ID)")
	clientSecret := flag.String("client-secret", "", "Client secret used to authenticate (or set AZURE_CLIENT_SECRET)")
	tenantId := flag.String("tenant", "", "Azure tenant where the Azure DNS is located (or set AZURE_TENANT_ID)")
	flag.Parse()

	c := &DynDnsConfig{
		subscriptionId: *subscription,
		resourceGroup:  *resourceGroup,
		zoneName:       *zoneName,
		recordName:     *recordName,
		clientId:       *clientId,
		clientSecret:   *clientSecret,
		tenantId:       *tenantId,
	}
	result, err := updateRecord(c)
	if err != nil {
		log.Fatal(err)
	}

	r, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", r)
}

func updateRecord(config *DynDnsConfig) (dns.RecordSet, error) {
	ip, err := getIP()
	if err != nil {
		return dns.RecordSet{}, errors.New("Failed to retrieve public IP: " + err.Error())
	}
	client := dns.NewRecordSetsClient(config.subscriptionId)
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		creds := auth.NewClientCredentialsConfig(config.clientId, config.clientSecret, config.tenantId)
		authorizer, err = creds.Authorizer()
		if err != nil {
			return dns.RecordSet{}, err
		}
	}

	client.Authorizer = authorizer

	record := dns.RecordSet{
		Name: &config.recordName,
		RecordSetProperties: &dns.RecordSetProperties{
			TTL:      to.Int64Ptr(300),
			ARecords: &[]dns.ARecord{{Ipv4Address: &ip}},
		},
	}
	result, err := client.CreateOrUpdate(context.Background(), config.resourceGroup, config.zoneName, config.recordName, dns.A, record, "", "")
	if err != nil {
		return dns.RecordSet{}, err
	}
	return result, nil
}

func getIP() (string, error) {
	req, err := http.Get("https://ifconfig.me")
	if err != nil {
		return "", err
	}
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
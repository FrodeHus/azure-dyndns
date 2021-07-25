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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/dns/mgmt/dns"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/pterm/pterm"
)

type Config struct {
	SubscriptionId string
	ResourceGroup  string
	ZoneName       string
	RecordName     string
	ClientId       string
	ClientSecret   string
	TenantId       string
	Service        bool
	Interval       int
}

func main() {
	subscription := flag.String("subscription-id", "", "ID of the subscription where the Azure DNS zone is located")
	resourceGroup := flag.String("resource-group", "", "Name of the resource group where the Azure DNS zone is located")
	zoneName := flag.String("zone", "", "Name of the Azure DNS zone")
	recordName := flag.String("record", "", "Name of the DNS record to update")
	clientId := flag.String("client-id", "", "Client ID of the service principal used to login (or set AZURE_CLIENT_ID)")
	clientSecret := flag.String("client-secret", "", "Client secret used to authenticate (or set AZURE_CLIENT_SECRET)")
	tenantId := flag.String("tenant", "", "Azure tenant where the Azure DNS is located (or set AZURE_TENANT_ID)")
	configFile := flag.String("config", "", "Path of the configuration file to use")
	asService := flag.Bool("service", false, "Periodically updates DNS records")
	interval := flag.Int("interval", 300, "Define how often the DNS record should be updated (in seconds) when running as a service")
	flag.Parse()

	var c Config
	if configFile != nil {
		config, err := readConfigFile(*configFile)
		if err != nil {
			log.Fatal("Failed to load configuration file: " + err.Error())
		}

		c = config
	} else {
		c = Config{
			SubscriptionId: *subscription,
			ResourceGroup:  *resourceGroup,
			ZoneName:       *zoneName,
			RecordName:     *recordName,
			ClientId:       *clientId,
			ClientSecret:   *clientSecret,
			TenantId:       *tenantId,
			Service:        *asService,
			Interval:       *interval,
		}
	}

	if c.Service {
		runService(&c)
	} else {
		result, err := updateRecord(&c)
		if err != nil {
			log.Fatal(err)
		}

		printResult(&result)
	}
}

func runService(config *Config) {
	ticker := time.NewTicker(time.Duration(config.Interval) * time.Second)
	done := make(chan bool)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		signal := <-sigs
		fmt.Printf("Got %s - Stopping service...\n", signal)
		done <- true
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			case _ = <-ticker.C:
				_, err := updateRecord(config)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}()
	<-done
}

func printResult(result *dns.RecordSet) {
	r, err := json.Marshal(result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", r)
}

func updateRecord(config *Config) (dns.RecordSet, error) {
	spinner, _ := pterm.DefaultSpinner.Start("Updating DNS record...")
	ip, err := getIP()
	if err != nil {
		spinner.Fail()
		return dns.RecordSet{}, errors.New("Failed to retrieve public IP: " + err.Error())
	}
	spinner.UpdateText("Got IP " + ip)
	client := dns.NewRecordSetsClient(config.SubscriptionId)
	authorizer, err := getAuthorizer(config)
	if err != nil {
		spinner.Fail()
		return dns.RecordSet{}, err
	}

	client.Authorizer = authorizer
	creator := "azure-dyndns-client (Go)"
	updatedtime := time.Now().String()
	record := dns.RecordSet{
		Name: &config.RecordName,
		RecordSetProperties: &dns.RecordSetProperties{
			TTL:      to.Int64Ptr(300),
			ARecords: &[]dns.ARecord{{Ipv4Address: &ip}},
			Metadata: map[string]*string{
				"createdBy": &creator,
				"updated":   &updatedtime,
			},
		},
	}
	result, err := client.CreateOrUpdate(context.Background(), config.ResourceGroup, config.ZoneName, config.RecordName, dns.A, record, "", "")
	if err != nil {
		spinner.Fail()
		return dns.RecordSet{}, err
	}
	spinner.Success("DNS record " + pterm.LightCyan(config.RecordName+"."+config.ZoneName) + " updated with IP " + pterm.LightCyan(ip))
	return result, nil
}

func getAuthorizer(config *Config) (autorest.Authorizer, error) {
	if config.ClientId == "" || config.ClientSecret == "" || config.TenantId == "" {
		return auth.NewAuthorizerFromEnvironment()
	}
	creds := auth.NewClientCredentialsConfig(config.ClientId, config.ClientSecret, config.TenantId)
	authorizer, err := creds.Authorizer()
	return authorizer, err
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

func readConfigFile(file string) (Config, error) {
	jsonFile, err := os.Open(file)
	if err != nil {
		return Config{}, err
	}

	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return Config{}, err
	}
	var config Config
	json.Unmarshal(bytes, &config)
	return config, nil
}

package azuredyndnsclient

type Config struct {
	subscriptionId string
	resourceGroup  string
	zoneName       string
	recordName     string
	clientId       string
	clientSecret   string
	tenantId       string
}

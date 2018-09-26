package deployer

import (
	"context"
	"log"
	"time"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

var (
	ResourceGroupNamePrefix string
	SubscriptionId string
	Authorizer autorest.Authorizer
	Context context.Context
	Location string
	SecondsBetweenRGDeployments = 2
	SecondsForFirstStatusCheckSleep = 90
	SecondsBetweenVMStatusCheck = 5
	SecondsBetweenEventMonitorCheck = 20
	//TimeoutSecondsForVMProvision = 600
	TimeoutSecondsForVMProvision = 2700
)

func Initialize(subscriptionId string, uniquePrefix string, startTime time.Time, location string, vmCount int) {
	ResourceGroupNamePrefix = GetResourceGroupPrefix(uniquePrefix, startTime)
	SubscriptionId = subscriptionId
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	CheckError("Failed to get OAuth config: %v", err)
	Authorizer = authorizer
	Context = context.Background()
	Location = location

	// adjust to keep under the throttling limits, about 30,000 per 30 minutes
	if vmCount > 500 {
		log.Printf("using settings for >500 vms")
		SecondsBetweenRGDeployments = 8
		SecondsForFirstStatusCheckSleep = 220
		SecondsBetweenVMStatusCheck = 60
	} else if vmCount > 300 {
		log.Printf("using settings for >300 vms")
		SecondsBetweenRGDeployments = 4
		SecondsForFirstStatusCheckSleep = 120
		SecondsBetweenVMStatusCheck = 20
	} else if vmCount > 100 {
		log.Printf("using settings for >300 vms")
		SecondsBetweenRGDeployments = 3
		SecondsForFirstStatusCheckSleep = 100
		SecondsBetweenVMStatusCheck = 10
	} else {
		log.Printf("using settings for <100 vms")
	}
	SecondsBetweenRGDeployments = 0
}
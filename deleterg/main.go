package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
)

const (	
	DefaultSecondsBetweenRGDeletes = 7
	MinimumResourceGroupPrefixLength = 5

	AZURE_TENANT_ID = "AZURE_TENANT_ID"
	AZURE_CLIENT_ID = "AZURE_CLIENT_ID"
	AZURE_CLIENT_SECRET = "AZURE_CLIENT_SECRET"
	AZURE_SUBSCRIPTION_ID = "AZURE_SUBSCRIPTION_ID"
	AZURE_LOCATION_DEFAULT = "AZURE_LOCATION_DEFAULT"
)

var (
	ctx        = context.Background()
	authorizer autorest.Authorizer
	groupsClient resources.GroupsClient
)

func usage(errs ...error) {
	for _, err := range errs {
		fmt.Fprintf(os.Stderr, "error: %s\n\n", err.Error())
	}
	fmt.Fprintf(os.Stderr, "usage: %s [OPTIONS] RESOURCE_GROUP_PREFIX\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tdelete all resource groups with a prefix\n")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "required environment variables:\n")
	fmt.Fprintf(os.Stderr, "\t%s\n", AZURE_TENANT_ID)
	fmt.Fprintf(os.Stderr, "\t%s\n", AZURE_CLIENT_ID)
	fmt.Fprintf(os.Stderr, "\t%s\n", AZURE_CLIENT_SECRET)
	fmt.Fprintf(os.Stderr, "\t%s\n", AZURE_SUBSCRIPTION_ID)
	fmt.Fprintf(os.Stderr, "\t%s\n", AZURE_LOCATION_DEFAULT)
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "options:\n")
	flag.PrintDefaults()
}

func initializeApplicationVariables() (string, string, int, bool) {
	var secondsBetweenRGDeletes = flag.Int("secondsBetweenRGDelete", DefaultSecondsBetweenRGDeletes, "seconds between resource group deletion for the purpose of avoiding throttling" )
	var waitForCompletion = flag.Bool("waitForCompletion", true, "monitor deletion and wait for completion")

	flag.Parse()

	if envVarsAvailable := verifyEnvVars(); !envVarsAvailable {
		usage()
		os.Exit(1)
	}

	if argCount := len(flag.Args()); argCount == 0 {
		usage()
		os.Exit(1)
	}

	subscriptionID := os.Getenv(AZURE_SUBSCRIPTION_ID)
	
	prefix := flag.Arg(0)
	if len(prefix) < MinimumResourceGroupPrefixLength {
		fmt.Fprintf(os.Stderr, "ERROR: Resource Group Prefix must have minimum length of %d\n", MinimumResourceGroupPrefixLength)
		usage()
		os.Exit(2)
	}

	return subscriptionID, prefix, *secondsBetweenRGDeletes, *waitForCompletion
}

// Authenticate with the Azure services using file-based authentication
func initializeGroupsClient(subscriptionId string) {
	var err error
	authorizer, err = auth.NewAuthorizerFromEnvironment()
	if err != nil {
		log.Fatalf("Failed to get OAuth config: %v", err)
	}

	groupsClient = resources.NewGroupsClient(subscriptionId)
	groupsClient.Authorizer = authorizer
}

func verifyEnvVar(envvar string) bool {
	if _, available := os.LookupEnv(envvar); !available {
		fmt.Fprintf(os.Stderr, "ERROR: Missing Environment Variable %s\n", envvar)
		return false
	}
	return true
}

func verifyEnvVars() bool {
	available := true
	available = available && verifyEnvVar(AZURE_TENANT_ID)
	available = available && verifyEnvVar(AZURE_CLIENT_ID)
	available = available && verifyEnvVar(AZURE_CLIENT_SECRET)
	available = available && verifyEnvVar(AZURE_SUBSCRIPTION_ID)
	available = available && verifyEnvVar(AZURE_LOCATION_DEFAULT)
	return available
}

// listGroups gets an interator that gets all resource groups in the subscription
func listGroups(ctx context.Context) (resources.GroupListResultIterator, error) {
	return groupsClient.ListComplete(ctx, "", nil)
}

func deleteAllGroupsWithPrefix(SubscriptionID string, prefix string, secondsBetweenRGDeletes int) {
	for list, err := listGroups(ctx); list.NotDone(); err = list.Next() {
		if err != nil {
			log.Fatalf("got error: %s", err)
		}
		rgName := *list.Value().Name
		if strings.HasPrefix(rgName, prefix) {
			log.Printf("deleting group '%s'\n", rgName)
			_, err := groupsClient.Delete(ctx, rgName)
			if err != nil {
				log.Printf("got error: %v", err)
			}
			time.Sleep(time.Duration(secondsBetweenRGDeletes) * time.Second)
		}
	}
}

func waitForResourceGroupDeletion(SubscriptionID string, prefix string) {
	for {
		count := 0
		for list, err := listGroups(ctx); list.NotDone(); err = list.Next() {
			if err != nil {
				log.Printf("got error: %s", err)
			}
			rgName := *list.Value().Name
			if strings.HasPrefix(rgName, prefix) {
				count++
			}
		}
		if count == 0 {
			log.Printf("completed resource group deletion of groups with prefix %s", prefix)
			return
		}
		log.Printf("%d resource groups with prefix %s remain", count, prefix)
		time.Sleep(10 * time.Second)
	}
}

func main() {
	subscriptionID, prefix, secondsBetweenRGDeletes, waitForCompletion := initializeApplicationVariables()

	initializeGroupsClient(subscriptionID)

	deleteAllGroupsWithPrefix(subscriptionID, prefix, secondsBetweenRGDeletes)

	if waitForCompletion {
		waitForResourceGroupDeletion(subscriptionID, prefix)
	}
}
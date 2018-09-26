package main

import (
	"azuredeployer/deployer"
	"fmt"
	"flag"
	"log"
	"os"
	"time"
)
/*
const (	
	AZURE_TENANT_ID = "AZURE_TENANT_ID"
	AZURE_CLIENT_ID = "AZURE_CLIENT_ID"
	AZURE_CLIENT_SECRET = "AZURE_CLIENT_SECRET"
	AZURE_SUBSCRIPTION_ID = "AZURE_SUBSCRIPTION_ID"
	AZURE_LOCATION_DEFAULT = "AZURE_LOCATION_DEFAULT"
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
}*/

func usage(errs ...error) {
	for _, err := range errs {
		fmt.Fprintf(os.Stderr, "error: %s\n\n", err.Error())
	}
	fmt.Fprintf(os.Stderr, "usage: %s [OPTIONS] ResourceGroupPrefix ResourceGroupCount VMsPerResourceGroup\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "       deploy the VM")
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "options:\n")
	flag.PrintDefaults()
}

func main() {
	startTime := time.Now()
/*
	var resourceGroupPrefix string
	var resourceGroupCount int
	var vmsPerResourceGroup int

	var statisticsPath = flag.String("statistics-path", )

	DeploymentRunBasePath = "C:\\project\\deploymentruns\\loosePatformImageCopy"
	LooseVMPath = "C:\\scratch\\arm\\boottests\\loosevm"
	TemplateName = "azuredeploy.json"
	TemplateParametersName = "azuredeploy.parameters.json"
	VMExtensionName = "configureagent"

	--template-file win10-azuredeploy.json --parameters
*/

	// 1000
	//resourceGroupCount := 50
	//vmsPerResourceGroup := 20
	
	// good ratio for vmas 40x25
	//resourceGroupCount := 40
	//vmsPerResourceGroup := 25
	// vmss
	//resourceGroupCount := 2
	//vmsPerResourceGroup := 500
	
	// vmss custom
	//resourceGroupCount := 4
	//vmsPerResourceGroup := 250

	//resourceGroupCount := 40
	//vmsPerResourceGroup := 25
	
	resourceGroupCount := 1
	vmsPerResourceGroup := 10

	//deployer.Initialize("YOUR_SUBSCRIPTION_ID", "uniquename", startTime, "westus")
	deployer.Initialize("YOUR_SUBSCRIPTION_ID", "uniquename", startTime, "eastus", resourceGroupCount * vmsPerResourceGroup)
	log.Printf("starting deployment\n")
	log.Printf("\tSubscriptionId: %s\n", deployer.SubscriptionId)
	log.Printf("\tResourceGroupPrefix: %s\n", deployer.ResourceGroupNamePrefix)
	
	deploymentRun := deployer.InitializeDeploymentRun(
		deployer.LooseVMPath,
		resourceGroupCount,
		vmsPerResourceGroup,
		startTime) 

	deploymentRun.Run()

	log.Printf("deployment complete\n")
}
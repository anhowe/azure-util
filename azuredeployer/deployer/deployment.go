package deployer

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest/to"
)

func InitializeDeployment(groupId int, vmCount int, start time.Time, runPath string) *Deployment {
	instances := make([]*VMInstance, 0, vmCount)
	
	for i :=0; i < vmCount; i++ {
		vm := InitializeVMInstance(
			GetResourceGroupName(ResourceGroupNamePrefix, groupId, vmCount),
			i,
			start)
		instances = append(instances, vm)
	}

	return &Deployment{
		GroupId: groupId,
		VMCount: vmCount,
		TemplatePath: GetTemplatePath(runPath),
		TemplateParametersPath: GetTemplateParametersPath(runPath),
		VMInstances: instances,
		DeploymentComplete: false,
		Started: false,
	}
}

func (d *Deployment) Deploy(deploymentSyncWaitGroup *sync.WaitGroup) {
	defer deploymentSyncWaitGroup.Done()

	d.Started = true

	abortChannel := make(chan int)

	vmInstanceSyncWaitGroup := sync.WaitGroup{}
	
	vmInstanceSyncWaitGroup.Add(d.VMCount)

	for _, vm := range d.VMInstances {
		go vm.VMMonitor(&vmInstanceSyncWaitGroup, abortChannel)
	}
	
	d.deployTemplate()

	// abort running instances, wait for completion
	close(abortChannel)
	vmInstanceSyncWaitGroup.Wait()
	log.Printf("deployment finished\n")
	d.DeploymentComplete = true
}

func (d *Deployment) deployTemplate() {
	template, params := d.getTemplateInput()
	if template == nil {
		log.Fatalf("failure reading template %s\n", d.TemplatePath)
	}
	if params == nil {
		log.Fatalf("failure reading params %s\n", d.TemplateParametersPath)
	}

	d.createResourceGroup()

	d.createDeployment(template, params)
}

func (d *Deployment) getTemplateInput() (*map[string]interface{}, *map[string]interface {}) {
	if len(d.VMInstances) == 0 {
		return nil, nil
	}

	template, err := ReadJSON(d.TemplatePath)
	CheckError("unable to read template", err)
	
	params, err := ReadJSON(d.TemplateParametersPath)
	CheckError("unable to read template parameters", err)
	
	(*params)["uniquename"] = map[string]string{
		"value": d.VMInstances[0].ResourceGroupName,
	}
	(*params)["vmCount"] = map[string]int{
		"value": d.VMCount,
	}
	return template, params
}

func (d *Deployment) createResourceGroup() {
	if len(d.VMInstances) == 0 {
		return
	}

	groupsClient := resources.NewGroupsClient(SubscriptionId)
	groupsClient.Authorizer = Authorizer
	
	_, err := groupsClient.CreateOrUpdate(
		Context,
		d.VMInstances[0].ResourceGroupName,
		resources.Group{
			Location: to.StringPtr(Location)})
	CheckError(fmt.Sprintf("error creating rg %s", d.VMInstances[0].ResourceGroupName), err)
}

func (d *Deployment) createDeployment(template *map[string]interface{}, params *map[string]interface{}) {
	if len(d.VMInstances) == 0 {
		return
	}

	deploymentsClient := resources.NewDeploymentsClient(SubscriptionId)
	deploymentsClient.Authorizer = Authorizer
	// the default time of 15 minutes is too short for the deployment client, so we'll set for two hours
	// the azure SDK defaults are set here github.com/Azure/go-autorest/autorest/client.go
	deploymentsClient.PollingDuration = time.Hour * 2

	log.Printf("starting deployment for rg %s\n", d.VMInstances[0].ResourceGroupName)
	deploymentFuture, err := deploymentsClient.CreateOrUpdate(
		Context,
		d.VMInstances[0].ResourceGroupName,
		fmt.Sprintf("dep-%s", d.VMInstances[0].ResourceGroupName),
		resources.Deployment{
			Properties: &resources.DeploymentProperties{
				Template:   template,
				Parameters: params,
				Mode:       resources.Incremental,
			},
		},
	)
	if err != nil {
		log.Printf("deployment call failed for RG %s: %v\n", d.VMInstances[0].ResourceGroupName, err)
		return
	}

	time.Sleep(time.Duration(SecondsForFirstStatusCheckSleep) * time.Second)

	err = deploymentFuture.Future.WaitForCompletion(Context, deploymentsClient.BaseClient.Client)
	if err != nil {
		log.Printf("deployment failure for RG %s: %v\n", d.VMInstances[0].ResourceGroupName, err)
	}
	log.Printf("deployment complete for rg %s\n", d.VMInstances[0].ResourceGroupName)
}
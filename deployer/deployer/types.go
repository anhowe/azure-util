package deployer

import (
	"sync"
	"time"
)

// DeploymentRun describes a full run of a deployment
type DeploymentRun struct {
	RunPath string
	ResourceGroupCount int
	SecondsBetweenRGDeployments int
	VMsPerResourceGroup int
	StartTime time.Time
	
	// internal variables
	eventMonitorReceiver sync.WaitGroup
	deploymentReceivers sync.WaitGroup
	deployments []*Deployment
}

// Deployment represents a single resource group deployment
type Deployment struct {
	GroupId int
	VMCount int
	TemplatePath string
	TemplateParametersPath string
	VMInstances []*VMInstance
	DeploymentComplete bool
	Started bool
}

type VMInstance struct {
	ResourceGroupName string
	Name string
	Status string
	Start time.Time
	VMEnd time.Time
	EndEvent time.Time
}
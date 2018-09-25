package deployer

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

func InitializeDeploymentRun(path string, rgCount int, vmPerRGCount int, start time.Time) *DeploymentRun {
	deployments := make([]*Deployment, 0, rgCount)

	for i := 0; i < rgCount; i++ {
		deployment := InitializeDeployment(i, vmPerRGCount, start, path)
		deployments = append(deployments, deployment)
	}

	return &DeploymentRun{
		RunPath: path,
		ResourceGroupCount: rgCount,
		VMsPerResourceGroup: vmPerRGCount,
		SecondsBetweenRGDeployments: SecondsBetweenRGDeployments,
		StartTime: start,
		eventMonitorReceiver: sync.WaitGroup{},
		deploymentReceivers: sync.WaitGroup{},
		deployments: deployments,
	}
}

// Run starts the deployment of virtual machines
func (d *DeploymentRun) Run() {
	log.Printf("Starting VM EventMonitor\n")
	d.StartVMEventMonitor()

	log.Printf("Starting Deployments\n")
	d.StartDeployments()
	
	log.Printf("Waiting for Deployments to finish\n")
	d.WaitForDeploymentsToFinish()

	log.Printf("Waiting for VM EventMonitor to finish\n")
	d.WaitForVMEventMonitorToFinish()
	
	outputCsvPath := d.GetCSVPath()
	log.Printf("Writing deployment report %s\n", outputCsvPath)
	d.WriteReport(outputCsvPath)
}

func (d *DeploymentRun) StartVMEventMonitor() {
	if UseEventMonitor {
		d.eventMonitorReceiver.Add(1)
		go VMEventMonitor(&d.eventMonitorReceiver, d)
	}
}

func (d *DeploymentRun) StartDeployments() {
	d.deploymentReceivers.Add(d.ResourceGroupCount)

	for _, deployment := range d.deployments {
		go deployment.Deploy(&d.deploymentReceivers)
		time.Sleep(time.Duration(d.SecondsBetweenRGDeployments) * time.Second)
	}
}

func (d *DeploymentRun) WaitForDeploymentsToFinish() {
	log.Printf("Waiting for %d deployments to finish", d.ResourceGroupCount)
	d.deploymentReceivers.Wait()
}

func (d *DeploymentRun) WaitForVMEventMonitorToFinish() {
	if UseEventMonitor {
		log.Printf("Waiting for vm event monitor to finish")
		d.eventMonitorReceiver.Wait()
	}
}

func (d *DeploymentRun) WriteReport(outputCsvPath string) {
		// write the csv file
		file, err := os.Create(outputCsvPath)
		CheckError("cannot create file", err)
		defer file.Close()
	
		w := csv.NewWriter(file)

		// write the header
		columnCount := 5
		header := make([]string, 0, columnCount)
		header = append(header, "ResourceGroup")
		header = append(header, "VMName")
		header = append(header, "VMStatus")
		header = append(header, "VMStartUpTime")
		header = append(header, "VMProvisionTime")
		err = w.Write(header)
		CheckError("error writing header to csv:", err)
	
		for _, deployment := range d.deployments {
			for _, vm := range deployment.VMInstances {
				row := make([]string, 0, columnCount)
				row = append(row, vm.ResourceGroupName)
				row = append(row, vm.Name)
				row = append(row, vm.Status)
				row = append(row, fmt.Sprintf("%d", vm.GetVMStartupSeconds()))
				row = append(row, fmt.Sprintf("%d", vm.GetVMProvisionSeconds()))
				err := w.Write(row)
				CheckError("error writing record to csv:", err)
			}
		}
	
		// Write any buffered data to the underlying writer (standard output).
		w.Flush()
		CheckError("error flushing", w.Error())	
}

func (d *DeploymentRun) GetCSVPath() string {
	t := time.Now()
	projectFile := fmt.Sprintf("%02d-%02d-%02d-%02d%02d%02d-%d-nodes.csv", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), d.ResourceGroupCount * d.VMsPerResourceGroup)
	return path.Join(DeploymentRunBasePath, projectFile)
}
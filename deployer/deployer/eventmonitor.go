package deployer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Azure/azure-amqp-common-go/sas"
	eventhubs "github.com/Azure/azure-event-hubs-go"
)

func VMEventMonitor(vmInstanceSyncWaitGroup *sync.WaitGroup, d *DeploymentRun) {
	log.Printf("starting VM Event Monitor\n")
	defer vmInstanceSyncWaitGroup.Done()

	ticker := time.NewTicker(time.Duration(SecondsBetweenEventMonitorCheck) * time.Second)
	defer ticker.Stop()

	// set up wait group to wait for expected message
	eventReceived := make(chan struct{})

	hub, info := SetupEventHub()

	// declare handler for incoming events
	handler := CreateHandler(eventReceived, d)

	RegisterEventHubReceiveHandlers(hub, info, handler)

	for {
		select {
		case <-eventReceived:
			// reset the ticker
			ticker.Stop()
			ticker = time.NewTicker(time.Duration(SecondsBetweenEventMonitorCheck) * time.Second)
		case <-ticker.C:
		}

		// check for completion
		vmTotalCount := 0
		vmCountComplete := 0
		vmInflight := 0

		for _, deployment := range d.deployments {
			if deployment.DeploymentComplete {
				now := time.Now()
				firstDeploymentCompletionTime := deployment.VMInstances[0].VMEnd
				if int(now.Sub(firstDeploymentCompletionTime).Seconds()) > TimeoutSecondsForVMProvision {
					log.Printf("Timeout exceeded for deployment %d not waiting longer", deployment.GroupId)
					vmTotalCount += deployment.VMCount
					vmCountComplete += deployment.VMCount
					vmInflight += deployment.VMCount 
					continue
				}
			}
			for _, vm := range deployment.VMInstances {
				vmTotalCount++
				if !vm.EndEvent.IsZero() || vm.IsFailed() {
					// vm has completed
					vmCountComplete++
				} else if deployment.Started {
					// inflight
					vmInflight++
				} else {
					// not started
				}
			}
		}
		log.Printf("(%d / %d) VMs completed provisioning, %d inflight", vmCountComplete, vmTotalCount, vmInflight)
		if vmTotalCount == vmCountComplete {
			return
		}
	}
}

func SetupEventHub() (*eventhubs.Hub, *eventhubs.HubRuntimeInformation) {
	// setup the handler
	provider, err := sas.NewTokenProvider(sas.TokenProviderWithKey(EventHubSenderKeyName, EventHubSenderKey))
	if err != nil {
		log.Fatalf("failed to get token provider: %s\n", err)
	}

	// get an existing hub
	hub, err := eventhubs.NewHub(EventHubNSName, EventHubHubName, provider)
	defer hub.Close(Context)
	if err != nil {
		log.Fatalf("failed to get hub: %s\n", err)
	}

	// get info about partitions in hub
	info, err := hub.GetRuntimeInformation(Context)
	if err != nil {
		log.Fatalf("failed to get runtime info: %s\n", err)
	}
	log.Printf("partition IDs: %s\n", info.PartitionIDs)

	return hub, info
}

func CreateHandler(eventReceived chan struct{}, d *DeploymentRun) (handler eventhubs.Handler) {
	return func(ctx context.Context, event *eventhubs.Event) error {
		fmt.Printf("received: %s\n", string(event.Data))
		// only mark for VMs
		found := false

IterateDeployments:
		for _, deployment := range d.deployments {
			for _, vm := range deployment.VMInstances {
				// for VMSS, just mark the first node that does not have a date set
				if vm.Name == string(event.Data) || (IsVMSS && vm.EndEvent.IsZero()) {
					found = true
					vm.EndEvent = time.Now()
					break IterateDeployments
				}
			}
		} 
		if !found {
			log.Printf("WARNING: event not found %s", event.Data)
		}
		// notify channel that event was received
		eventReceived <- struct{}{}
		return nil
	}
}

func RegisterEventHubReceiveHandlers(hub *eventhubs.Hub, info *eventhubs.HubRuntimeInformation, handler eventhubs.Handler) {
	for _, partitionID := range info.PartitionIDs {
		_, err := hub.Receive(
			Context,
			partitionID,
			handler,
			//eventhubs.ReceiveWithStartingOffset(persist.StartOfStream),
			eventhubs.ReceiveWithLatestOffset(),
		)
		if err != nil {
			log.Fatalf("failed to receive for partition ID %s: %s\n", partitionID, err)
		}
	}
}
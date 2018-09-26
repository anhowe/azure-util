package deployer

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-06-01/compute"
)

func InitializeVMInstance(rgName string, vmId int, start time.Time) *VMInstance {
	return &VMInstance{
		ResourceGroupName: rgName,
		Name: GetVMName(rgName, vmId),
		Status: VM_NOSTATUS,
		Start: start,
		VMEnd: time.Time{},
		EndEvent: time.Time{},
	}
}

func (v *VMInstance) GetVMStartupSeconds() uint {
	//if v.VMEnd.IsZero() {
	//	v.VMEnd = time.Now()
	//}
	return uint(v.VMEnd.Sub(v.Start).Seconds())
}

func (v *VMInstance) GetVMProvisionSeconds() uint {
	if UseEventMonitor {
		if v.EndEvent.IsZero() {
			v.EndEvent = time.Now()
		}
		return uint(v.EndEvent.Sub(v.Start).Seconds())
	} else {
		return v.GetVMStartupSeconds() 
	}
}

func (v *VMInstance) VMMonitor(vmInstanceSyncWaitGroup *sync.WaitGroup, abort chan int) {
	log.Printf("starting VM %s in rg %s\n", v.Name, v.ResourceGroupName)
	defer vmInstanceSyncWaitGroup.Done()
	time.Sleep(time.Duration(SecondsForFirstStatusCheckSleep) * time.Second)
	for {
		if !IsVMSS {
			// get VM state for VMs only
			finished := v.checkVMState()
			//finished := v.checkVMExtensionState()
			if finished {
				v.VMEnd = time.Now()
				log.Printf("finished VM %s in rg %s\n", v.Name, v.ResourceGroupName)
				return
			}
		}
		
		// check for early abort or sleep
		select {
		case <-abort:
			v.VMEnd = time.Now()
			log.Printf("finished VM %s in rg %s\n", v.Name, v.ResourceGroupName)
/*			if !v.IsSucceeded() {
				log.Printf("WARNING: no extension status, falling back to VM Status\n")
				v.checkVMState()
			}*/
			return 
		default:
			time.Sleep(time.Duration(SecondsBetweenVMStatusCheck) * time.Second)
		}
	}
}

func (v *VMInstance) checkVMState() bool {
	vmClient := compute.NewVirtualMachinesClient(SubscriptionId)
	vmClient.Authorizer = Authorizer
	vm, err := vmClient.Get(Context, v.ResourceGroupName, v.Name, compute.InstanceView)
	if err != nil {
		//log.Printf("error getting context VM %s in rg %s: %v\n", v.Name, v.ResourceGroupName, err)
		return false
	}
	v.Status = *vm.VirtualMachineProperties.ProvisioningState
	return v.IsSucceeded()
}

/*func (v *VMInstance) checkVMExtensionState() bool {
	vmExtensionClient := compute.NewVirtualMachineExtensionsClient(SubscriptionId)
	vmExtensionClient.Authorizer = Authorizer
	vmExtension, err := vmExtensionClient.Get(Context, v.ResourceGroupName, v.Name, VMExtensionName, "")
	if err != nil {
		//log.Printf("error getting context VM %s in rg %s: %v\n", v.Name, v.ResourceGroupName, err)
		return false
	}
	v.Status = *vmExtension.VirtualMachineExtensionProperties.ProvisioningState
	return v.IsSucceeded()
}*/

func (v *VMInstance) IsSucceeded() bool {
	return strings.Compare(v.Status, VM_SUCCEEDED) == 0
}

func (v *VMInstance) IsFailed() bool {
	return strings.Compare(v.Status, VM_FAILED) == 0
}

func GetVMName(rgName string, vmId int) string {
	return fmt.Sprintf("vm-%s-%d", rgName, vmId)
}
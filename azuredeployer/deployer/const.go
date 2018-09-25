package deployer

import (
	"path"
)

const (	
	//DeploymentRunBasePath = "C:\\project\\deploymentruns\\loosePatformImageCopy"
	//LooseVMPath = "C:\\scratch\\arm\\boottests\\loosevm"
	//DeploymentRunBasePath = "C:\\project\\deploymentruns\\vmas2"
	//LooseVMPath = "C:\\scratch\\arm\\boottests\\vmas2"
	//DeploymentRunBasePath = "C:\\project\\deploymentruns\\vmss"
	//LooseVMPath = "C:\\scratch\\arm\\boottests\\vmss"
	//DeploymentRunBasePath = "C:\\project\\deploymentruns\\vmascustomimage"
	//LooseVMPath = "C:\\scratch\\arm\\boottests\\vmascustomimagecopy"
	DeploymentRunBasePath = "C:\\project\\deploymentruns\\vmascustomimage"
	LooseVMPath = "C:\\scratch\\arm\\boottests\\vmascustomimagecopyavere6node"
	//DeploymentRunBasePath = "C:\\project\\deploymentruns\\vmsscustomimageavere"
	//LooseVMPath = "C:\\scratch\\arm\\boottests\\vmsscustomimagecopyavere"
	//DeploymentRunBasePath = "C:\\project\\deploymentruns\\vmsscustomimage"
	//LooseVMPath = "C:\\scratch\\arm\\boottests\\vmsscustomimagecopy"
	//DeploymentRunBasePath = "C:\\project\\deploymentruns\\loosePatformImageCopy"
	//LooseVMPath = "C:\\scratch\\arm\\boottests\\loosevmcustomimagecopy"
	//DeploymentRunBasePath = "C:\\project\\deploymentruns\\loosePatformImageCopy"
	//LooseVMPath = "C:\\scratch\\arm\\boottests\\loosevmcustomimagecopyavere"
	TemplateName = "azuredeploy.json"
	TemplateParametersName = "azuredeploy.parameters.json"
	IsVMSS = false
	UseEventMonitor = true

	EventHubSenderKeyName = "EventHubSenderKeyName"
	EventHubSenderKey = "EventHubSenderKey"
	EventHubNSName = "EventHubNSName"
	EventHubHubName = "EventHubHubName"
)

// VM Status
const (
	VM_NOSTATUS = "Nostatus"
	VM_SUCCEEDED = "Succeeded"
	VM_FAILED = "Failed"
)

func GetTemplatePath(runPath string) string {
	return path.Join(runPath, TemplateName)
}

func GetTemplateParametersPath(runPath string) string {
	return path.Join(runPath, TemplateParametersName)
}

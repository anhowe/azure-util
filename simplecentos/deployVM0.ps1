$resourceGroup="anhowe0308centosd"
New-AzureRmResourceGroup -Force -Name $resourceGroup -Location "canadacentral"
New-AzureRmResourceGroupDeployment -Name $resourceGroup -ResourceGroupName $resourceGroup -TemplateFile ./azuredeploy.json 
#-TemplateParameterFile azuredeploy.parameters.anhowe.json

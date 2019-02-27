$resourceGroup="anhowe0227lina"
New-AzureRmResourceGroup -Force -Name $resourceGroup -Location "eastus"
New-AzureRmResourceGroupDeployment -Name $resourceGroup -ResourceGroupName $resourceGroup -TemplateFile ./azuredeploy.json -TemplateParameterFile azuredeploy.parameters.anhowe.json

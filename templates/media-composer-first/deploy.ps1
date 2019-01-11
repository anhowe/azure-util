$resourceGroup="anhowe0111c"
New-AzureRmResourceGroup -Force -Name $resourceGroup -Location "eastus"
New-AzureRmResourceGroupDeployment -Name $resourceGroup -ResourceGroupName $resourceGroup -TemplateFile ./azuredeploy-auto.json -TemplateParameterFile azuredeploy-auto.parameters.anhowe.json

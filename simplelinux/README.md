# Boot a simple windows VM with managed identity.

This deploys a simple linux VM with managed identity.

<a href="https://portal.azure.com/#create/Microsoft.Template/uri/https://raw.githubusercontent.com/anhowe/azure-util/master/simplelinux/azuredeploy.json" target="_blank">
    <img src="http://azuredeploy.net/deploybutton.png"/>
</a>

## Installing the AZ CLI

Here are the instructions to use az to use the managed identity:

  1. ssh to the node
  1. run `curl -L https://aka.ms/InstallAzureCli | bash` and accept all defaults
  1. run `source ~/.bashrc` to be able to use in current shell
  1. run `az login --identity` to login using the VM identity
  1. run `az group list` to test that az has access to the current resource group
#!/usr/bin/python
import argparse
import base64
import gzip
import json
import os
import shutil
import subprocess
import StringIO
import sys

def buildb64GzipStringFromFile(file):
    # read the script file
    with open(file) as f:
        content = f.read()
    compressedbuffer=StringIO.StringIO()

    # gzip the script file
    # mtime=0 sets a fixed timestamp in GZip header to the Epoch which is January 1st, 1970
    # Make sure it doens't change unless the stream changes 
    with gzip.GzipFile(fileobj=compressedbuffer, mode='wb', mtime=0) as f:
        f.write(content)
    b64GzipStream=base64.b64encode(compressedbuffer.getvalue())

    return b64GzipStream

# processes a Yaml file to be included properly in ARM template
def convertToOneArmTemplateLine(clusterYamlFile):
    # remove the \r\n and include \n in body and escape " to \"
    return  clusterYamlFile.replace("\n", "\\n").replace('"', '\\"')

# Loads the base ARM template file and injects the Yaml for the shell scripts into it.
def processBaseTemplate(baseTemplatePath,
                        clusterInstallScript,
                        jumpboxTemplatePath = None,
                        linuxJumpboxInstallScript = None,
                        swarmWindowsAgentInstallScript = None,
                        additionalFiles = [],
                        windowsAgentDiagnosticsExtensionTemplatePath = None):

    #String to replace in JSON file
    CLUSTER_YAML_REPLACE_STRING  = "#clusterCustomData"
    
    # Load Base Template
    armTemplate = []
    with open(baseTemplatePath) as f:
        armTemplate = f.read()
        
    # Generate cluster Yaml file for ARM
    clusterYamlFile = convertToOneArmTemplateLine(buildb64GzipStringFromFile(clusterInstallScript))
    armTemplate = armTemplate.replace(CLUSTER_YAML_REPLACE_STRING, clusterYamlFile)

    # Make sure the final string is valid JSON
    try:
        json_object = json.loads(armTemplate)
    except ValueError, e:
        print e
        errorFileName = baseTemplatePath + ".err"
        with open(errorFileName, "w") as f:
            f.write(armTemplate)
        print "Invalid armTemplate saved to: " + errorFileName
        raise

    return armTemplate;

if __name__ == "__main__":
    # Parse Arguments
    parser = argparse.ArgumentParser()
    parser.add_argument("-o", "--output_directory",  help="Directory to write templates files to.  Default is current directory.")
    
    args = parser.parse_args()

    if (args.output_directory == None) :
        args.output_directory = os.getcwd()

    args.output_directory = os.path.expandvars(os.path.normpath(args.output_directory))

    if ( os.path.exists(args.output_directory) == False ):
        os.mkdir(args.output_directory)

    # Input Arm Template Artifacts to be processed in
    # Note:  These files are not useable ARM templates on their own or valid JSON
    # They require processing by this script.
    ARM_INPUT_TEMPLATE_TEMPLATE                  = "base-template.json"
    
    # Shell Scripts to load into YAML
    VDBENCH_INSTALL_SCRIPT = "setupMachine.ps1"
    
    # Output ARM Template Files.  WIll Also Output name.parameters.json for each
    ARM_OUTPUT_TEMPLATE                                   = "mainTemplate.json"
    MARKETPLACE_UI_DEFINITION                             = "createUiDefinition.json"
    ARM_OUTPUT_TEMPLATE_FINAL                             = "../azuredeploy-auto.json"
    
    # build the ARM template for jumpboxless
    with open(os.path.join(args.output_directory, ARM_OUTPUT_TEMPLATE), "w") as armTemplate:
        clusterTemplate = processBaseTemplate(
            baseTemplatePath=ARM_INPUT_TEMPLATE_TEMPLATE, 
            clusterInstallScript=VDBENCH_INSTALL_SCRIPT)
        armTemplate.write(clusterTemplate)

    MARKETPLACE_ZIP                                   = "marketplace.zip"

    # zipfile format is not compatible with Azure Marketplace so break out to powershell
    subprocess.call(["C:\\WINDOWS\\system32\\WindowsPowerShell\\v1.0\\powershell.exe", "-Command", "Compress-Archive -Path %s, %s -Force -DestinationPath %s" % (ARM_OUTPUT_TEMPLATE, MARKETPLACE_UI_DEFINITION, MARKETPLACE_ZIP)])
    #    zipf = zipfile.ZipFile(MARKETPLACE_ZIP, 'w', zipfile.ZIP_DEFLATED)
    #    zipf.write(ARM_OUTPUT_TEMPLATE)
    #    zipf.write(MARKETPLACE_UI_DEFINITION)
    #    zipf.close()

    shutil.move(ARM_OUTPUT_TEMPLATE, ARM_OUTPUT_TEMPLATE_FINAL)
    
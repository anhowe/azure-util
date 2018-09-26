#!/usr/bin/python
import base64
import os
import gzip
import re
import StringIO
import sys

# goal "commandToExecute": "[variables('jumpboxWindowsCustomScript')]"

def convertToOneArmTemplateLine(file):
    with open(file) as f:
        content = f.read()

    # convert to one line
    content = content.replace("\\", "\\\\")
    content = content.replace("\r\n", "\\n")
    content = content.replace("\n", "\\n")
    content = content.replace('"', '\\"')

    # replace {{{ }}} with variable names
    return re.sub(r"{{{([^}]*)}}}", r"',variables('\1'),'", content)

def usage():
    print
    print "    usage: %s file1" % os.path.basename(sys.argv[0])
    print
    print "    builds a one line string to send to commandToExecute"

if __name__ == "__main__":
    if len(sys.argv)!=2:
        usage()
        sys.exit(1)

    file = sys.argv[1]
    if not os.path.exists(file):
        print "Error: file %s does not exist"
        sys.exit(2)

    # build the yml file for cluster
    oneline = convertToOneArmTemplateLine(file)

    print "\"customData\": \"[base64(concat('%s'))]\"," % oneline
    #print '"commandToExecute": "powershell.exe -ExecutionPolicy Unrestricted -command \\"%s\\""' % (oneline)
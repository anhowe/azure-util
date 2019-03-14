#!/bin/bash -x

set -x
set -e

ARM_ENDPOINT=https://management.azure.com/metadata/endpoints?api-version=2017-12-01

function wait_arm_endpoint() {
    # ensure the arm endpoint is reachable
    # https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service#getting-azure-environment-where-the-vm-is-running
    if ! retrycmd_if_failure 12 2 curl -m 5 -o /dev/null $ARM_ENDPOINT ; then
        echo "no internet! arm endpoint $ARM_ENDPOINT not reachable.  Please see https://aka.ms/averedocs on what endpoints are required."
        exit 1
    fi
}

function sleep_10_minutes() {
    retries=120
    apt_update_output=/tmp/apt-get-update.out
    for i in $(seq 1 $retries); do
        echo "`date` - tick $i - sleeping for 10 minutes before returning"
        sleep 5
    done
    echo Executed apt-get update $i times
}

function apt_get_update() {
    retries=10
    apt_update_output=/tmp/apt-get-update.out
    for i in $(seq 1 $retries); do
        timeout 120 apt-get update 2>&1 | tee $apt_update_output | grep -E "^([WE]:.*)|([eE]rr.*)$"
        [ $? -ne 0  ] && cat $apt_update_output && break || \
        cat $apt_update_output
        if [ $i -eq $retries ]; then
            return 1
        else sleep 30
        fi
    done
    echo Executed apt-get update $i times
}

function apt_get_install() {
    retries=$1; wait_sleep=$2; timeout=$3; shift && shift && shift
    for i in $(seq 1 $retries); do
        # timeout occasionally freezes
        #echo "timeout $timeout apt-get install --no-install-recommends -y ${@}"
        #timeout $timeout apt-get install --no-install-recommends -y ${@}
        apt-get install --no-install-recommends -y ${@}
        echo "completed"
        [ $? -eq 0  ] && break || \
        if [ $i -eq $retries ]; then
            return 1
        else
            sleep $wait_sleep
            apt_get_update
        fi
    done
    echo Executed apt-get install --no-install-recommends -y \"$@\" $i times;
}

function config_linux() {
	#hostname=`hostname -s`
	#sudo sed -ie "s/127.0.0.1 localhost/127.0.0.1 localhost ${hostname}/" /etc/hosts
	export DEBIAN_FRONTEND=noninteractive  
	apt_get_update
	apt_get_install 20 10 180 curl dirmngr python-pip nfs-common
    apt remove --purge -y python-keyring
    pip install --requirement /opt/avere/python_requirements.txt
}

function main() {
    echo "wait arm endpoint"
    wait_arm_endpoint

    echo "configure linux"
    config_linux

    sleep_10_minutes
}

main
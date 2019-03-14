#!/bin/bash -x

set -x
set -e

function sleep_10_minutes() {
    retries=120
    apt_update_output=/tmp/apt-get-update.out
    for i in $(seq 1 $retries); do
        echo "`date` - tick $i - sleeping for 10 minutes before returning"
        sleep 5
    done
    echo Executed apt-get update $i times
}

sleep_10_minutes

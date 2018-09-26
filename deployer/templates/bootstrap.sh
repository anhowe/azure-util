#!/bin/bash

#
# The following script mounts a default folder round robined across
# the vFXT ip addresses.
#
# Save this script to any Avere vFXT volume, for example:
#     /bootstrap/bootstrap.sh
#
# The following environment variables must be set:
#     NFS_IP_CSV="10.0.0.22,10.0.0.23,10.0.0.24"
#     NFS_PATH=/msazure
#     BASE_DIR=/nfs
#

function retrycmd_if_failure() {
    retries=$1; max_wait_sleep=$2; shift && shift
    for i in $(seq 1 $retries); do
        ${@}
        [ $? -eq 0  ] && break || \
        if [ $i -eq $retries ]; then
            return 1
        else
            sleep $(($RANDOM % $max_wait_sleep))
        fi
    done
}

function mount_round_robin() {
    # to ensure the nodes are spread out somewhat evenly the default
    # mount point is based on this node's IP octet4 % vFXT node count.
    declare -a AVEREVFXT_NODES="($(echo ${NFS_IP_CSV} | sed "s/,/ /g"))"
    OCTET4=$((`ifconfig | sed -En 's/127.0.0.1//;s/.*inet (addr:)?(([0-9]*\\.){3}[0-9]*).*/\\2/p' | sed -e 's/^.*\.\([0-9]*\)/\1/' | sed 's/[^0-9]*//g'`))
    DEFAULT_MOUNT_INDEX=$((${OCTET4} % ${#AVEREVFXT_NODES[@]}))
    ROUND_ROBIN_IP=${AVEREVFXT_NODES[${DEFAULT_MOUNT_INDEX}]}

    DEFAULT_MOUNT_POINT="${BASE_DIR}/default"

    # no need to write again if it is already there
    if ! grep --quiet "${DEFAULT_MOUNT_POINT}" /etc/fstab; then
        echo "${ROUND_ROBIN_IP}:${NFS_PATH}    ${DEFAULT_MOUNT_POINT}    nfs hard,nointr,proto=tcp,mountproto=tcp,retry=30 0 0" >> /etc/fstab
        mkdir -p "${DEFAULT_MOUNT_POINT}"
        chown nobody:nogroup "${DEFAULT_MOUNT_POINT}"
    fi
    if ! grep -qs "${DEFAULT_MOUNT_POINT} " /proc/mounts; then
        retrycmd_if_failure 12 20 mount "${DEFAULT_MOUNT_POINT}" || exit 1
    fi
}

function write_parallelcp() {
    FILENAME=/usr/bin/parallelcp
    sudo touch $FILENAME
    sudo chmod 755 $FILENAME
    sudo /bin/cat <<EOM >$FILENAME
#!/bin/bash

display_usage() {
    echo -e "\nUsage: \$0 SOURCE_DIR DEST_DIR\n"
}

if [  \$# -le 1 ] ; then
    display_usage
    exit 1
fi

if [[ ( \$# == "--help") ||  \$# == "-h" ]] ; then
    display_usage
    exit 0
fi

SOURCE_DIR="\$1"
DEST_DIR="\$2"

if [ ! -d "\$SOURCE_DIR" ] ; then
    echo "Source directory \$SOURCE_DIR does not exist, or is not a directory"
    display_usage
    exit 2
fi

if [ ! -d "\$DEST_DIR" ] && ! mkdir -p \$DEST_DIR ; then
    echo "Destination directory \$DEST_DIR does not exist, or is not a directory"
    display_usage
    exit 2
fi

if [ ! -w "\$DEST_DIR" ] ; then
    echo "Destination directory \$DEST_DIR is not writeable, or is not a directory"
    display_usage
    exit 3
fi

if ! which parallel > /dev/null ; then
    sudo apt-get update && sudo apt install -y parallel
fi

DIRJOBS=225
JOBS=225
find \$SOURCE_DIR -mindepth 1 -type d -print0 | sed -z "s:\$SOURCE_DIR\/::" | parallel --will-cite -j\$DIRJOBS -0 "mkdir -p \$DEST_DIR/{}"
find \$SOURCE_DIR -mindepth 1 ! -type d -print0 | sed -z "s:\$SOURCE_DIR\/::" | parallel --will-cite -j\$JOBS -0 "cp -P \$SOURCE_DIR/{} \$DEST_DIR/{}"
EOM
}

function main() {
    echo "mount round robin default path"
    mount_round_robin

    # add extra bootstrap and installation code here
    # this could be:
    #  - installation bash scripts
    #  - chef and puppet scripts
    #  - ansible scripts
    # when pulling content from the NFS server, ensure to use the round robin path, listed
    # under the default path, something similar to DEFAULT_MOUNT_POINT="${BASE_DIR}/default"
    DEFAULT_MOUNT_POINT="${BASE_DIR}/default"
    SENDER="${DEFAULT_MOUNT_POINT}/bootstrap/eventhubsender"
    VMNAME=$(retrycmd_if_failure 60 5 curl -s -H Metadata:true "http://169.254.169.254/metadata/instance/compute/name?api-version=2017-08-01&format=text")

    #write_parallelcp
    #retrycmd_if_failure 60 5 apt install -y parallel

    #TARGET=/mnt/tools
    TARGET=/opt/tools

    # 14 GB
    #retrycmd_if_failure 60 5 /usr/bin/parallelcp /nfs/default/tools ${TARGET}
    # 5 GB
    #retrycmd_if_failure 60 5 /usr/bin/parallelcp /nfs/default/tools5GB ${TARGET}
    # 1 GB
    #retrycmd_if_failure 60 5 /usr/bin/parallelcp /nfs/default/tools1GB ${TARGET}
    echo "copy directly"
    #retrycmd_if_failure 60 5 /usr/bin/parallelcp ${TARGET} /mnt/tools2
    retrycmd_if_failure 60 5 /usr/bin/parallelcp /nfs/default/tools5GB /mnt/tools2
    echo "copycomplete"
    retrycmd_if_failure 60 5 $SENDER "${VMNAME}"
}

main

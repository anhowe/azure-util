#!/bin/bash

display_usage() { 
    echo -e "\nUsage: $0 SOURCE_DIR\n" 
    echo ""
    echo "\tcreate 1TB of files under the SOURCE_DIR"
}

if [ $# -le 0 ] ; then 
    display_usage
    exit 1
fi 

if [[ ( $# == "--help") ||  $# == "-h" ]] ; then 
    display_usage
    exit 0
fi

SOURCE_DIR="$1"

if [ ! -d "$SOURCE_DIR" ] ; then
    echo "ERROR: Source directory $SOURCE_DIR does not exist, or is not a directory"
    display_usage
    exit 2
fi

while [ $(du -s $SOURCE_DIR | awk '{print $1}') -le 1048576 ] ; do
    PATH1=`cat /dev/urandom| tr -dc 'a-z'|head -c 1`
    PATH2=`cat /dev/urandom| tr -dc 'a-z'|head -c 1`
    FILE=`cat /dev/urandom| tr -dc 'a-z'|head -c 1`
    DIRPATH=$SOURCE_DIR/${PATH1}/${PATH2}
    mkdir -p $DIRPATH
    FULLPATH=$DIRPATH/$FILE
    echo $FULLPATH
    head -c 10M < /dev/urandom > $FULLPATH
    ls -l $FULLPATH
done
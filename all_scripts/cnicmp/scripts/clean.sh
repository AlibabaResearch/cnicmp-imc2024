#!/bin/bash

usage() {
    cat << EOF
Usage:
./clean.sh namespace
EOF
}

if [[ $# -ne 1 ]]; then
    usage
    exit 1
fi

ns=$1


for container in $(crictl ps -q); do
	#echo $container
	crictl stop $container > /dev/null
	crictl rm -f $container > /dev/null
done

for pod in $(crictl pods --namespace $ns -q); do
	#echo $pod
	#crictl stopp $pod > /dev/null
	crictl rmp -f $pod > /dev/null
done

#ovs-vsctl del-br br1
#ovs-vsctl add-br br1
#ovs-vsctl -- --all destroy qos
# crictl rmp -f $(crictl pods --namespace $ns -q)
ifconfig lo 127.0.0.1 up

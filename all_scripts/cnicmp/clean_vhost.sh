pkill -9 close
pkill -9 python3
pkill -9 ovs-server
pkill -9 ovs_server
pkill -9 ovs
pkill -9 qemu
pkill -9 containerd
bash ./setup/ovs_setup.sh
bash ./setup/ovs_setup.sh
bash del_netns.sh
./scripts/clean.sh test


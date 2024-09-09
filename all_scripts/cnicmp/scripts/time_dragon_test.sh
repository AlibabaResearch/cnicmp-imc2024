#! /bin/bash
DIR=$(dirname $0)
kata_config="/opt/kata/share/defaults/kata-containers/configuration-dragonball.toml"
containerd_config="/opt/kata/share/defaults/kata-containers/configuration-dragonball.toml"
crictl_config="/etc/crictl.yaml"
source $DIR/time_test.conf

sed -r -i "s/default_memory = [0-9]+/default_memory = 168/g" $kata_config
sed -r -i "s/snapshotter = \".*\"/snapshotter = \"devmapper\"/g" ${containerd_config}
systemctl restart containerd
sleep 10

#exit 1

start_date=$(date +%m%d%H%M)
base_dir=$(printf "time_dragonball_%s" $start_date)
mkdir -p ${base_dir}
version="${base_dir}/versions.txt"
uname -r >> $version
containerd --version >> $version
crictl --version >> $version
kata-runtime --version >> $version
firecracker --version >> $version
cp $kata_config $base_dir/
cp $containerd_config $base_dir/
cp $crictl_config $base_dir/

#exit 1

for c in ${concurency[@]}; do
    echo "--- dragon $c"
    export result_dir=$(printf "%s/con_%03d" $base_dir $c)
    $DIR/closedloop.sh $c 5 rund
    if [[ $? != 0 ]]; then
        exit $?
    fi
done

sed -r -i "s/snapshotter = \".*\"/snapshotter = \"overlayfs\"/g" ${containerd_config}
systemctl restart containerd
sleep 10
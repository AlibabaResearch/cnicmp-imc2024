#!/usr/bin/env bash
set -e
cd "$(dirname "$0")"

if [ "$(uname)" == "Darwin" ]; then
	export GOOS="${GOOS:-linux}"
fi

export GOFLAGS="${GOFLAGS} -mod=vendor"

mkdir -p "${PWD}/bin"

echo "Building plugins ${GOOS}"
PLUGINS="plugins/meta/* plugins/main/* plugins/ipam/* plugins/ovs/*"
for d in $PLUGINS; do
	if [ -d "$d" ]; then
		plugin="$(basename "$d")"
		if [ "${plugin}" != "windows" ]; then
			echo "   building $plugin"
			${GO:-go} build -o "${PWD}/bin/$plugin" "$@" ./"$d"
		fi
	fi
done

echo "Building ovs-server"
${GO:-go} build -o "${PWD}/bin/ovs-server" "$@" "./pkg/ovs-cni/server"

CNI_BIN_PATH="/opt/cni/bin"
OVS_CNI_NAMES=("ovs-internal" "ovs-ipvlan" "ovs-macvlan" "ovs-ipvtap" "ovs-macvtap")
for cni in "${OVS_CNI_NAMES[@]}"
do
	echo "Copying ${cni} to ${CNI_BIN_PATH} ..."
	cp "${PWD}/bin/${cni}" "${CNI_BIN_PATH}/${cni}"
done
pid=$(crictl runp --runtime=rund /tmp/pod_test/_pod_config_001.yaml)

while true; do
    if crictl inspectp $pid | grep -q "SANDBOX_READY"; then
        break
    else
        sleep 0.01
    fi
done
ip=$(crictl inspectp $pid | grep -e "10.88" | awk -F '\"' '{print $4}' | head -n 1)

crictl inspectp $pid | grep -e IP -e Mac


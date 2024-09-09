pid=$(crictl runp --runtime=kata ./net_test.yaml)
#echo $pid
cid=$(crictl create --no-pull $pid tp-iperf.json ./net_test.yaml)
#echo $cid
crictl start $cid

while true; do
    if crictl ps | grep -q "Running"; then
        break
    else
        sleep 0.01
    fi
done

crictl inspectp $pid | grep -e IP -e Mac
#bash net_test_apply2.sh $cid
ovs-vsctl show | grep veth
rm -f /home/cni/cnicmp_logs/cni_logs/ovs_cni/ovs.log

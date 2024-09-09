DIR=$(dirname $0)/../cnicmp/
res_file="$DIR/../single_pod/dataplane_logs/${1}-res"
peer_ip=$2
monitor_tool="$DIR/scripts/ovs_test_scripts/monitor_resource2.py"
tc_conf_tool="$DIR/scripts/ovs_test_scripts/tc_conf.py"

cid=$(ssh $peer_ip "crictl ps | tail -n 1" | awk '{print $1}')
#$(crictl create --no-pull $pid tp-iperf.json /tmp/pod_test/net_test.yaml)
echo $cid
cid2=$(crictl ps | tail -n 1 | awk '{print $1}')
#$(crictl create --no-pull $pid2 tp-iperf1.json /tmp/pod_test/net_test2.yaml)
echo $cid2
server_ip=$3
#$(ssh $peer_ip "crictl inspectp $pid | grep -e '\"IP\"' | awk 'BEGIN{FS="\""} {print $4}'")
echo $server_ip

echo "start testing"
#start monitoring
#ssh $peer_ip "python3 ${tc_conf_tool} ${res_file}"
#python3 ${tc_conf_tool} ${res_file}

# 1. throughput test
python3 $monitor_tool ${res_file}.client $cid2 &
ssh $peer_ip "python3 ${monitor_tool} ${res_file}.server ${cid} &" &
ssh $peer_ip "crictl exec ${cid} iperf3 -s &" &
sleep 2
crictl exec $cid2 iperf3 -c $server_ip -t 10 #tcp_lat
mon_id=$(ssh $peer_ip "ps -ef | grep  ${monitor_tool} | head -n 1" |  awk '{print $2}')
ssh $peer_ip "kill -9 ${mon_id}"
mon_id=$(ssh $peer_ip "ps -ef | grep  ${monitor_tool} | head -n 1" |  awk '{print $2}')
ssh $peer_ip "kill -9 ${mon_id}"
kill -9 `ps -ef | grep  ${monitor_tool} | head -n 1 |  awk '{print $2}'`
kill -9 `ps -ef | grep  ${monitor_tool} | head -n 1 |  awk '{print $2}'`
kill -9 `ps -ef | grep  ${monitor_tool} | head -n 1 |  awk '{print $2}'`
kill -9 `ps -ef | grep  ${monitor_tool} | head -n 1 |  awk '{print $2}'`
# 2. pps test
#python3 $monitor_tool ${res_file}.pps.client $cid2 &
#python3 $monitor_tool ${res_file}.pps.server $cid &
#crictl exec $cid2 iperf3 -c $server_ip -l 64 -t 10
#kill -9 `ps -ef | grep  $monitor_tool | head -n 1 |  awk '{print $2}'`
#kill -9 `ps -ef | grep  $monitor_tool | head -n 1 |  awk '{print $2}'`

# 3. latency test
ssh $peer_ip "crictl exec ${cid} netserver &" &
sleep 2
crictl exec $cid2 netperf -H $server_ip -l 10 -t TCP_RR -- -d rr -O "THROUGHPUT, THROUGHPUT_UNITS, MIN_LATENCY, MAX_LATENCY, MEAN_LATENCY, STDDEV_LATENCY"

#end 


res_file=$1
DIR=$(dirname $0)/../cnicmp/
monitor_tool="$DIR/scripts/ovs_test_scripts/monitor_resource2.py"
tc_conf_tool="$DIR/scripts/ovs_test_scripts/tc_conf.py"

pid=$(crictl runp --runtime=kata $DIR/../single_pod/net_test.yaml)
#echo $pid
cid=$(crictl create --no-pull $pid tp-iperf.json $DIR/../single_pod/net_test.yaml)
#echo $cid
crictl start $cid

while true; do
    if crictl ps | grep perf0 | grep -q "Running"; then
        break
    else
        sleep 0.01
    fi
done

pid2=$(crictl runp --runtime=kata $DIR/../single_pod/net_test2.yaml)
#echo $pid
cid2=$(crictl create --no-pull $pid2 tp-iperf1.json $DIR/../single_pod/net_test2.yaml)
#echo $cid
crictl start $cid2

while true; do
    if crictl ps | grep perf1 | grep -q "Running"; then
        break
    else
        sleep 0.01
    fi
done

server_ip=$(crictl inspectp $pid | grep -e '"IP"' | awk 'BEGIN{FS="\""} {print $4}')
echo $server_ip

echo "start testing - round $2..."
#start monitoring
python3 $tc_conf_tool $res_file

# 1. throughput test
python3 $monitor_tool ${res_file}.client $cid2 &
python3 $monitor_tool ${res_file}.server $cid &
crictl exec $cid iperf3 -s &
sleep 2
crictl exec $cid2 iperf3 -c $server_ip -t 10 #tcp_lat
kill -9 `ps -ef | grep  $monitor_tool | head -n 1 |  awk '{print $2}'`
kill -9 `ps -ef | grep  $monitor_tool | head -n 1 |  awk '{print $2}'`

# 2. pps test
python3 $monitor_tool ${res_file}.pps.client $cid2 &
python3 $monitor_tool ${res_file}.pps.server $cid &
crictl exec $cid2 iperf3 -c $server_ip -l 64 -t 10
kill -9 `ps -ef | grep  $monitor_tool | head -n 1 |  awk '{print $2}'`
kill -9 `ps -ef | grep  $monitor_tool | head -n 1 |  awk '{print $2}'`

# 3. latency test
crictl exec $cid netserver &
sleep 2
crictl exec $cid2 netperf -H $server_ip -l 10 -t TCP_RR -- -d rr -O "THROUGHPUT, THROUGHPUT_UNITS, MIN_LATENCY, MAX_LATENCY, MEAN_LATENCY, STDDEV_LATENCY"

#end 


TIME_MILLI="date +%s%3N"
start_time=$($TIME_MILLI)

pid=$(crictl runp --runtime=kata /tmp/pod_test/_pod_config_001.yaml)

while true; do
    if crictl inspectp $pid | grep -q "SANDBOX_READY"; then
        break
    else
        sleep 0.01
    fi
done

end_time1=$($TIME_MILLI)

ip=$(crictl inspectp $pid | grep -e "10.88" | awk -F '\"' '{print $4}' | head -n 1)

end_time2=$($TIME_MILLI)

while true; do
    if crictl exec --timeout 1 $1 ping $ip -c 1 | grep -q "ttl"; then
        break
    else
        sleep 0.01
    fi
done

end_time=$($TIME_MILLI)
total_latency=$(($end_time - $start_time))
cmp_t1=$(($end_time1 - $start_time))
cmp_t2=$(($end_time2 - $start_time))

echo $total_latency
echo $cmp_t1
echo $cmp_t2

crictl inspectp $pid | grep -e IP -e Mac


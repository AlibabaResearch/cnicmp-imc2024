#!/bin/bash

usage() {
    cat << EOF
Usage:
./closedloop.sh num_client num_round runtime
EOF
}

if [[ $# -ne 3 ]]; then
    usage
    exit 1
fi

client_num=$1
req_num=$2
runtime=$3

DIR=$(dirname $0)
result_dir=${result_dir:-$(printf "%s_%03d_%s" ${runtime} ${client_num} $(date +%m%d%H%M))}
ns=${ns:-"test"}
prewarm_ns="$ns-prewarm"
tmp_dir="/tmp/pod_test"
kata_log_dir="/home/cni/cnicmp_logs/kata_logs/tmp"
containerd_log_dir="/home/cni/cnicmp_logs/containerd_logs/tmp"
log_tool_dir="$(dirname `readlink -f $0`)/../analysis/log_tool.py"
mon_tool_dir="/home/cni/nsdi/scripts/mon_cpu.py"
cpustat_out=$result_dir/cpu.txt

start_cpu_load() {
    local stat=$(grep 'cpu ' /proc/stat)
    idle1=$(echo $stat | awk '{print $5+$6}')
    total1=$(echo $stat | awk '{print $2+$3+$4+$5+$6+$7+$8}')
}

end_cpu_load() {
    local stat=$(grep 'cpu ' /proc/stat)
    idle2=$(echo $stat | awk '{print $5+$6}')
    total2=$(echo $stat | awk '{print $2+$3+$4+$5+$6+$7+$8}')

    idle=$(($idle2 - $idle1))
    total=$(($total2 - $total1))
    busy=$(($total - $idle))
    echo $busy $total | awk '{print $1,$2,$1/$2*100}' >> $cpustat_out
}

gen_config() {
    pod_config=()
    file=()
    for i in $(seq ${client_num}); do
        local iii=$(printf "%03d" ${i})
        file[$i]="${result_dir}/client${iii}.txt"
        pod_config[$i]="${tmp_dir}/_pod_config_${iii}.yaml"
        cat > ${pod_config[$i]} << EOF
metadata:
  name: sandbox-${iii}
  namespace: $ns
  uid: busybox-sandbox
  attempt: 1
log_directory: $tmp_dir
linux:
  security_context:
    namespace_options:
      network: 2
EOF
    done
}

TIME_MILLI="date +%s%3N" # in milliseconds
# lyz: change the ready condition - until icmp reply
client() {
    # client sandbox.yaml output_filename
    curl -s "http://127.0.0.1:5005/get_lock?op=obtain&stage=total" > log
    #curl -s http://127.0.0.1:5005/get_lock?op=obtain&stage=total > log
    local pod_config=$1
    local output=$2

    local start_time=$($TIME_MILLI)
    pid=$(crictl runp --runtime=$runtime $pod_config)
    while true; do # wait for pod ready
        if crictl inspectp $pid | grep -q "SANDBOX_READY"; then
            break
        else
            sleep 0.01
        fi
    done
    curl -s "http://127.0.0.1:5005/get_lock?op=release&stage=total" > log
    #curl -s http://127.0.0.1:5005/get_lock?op=release&stage=total > log
    local end_time=$($TIME_MILLI)
    local total_latency=$(($end_time - $start_time))
    echo $pid $start_time $end_time $total_latency >> $output
}

run_once() {
    local pid=()
    for c in $(seq ${client_num}); do
	fid=$c+400 
        client ${pod_config[$c]} ${file[$c]} &
        pid+=($!)
    done
    wait ${pid[@]}
}

run_multi() {
    for i in $(seq ${req_num}); do
        local datestr=$(date "+%Y%m%d%H%M%S")
        touch ${kata_log_dir}/kata-${datestr}.txt
        touch ${containerd_log_dir}/containerd-${datestr}.txt
        local iii=$(printf "%03d" $i)
        printf "round $iii start"
        python3 ${mon_tool_dir} &
	start_cpu_load
        local start_time=$($TIME_MILLI)
        run_once
        local end_time=$($TIME_MILLI)
        local total_latency=$(($end_time - $start_time))
        end_cpu_load
        printf " $total_latency clean"
        sleep 5
	kill -9 `ps -ef | grep  mon_cpu | head -n 1 |  awk '{print $2}'`
        python3 ${log_tool_dir}
        clean
        echo " finish " $total_latency
    done
}


# lyz: launch a container to test network setup
#launch_net_container() {
#    $DIR/net_test.sh
#}

clean() {
    timeout 3000 $DIR/clean.sh $ns
    if [[ $? != 0 ]]; then
        printf "\nremove pods failed\n" >&2
        exit 1
    fi
}

clean_config() {
    # rm -r ${tmp_dir}
    echo "clean"
}

prewarm() {
    pod_config_prewarm="${tmp_dir}/_pod_config_prewarm.yaml"
    cat > $pod_config_prewarm << EOF
metadata:
  name: sandbox
  namespace: $prewarm_ns
  uid: busybox-sandbox
  attempt: 1
log_directory: $tmp_dir
linux:
  security_context:
    namespace_options:
      network: 2
EOF
    client $pod_config_prewarm "/dev/null"
    $DIR/clean.sh $ns
}

mkdir -p ${result_dir}
if [[ ! -d ${tmp_dir} ]]; then
    mkdir ${tmp_dir}
fi

mkdir -p ${kata_log_dir}
mkdir -p ${containerd_log_dir}
if [ "$(ls -A ${kata_log_dir})" ]; then
    rm ${kata_log_dir}/*
fi
if [ "$(ls -A ${containerd_log_dir})" ]; then
    rm ${containerd_log_dir}/*
fi

clean
gen_config
#prewarm
run_multi
clean_config

awk '{t+=$4}END{print t/NR}' ${file[@]} >> ${result_dir}/latency.txt

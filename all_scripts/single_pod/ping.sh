pid=$(ctr tasks ls | grep hello | awk '{print $2}')
echo $pid
netnspath=/proc/$pid/ns/net
echo $netnspath
CNI_PATH=/opt/cni/bin ./exec-plugins.sh add $pid $netnspath

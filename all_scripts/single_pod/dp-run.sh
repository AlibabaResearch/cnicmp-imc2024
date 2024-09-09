cni_name=$1
test_round=5
DIR=$(dirname $0)/../cnicmp/
log_path="$DIR/../single_pod/dataplane_logs/"
clean_tool1="bash $DIR/scripts/clean.sh test"
clean_tool2="bash $DIR/del_netns.sh"
clean_tool3="bash $DIR/add_br0_netdev.sh"
clean_tool4="bash $DIR/add_internal.sh"
rebuild_tool=""
tp_file="${log_path}${cni_name}-tp.log"
rm -r $tp_file
for i in $(seq $test_round); do
	$clean_tool3
        $clean_tool4
	echo "start testing - round ${i}..."
	resource_file="${log_path}${cni_name}-res-${i}.log"
	rm -r $resource_file
	bash ./dp-test.sh $resource_file $i >> $tp_file
        $clean_tool1
	$clean_tool2
	$clean_tool1
done


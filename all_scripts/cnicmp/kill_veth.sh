pkill -9 closed
pkill -9 python3
pkill -9 virtiofsd
python3 ./scripts/ovs_test_scripts/del_all_netns.py 
bash ./scripts/clean.sh test

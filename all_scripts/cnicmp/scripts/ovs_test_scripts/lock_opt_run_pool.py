from bayes_opt import BayesianOptimization
import os
current_file_path = os.path.abspath(__file__)
root_dir = os.path.dirname(current_file_path) + "/../../"

def get_time_cost(output_lines):
    for line in output_lines:
        if "finish" in line:
            return int(line.split()[1])
    return 99999999

def clean_containers():
    os.system(f"bash {root_dir}/add_br0_netdev.sh")
    os.system(f"bash {root_dir}/scripts/clean.sh test")
    os.system(f"bash {root_dir}/del_netns.sh")
    os.system(f"bash {root_dir}/scripts/clean.sh test")
    os.system("ovs-vsctl -- --all destroy qos")
    os.system("pkill -9 ovs-server")

concurrency = [1, 20, 40, 100, 150, 200, 300, 400]
def translate_param(lock_n):
    lock = int(lock_n)
    lock = min(400, lock)
    lock = max(1, lock)
    #lock = concurrency[int(lock_n)]
    #print(lock)
    return lock

def run_launch(lock_n1, lock_n2, lock_n3, lock_n4, lock_n5):#, lock_total):
    clean_containers()
    lock_n1 = translate_param(lock_n1)
    lock_n2 = translate_param(lock_n2)
    lock_n3 = translate_param(lock_n3)
    lock_n4 = translate_param(lock_n4)
    lock_n5 = translate_param(lock_n5)
    lock_total = 400 #translate_param(lock_total)
    # start lock server 
    os.system("nohup python3 {}/scripts/ovs_test_scripts/lock_opt_server.py {} {} {} {} {} {} > flask.log &".format(root_dir, lock_n1, lock_n2, lock_n3, lock_n4, lock_n5, lock_total))
    print("nohup python3 {}/scripts/ovs_test_scripts/lock_opt_server.py {} {} {} {} {} {} > flask.log &".format(root_dir, lock_n1, lock_n2, lock_n3, lock_n4, lock_n5, lock_total))
    #exit()
    # start tests
    cmd = "python3 tenant-test-pool.py ovs-dev-pool 400 1 0 0 kata > raw.log" #"bash /home/cni/cnicmp/scripts/time_kata_test_net.sh"
    os.system(cmd)
    with open("./raw.log", "r") as f:
        raw_results = f.readlines()
    #raw_results = os.popen(cmd).readlines()
    print(raw_results)
    time_cost = get_time_cost(raw_results)
    print("{} {} {} {} {} {} Obtained cost: {} ".format(lock_n1, lock_n2, lock_n3, lock_n4, lock_n5, lock_total, time_cost))
    # stop lock server
    for i in range(2):
        os.system("kill -9 `ps -ef | grep  lock_opt_server | head -n 1 |  awk '{print $2}'`")
    return -time_cost # in ms

def run_launch_only_total(lock_total):
    lock_n1 = 400
    lock_n2 = 400
    lock_n3 = 400
    lock_n4 = 400
    lock_n5 = 400
    lock_total = translate_param(lock_total)
    # start lock server 
    os.system("nohup python3 {}/scripts/ovs_test_scripts/lock_opt_server.py {} {} {} {} {} {} > flask.log &".format(root_dir, lock_n1, lock_n2, lock_n3, lock_n4, lock_n5, lock_total))
    print("nohup python3 {}/scripts/ovs_test_scripts/lock_opt_server.py {} {} {} {} {} {} > flask.log &".format(root_dir, lock_n1, lock_n2, lock_n3, lock_n4, lock_n5, lock_total))
    #exit()
    # start tests
    cmd = "python3 tenant-test.py ovs-dev-pool 400 1 0 0 > raw.log" #"bash /home/cni/cnicmp/scripts/time_kata_test_net.sh"
    os.system(cmd)
    with open("./raw.log", "r") as f:
        raw_results = f.readlines()
    #raw_results = os.popen(cmd).readlines()
    print(raw_results)
    time_cost = get_time_cost(raw_results)
    print("{} {} {} {} {} {} Obtained cost: {} ".format(lock_n1, lock_n2, lock_n3, lock_n4, lock_n5, lock_total, time_cost))
    # stop lock server
    for i in range(2):
        os.system("kill -9 `ps -ef | grep  lock_opt_server | head -n 1 |  awk '{print $2}'`")
    #clean_containers()
    return -time_cost # in ms

pbounds = {'lock_n1': (1, 401), 'lock_n2': (1, 401), 'lock_n3': (1, 401),
            'lock_n4': (1, 401), 'lock_n5': (1, 401)}#, 'lock_total': (1, 401)}

# simplify 1 - discretized choices
# pbounds = {'lock_n1': (0, len(concurrency) - 0.001), 'lock_n2': (0, len(concurrency) - 0.001), 'lock_n3': (0, len(concurrency) - 0.001),
#             'lock_n4': (0, len(concurrency) - 0.001), 'lock_n5': (0, len(concurrency) - 0.001)}#, 'lock_total': (0, len(concurrency) - 0.001)}

# simplify 2 - adjust only total concurrency
#pbounds = {'lock_total': (1, 401)}

optimizer = BayesianOptimization(run_launch, pbounds)

#optimizer.initialize([[400, 400, 400, 400, 400, 100]])

optimizer.maximize(init_points=5, n_iter=50)

print(optimizer.max)

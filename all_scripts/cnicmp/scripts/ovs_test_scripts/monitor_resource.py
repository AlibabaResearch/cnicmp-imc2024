import os 
import time
import sys

marker = sys.argv[1] 
cid = sys.argv[2]
interval = 0.1

test_id= int(time.time())
log_path = marker#"/home/cni/cnicmp_logs/dataplane_logs/res-{}.log".format(marker)
cmd = "crictl stats {}".format(cid)

with open(log_path, "w") as f:
    while True:
        ts = int(time.time() * 1000)
        line = os.popen(cmd).readlines()[1].split()
        #print(line)
        cpu_util = line[1]
        mem_util = line[2]
        f.write("{}\t{}\t{}\n".format(ts, cpu_util, mem_util))
        f.flush()
        time.sleep(interval)


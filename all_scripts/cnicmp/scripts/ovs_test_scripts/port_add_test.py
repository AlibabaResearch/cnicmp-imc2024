import os
import sys
import threading
import time

br_name = sys.argv[1]
port_type = sys.argv[2]
port_num = eval(sys.argv[3])

class MyThreading(threading.Thread):
    def __init__(self, pid):
        super(MyThreading, self).__init__()
        self.id = pid

    def run(self):
        os.system("ovs-vsctl add-port {} port{} -- set Interface port{} type={}".format(br_name, self.id, self.id, port_type))

os.system("ovs-vsctl del-br {}".format(br_name))
os.system("ovs-vsctl add-br {} -- set bridge {} datapath_type=netdev".format(br_name, br_name))
threads = []

start_t = time.time()
for pid in range(port_num):
    t = MyThreading(pid)
    threads.append(t)
    t.start()
for t in threads: t.join()

end_t = time.time()
#os.system("ovs-vsctl del-br {}".format(br_name))
x = os.popen("ovs-vsctl show | grep Port | wc -l").readlines()
print(x[0].strip())
end_t2 = time.time()

print("Creating {} {} ports takes {}s in total.".format(port_num, port_type, end_t - start_t))
print("Deletion takes {}s in total.".format(end_t2 - end_t))

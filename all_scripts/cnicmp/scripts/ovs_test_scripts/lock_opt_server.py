import sys
import os
current_file_path = os.path.abspath(__file__)
current_dir = os.path.dirname(current_file_path)
sys.path.append(current_dir)
from flask import Flask, request
import threading
import random
import time
import logging

app = Flask(__name__)
app.logger.setLevel(logging.ERROR)

cni_stage_num = int(sys.argv[1])
kata_resource1_stage_num = int(sys.argv[2])
kata_interface_stage_num = int(sys.argv[3])
kata_start_vm_stage_num = int(sys.argv[4])
kata_resource2_stage_num = int(sys.argv[5])
total_stage_num = int(sys.argv[6])

# config the two parameters if new stages need locking
lock_labels = ["cniAdd", "resource1", "kataInterface", "startVM", "resource2", "total"]
lock_nums = [cni_stage_num, kata_resource1_stage_num, kata_interface_stage_num,
            kata_start_vm_stage_num, kata_resource2_stage_num, total_stage_num]
# corresponding
assert len(lock_labels) == len(lock_nums)
stage_num = len(lock_labels)
access_locks1 = [threading.Lock() for i in range(stage_num)]
access_locks2 = [threading.Lock() for i in range(stage_num)]
release_counts = [0 for i in range(stage_num)]
obtain_counts = [0 for i in range(stage_num)]

BARRIER_MAX = 1
barr_num = 0
barr_num2 = 0
barr_lock = threading.Lock()

# wget http://127.0.0.1:5000/get_lock?op=obtain&stage=obtain
@app.route('/get_lock')
def get_lock():
    global access_locks1, access_locks2
    global release_counts, obtain_counts
    op_type = request.args.get('op')
    stage_name = request.args.get('stage')
    assert stage_name in lock_labels
    sid = lock_labels.index(stage_name)
    if op_type == "obtain":
        access_locks1[sid].acquire()
        while lock_nums[sid] == 0:
            time.sleep(0.1)
        access_locks2[sid].acquire()
        lock_nums[sid] -= 1
        obtain_counts[sid] += 1
        print("obtain lock: {} - remain{} - obtained{}".format(stage_name, lock_nums[sid], obtain_counts[sid]))
        access_locks2[sid].release()
        access_locks1[sid].release()
    elif op_type == "release":
        access_locks2[sid].acquire()
        lock_nums[sid] += 1
        release_counts[sid] += 1
        print("release lock: {} - remain{} -  released{}".format(stage_name, lock_nums[sid], release_counts[sid]))
        access_locks2[sid].release()
    else:
        raise Exception("Unsupported lock op")
    return "success"


@app.route('/barrierVM')
def barrier():
    global barr_num2
    barr_lock.acquire()
    barr_num2 += 1
    print("barrierVM num:", barr_num2)
    barr_lock.release()
    while barr_num2 % BARRIER_MAX != 0:
        time.sleep(0.05)
    return "success"

if __name__ == '__main__':
    app.run(debug=False, threaded=True, processes=1, port=5005)

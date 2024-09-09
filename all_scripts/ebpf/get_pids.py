import subprocess
import re
import time

interval = 0.1
num_of_process = 100

pids = []

def get_container_pids():
    global pids
    # 建立ps和grep命令
    ps_command = ['ps', '-ef']
    grep_command = ['grep', '[c]ontainerd-shim']

    # 执行ps命令
    ps = subprocess.Popen(ps_command, stdout=subprocess.PIPE)

    # 将ps的输出作为grep的输入，并执行grep命令
    grep = subprocess.Popen(grep_command, stdin=ps.stdout, stdout=subprocess.PIPE)

    # 关闭ps的输出，以便grep可以从管道读取输入
    ps.stdout.close()

    # 获取grep的输出
    output, _ = grep.communicate()

    # 定义用于匹配PID的正则表达式
    pid_re = re.compile(r'\s*(\d+)\s+')

    # 解析输出并打印PID
    for line in output.decode('utf-8').splitlines():
        print(line)
        pid = line.split()[1]
        print("PID:", pid)
        if pid not in pids: pids.append(pid) 
        # match = pid_re.match(line)
        # if match:
        #     pid = match.group(1)
        #     print("PID:", pid)

while len(pids) < num_of_process:
    get_container_pids()
    time.sleep(interval)


# write to a log file
msg = "\t".join(pids) + "\n"
with open("./pids", "w") as f:
    f.write(msg)


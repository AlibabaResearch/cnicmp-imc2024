import os

netns_name = []
for line in os.popen("ip netns show | grep -e cni- -e cnitest").readlines():
    #print(line)
    netns_name.append(line.split()[0])
    name = line.split()[0]
    os.system("ip netns del {}".format(name))

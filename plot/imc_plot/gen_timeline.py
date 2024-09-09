import os 

ids = [0, 1, 2, 3, 4, 6, -5, -6, -3, -2, -1 ]

for id in ids:
    os.system("python3 ./pooling_results_breakdown_fixed.py {}".format(id))

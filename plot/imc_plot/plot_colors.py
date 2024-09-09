import numpy as np

colors = [
    [8, 29, 89],
    [33, 49, 140],
    [34, 84, 163],
    [30, 128, 184],
    [48, 165, 194],
    [91, 191, 192],
    [149, 213, 184],
    [205, 235, 179],
    [238, 248, 180],
    [254, 254, 215]
]

colors = np.array(colors) / 256
circled_num = [chr(9312 + i) for i in range(20)]

def get_colors(num):
    sample_interval = int(len(colors)/num) 
    sample_offset = 0#len(colors) % num
    if sample_interval == 1: sample_offset = 0
    res = []
    i = sample_offset
    while i < len(colors):
        res.append(colors[i])
        i += sample_interval
    return res

colors2 = [np.array([34, 84, 163])/256, "darkorange", "purple",  "olive", "black", "darkgreen", "firebrick", "teal", "chocolate"]

colors_pack1 = [
    [255,127,0],
    [55,126,184],
    [77,175,74],
    [152,78,163],
    [228,26,28],
    [255,255,51],
    [166,86,40],
    [247,129,191],
    [153,153,153]
]
colors_pack1 = np.array(colors_pack1) / 256

colors_pack2 = [
    [213,62,79],
    [244,109,67],
    [253,174,97],
    [254,224,139],
    [255,255,191],
    [230,245,152],
    [171,221,164],
    [102,194,165],
    [50,136,189]
]
colors_pack2 = np.array(colors_pack2) / 256


colors_pack3 = [
    [165,0,38],
    [215,48,39],
    [244,109,67],
    [253,174,97],
    [254,224,144],
    #[171,217,233],
    [91, 191, 192],
    [116,173,209],
    [69,117,180],
    [49,54,149]
]
colors_pack3 = np.array(colors_pack3) / 256

hatch_patterns_raw = ["/",  "-", "+", "\\", "*", "|", ".",  "x", "o"]
hatch_patterns_density = 2
hatch_patterns = [i * hatch_patterns_density for i in hatch_patterns_raw] 

def sample_colors(colors, num):
    res = []
    for i in range(num):
        if i % 2 == 0:
            res.append(colors[-int(i/2)-1])
        else:
            res.append(colors[int(i/2)])
    return res

def sample_colors2(colors, internal):
    res = []
    for i in range(0, len(colors), internal):
        res.append(colors[i])
    return res
import psutil

# 获取内存使用情况
memory = psutil.virtual_memory()

# 获取总内存大小
total_memory = memory.total/(1024*1024)  # 转换为MB
# 获取已使用内存大小
used_memory = memory.used/(1024*1024)  # 转换为MB
# 获取可用内存大小
available_memory = memory.available/(1024*1024)  # 转换为MB
# 获取内存使用率
memory_percent = memory.percent

cpu_percent = psutil.cpu_percent(interval=0.1, percpu=False)

print("{}\t{}".format(cpu_percent, memory_percent))
print(f"总的内存大小：{total_memory}MB")
print(f"已使用内存大小：{used_memory}MB")
print(f"可用内存大小：{available_memory}MB")
print(f"内存使用率：{memory_percent}%")

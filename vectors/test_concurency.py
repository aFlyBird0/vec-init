import re

# 打开文件并读取内容
with open('init/name.fvecs.redis', 'r') as file:
    lines = file.readlines()

# 从每行中提取ID
ids = list()
for line in lines:
    match = re.search('<(\d+)>', line)
    if match:
        id = int(match.group(1))
        ids.append(id)

# 找到缺失或重复的ID
missing_ids = set(range(10000)) - set(ids)
duplicate_ids = [id for id in ids if list(ids).count(id) > 1]

# 输出结果
print("Missing IDs: ", missing_ids)
print("Duplicate IDs: ", duplicate_ids)

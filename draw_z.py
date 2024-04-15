import jsonlines
import matplotlib.pyplot as plt
import sys
import os

# 从命令行参数获取JSONL文件路径
def get_jsonl_file_path():
    if len(sys.argv) != 2:
        print("Usage: python script.py <jsonl_file>")
        sys.exit(1)
    return sys.argv[1]

# 读取JSONL文件并计算订单数量的变化
def process_jsonl_file(file_path):
    order_changes = []
    with jsonlines.open(file_path) as reader:
        order_count = 0
        for obj in reader:
            if obj["OrderType"] == "New":
                order_count += 1
            elif obj["OrderType"] == "Cancel":
                order_count -= 1
            order_changes.append(order_count)
    return order_changes

# 绘制折线图并保存为同名的jpg文件
def plot_order_changes(order_changes, file_path):
    plt.plot(range(1, len(order_changes) + 1), order_changes, marker='o', linestyle='-')
    plt.xlabel('Order Number')
    plt.ylabel('Orders on Order Book')
    plt.title('Order Book Dynamics')
    plt.grid(True)
    plt.savefig(os.path.splitext(file_path)[0] + "_num.jpg")
    plt.close()

# 主函数
def main():
    file_path = get_jsonl_file_path()
    order_changes = process_jsonl_file(file_path)
    plot_order_changes(order_changes, file_path)

if __name__ == "__main__":
    main()

import sys
import matplotlib.pyplot as plt
import pandas as pd

def plot_from_csv(csv_filename):
    # 读取CSV文件
    df = pd.read_csv(csv_filename)

    # 根据CostMillisecond排序
    df = df.sort_values(by='CostMillisecond')

    # 折线图
    plt.figure(figsize=(10, 5))
    plt.plot(df['ClientOrderID'], df['CostMillisecond'], marker='o', linestyle='-')
    plt.title('Line Plot of CostMillisecond')
    plt.xlabel('ClientOrderID')
    plt.ylabel('CostMillisecond')
    plt.xticks(rotation=45)
    plt.grid(True)
    plt.tight_layout()

    # 保存折线图为同名的JPEG文件
    plt.savefig(csv_filename.replace('.csv', '_line_plot.jpg'))

    # 柱状图
    plt.figure(figsize=(10, 5))
    plt.bar(df['ClientOrderID'], df['CostMillisecond'])
    plt.title('Bar Plot of CostMillisecond')
    plt.xlabel('ClientOrderID')
    plt.ylabel('CostMillisecond')
    plt.xticks(rotation=45)
    plt.grid(True)
    plt.tight_layout()

    # 保存柱状图为同名的JPEG文件
    plt.savefig(csv_filename.replace('.csv', '_plot.jpg'))

    plt.close('all')

if __name__ == "__main__":
    # 检查命令行参数是否正确
    if len(sys.argv) != 2:
        print("Usage: python script.py <csv_filename>")
        sys.exit(1)

    csv_filename = sys.argv[1]

    # 绘制折线图和柱状图并保存为图片
    plot_from_csv(csv_filename)

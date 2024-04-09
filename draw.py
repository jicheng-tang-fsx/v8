import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import os
import sys

def generate_and_save_histogram(csv_file_path):
    """
    读取CSV文件，绘制耗时的直方图，并以高分辨率保存为同名的JPG文件。
    直方图的bins以数量1为尺度，更密集。图形布局将被优化以增加精确性。

    参数:
    csv_file_path: str，CSV文件的路径。
    """
    # 载入CSV文件
    data = pd.read_csv(csv_file_path)

    # 计算bins的数量，基于最大值和最小值的差，加1确保覆盖所有可能的值
    min_value = data['CostMillisecond'].min()
    max_value = data['CostMillisecond'].max()
    bins = int(max_value - min_value + 1)

    # 使用Seaborn绘制直方图
    plt.figure(figsize=(10, 6))
    sns.histplot(data['CostMillisecond'], kde=False, color='skyblue', bins=bins, alpha=0.7)
    plt.title('Cost Millisecond Distribution')
    plt.xlabel('Cost Millisecond')
    plt.ylabel('Frequency')

    # 优化图形布局
    plt.tight_layout()

    # 构建保存文件的路径，将.csv扩展名替换为.jpg
    save_file_path = os.path.splitext(csv_file_path)[0] + '.jpg'

    # 保存图表为JPG文件，提高分辨率
    plt.savefig(save_file_path, dpi=300)

    # 清除当前图形
    plt.close()

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python script_name.py <path_to_csv_file>")
    else:
        csv_file_path = sys.argv[1]
        generate_and_save_histogram(csv_file_path)

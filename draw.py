import pandas as pd
import matplotlib.pyplot as plt
import sys

def plot_cost_millisecond(csv_file):
    # 读取CSV文件
    df = pd.read_csv(csv_file)
    
    # 创建一个包含4个子图的图形窗口
    fig, axs = plt.subplots(4, 1, figsize=(10, 20))
    
    # 定义时间字段列表
    time_fields = ['OmsCostTime1', 'MatchCostTime', 'OmsCostTime2', 'TotalCostTime']
    
    # 为每个时间字段绘制折线图
    for i, field in enumerate(time_fields):
        axs[i].plot(df[field], marker='o', linestyle='-', markersize=4)
        axs[i].set_title(f'{field} Over Time')
        axs[i].set_xlabel('Order Index')
        axs[i].set_ylabel(f'{field} (ms)')
    
    # 调整子图之间的间距
    plt.tight_layout()
    
    # 保存图像为JPG，文件名与CSV文件同名
    output_file = csv_file.rsplit('.', 1)[0] + '.jpg'
    plt.savefig(output_file, format='jpg', dpi=150)
    plt.close()
    print(f'Plot saved as {output_file}')

if __name__ == '__main__':
    if len(sys.argv) != 2:
        print("Usage: python script.py <path_to_csv_file>")
    else:
        csv_file = sys.argv[1]
        plot_cost_millisecond(csv_file)

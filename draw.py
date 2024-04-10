import pandas as pd
import matplotlib.pyplot as plt
import sys

def plot_cost_millisecond(csv_file):
    # 读取CSV文件
    df = pd.read_csv(csv_file)
    
    # 创建图形和轴对象，增加图形的高度以便于文本放置
    fig, ax = plt.subplots(figsize=(10, 8))
    
    # 定义时间字段和颜色
    time_fields = ['OmsCostTime1', 'MatchCostTime', 'OmsCostTime2', 'TotalCostTime']
    colors = ['blue', 'green', 'red', 'purple']  # 为每个折线图指定颜色
    
    # 为每个时间字段绘制折线图
    for field, color in zip(time_fields, colors):
        ax.plot(df[field], marker='o', linestyle='-', markersize=4, color=color, label=field)
    
    # 添加图例
    ax.legend()
    
    # 添加图表标题和坐标轴标签
    ax.set_title('Cost Milliseconds Over Time')
    ax.set_xlabel('Order Index')
    ax.set_ylabel('Cost Milliseconds')
    
    # 调整布局以确保底部文本可见
    plt.subplots_adjust(bottom=0.25)
    
    # 在调整后的图形下方添加文本说明
    ax.text(0.5, -0.15, "OmsCostTime1: Delay in processing orders from clients.\n"
                        "OmsCostTime2: Delay in processing returns from the matching engine to clients.",
            transform=ax.transAxes, fontsize=10, color='black', ha='center', va='top')
    
    # 保存图像为JPG，文件名与CSV文件同名
    output_file = csv_file.rsplit('.', 1)[0] + '.jpg'
    plt.savefig(output_file, format='jpg', dpi=1000)
    plt.close()
    print(f'Plot saved as {output_file}')

if __name__ == '__main__':
    if len(sys.argv) != 2:
        print("Usage: python script.py <path_to_csv_file>")
    else:
        csv_file = sys.argv[1]
        plot_cost_millisecond(csv_file)

import pandas as pd
import matplotlib.pyplot as plt
import sys

def plot_cost_millisecond(csv_file):
    # 读取CSV文件
    df = pd.read_csv(csv_file)
    
    # 绘制折线图
    plt.figure(figsize=(10, 6))
    plt.plot(df['CostMillisecond'], marker='o', linestyle='-', markersize=4)
    plt.title('Cost Millisecond Over Time')
    plt.xlabel('Order Index')
    plt.ylabel('Cost Millisecond')
    
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

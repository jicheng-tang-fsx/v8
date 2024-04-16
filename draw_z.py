import pandas as pd
import matplotlib.pyplot as plt
import sys

def read_and_plot(file_name):
    # Read the JSONL data into a DataFrame
    df = pd.read_json(file_name, lines=True)

    # Assign numerical values to order types for calculation: +1 for New, -1 for Cancel
    df['OrderChange'] = df['OrderType'].map({'New': 1, 'Cancel': -1})

    # Calculate the cumulative sum to reflect the running total of active orders
    df['CumulativeOrders'] = df['OrderChange'].cumsum()

    # Plot the data
    plt.figure(figsize=(10, 5))
    plt.plot(df['LogTime'], df['CumulativeOrders'], marker='o', linestyle='-', color='b')
    plt.title('Change in Order Numbers Over Time')
    plt.xlabel('Time')  # Set x-axis label to 'Time'
    plt.ylabel('Cumulative Order Changes')

    # Remove all x-axis tick labels (date and time values) and set a general label
    plt.xticks([])  # This removes all x-axis tick marks and labels

    plt.grid(True)  # Enable grid for easier visual alignment
    plt.tight_layout()  # Adjust layout to make sure everything fits without overlap

    # Save the plot as a JPG file with the same name as the input file
    output_file_name = file_name.rsplit('.', 1)[0] + '_num.jpg'
    plt.savefig(output_file_name)
    print(f"Plot saved as {output_file_name}")

if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python script_name.py <filename>")
    else:
        file_name = sys.argv[1]
        read_and_plot(file_name)

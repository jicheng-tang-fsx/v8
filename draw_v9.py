import pandas as pd
import matplotlib.pyplot as plt
import json
import argparse
import os

def load_jsonl(input_path):
    """Load data from a JSONL file into a pandas DataFrame."""
    data = []
    with open(input_path, 'r') as file:
        for line in file:
            data.append(json.loads(line))
    return pd.DataFrame(data)

def plot_histogram(data, output_path):
    """Plot histogram and save it as a JPEG file, displaying minimal x-axis information."""
    fig, ax = plt.subplots(figsize=(10, 6))
    data.plot(kind='bar', ax=ax, color=['skyblue', 'salmon'], alpha=0.75)
    
    # Remove x-axis labels and ticks
    ax.set_xticklabels([])
    ax.set_xticks([])

    # Set x-axis label to "Time"
    ax.set_xlabel('Time')
    
    # Custom title including symbol information and description
    ax.set_title('Orders per Second - symbol=7203, HRT send order behavior')
    ax.set_ylabel('Number of Orders')
    plt.tight_layout()
    plt.savefig(output_path)
    plt.close()

def main(input_path):
    """Main function to load, process data, and save plot."""
    df = load_jsonl(input_path)

    # Convert LogTime to datetime
    df['LogTime'] = pd.to_datetime(df['LogTime'], format='%m/%d/%Y %H:%M:%S.%f')

    # Truncate milliseconds to group by second
    df['second'] = df['LogTime'].dt.floor('S')

    # Count "New" and "Cancel" orders per second
    summary = df.groupby(['second', 'OrderType']).size().unstack(fill_value=0)

    # Ensure all necessary columns exist
    if 'New' not in summary.columns:
        summary['New'] = 0
    if 'Cancel' not in summary.columns:
        summary['Cancel'] = 0

    # Generate the output filename from the input filename
    output_filename = f"{os.path.splitext(input_path)[0]}.jpg"

    # Plot and save the histogram
    plot_histogram(summary, output_filename)
    print(f"Plot saved as {output_filename}")

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description="Process JSONL files and plot data.")
    parser.add_argument('input_file', metavar='input.jsonl', type=str, help='Input JSONL file path')
    
    args = parser.parse_args()
    
    main(args.input_file)

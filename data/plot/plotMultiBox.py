import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import os

def load_and_process_file(file_path):
    df = pd.read_csv(file_path, header=None, names=['event_number', 'arrival_ts', 'completion_ts'])
    df['delay'] = (df['completion_ts'] - df['arrival_ts']) / 1_000_000
    return df['delay']

def plot_min_max_comparison(file_paths, labels, output_path):
    colors = ['#1f77b4', '#ff7f0e', '#2ca02c', '#d62728', '#9467bd', '#8c564b']
    stats = []
    for file_path in file_paths:
        delays = load_and_process_file(file_path)
        stats.append({
            'min': np.percentile(delays, 1),
            'max': np.percentile(delays, 99.99)
        })

    plt.figure(figsize=(10, 6))
    rect_width = 0.1
    positions = np.arange(1, len(labels) + 1) * 0.25

    for i, (stat, pos, color) in enumerate(zip(stats, positions, colors)):
        height = stat['max'] - stat['min']
        bottom = stat['min']
        plt.bar(x=pos, height=height, bottom=bottom, width=rect_width,
               color=color, alpha=0.7, label=labels[i])
        cap_width = rect_width * 1.2
        plt.hlines(y=stat['min'], xmin=pos-cap_width/2, xmax=pos+cap_width/2, color=color, linewidth=2)
        plt.hlines(y=stat['max'], xmin=pos-cap_width/2, xmax=pos+cap_width/2, color=color, linewidth=2)

    plt.title('Min-Max Delay Comparison')
    plt.ylabel('Processing Delay (ms)')
    plt.grid(True, alpha=0.3)
    plt.xticks(positions, labels, rotation=45 if max(len(label) for label in labels) > 10 else 0)
    plt.legend(bbox_to_anchor=(1.05, 1), loc='upper left')
    plt.xlim(positions[0] - 0.25, positions[-1] + 0.25)
    plt.tight_layout()
    plt.savefig(output_path, format='pdf', bbox_inches='tight', dpi=300)
    plt.close()

if __name__ == "__main__":
    files = [
        "../data/testResults/1.02.2025/NonSimulation_runs/sepsis/delay_result_sepsis.csv",
        "../data/testResults/1.02.2025/NonSimulation_runs/bpic2012/delay_result_bpic2012.csv",
        "../data/testResults/1.02.2025/NonSimulation_runs/trafficFines/delay_result_motivating_GCdefault.csv"
    ]
    labels = ["Sepsis", "BPIC12", "RTF"]
    output_path = "../data/testResults/1.02.2025/NonSimulation_runs/delay_comparison_boxplot.pdf"

    plot_min_max_comparison(files, labels, output_path)
    print(f"Min-max comparison plot saved to: {output_path}")

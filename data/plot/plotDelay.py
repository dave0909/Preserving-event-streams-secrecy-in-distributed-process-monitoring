import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import os

def analyze_event_delays(csv_path, output_dir):
    # Read CSV file without headers
    df = pd.read_csv(csv_path, header=None, names=['event_number', 'arrival_ts', 'completion_ts'])

    # Sort the dataframe by arrival timestamp
    df = df.sort_values('arrival_ts')

    # Convert nanoseconds to milliseconds for better readability
    df['delay'] = (df['completion_ts'] - df['arrival_ts']) / 1_000_000  # Convert to milliseconds

    # Calculate percentage based on sorted order
    total_events = len(df)
    df['completion_percentage'] = (df.index + 1) * 100 / total_events

    # Calculate moving average for delay
    window_size = max(1, int(total_events * 0.05))  # Use 5% of total events as window
    df['delay_ma'] = df['delay'].rolling(window=window_size, center=True).mean()

    # Calculate metrics
    metrics = {
        'average_delay': df['delay'].mean(),
        'median_delay': np.median(df['delay']),
        'max_delay': df['delay'].max(),
        'min_delay': df['delay'].min(),
        'percentile_1': np.percentile(df['delay'], 1),
        'percentile_5': np.percentile(df['delay'], 5),
        'percentile_10': np.percentile(df['delay'], 10),
        'percentile_99': np.percentile(df['delay'], 99.99),
        'total_events': total_events
    }

    # Save metrics to CSV
    metrics_df = pd.DataFrame([metrics])
    metrics_output_path = os.path.join(output_dir, 'delay_metrics.csv')
    metrics_df.to_csv(metrics_output_path, index=False)

    # Figure 1: Main delay plot
    plt.figure(figsize=(12, 6))
    plt.plot(df['completion_percentage'], df['delay'], marker='', linestyle='-',
             color='blue', linewidth=1.5, label='Delay')
    plt.plot(df['completion_percentage'], df['delay_ma'], color='red',
             linewidth=2, label=f'Moving Average (window={window_size})')

    # Add percentile and median lines
    plt.axhline(y=metrics['percentile_1'], color='r', linestyle='--', alpha=0.5,
                label=f'1st percentile ({metrics["percentile_1"]:.2f}ms)')
    plt.axhline(y=metrics['percentile_5'], color='g', linestyle='--', alpha=0.5,
                label=f'5th percentile ({metrics["percentile_5"]:.2f}ms)')
    plt.axhline(y=metrics['percentile_10'], color='orange', linestyle='--', alpha=0.5,
                label=f'10th percentile ({metrics["percentile_10"]:.2f}ms)')
    plt.axhline(y=metrics['percentile_99'], color='magenta', linestyle='--', alpha=0.7,
                    label=f'99th percentile ({metrics["percentile_99"]:.2f}ms)')
    plt.axhline(y=metrics['median_delay'], color='purple', linestyle='--', alpha=0.7,
                label=f'Median ({metrics["median_delay"]:.2f}ms)')


    plt.title('Event Processing Delay vs Run Completion')
    plt.xlabel('Run Completion (%)')
    plt.ylabel('Processing Delay (ms)')
    plt.grid(True, alpha=0.3)
    #plt.legend(loc='center left', bbox_to_anchor=(1, 0.5))

    # Add statistics as text
    stats_text = (f'Average Delay: {metrics["average_delay"]:.2f}ms\n'
                 f'Median Delay: {metrics["median_delay"]:.2f}ms\n'
                 f'Max Delay: {metrics["max_delay"]:.2f}ms\n'
                 f'Min Delay: {metrics["min_delay"]:.2f}ms\n'
                 f'1st percentile: {metrics["percentile_1"]:.2f}ms\n'
                 f'5th percentile: {metrics["percentile_5"]:.2f}ms\n'
                 f'10th percentile: {metrics["percentile_10"]:.2f}ms\n'
                 f'99th percentile: {metrics["percentile_99"]:.2f}ms\n')
    plt.text(0.02, 0.98, stats_text,
             transform=plt.gca().transAxes,
             verticalalignment='top',
             bbox=dict(boxstyle='round', facecolor='white', alpha=0.8))

    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'delay_main.pdf'), format='pdf', bbox_inches='tight', dpi=300)
    plt.close()

    # Figure 2: Box plot (without outliers)
    plt.figure(figsize=(8, 6))
    plt.boxplot(df['delay'], vert=True, whis=[1, 99.99], showfliers=False)
    plt.title('Delay Distribution (Box Plot)')
    plt.ylabel('Processing Delay (ms)')
    plt.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'delay_boxplot.pdf'), format='pdf', bbox_inches='tight', dpi=300)
    plt.close()

    # Figure 3: Moving average trend
    plt.figure(figsize=(10, 6))
    plt.plot(df['completion_percentage'], df['delay_ma'], color='red',
             linewidth=2, label='Moving Average Trend')
    plt.title(f'Moving Average Trend (Window Size: {window_size} events)')
    plt.xlabel('Run Completion (%)')
    plt.ylabel('Processing Delay (ms)')
    plt.grid(True, alpha=0.3)
    plt.legend()
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'delay_trend.pdf'), format='pdf', bbox_inches='tight', dpi=300)
    plt.close()

    return df, metrics_df


if __name__ == "__main__":
    csv_file_path = "../output/delay_result.csv"
    output_dir = "../output/"
    #csv_file_path = "../data/testResults/1.02.2025/NonSimulation_runs/bpic2012/delay_result_bpic2012.csv"
    #output_dir = "../data/testResults/1.02.2025/NonSimulation_runs/bpic2012/"
    #csv_file_path = "../data/testResults/1.02.2025/NonSimulation_runs/trafficFines/delay_result_motivating_GCdefault.csv"
    #output_dir = "../data/testResults/1.02.2025/NonSimulation_runs/trafficFines/"
    #csv_file_path = "../data/testResults/1.02.2025/NonSimulation_runs/sepsis/delay_result_sepsis.csv"
    #output_dir = "../data/testResults/1.02.2025/NonSimulation_runs/sepsis/"
    #csv_file_path = "../data/testResults/1.02.2025/NonSimulation_runs/motivating/delay_result_motivating.csv"
    #output_dir = "../data/testResults/1.02.2025/NonSimulation_runs/motivating/"
    #csv_file_path = "../data/output/delay_result.csv"

    #csv_file_path = "../testResults/19.02.2025/teeMode/bpic2012/delay_result_bpic2012.csv"
    #output_dir = "../testResults/19.02.2025/teeMode/bpic2012/"
    #csv_file_path = "../data/testResults/1.02.2025/NonSimulation_runs/trafficFines/delay_result_motivating_GCdefault.csv"
    #output_dir = "../data/testResults/1.02.2025/NonSimulation_runs/trafficFines/"
    #csv_file_path = "../data/testResults/1.02.2025/NonSimulation_runs/sepsis/delay_result_sepsis.csv"
    #output_dir = "../data/testResults/1.02.2025/NonSimulation_runs/sepsis/"
    #csv_file_path = "../19.02.2025/teeMode/motivating/delay_result_motivating.csv"
    #output_dir = "../data/testResults/1.02.2025/NonSimulation_runs/motivating/"
    df, metrics_df = analyze_event_delays(csv_file_path, output_dir)
    print(f"Analysis completed. Output files saved in: {output_dir}")


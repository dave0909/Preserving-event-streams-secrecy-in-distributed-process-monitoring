import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
import os

def analyze_event_delays(csv_path, output_pdf_path):
    # Read CSV file without headers
    df = pd.read_csv(csv_path, header=None, names=['event_number', 'arrival_ts', 'completion_ts'])

    # Sort the dataframe by arrival timestamp
    df = df.sort_values('arrival_ts')

    # Convert nanoseconds to milliseconds for better readability
    df['delay'] = (df['completion_ts'] - df['arrival_ts']) / 1_000_000  # Convert to milliseconds

    # Calculate percentage based on sorted order
    total_events = len(df)
    df['completion_percentage'] = (df.index + 1) * 100 / total_events

    # Calculate all metrics
    metrics = {
        'average_delay': df['delay'].mean(),
        'median_delay': np.median(df['delay']),
        'max_delay': df['delay'].max(),
        'min_delay': df['delay'].min(),
        'percentile_1': np.percentile(df['delay'], 1),
        'percentile_5': np.percentile(df['delay'], 5),
        'percentile_10': np.percentile(df['delay'], 10),
        'total_events': total_events
    }

    # Save metrics to CSV
    metrics_df = pd.DataFrame([metrics])
    metrics_output_path = os.path.splitext(output_pdf_path)[0] + '_metrics.csv'
    metrics_df.to_csv(metrics_output_path, index=False)

    # Create the plot
    plt.figure(figsize=(12, 6))

    # Plot main delay line
    plt.plot(df['completion_percentage'], df['delay'], marker='', linestyle='-', color='blue',
             linewidth=1.5, label='Delay')

    # Add percentile and median lines
    plt.axhline(y=metrics['percentile_1'], color='r', linestyle='--', alpha=0.5,
                label=f'1st percentile ({metrics["percentile_1"]:.2f}ms)')
    plt.axhline(y=metrics['percentile_5'], color='g', linestyle='--', alpha=0.5,
                label=f'5th percentile ({metrics["percentile_5"]:.2f}ms)')
    plt.axhline(y=metrics['percentile_10'], color='orange', linestyle='--', alpha=0.5,
                label=f'10th percentile ({metrics["percentile_10"]:.2f}ms)')
    plt.axhline(y=metrics['median_delay'], color='purple', linestyle='--', alpha=0.7,
                label=f'Median ({metrics["median_delay"]:.2f}ms)')

    # Customize the plot
    plt.title('Event Processing Delay vs Run Completion')
    plt.xlabel('Run Completion (%)')
    plt.ylabel('Processing Delay (ms)')
    plt.grid(True, alpha=0.3)

    # Add legend
    plt.legend(loc='center left', bbox_to_anchor=(1, 0.5))

    # Add statistics as text
    stats_text = (f'Average Delay: {metrics["average_delay"]:.2f}ms\n'
                 f'Median Delay: {metrics["median_delay"]:.2f}ms\n'
                 f'Max Delay: {metrics["max_delay"]:.2f}ms\n'
                 f'Min Delay: {metrics["min_delay"]:.2f}ms\n'
                 f'1st percentile: {metrics["percentile_1"]:.2f}ms\n'
                 f'5th percentile: {metrics["percentile_5"]:.2f}ms\n'
                 f'10th percentile: {metrics["percentile_10"]:.2f}ms')

    plt.text(0.02, 0.98, stats_text,
             transform=plt.gca().transAxes,
             verticalalignment='top',
             bbox=dict(boxstyle='round', facecolor='white', alpha=0.8))

    # Adjust layout to accommodate legend
    plt.tight_layout()

    # Save the plot as PDF
    plt.savefig(output_pdf_path, format='pdf', bbox_inches='tight', dpi=300)

    # Close the plot to free memory
    plt.close()

    return df, metrics_df

# Example usage
if __name__ == "__main__":
    csv_file_path = "../data/output/delay_result.csv"
    pdf_output_path = "../data/output/event_delay_analysis.pdf"
    df, metrics_df = analyze_event_delays(csv_file_path, pdf_output_path)
    print(f"Metrics saved to: {os.path.splitext(pdf_output_path)[0] + '_metrics.csv'}")
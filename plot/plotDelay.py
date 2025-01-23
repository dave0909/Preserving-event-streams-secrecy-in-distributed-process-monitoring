import pandas as pd
import matplotlib.pyplot as plt
import numpy as np

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

    # Create the plot
    plt.figure(figsize=(12, 6))
    plt.plot(df['completion_percentage'], df['delay'], marker='', linestyle='-', color='blue', linewidth=1.5)

    # Customize the plot
    plt.title('Event Processing Delay vs Run Completion')
    plt.xlabel('Run Completion (%)')
    plt.ylabel('Processing Delay (ms)')
    plt.grid(True, alpha=0.3)

    # Add some statistics as text
    avg_delay = df['delay'].mean()
    max_delay = df['delay'].max()
    min_delay = df['delay'].min()

    stats_text = f'Average Delay: {avg_delay:.2f}ms\nMax Delay: {max_delay:.2f}ms\nMin Delay: {min_delay:.2f}ms'
    plt.text(0.02, 0.98, stats_text,
             transform=plt.gca().transAxes,
             verticalalignment='top',
             bbox=dict(boxstyle='round', facecolor='white', alpha=0.8))
    # Adjust layout
    plt.tight_layout()
    # Save the plot as PDF
    plt.savefig(output_pdf_path, format='pdf', bbox_inches='tight', dpi=300)
    # Close the plot to free memory
    plt.close()
    return df

# Example usage
if __name__ == "__main__":
    #csv_file_path = "../data/testResults/22.01.205/trafficFines/delay_result_trafficFines_GCdefault.csv"
    csv_file_path = "../data/output/delay_result.csv"
    pdf_output_path = "../data/output/event_delay_analysis.pdf"
    # Replace with your CSV file path
    df = analyze_event_delays(csv_file_path, pdf_output_path)
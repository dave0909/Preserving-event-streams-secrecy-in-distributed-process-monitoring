import numpy as np
import pandas as pd
import matplotlib.pyplot as plt
from datetime import datetime
from scipy.optimize import curve_fit

# Read the CSV file
df = pd.read_csv('../data/output/memory_usage.csv', decimal='.', header=0)
#df = pd.read_csv('../data/testResults/24.01.2025/trafficFines/memory_usage_trafficFines_GCdefault.csv', decimal='.', header=0)
# Convert microseconds to datetime
df['Timestamp'] = df['Timestamp'].apply(lambda x: datetime.utcfromtimestamp(x/1_000))

# Calculate time metrics
start_time = df['Timestamp'].min()
df['Durata (Secondi)'] = (df['Timestamp'] - start_time).dt.total_seconds()
total_runtime_seconds = df['Durata (Secondi)'].max() - df['Durata (Secondi)'].min()

# Normalize duration to percentage
df['Durata Normalizzata'] = (df['Durata (Secondi)'] - df['Durata (Secondi)'].min()) / total_runtime_seconds
df['Durata Normalizzata'] = df['Durata Normalizzata'] * 100

# Convert Bytes to MegaBytes
df['Memory usage (MB)'] = df['Memory Usage'] / 1048576

# Calculate memory statistics
memory_stats = {
    'Min Memory (MB)': df['Memory usage (MB)'].min(),
    'Max Memory (MB)': df['Memory usage (MB)'].max(),
    'Average Memory (MB)': df['Memory usage (MB)'].mean(),
    'Median Memory (MB)': df['Memory usage (MB)'].median(),
    'Standard Deviation (MB)': df['Memory usage (MB)'].std(),
    'Total Runtime (seconds)': total_runtime_seconds
}

# Convert stats to DataFrame and save to CSV
stats_df = pd.DataFrame([memory_stats])
stats_df.to_csv('../data/output/memory_stats.csv', index=False)

# Group by normalized duration and calculate mean memory usage
result = df.groupby('Durata Normalizzata')['Memory usage (MB)'].mean().reset_index()

# Plot configuration
plt.style.use("seaborn-v0_8-bright")
plt.figure(figsize=(16, 9))

# Plot memory usage trend
plt.plot(result['Durata Normalizzata'],
         result['Memory usage (MB)'],
         color='steelblue',
         linewidth=2,
         label="Memory usage trend",
         marker="")

# Configure plot aesthetics
plt.xlabel('Run completion percentage', fontsize=30, labelpad=15)
plt.ylabel('Memory usage (MB)', fontsize=30, labelpad=15)
plt.legend(loc='upper right', fontsize=18, edgecolor="black", fancybox=False)
plt.grid(True, linestyle='--')
plt.xticks(fontsize=30)
plt.yticks(fontsize=30)
plt.xlim([0, 101])
plt.ylim(0, 300)
plt.tight_layout()

# Save the plot
plt.savefig('../data/output/memory_usage.pdf')
# plt.show()
import pandas as pd
import matplotlib.pyplot as plt

# Read CSV files
df_simulation = pd.read_csv('../testResults/24.01.2025/trafficFines/memory_usage_trafficFines_GCdefault.csv', decimal='.', header=0)
df_sepsis = pd.read_csv('../testResults/24.01.2025/trafficFines/memory_usage_trafficFines_GC10ms.csv', decimal='.', header=0)
df_volvo = pd.read_csv('../testResults/24.01.2025/trafficFines/memory_usage_trafficFines_GC50ms.csv', decimal='.', header=0)

# Convert timestamps from milliseconds to seconds
df_simulation['Timestamp'] /= 1000
df_sepsis['Timestamp'] /= 1000
df_volvo['Timestamp'] /= 1000

# Add initial zero points
df_simulation.loc[-1] = [df_simulation['Timestamp'].min() - 1, 0]
df_simulation.index = df_simulation.index + 1
df_simulation.sort_index(inplace=True)

df_sepsis.loc[-1] = [df_sepsis['Timestamp'].min() - 1, 0]
df_sepsis.index = df_sepsis.index + 1
df_sepsis.sort_index(inplace=True)

#df_volvo = df_volvo[:len(df_simulation)]
df_volvo.loc[-1] = [df_volvo['Timestamp'].min() - 1, 0]
df_volvo.index = df_volvo.index + 1
df_volvo.sort_index(inplace=True)

# Calculate time progression as percentage for each dataset
def calculate_percentage(df):
    min_time = df['Timestamp'].min()
    max_time = df['Timestamp'].max()
    total_duration = max_time - min_time
    df['Completion Percentage'] = ((df['Timestamp'] - min_time) / total_duration) * 100
    df['Duration'] = total_duration
    return df

# Apply percentage calculation to each dataset
df_simulation = calculate_percentage(df_simulation)
df_sepsis = calculate_percentage(df_sepsis)
df_volvo = calculate_percentage(df_volvo)

# Convert Bytes to MegaBytes
df_simulation['Memory usage (MB)'] = df_simulation['Memory Usage'] / 1048576
df_sepsis['Memory usage (MB)'] = df_sepsis['Memory Usage'] / 1048576
df_volvo['Memory usage (MB)'] = df_volvo['Memory Usage'] / 1048576

# Group by percentage for smoother plotting
result = df_simulation.groupby('Completion Percentage')['Memory usage (MB)'].mean().reset_index()
result_sepsis = df_sepsis.groupby('Completion Percentage')['Memory usage (MB)'].mean().reset_index()
result_volvo = df_volvo.groupby('Completion Percentage')['Memory usage (MB)'].mean().reset_index()

# Convert durations to 9m:10s format
def format_duration(seconds):
    minutes = seconds // 60
    remaining_seconds = seconds % 60
    return f"{int(minutes)}m:{int(remaining_seconds):02d}s"

# Calculate durations in this format for each dataset
duration_simulation = format_duration(df_simulation['Duration'].iloc[0])
duration_sepsis = format_duration(df_sepsis['Duration'].iloc[0])
duration_volvo = format_duration(df_volvo['Duration'].iloc[0])

# Get the last y-values for each plot
last_simulation_y = result['Memory usage (MB)'].iloc[-1]
last_volvo_y = result_volvo['Memory usage (MB)'].iloc[-1]
last_sepsis_y = result_sepsis['Memory usage (MB)'].iloc[-1]

# Plotting
plt.style.use("seaborn-v0_8-bright")
plt.figure(figsize=(16, 9))

# Add lines
plt.plot(result['Completion Percentage'], result['Memory usage (MB)'],
         label='Default garbage collection', color='#FF204E',
         linewidth=1, marker='', markersize=2, alpha=1)
plt.plot(result_volvo['Completion Percentage'], result_volvo['Memory usage (MB)'],
         label='Custom garbage collection 1', color='#1abc9c',
         linewidth=1, marker='', markersize=2, alpha=0.8)
plt.plot(result_sepsis['Completion Percentage'], result_sepsis['Memory usage (MB)'],
         label='Custom garbage collection 2', color='blue',
         linewidth=1, marker='', markersize=2, alpha=0.7)

plt.axhline(y=175, color='black', linestyle='dashed',
            label="Total log size", linewidth=4)

# Add duration text outside the chart (right of the y-axis)
text_offset = 5  # Offset to position the text outside the chart
plt.text(100.3, last_simulation_y, f"{duration_simulation}",
         color='#FF204E', fontsize=24, va='center', ha='left')
plt.text(100.3, last_volvo_y+13, f"{duration_volvo}",
         color='#1abc9c', fontsize=24, va='center', ha='left')
plt.text(100.3, last_sepsis_y-5, f"{duration_sepsis}",
         color='blue', fontsize=24, va='center', ha='left')

# Customize the plot
plt.xticks(fontsize=34)
plt.yticks(fontsize=34)
plt.xlabel('Run completion percentage', fontsize=34, labelpad=15)
plt.ylabel('Memory usage [MB]', fontsize=34, labelpad=15)
plt.grid(True, linestyle='--')

# Set axis limits
plt.xlim([0, 100])  # Extend x-axis to make space for text
plt.ylim([0, 300])

plt.legend(loc='upper left', fontsize=27)
plt.tight_layout()

plt.savefig('../charts/memoryusageMultiplot_GC.pdf')
exit()

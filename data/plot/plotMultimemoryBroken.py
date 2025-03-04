import pandas as pd
import matplotlib.pyplot as plt
from datetime import datetime

# Load datasets
df_simulation = pd.read_csv('../data/testResults/1.02.2025/NonSimulation_runs/motivating/memory_usage_motivating.csv', decimal='.', header=0)
df_sepsis = pd.read_csv('../data/testResults/1.02.2025/NonSimulation_runs/sepsis/memory_usage_sepsis.csv', decimal='.', header=0)
df_volvo = pd.read_csv('../data/testResults/1.02.2025/NonSimulation_runs/trafficFines/memory_usage_motivating_GCdefault.csv', decimal='.', header=0)
df_bpic2012 = pd.read_csv('../data/testResults/1.02.2025/NonSimulation_runs/bpic2012/memory_usage_bpic2012.csv', decimal='.', header=0)

# Add initial zero row
def add_initial_zero(df):
    df.loc[-1] = [min(df['Timestamp']) - 1, 0]
    df.index = df.index + 1
    df.sort_index(inplace=True)
    return df

df_simulation = add_initial_zero(df_simulation)
df_sepsis = add_initial_zero(df_sepsis)
df_volvo = add_initial_zero(df_volvo)
df_bpic2012 = add_initial_zero(df_bpic2012)

# Convert timestamp to datetime
def convert_timestamp(df):
    df['Timestamp'] = df['Timestamp'].apply(lambda x: datetime.utcfromtimestamp(x/1000))
    return df

df_simulation = convert_timestamp(df_simulation)
df_sepsis = convert_timestamp(df_sepsis)
df_volvo = convert_timestamp(df_volvo)
df_bpic2012 = convert_timestamp(df_bpic2012)

# Normalize durations
def normalize_duration(df, start_time):
    df['Durata (Seconds)'] = (df['Timestamp'] - start_time).dt.total_seconds()
    total_runtime = df['Durata (Seconds)'].max() - df['Durata (Seconds)'].min()
    df['Durata Normalizzata'] = (df['Durata (Seconds)'] - df['Durata (Seconds)'].min()) / total_runtime * 100
    return df

df_simulation = normalize_duration(df_simulation, df_simulation['Timestamp'].min())
df_sepsis = normalize_duration(df_sepsis, df_sepsis['Timestamp'].min())
df_volvo = normalize_duration(df_volvo, df_volvo['Timestamp'].min())
df_bpic2012 = normalize_duration(df_bpic2012, df_bpic2012['Timestamp'].min())

# Convert memory usage
def convert_memory_usage(df):
    df['Memory usage (MB)'] = df['Memory Usage'] / 1048576
    return df

df_simulation = convert_memory_usage(df_simulation)
df_sepsis = convert_memory_usage(df_sepsis)
df_volvo = convert_memory_usage(df_volvo)
df_bpic2012 = convert_memory_usage(df_bpic2012)

# Aggregate memory usage
result = df_simulation.groupby('Durata Normalizzata')['Memory usage (MB)'].mean().reset_index()
result_sepsis = df_sepsis.groupby('Durata Normalizzata')['Memory usage (MB)'].mean().reset_index()
result_volvo = df_volvo.groupby('Durata Normalizzata')['Memory usage (MB)'].mean().reset_index()
result_bpic2012 = df_bpic2012.groupby('Durata Normalizzata')['Memory usage (MB)'].mean().reset_index()

# Create figure with broken y-axis
fig, (ax1, ax2) = plt.subplots(2, 1, sharex=True, figsize=(16, 9), gridspec_kw={'height_ratios': [3, 1]})

# Upper plot (y > 30)
ax1.plot(result_bpic2012['Durata Normalizzata'], result_bpic2012['Memory usage (MB)'], label='BPIC2012', color='green', linewidth=4)
ax1.plot(result_volvo['Durata Normalizzata'], result_volvo['Memory usage (MB)'], label='Road Traffic Fines', color='#FF204E', linewidth=4)
ax1.plot(result['Durata Normalizzata'], result['Memory usage (MB)'], label='Motivating Scenario', color='steelblue', linewidth=4)
ax1.plot(result_sepsis['Durata Normalizzata'], result_sepsis['Memory usage (MB)'], label='Sepsis', color='#ff7f00', linewidth=4)
ax1.set_ylim(10, 300)
ax1.spines.bottom.set_visible(False)
ax1.tick_params(bottom=False)
ax1.grid(True, linestyle='--')

# Lower plot (y < 20)
ax2.plot(result_bpic2012['Durata Normalizzata'], result_bpic2012['Memory usage (MB)'], color='green', linewidth=4)
ax2.plot(result_volvo['Durata Normalizzata'], result_volvo['Memory usage (MB)'], color='#FF204E', linewidth=4)
ax2.plot(result['Durata Normalizzata'], result['Memory usage (MB)'], color='steelblue', linewidth=4)
ax2.plot(result_sepsis['Durata Normalizzata'], result_sepsis['Memory usage (MB)'], color='#ff7f00', linewidth=4)
ax2.set_ylim(0, 6)
ax2.spines.top.set_visible(False)
ax2.grid(True, linestyle='--')

# Broken axis diagonal lines
d = .015
kwargs = dict(transform=ax1.transAxes, color='k', clip_on=False)
ax1.plot((-d, +d), (-d, +d), **kwargs)
ax1.plot((1 - d, 1 + d), (-d, +d), **kwargs)
kwargs.update(transform=ax2.transAxes)
ax2.plot((-d, +d), (1 - d, 1 + d), **kwargs)
ax2.plot((1 - d, 1 + d), (1 - d, 1 + d), **kwargs)

# Labels and legend
ax2.set_xlabel('Run completion percentage', fontsize=30, labelpad=15)
fig.text(0.02, 0.5, 'Memory usage [MB]', va='center', rotation='vertical', fontsize=30)
ax1.legend(loc='upper left', fontsize=25)
plt.xticks(fontsize=30)
plt.yticks(fontsize=30)
plt.xlim([-3, 100])
plt.tight_layout()

# Save and show
plt.savefig('../data/testResults/27.01.2025/memoryusageMultiplot_now.pdf')
plt.show()

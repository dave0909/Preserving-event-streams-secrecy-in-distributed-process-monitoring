import numpy as np  # Assicurati che numpy sia importato
import pandas as pd
import matplotlib.pyplot as plt
from datetime import datetime
from scipy.optimize import curve_fit

df = pd.read_csv('../data/testResults/v1/memory_usage_trafficFines_v1.csv', decimal='.', header=0)

# SECONDS
df['Timestamp'] = df['Timestamp'].apply(lambda x: datetime.utcfromtimestamp(x/1000))
start_time = df['Timestamp'].min()
df['Durata (Secondi)'] = (df['Timestamp'] - start_time).dt.total_seconds()
total_runtime_seconds = df['Durata (Secondi)'].max() - df['Durata (Secondi)'].min()

# Normalize 'Durata (Secondi)'
df['Durata Normalizzata'] = (df['Durata (Secondi)'] - df['Durata (Secondi)'].min()) / total_runtime_seconds
df['Durata Normalizzata'] = df['Durata Normalizzata'] * 100

# Convert Bytes in MegaBytes
df['Memory usage (MB)'] = df['Memory Usage'] / 1048576

# Unify the dataset
result = df.groupby('Durata Normalizzata')['Memory usage (MB)'].mean().reset_index()


# PLOT
plt.style.use("seaborn-v0_8-bright")
plt.figure(figsize=(16, 9))

# Original trend
plt.plot(result['Durata Normalizzata'], result['Memory usage (MB)'], color='steelblue', linewidth=3, label="Memory usage trend")

# Logarithmic curve
#plt.plot(x_data, log_fit_y, color='darkorange', linestyle='--', linewidth=2, label="Logarithmic interpolation")

plt.axhline(y=181, color='red', linestyle='--', linewidth=3, label="Total log size")

plt.xlabel('Run completion percentage', fontsize=30, labelpad=15)
plt.ylabel('Memory usage (MB)', fontsize=30, labelpad=15)
plt.legend(loc='upper right', fontsize=18, edgecolor="black", fancybox=False)
plt.grid(True, linestyle='--')
plt.xticks(fontsize=30)
plt.yticks(fontsize=30)

plt.xlim([0, 101])
plt.ylim(0, 300)
plt.tight_layout()
plt.savefig('../data/testResults/memory_usage_trafficFines_v1.pdf')
# plt.show()
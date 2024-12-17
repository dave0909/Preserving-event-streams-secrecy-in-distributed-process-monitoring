import pandas as pd
#import numpy as np
import matplotlib.pyplot as plt
from datetime import datetime



# For multilog test
#df_simulation = pd.read_csv('../data/testResults/v1/memory_usage_motivating_v1.csv', decimal='.', header=0)
#df_sepsis = pd.read_csv('../data/testResults/v1/memory_usage_sepsis_v1.csv', decimal='.', header=0,)
#df_volvo = pd.read_csv('../data/testResults/success1/memory_usage_trafficFines.csv', decimal='.', header=0,)

#For trafficFinesTests
df_simulation = pd.read_csv('../data/testResults/v1/memory_usage_trafficFines_v1.csv', decimal='.', header=0)
df_sepsis = pd.read_csv('../data/testResults/v1/memory_usage_trafficFines_v1.1.csv', decimal='.', header=0,)
df_volvo = pd.read_csv('../data/testResults/success1/memory_usage_trafficFines.csv', decimal='.', header=0,)



# add a row in the dataframe (timestamp: min(df['Timestamp']) - 1, Memory Usage: 0)
df_simulation.loc[-1] = [min(df_simulation['Timestamp']) - 1, 0]  # adding a row
df_simulation.index = df_simulation.index + 1  # shifting index
df_simulation.sort_index(inplace=True)
#df_simulation = df_simulation.iloc[::1000]

#df_sepsis = pd.read_csv('../data/testResults/v1/memory_usage_sepsis_v1.csv', decimal='.', header=0,)
df_sepsis = df_sepsis._append({'Timestamp': min(df_sepsis['Timestamp']) - 1, 'Memory Usage': 0}, ignore_index=True)
df_sepsis.loc[-1] = [min(df_sepsis['Timestamp']) - 1, 0]  # adding a row
df_sepsis.index = df_sepsis.index + 1  # shifting index
df_sepsis.sort_index(inplace=True)
#df_sepsis = df_sepsis.iloc[::20]

df_volvo=df_volvo[:len(df_simulation)]
df_volvo = df_volvo._append({'Timestamp': min(df_volvo['Timestamp']) - 1, 'Memory Usage': 0}, ignore_index=True)
df_volvo.loc[-1] = [min(df_volvo['Timestamp']) - 1, 0]  # adding a row
df_volvo.index = df_volvo.index + 1  # shifting index
df_volvo.sort_index(inplace=True)


# Convert in datetime
df_simulation['Timestamp'] = df_simulation['Timestamp'].apply(lambda x: datetime.utcfromtimestamp(x/1000))
df_sepsis['Timestamp'] = df_sepsis['Timestamp'].apply(lambda x: datetime.utcfromtimestamp(x/1000))
df_volvo['Timestamp'] = df_volvo['Timestamp'].apply(lambda x: datetime.utcfromtimestamp(x/1000))

#df_volvo = df_volvo.iloc[::1000]
#df_simulation = df_simulation.iloc[::1]

# Calculate first boot timestamp
start_time = df_simulation['Timestamp'].min()
start_time_sepsis = df_sepsis['Timestamp'].min()
start_time_volvo = df_volvo['Timestamp'].min()

# Transform timestamps into seconds
df_simulation['Durata (Seconds)'] = (df_simulation['Timestamp'] - start_time).dt.total_seconds()
df_sepsis['Durata (Seconds)'] = (df_sepsis['Timestamp'] - start_time_sepsis).dt.total_seconds()
df_volvo['Durata (Seconds)'] = (df_volvo['Timestamp'] - start_time_volvo).dt.total_seconds()


# Calculate total runtime
total_runtime_seconds_simulation = df_simulation['Durata (Seconds)'].max() - df_simulation['Durata (Seconds)'].min()
total_runtime_seconds_sepsis = df_sepsis['Durata (Seconds)'].max() - df_sepsis['Durata (Seconds)'].min()
total_runtime_seconds_volvo = df_volvo['Durata (Seconds)'].max() - df_volvo['Durata (Seconds)'].min()

# Normalize 'Durata (Secondi)' Simulation
df_simulation['Durata Normalizzata'] = (df_simulation['Durata (Seconds)'] - df_simulation['Durata (Seconds)'].min()) / total_runtime_seconds_simulation
df_simulation['Durata Normalizzata'] = df_simulation['Durata Normalizzata'] * 100
# Normalize 'Durata (Secondi)' Sepsis
df_sepsis['Durata Normalizzata'] = (df_sepsis['Durata (Seconds)'] - df_sepsis['Durata (Seconds)'].min()) / total_runtime_seconds_sepsis
df_sepsis['Durata Normalizzata'] = df_sepsis['Durata Normalizzata'] * 100
# Normalize 'Durata (Secondi)' Volvo
df_volvo['Durata Normalizzata'] = (df_volvo['Durata (Seconds)'] - df_volvo['Durata (Seconds)'].min()) / total_runtime_seconds_volvo
df_volvo['Durata Normalizzata'] = df_volvo['Durata Normalizzata'] * 100



# Convert Bytes in MegaBytes
df_simulation['Memory usage (MB)'] = df_simulation['Memory Usage'] / 1048576
df_sepsis['Memory usage (MB)'] = df_sepsis['Memory Usage'] / 1048576
df_volvo['Memory usage (MB)'] = df_volvo['Memory Usage'] / 1048576

# Unify the dataset
result = df_simulation.groupby('Durata Normalizzata')['Memory usage (MB)'].mean().reset_index()
result_sepsis = df_sepsis.groupby('Durata Normalizzata')['Memory usage (MB)'].mean().reset_index()
result_volvo = df_volvo.groupby('Durata Normalizzata')['Memory usage (MB)'].mean().reset_index()

pd.options.display.float_format = '{:.2f}'.format
# print(result)

# PLOT
plt.style.use("seaborn-v0_8-bright")
plt.figure(figsize=(16,9))

plt.plot(result['Durata Normalizzata'], result['Memory usage (MB)'], label='Motivating scenario', color='blue', linewidth=4, marker = '', markersize=2 , alpha=0.7)

# Create a line plot for the dataset volvo
plt.plot(result_volvo['Durata Normalizzata'], result_volvo['Memory usage (MB)'], label='Traffic Road Fines (sampled)', color='red', linewidth=2, marker = '', markersize=2, alpha=0.7)

# Create a line plot for the dataset sepsis
plt.plot(result_sepsis['Durata Normalizzata'], result_sepsis['Memory usage (MB)'], label='Sepsis', color='green', linewidth=2, marker='', markersize=2, alpha =0.7)

plt.xticks(fontsize=30)
plt.yticks(fontsize=30)
plt.xlabel('Run completion percentage', fontsize = 30, labelpad= 15)
plt.ylabel('Memory usage (MB)', fontsize = 30,  labelpad= 15)
plt.grid(True, linestyle='--')
plt.tight_layout()

plt.xlim([0, 100])
plt.ylim([0,300])


plt.legend (loc='upper left', fontsize=25)

#plt.fill_between(result['Durata Normalizzata'],result['Memory usage (MB)'], color = 'azure')
plt.tight_layout()
#plt.savefig('/Users/luca/Documents/PythonProjects/TEE_Evaluation/test_memoryusage/memoryusage3.pdf')
plt.savefig('../data/testResults/memoryusageMultiplot_garbage.pdf')
#plt.show()
exit()

#SEPSIS STATE UPDATE AVERAGE TIMING
#Time from start of the run:771.543776, Current average duration (ms): 0.002753, Min duration (ms): 0.001000, Max duration (ms): 0.017000

#MOTIVATING STATE UPDATE AVERAGE TIMING
#Time from start of the run:1816.785193, Current average duration (ms): 0.003381, Min duration (ms): 0.001000, Max duration (ms): 0.020000

import pandas as pd
#import numpy as np
import matplotlib.pyplot as plt
from datetime import datetime

df = pd.read_csv('../data/testResults/v1/memory_usage_trafficFines_v1.csv', decimal='.', header=0)
# SECONDS
df['Timestamp'] = df['Timestamp'].apply(lambda x: datetime.utcfromtimestamp(x/1000))
#Right know timestamps are taken every 50ms. consider points with timestamp every 200ms
#df = df.iloc[::500, :]

# Calculate first boot timestamp
start_time = df['Timestamp'].min()
# Transform timestamps into seconds
df['Durata (Secondi)'] = (df['Timestamp'] - start_time).dt.total_seconds()
# Calculate total runtime
total_runtime_seconds = df['Durata (Secondi)'].max() - df['Durata (Secondi)'].min()


# Normalize 'Durata (Secondi)'
df['Durata Normalizzata'] = (df['Durata (Secondi)'] - df['Durata (Secondi)'].min()) / total_runtime_seconds
df['Durata Normalizzata'] = df['Durata Normalizzata'] * 100



# Convert Bytes in MegaBytes
df['Memory usage (MB)'] = df['Memory Usage'] / 1048576

# Unify the dataset
result = df.groupby('Durata Normalizzata')['Memory usage (MB)'].mean().reset_index()
pd.options.display.float_format = '{:.2f}'.format
# print(result)


#diff_seconds_segment_received = (first_segment_recieved - inizialization)/1000
#diff_seconds_norm = (diff_seconds_segment_received - df['Durata (Secondi)'].min()) / total_runtime_seconds
#diff_seconds_norm = diff_seconds_norm * 100
#print(diff_seconds_norm)

#diff_seconds_comp = (first_computation - inizialization)/1000
#diff_seconds_norm_comp = ( diff_seconds_comp - df['Durata (Secondi)'].min()) / total_runtime_seconds
#diff_seconds_norm_comp = diff_seconds_norm_comp * 100
#print(diff_seconds_norm_comp)


#diff_seconds_att = (attestation - inizialization)/1000
#diff_seconds_norm_att = ( diff_seconds_att - df['Durata (Secondi)'].min()) / total_runtime_seconds
#diff_seconds_norm_att = diff_seconds_norm_att * 100
#print(diff_seconds_norm_att)


# PLOT
plt.style.use("seaborn-v0_8-bright")
plt.figure(figsize=(16,9))

#plt.plot(result['Durata Normalizzata'],result['Memory usage (MB)'], color = 'purple',linewidth=4, marker='.')
plt.plot(result['Durata Normalizzata'],result['Memory usage (MB)'], color = 'steelblue',linewidth=2, marker='')
# plt.plot(result_2['Durata Normalizzata'],result_2['Memory usage (MB)'].fillna(0), color = 'purple',linewidth=3, marker='.')


plt.xticks(fontsize=30)
plt.yticks(fontsize=30)

plt.xlabel('Run completion percentage', fontsize = 30, labelpad= 15)
plt.ylabel('Memory usage (MB)', fontsize = 30,  labelpad= 15)
plt.grid(True, linestyle='--')
plt.xlim([0, 101])
plt.ylim(0, 300)
plt.tight_layout()
#plt.legend(['Memory usage trend', 'First attestation', 'First segment recieved', 'First computation'], loc='upper left', fontsize=25, framealpha=1)
# plt.legend(['Memory usage trend', 'First segment recieved', 'First computation', 'First attestation'], loc='upper right', fontsize=18, framealpha=1, edgecolor="black", fancybox=False)

plt.fill_between(result['Durata Normalizzata'],result['Memory usage (MB)'], color = 'azure')
plt.tight_layout()
plt.savefig('../data/testResults/memoryusage_trafficFines_v1.pdf')
#plt.show()
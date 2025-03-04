import pandas as pd
import matplotlib.pyplot as plt
import numpy as np

# Load the data
file_path = "../data/testResults/1.02.2025/TEEvsNAT_memory.csv"
if file_path.endswith(".ods"):
    df = pd.read_excel(file_path, engine="odf", sheet_name="Sheet1")
elif file_path.endswith(".csv"):
    df = pd.read_csv(file_path)
else:
    raise ValueError("Unsupported file format")

# Compute the difference
df['Delta'] = df['TEE'] - df['Simulation']

# Define bar width and x locations
x = np.arange(len(df))  # the label locations
width = 0.2  # the width of the bars

# Create the figure and axes with broken y-axis
fig, (ax, ax2) = plt.subplots(2, 1, sharex=True, figsize=(10, 6), gridspec_kw={'height_ratios': [3, 1]})

# Define break points
ymax = df[['Simulation', 'TEE', 'Delta']].max().max()
break_start = 4.5
break_end = 10

# Plot bars
rects1 = ax.bar(x - width, df['Simulation'], width, label='Native mode', color='blue')
rects2 = ax.bar(x, df['TEE'], width, label='TEE mode', color='orange')
rects3 = ax.bar(x + width, df['Delta'], width, label='Delta', color='magenta')

ax2.bar(x - width, df['Simulation'], width, color='blue', label='Native mode')
ax2.bar(x, df['TEE'], width, color='orange', label='TEE mode')
ax2.bar(x + width, df['Delta'], width, color='magenta', label='Delta')

# Adjust y-limits for broken axis
ax.set_ylim(break_end, ymax+10)
ax2.set_ylim(0, break_start)
ax2.set_yticks([2, 4])

# Hide spines between axes
ax.spines['bottom'].set_visible(False)
ax2.spines['top'].set_visible(False)
ax.xaxis.set_visible(False)
ax2.xaxis.set_visible(True)
ax.tick_params(labeltop=False, labelsize=15)
ax2.tick_params(labelbottom=True, labelsize=15)

# Add diagonal lines
d = .015  # size of diagonal cut
kwargs = dict(transform=ax.transAxes, color='black', clip_on=False)
ax.plot((-d, +d), (-d, +d), **kwargs)
ax.plot((1 - d, 1 + d), (-d, +d), **kwargs)
kwargs.update(transform=ax2.transAxes)
ax2.plot((-d, +d), (1 - d, 1 + d), **kwargs)
ax2.plot((1 - d, 1 + d), (1 - d, 1 + d), **kwargs)

# Add numerical values above each Delta bar in both axes
for rect in rects3:
    height = rect.get_height()
    if height > break_end:
        ax.text(rect.get_x() + rect.get_width()/2 + 0.03, height, f'{height:.2f}', ha='center', va='bottom', fontsize=12, fontweight='bold')
    elif height < break_start:
        ax2.text(rect.get_x() + rect.get_width()/2, height, f'{height:.2f}', ha='center', va='bottom', fontsize=12, fontweight='bold')

# Labels, title, and custom x-axis labels
ax2.set_xlabel('Event log', fontsize=18, labelpad=18)
ax.set_ylabel('Memory usage [MB]', fontsize=18, labelpad=18)
ax2.set_xticks(x)
ax2.set_xticklabels(df['Log'], rotation=0, ha='center')

plt.xticks(fontsize=15)
plt.yticks(fontsize=15)
ax.legend(loc='upper left', fontsize=15, edgecolor="black", fancybox=False)
ax.grid(axis='y', alpha=0.4)
ax2.grid(axis='y', alpha=0.4)
plt.tight_layout()
plt.savefig("../data/testResults/1.02.2025/comparison_plot_total.pdf")

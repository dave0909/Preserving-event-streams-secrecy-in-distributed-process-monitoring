import matplotlib.pyplot as plt
import numpy as np

# Data from the table
logs = ['SC', 'Sepsis', 'BPIC12', 'RTF']
memory_usage = [4.02, 2.50, 21.31, 105.02]
lower = [3, 1, 15, 100]  # Mem. usage in MB
log_size = [7.34, 5.18, 70.60, 175.00]  # Log size as error bars

# Create two separate plots, one for Sepsis-SC and one for RTF and BPIC12
fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(9, 7))

# First plot for Sepsis-SC
bar_width = 0.1
bar_positions1 = [0, 0.15]  # Position bars closer together
bars1 = ax1.bar(bar_positions1, memory_usage[:2], color=['steelblue', '#ff7f00'], width=bar_width)

# Add dashed lines for log size (one for each bar), with the same color as the bars
for i, bar in enumerate(bars1):
    ax1.plot([bar.get_x(), bar.get_x() + bar.get_width()], [log_size[i], log_size[i]], '--', lw=4, color=bar.get_facecolor())  # Dashed line from memory usage to log size

# Add text above the dashed lines for log size
for i, pos in enumerate(bar_positions1):
    ax1.text(pos, log_size[i] + 0.1, f'{log_size[i]:.2f}', ha='center', fontsize=22, fontweight='bold')

# Add memory usage text in the center of the bars
for i, pos in enumerate(bar_positions1):
    ax1.text(pos, memory_usage[i] + 0.1, f'{memory_usage[i]:.2f}', ha='center', color='black', fontsize=22, fontweight='bold')

ax1.set_xticks(bar_positions1)
ax1.set_xticklabels(logs[:2])
ax1.set_ylim(0, 11)  # Scale limit for Sepsis-SC
ax1.set_ylabel('Average memory usage [MB]', fontsize=29)

# Second plot for BPIC12 and RTF
bar_positions2 = [0, 0.15]  # Position bars closer together
bars2 = ax2.bar(bar_positions2, memory_usage[2:], color=['green', '#FF204E'], width=bar_width)

# Add dashed lines for log size (one for each bar), with the same color as the bars
for i, bar in enumerate(bars2):
    ax2.plot([bar.get_x(), bar.get_x() + bar.get_width()], [log_size[2 + i], log_size[2 + i]], '--', lw=4, color=bar.get_facecolor())  # Dashed line from memory usage to log size

# Add text above the dashed lines for log size
for i, pos in enumerate(bar_positions2):
    ax2.text(pos, log_size[2 + i] + 2.5, f'{log_size[2 + i]:.2f}', ha='center', fontsize=21, fontweight='bold')

# Add memory usage text in the center of the bars
for i, pos in enumerate(bar_positions2):
    ax2.text(pos, memory_usage[2 + i] + 2.5, f'{memory_usage[2 + i]:.2f}', ha='center', color='black', fontsize=21, fontweight='bold')

ax2.set_xticks(bar_positions2)
ax2.set_xticklabels(logs[2:])
ax2.set_ylim(0, 237)  # Scale limit for Sepsis-SC

# Add legend for the dashed lines representing the log sizes
ax1.plot([], [], '--', color='steelblue', label='SC size')
ax1.plot([], [], '--', color='#ff7f00', label='Sepsis size')
ax2.plot([], [], '--', color='green', label='BPIC12 size')
ax2.plot([], [], '--', color='#FF204E', label='RTF size')

# Add the legend
ax1.legend(loc='upper left', fontsize=23)
ax2.legend(loc='upper left', fontsize=23)

ax1.xaxis.set_tick_params(labelsize=28)
ax2.xaxis.set_tick_params(labelsize=28)

ax1.yaxis.set_tick_params(labelsize=28)
ax2.yaxis.set_tick_params(labelsize=28)

# Reduce space between the charts
plt.tight_layout(pad=1)

# Save the chart as a PDF file
plt.savefig("../charts/plotMemoryLog.pdf", format="pdf")
import pandas as pd
import matplotlib.pyplot as plt
import numpy as np

# Load the data
file_path = "../testResults/19.02.2025/TEEvsSIM.csv"
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
width = 0.26  # slightly wider bars to reduce space between groups
gap = 0.00    # reduced gap between bars within a group

# Create the plot with square aspect ratio
fig, ax = plt.subplots(figsize=(11, 8))  # Square figure

# Position bars closer together
rects1 = ax.bar(x - width - gap/2, df['Simulation'], width, label='Native mode', color='blue')
rects2 = ax.bar(x, df['TEE'], width, label='TEE mode', color='orange')
rects3 = ax.bar(x + width + gap/2, df['Delta'], width, label='Overhead', color='magenta')

# Add numerical values above each Delta bar
index=0
for rect in rects3:
    height = rect.get_height()
    #Compute the percentage of the overhead with respect to the respective native mode
    percentage = int(height / df['Simulation'][index] * 100)
    ax.text(rect.get_x() + rect.get_width()/2 + 0.07, height, f'+{percentage:.0f}%', ha='center', va='bottom', fontsize=22, fontweight='bold')
    index+=1
# Labels, title, and custom x-axis labels
#ax.set_xlabel('Event log', fontsize=25, labelpad=18)
ax.set_ylabel('Average observation lag [ms]', fontsize=30, labelpad=18)
ax.set_xticks(x)
ax.set_xticklabels(df['Log'], rotation=0, ha='center')
plt.xticks(fontsize=30)
plt.yticks(fontsize=30)

# Legend
plt.legend(loc='upper left', fontsize=22, edgecolor="black", fancybox=False)

# Grid and layout
plt.grid(axis='y', alpha=0.4)
plt.tight_layout()

# Save the plot as a PDF
plt.savefig("../charts/barplotdelay2.pdf")
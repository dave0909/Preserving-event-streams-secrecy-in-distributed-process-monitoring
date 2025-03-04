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
width = 0.2  # the width of the bars

# Create the plot
fig, ax = plt.subplots(figsize=(10, 6))
rects1 = ax.bar(x - width, df['Simulation'], width, label='Native mode', color='blue')
rects2 = ax.bar(x, df['TEE'], width, label='TEE mode', color='orange')
rects3 = ax.bar(x + width, df['Delta'], width, label='Delta', color='magenta')

# Add numerical values above each Delta bar
for rect in rects3:
    height = rect.get_height()
    ax.text(rect.get_x() + rect.get_width()/2, height, f'{height:.2f}', ha='center', va='bottom', fontsize=12, fontweight='bold')

# Labels, title, and custom x-axis labels
ax.set_xlabel('Event log', fontsize=20, labelpad=18)
ax.set_ylabel('Average observation lag [ms]', fontsize=21, labelpad=18)
ax.set_xticks(x)
ax.set_xticklabels(df['Log'], rotation=0, ha='center')

plt.xticks(fontsize=19)
plt.yticks(fontsize=20)
ax.legend()
plt.legend(loc='upper left', fontsize=15, edgecolor="black", fancybox=False)
#plt.grid(True, linestyle='--')
# Save the plot as a PDF
plt.grid(axis='y', alpha=0.4)
plt.tight_layout()
plt.savefig("../charts/barplotdelay.pdf")

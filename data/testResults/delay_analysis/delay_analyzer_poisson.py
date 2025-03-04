import pm4py
from pm4py.objects.log.importer.xes import importer as xes_importer
import numpy as np
from collections import Counter
import csv
from scipy import stats

def analyze_global_delays(file_path, metrics_output_file):
    # Parse the XES log
    log = xes_importer.apply(file_path)

    # Extract all events with their timestamps
    all_events = []
    for trace in log:
        for event in trace:
            if "time:timestamp" in event:
                all_events.append(event["time:timestamp"].timestamp())

    # Sort all timestamps globally
    all_events.sort()

    # Compute delays between consecutive events in the sorted global list
    global_delays = np.diff(all_events)

    # Convert delays to milliseconds
    all_delays_ms = global_delays * 1e3
    non_zero_delays_ms = all_delays_ms[all_delays_ms > 0]

    # Poisson Process Analysis
    if len(non_zero_delays_ms) > 0:
        # KS Test for Exponential Fit
        ks_stat, ks_pval = stats.kstest(non_zero_delays_ms, 'expon', args=(np.mean(non_zero_delays_ms), np.std(non_zero_delays_ms)))

        # Variance-to-Mean Ratio (VMR) Test
        mean_count = np.mean(non_zero_delays_ms)
        var_count = np.var(non_zero_delays_ms)
        vmr = var_count / mean_count if mean_count > 0 else None
    else:
        ks_stat, ks_pval, vmr = None, None, None

    # Compute metrics
    zero_delays_count = np.sum(all_delays_ms == 0)
    total_delays = len(all_delays_ms)
    zero_delays_percentage = (zero_delays_count / total_delays * 100) if total_delays > 0 else 0

    metrics = {
        "total_events": len(all_events),
        "total_delays": total_delays,
        "zero_delays_count": zero_delays_count,
        "zero_delays_percentage": zero_delays_percentage,
        "non_zero_delays_count": len(non_zero_delays_ms),
        "mean": np.mean(all_delays_ms) if len(all_delays_ms) > 0 else 0.0,
        "std": np.std(all_delays_ms) if len(all_delays_ms) > 0 else 0.0,
        "ks_test_stat": ks_stat,
        "ks_test_pval": ks_pval,
        "vmr": vmr,
    }

    # Write metrics to CSV
    with open(metrics_output_file, mode='w', newline='') as csvfile:
        writer = csv.writer(csvfile)
        writer.writerow(["Metric", "Value"])
        for key, value in metrics.items():
            writer.writerow([key, value])

# Usage
if __name__ == "__main__":
    analyze_global_delays("../../xes/sepsis.xes", "delays_poisson_sepsis.csv")

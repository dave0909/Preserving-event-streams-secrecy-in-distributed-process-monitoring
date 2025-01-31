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

    # Calculate percentage of zero delays
    zero_delays_count = np.sum(all_delays_ms == 0)
    total_delays = len(all_delays_ms)
    zero_delays_percentage = (zero_delays_count / total_delays * 100) if total_delays > 0 else 0

    # Compute metrics including zero delays
    metrics = {
        "total_events": len(all_events),
        "total_delays": total_delays,
        "zero_delays_count": zero_delays_count,
        "zero_delays_percentage": zero_delays_percentage,
        "non_zero_delays_count": len(non_zero_delays_ms),
        "first_percentile": np.percentile(all_delays_ms, 1) if len(all_delays_ms) > 0 else 0.0,
        "fifth_percentile": np.percentile(all_delays_ms, 5) if len(all_delays_ms) > 0 else 0.0,
        "tenth_percentile": np.percentile(all_delays_ms, 10) if len(all_delays_ms) > 0 else 0.0,
        "first_quartile": np.percentile(all_delays_ms, 25) if len(all_delays_ms) > 0 else 0.0,
        "median": np.median(all_delays_ms) if len(all_delays_ms) > 0 else 0.0,
        "third_quartile": np.percentile(all_delays_ms, 75) if len(all_delays_ms) > 0 else 0.0,
        "mean": np.mean(all_delays_ms) if len(all_delays_ms) > 0 else 0.0,
        "mean_non_zero": np.mean(non_zero_delays_ms) if len(non_zero_delays_ms) > 0 else 0.0,
        "mode": Counter(all_delays_ms).most_common(1)[0][0] if len(all_delays_ms) > 0 else 0.0,
        "all_modes": [item[0] for item in Counter(all_delays_ms).most_common()],
        "std": np.std(all_delays_ms) if len(all_delays_ms) > 0 else 0.0,
        "std_non_zero": np.std(non_zero_delays_ms) if len(non_zero_delays_ms) > 0 else 0.0,
        "min": np.min(all_delays_ms) if len(all_delays_ms) > 0 else 0.0,
        "max": np.max(all_delays_ms) if len(all_delays_ms) > 0 else 0.0,
    }

    # Add additional statistics about the percentiles
    if len(all_delays_ms) > 0:
        # Count events below different percentiles
        events_below_first = np.sum(all_delays_ms < metrics["first_percentile"])
        events_below_fifth = np.sum(all_delays_ms < metrics["fifth_percentile"])
        events_below_tenth = np.sum(all_delays_ms < metrics["tenth_percentile"])

        # Add these counts to metrics
        metrics.update({
            "events_below_first_percentile": events_below_first,
            "events_below_first_percentile_percentage": (events_below_first / len(all_delays_ms)) * 100,
            "events_below_fifth_percentile": events_below_fifth,
            "events_below_fifth_percentile_percentage": (events_below_fifth / len(all_delays_ms)) * 100,
            "events_below_tenth_percentile": events_below_tenth,
            "events_below_tenth_percentile_percentage": (events_below_tenth / len(all_delays_ms)) * 100
        })

    # Write metrics to CSV
    with open(metrics_output_file, mode='w', newline='') as csvfile:
        writer = csv.writer(csvfile)
        writer.writerow(["Metric", "Value"])
        for key, value in metrics.items():
            if key == "all_modes":
                writer.writerow([key, ";".join(map(str, value))])
            else:
                writer.writerow([key, value])

# Usage
if __name__ == "__main__":
    analyze_global_delays("../../xes/bpic2012.xes", "delays_interinstance_bpic2012.csv")
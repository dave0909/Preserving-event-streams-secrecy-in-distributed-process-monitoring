import pm4py
from pm4py.objects.log.importer.xes import importer as xes_importer
import numpy as np
from collections import Counter
import csv
from scipy import stats

def analyze_event_log(file_path, metrics_output_file, batches_output_file):
    # Parse the XES log
    log = xes_importer.apply(file_path)

    # Data structure to hold delays for all traces
    delays = []
    batch_info = []  # To store batch information with trace IDs

    # Process each trace
    for trace_idx, trace in enumerate(log):
        trace_id = trace.attributes.get("concept:name", f"Trace_{trace_idx}")
        timestamps = []
        for event in trace:
            if "time:timestamp" in event:
                timestamps.append(event["time:timestamp"].timestamp())

        # Sort timestamps
        timestamps.sort()

        # Compute delays between consecutive events
        trace_delays = np.diff(timestamps)

        # Add non-zero delays to the main list
        delays.extend(delay for delay in trace_delays if delay > 0)

        # Compute lengths of batches with zero delays (simultaneous events)
        zero_delay_lengths = [1]
        for delay in trace_delays:
            if delay == 0:
                zero_delay_lengths[-1] += 1
            else:
                zero_delay_lengths.append(1)

        # Collect lengths greater than 1 (batches of simultaneous events)
        batch_start_idx = 0
        for length in zero_delay_lengths:
            if length > 1:
                batch_info.append({"trace_id": trace_id, "batch_size": length})
            batch_start_idx += length

    # Convert delays to milliseconds
    delays_ms = np.array(delays) * 1e3

    # Compute metrics including new percentiles
    metrics = {
        "first_percentile": np.percentile(delays_ms, 1) if len(delays_ms) > 0 else 0.0,
        "fifth_percentile": np.percentile(delays_ms, 5) if len(delays_ms) > 0 else 0.0,
        "tenth_percentile": np.percentile(delays_ms, 10) if len(delays_ms) > 0 else 0.0,
        "first_quartile": np.percentile(delays_ms, 25) if len(delays_ms) > 0 else 0.0,
        "median": np.median(delays_ms) if len(delays_ms) > 0 else 0.0,
        "third_quartile": np.percentile(delays_ms, 75) if len(delays_ms) > 0 else 0.0,
        "mean": np.mean(delays_ms) if len(delays_ms) > 0 else 0.0,
        "mode": Counter(delays_ms).most_common(1)[0][0] if len(delays_ms) > 0 else 0.0,
        "all_modes": [item[0] for item in Counter(delays_ms).most_common()],
        "std": np.std(delays_ms) if len(delays_ms) > 0 else 0.0,
        "min": np.min(delays_ms) if len(delays_ms) > 0 else 0.0,
        "max": np.max(delays_ms) if len(delays_ms) > 0 else 0.0,
    }

    # Add additional statistics about the percentiles
    if len(delays_ms) > 0:
        # Count events below different percentiles
        events_below_first = np.sum(delays_ms < metrics["first_percentile"])
        events_below_fifth = np.sum(delays_ms < metrics["fifth_percentile"])
        events_below_tenth = np.sum(delays_ms < metrics["tenth_percentile"])

        # Add these counts to metrics
        metrics.update({
            "events_below_first_percentile": events_below_first,
            "events_below_first_percentile_percentage": (events_below_first / len(delays_ms)) * 100,
            "events_below_fifth_percentile": events_below_fifth,
            "events_below_fifth_percentile_percentage": (events_below_fifth / len(delays_ms)) * 100,
            "events_below_tenth_percentile": events_below_tenth,
            "events_below_tenth_percentile_percentage": (events_below_tenth / len(delays_ms)) * 100
        })

    # Write metrics to CSV
    with open(metrics_output_file, mode='w', newline='') as csvfile:
        writer = csv.writer(csvfile)
        writer.writerow(["Metric", "Value"])  # Removed "(ms)" as not all metrics are in ms now
        for key, value in metrics.items():
            if key == "all_modes":
                writer.writerow([key, ";".join(map(str, value))])
            else:
                writer.writerow([key, value])

    # Write batch information to a separate CSV
    with open(batches_output_file, mode='w', newline='') as csvfile:
        writer = csv.writer(csvfile)
        writer.writerow(["Trace ID", "Batch Size"])
        for batch in batch_info:
            writer.writerow([batch["trace_id"], batch["batch_size"]])

# Usage
if __name__ == "__main__":
    analyze_event_log("../../xes/sepsis.xes", "delays_sepsis.csv", "batches_sepsis.csv")
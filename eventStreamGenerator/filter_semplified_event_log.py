import pm4py
import argparse
from pm4py.objects.log.importer.xes import importer as xes_importer
from pm4py.objects.log.exporter.xes import exporter as xes_exporter
from pm4py.algo.filtering.log.variants import variants_filter

# Set up argument parser
parser = argparse.ArgumentParser(description='Filter an event log using only the top k recurrent traces.')
parser.add_argument('log_path', type=str, help='Path to the XES event log file')
parser.add_argument('k', type=int, help='Number of top recurrent traces to keep')
parser.add_argument('output_path', type=str, help='Path to save the filtered XES event log file')
args = parser.parse_args()

# Load the event log
log = xes_importer.apply(args.log_path)

# Get the variants (unique traces) and their frequencies
variants = variants_filter.get_variants(log)
variant_counts = {variant: len(variants[variant]) for variant in variants}

# Sort the variants by frequency and get the top k
sorted_variants = sorted(variant_counts.items(), key=lambda item: item[1], reverse=True)
top_k_variants = [variant for variant, count in sorted_variants[:args.k]]

# Filter the log to include only the top k variants
filtered_log = variants_filter.apply(log, top_k_variants)

# Save the filtered event log
xes_exporter.apply(filtered_log, args.output_path)

print(f"Filtered log saved to {args.output_path}")
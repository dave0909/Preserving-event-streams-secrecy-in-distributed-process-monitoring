import pm4py
from pm4py.objects.log.importer.xes import importer as xes_importer
import argparse
from collections import defaultdict

def count_cases_within_n_events(log_path: str, n_events: int) -> int:
    """
    Count the number of unique cases that appear within the first n_events events of the log.

    Parameters:
    -----------
    log_path : str
        Path to the XES event log file
    n_events : int
        Number of events to consider from the start of the log

    Returns:
    --------
    int
        Number of unique cases within the first n_events events
    """
    # Import the XES log
    log = xes_importer.apply(log_path)

    event_count = 0
    cases_seen = set()

    # Iterate through cases and their events
    for case in log:
        for event in case:
            event_count += 1
            cases_seen.add(case.attributes['concept:name'])

            if event_count >= n_events:
                return len(cases_seen)

    # If we have fewer events than n_events, return total number of cases seen
    return len(cases_seen)

def main():
    # Set up argument parser
    parser = argparse.ArgumentParser(description='Count unique cases within first N events in XES log')
    parser.add_argument('log_path', type=str, help='Path to the XES event log file')
    parser.add_argument('n_events', type=int, help='Number of events to consider from start of log')

    # Parse arguments
    args = parser.parse_args()

    try:
        # Count cases within first n events
        result = count_cases_within_n_events(args.log_path, args.n_events)

        # Print results
        print(f"\nResults:")
        print(f"Number of unique cases within first {args.n_events} events: {result}")

    except Exception as e:
        print(f"Error processing log file: {str(e)}")

if __name__ == "__main__":
    main()
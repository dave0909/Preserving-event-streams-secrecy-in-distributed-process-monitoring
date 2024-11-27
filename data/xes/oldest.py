import pm4py
from pm4py.objects.log.importer.xes import importer as xes_importer
from datetime import datetime

# Load the XES log
log = xes_importer.apply("trafficFines.xes")

# Initialize variables to track the oldest event
oldest_event = None
oldest_timestamp = None

# Iterate through all events in the log
for trace in log:
    for event in trace:
        for key, value in event.items():
            if key == "time:timestamp":
                print(key, type(value))
                #the timestamp is already in datetime format
                if oldest_timestamp is None or value < oldest_timestamp:
                    oldest_event = event
                    oldest_timestamp = value
# Print the oldest event
print("Oldest event:")
print(oldest_event)
print("Timestamp:", oldest_timestamp)
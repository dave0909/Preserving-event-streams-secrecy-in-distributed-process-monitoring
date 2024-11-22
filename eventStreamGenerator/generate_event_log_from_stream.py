import queue
import socket
import threading
from pm4py.objects.log.importer.xes import importer as xes_importer
from pm4py.objects.log.exporter.xes import exporter as xes_exporter
from pm4py.objects.log.obj import EventLog, Trace, Event

# Dictionary to hold message queues for each port
port_queues = {
    8085: queue.Queue(),
    8086: queue.Queue(),
    8087: queue.Queue()
}
orgs_port = {'organization_A': 8085, 'organization_B': 8086, 'organization_C': 8087}

def handle_event_data(host, port, number_of_complete_cases):
    event_log = EventLog()
    case_count = 0
    traces = {}

    while case_count <= number_of_complete_cases:
        try:
            # Create a socket object
            s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            # Connect to the server
            s.connect((host, port))
            print(f"Connected to {host}:{port}")
            # Receive and print data continuously
            event_splitted = ""
            event_counter = 0
            while case_count <= number_of_complete_cases:
                data = s.recv(2048)
                if not data:
                    break
                xes_string = data.decode()
                event_string_list = xes_string.split('\n')
                if event_splitted != "":
                    event_string_list[0] = event_splitted + event_string_list[0]
                    event_splitted = ""
                for event in event_string_list:
                    if event != "":
                        try:
                            log = xes_importer.deserialize(event)
                            for trace in log:
                                trace_id = trace.attributes["concept:name"]
                                if trace_id not in traces:
                                    traces[trace_id] = Trace(attributes=trace.attributes)
                                for e in trace:
                                    traces[trace_id].append(Event(e))
                                    event_counter += 1
                                if len(traces[trace_id]) == 1:
                                    case_count += 1
                                if case_count > number_of_complete_cases:
                                    break
                        except Exception as e:
                            print(f"Error: {e}")
                            if type(e).__name__ == 'XMLSyntaxError':
                                event_splitted = event
                            continue
        except Exception as e:
            print(f"Error: {e}")

    # Add all traces to the event log, excluding the last event
    for trace_id, trace in traces.items():
        if len(trace) > 1:
            event_log.append(trace)

    # Write the event log to a new XES file
    xes_exporter.apply(event_log, "../data/xes/motivating.xes")
    print("Event log written to motivating.xes")
    print("Number of cases: ", len(event_log))
    print ("Number of events: ", event_counter)

if __name__ == "__main__":
    host = '127.0.0.1'
    port = 1234
    number_of_complete_cases = int(input("Enter the number of complete cases: "))
    handle_event_data(host, port, number_of_complete_cases)
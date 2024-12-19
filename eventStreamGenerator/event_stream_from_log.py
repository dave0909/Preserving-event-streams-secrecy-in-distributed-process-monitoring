import pm4py
import socket
import time
import argparse
from pm4py.objects.log.importer.xes import importer as xes_importer

# Set up argument parser
parser = argparse.ArgumentParser(description='Process an event log and send events over TCP.')
parser.add_argument('log_path', type=str, help='Path to the XES event log file')
args = parser.parse_args()

TCP_IP = 'localhost'
TCP_PORT = 1234
BUFFER_SIZE = 1024

# Load the event log from the provided path
log = xes_importer.apply(args.log_path)

# Function that returns a standard event at the end trace. The result value is a xes string
def create_final_event_xml_string(trace):
    trace_name = trace.attributes["concept:name"] if "concept:name" in trace.attributes else "case_unknown"
    event_xml = f"<org.deckfour.xes.model.impl.XTraceImpl><log openxes.version=\"1.0RC7\" xes.features=\"nested-attributes\" xes.version=\"1.0\" xmlns=\"http://www.xes-standard.org/\">"
    event_xml += f"<trace><string key=\"concept:name\" value=\"{trace_name}\"/><event>"
    event_xml += "<string key=\"concept:name\" value=\"__END__\"/><date key=\"time:timestamp\" value=\"2014-05-19T20:05:46.000+02:00\"/></event>"
    event_xml += "</trace></log></org.deckfour.xes.model.impl.XTraceImpl>\n"
    return event_xml

def create_event_xml_string(trace, event):
    trace_name = trace.attributes["concept:name"] if "concept:name" in trace.attributes else "case_unknown"
    event_xml = f"<org.deckfour.xes.model.impl.XTraceImpl><log openxes.version=\"1.0RC7\" xes.features=\"nested-attributes\" xes.version=\"1.0\" xmlns=\"http://www.xes-standard.org/\">"
    event_xml += f"<trace><string key=\"concept:name\" value=\"{trace_name}\"/><event>"
    for key, value in event.items():
        if isinstance(value, str):
            event_xml += f"<string key=\"{key}\" value=\"{value}\"/>"
        elif isinstance(value, bool):
            event_xml += f"<boolean key=\"{key}\" value=\"{value}\"/>"
        elif isinstance(value, int):
            event_xml += f"<int key=\"{key}\" value=\"{value}\"/>"
        elif isinstance(value, float):
            event_xml += f"<float key=\"{key}\" value=\"{value}\"/>"
        elif key == "time:timestamp":
            event_xml += f"<date key=\"{key}\" value=\"{value.isoformat()}\"/>"
        elif isinstance(value, pm4py.objects.log.obj.Event):
            event_xml += f"<date key=\"{key}\" value=\"{value.isoformat()}\"/>"
    event_xml += "</event></trace></log></org.deckfour.xes.model.impl.XTraceImpl>\n"
    return event_xml

# Set up TCP server to accept connections
with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server_socket:
    server_socket.bind((TCP_IP, TCP_PORT))
    server_socket.listen(1)
    print(f"Listening on {TCP_IP}:{TCP_PORT}...")
    while True:
        client_socket, client_address = server_socket.accept()
        print(f"Connection from {client_address} established.")
        try:
            for trace in log:
                for event in trace:
                    event_string = create_event_xml_string(trace, event)
                    client_socket.sendall(event_string.encode())
                    time.sleep(0.01)
                final_event_string = create_final_event_xml_string(trace)
                client_socket.sendall(final_event_string.encode())
                time.sleep(0.01)
        except Exception as e:
            print(f"An error occurred: {e}")
            client_socket.close()
        finally:
            client_socket.close()
            print(f"Connection from {client_address} closed.")
            break

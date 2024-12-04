# Description: This script reads an event log in XES format and sends the events as XML strings over UDP to a specified IP and port.

# import pm4py
# import socket
# import time
# from pm4py.objects.log.importer.xes import importer as xes_importer
#
# # Configure UDP connection details
# UDP_IP = 'localhost'  # Replace with the IP where clients will listen
# UDP_PORT = 9000       # Replace with your specific port
#
# # Load the XES event log
# log_path = '../data/xes/sepsis.xes'  # Replace with the path to your XES file
# log = xes_importer.apply(log_path)
#
# # Helper function to create the XML event string
# def create_event_xml_string(trace, event):
#     trace_name = trace.attributes["concept:name"] if "concept:name" in trace.attributes else "case_unknown"
#     event_xml = f"<org.deckfour.xes.model.impl.XTraceImpl><log openxes.version=\"1.0RC7\" xes.features=\"nested-attributes\" xes.version=\"1.0\" xmlns=\"http://www.xes-standard.org/\">"
#     event_xml += f"<trace><string key=\"concept:name\" value=\"{trace_name}\"/><event>"
#
#     # Add event attributes as XML elements
#     for key, value in event.items():
#         print(key, value)
#         if isinstance(value, str):
#             event_xml += f"<string key=\"{key}\" value=\"{value}\"/>"
#         elif isinstance(value, int):
#             event_xml += f"<int key=\"{key}\" value=\"{value}\"/>"
#         elif isinstance(value, float):
#             event_xml += f"<float key=\"{key}\" value=\"{value}\"/>"
#         elif key=="time:timestamp":
#             event_xml += f"<date key=\"{key}\" value=\"{value.isoformat()}\"/>"
#         elif isinstance(value, pm4py.objects.log.obj.Event):
#             event_xml += f"<date key=\"{key}\" value=\"{value.isoformat()}\"/>"
#
#     event_xml += "</event></trace></log></org.deckfour.xes.model.impl.XTraceImpl>"
#     return event_xml
#
# # Set up UDP socket
# with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as s:
#     for trace in log:
#         for event in trace:
#             # Create the XML string for the current event
#             event_string = create_event_xml_string(trace, event)
#
#             # Send the event string as a UDP packet
#             s.sendto(event_string.encode(), (UDP_IP, UDP_PORT))
#             print(f"Sent event: {event_string}")
#
#             # Wait between sends to simulate streaming
#             time.sleep(1)
# import pm4py
# from pm4py.objects.log.importer.xes import importer as xes_importer
# from pm4py.algo.discovery.inductive import algorithm as bpmn_discovery
# from pm4py.visualization.bpmn import visualizer as bpmn_visualizer
# import pm4py.objects.conversion.process_tree.variants.to_bpmn as to_bpmn
# import pm4py.objects.bpmn.exporter.exporter as exporter
# # Load the event log from an XES file
# log_path = ('../data/xes/sepsis.xes')  # Replace with the path to your XES file
# log = xes_importer.apply(log_path)
# bpmn=pm4py.discover_bpmn_inductive(log)
# output_path = "discovered_model.bpmn"  # Desired path and filename for the BPMN file
# bpmn_file = exporter.apply(bpmn, output_path)
# pm4py.view_bpmn(bpmn)







import pm4py
import socket
import time
from pm4py.objects.log.importer.xes import importer as xes_importer

TCP_IP = 'localhost'
TCP_PORT = 1234
BUFFER_SIZE = 1024

log_path = '../data/xes/motivating.xes'  # Replace with the path to your XES file
log = xes_importer.apply(log_path)
print(log[0])

#Funtion that returns a standard event at the end trace. The result value is a xes string
def create_final_event_xml_string(trace):
    trace_name = trace.attributes["concept:name"] if "concept:name" in trace.attributes else "case_unknown"
    event_xml = f"<org.deckfour.xes.model.impl.XTraceImpl><log openxes.version=\"1.0RC7\" xes.features=\"nested-attributes\" xes.version=\"1.0\" xmlns=\"http://www.xes-standard.org/\">"
    event_xml += f"<trace><string key=\"concept:name\" value=\"{trace_name}\"/><event>"
    event_xml+="<event><string key=\"concept:name\" value=\"EOT_EVENT\"/><date key=\"time:timestamp\" value=\"2014-05-19T20:05:46.000+02:00\"/></event>"
    event_xml += "</event></trace></log></org.deckfour.xes.model.impl.XTraceImpl>\n"
    return event_xml
def create_event_xml_string(trace, event):
    trace_name = trace.attributes["concept:name"] if "concept:name" in trace.attributes else "case_unknown"
    event_xml = f"<org.deckfour.xes.model.impl.XTraceImpl><log openxes.version=\"1.0RC7\" xes.features=\"nested-attributes\" xes.version=\"1.0\" xmlns=\"http://www.xes-standard.org/\">"
    event_xml += f"<trace><string key=\"concept:name\" value=\"{trace_name}\"/><event>"
    for key, value in event.items():
        if isinstance(value, str):
            event_xml += f"<string key=\"{key}\" value=\"{value}\"/>"
        #check if the value is a boolean
        elif isinstance(value, bool):
            event_xml += f"<boolean key=\"{key}\" value=\"{value}\"/>"
        elif isinstance(value, int):
            event_xml += f"<int key=\"{key}\" value=\"{value}\"/>"
        elif isinstance(value, float):
            event_xml += f"<float key=\"{key}\" value=\"{value}\"/>"
        elif key=="time:timestamp":
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
                print(trace.attributes["concept:name"])
                for event in trace:
                    event_string = create_event_xml_string(trace, event)
                    client_socket.sendall(event_string.encode())
                    print(f"Sent event: {event_string}")
                    time.sleep(0.01)
                #end_of_trace=create_final_event_xml_string(trace)
                #client_socket.sendall(event_string.encode())
        except Exception as e:
            print(f"An error occurred: {e}")
            client_socket.close()
        finally:
            client_socket.close()
            print(f"Connection from {client_address} closed.")

import queue
import socket
import threading
from pm4py.objects.log.importer.xes import importer as xes_importer
# Dictionary to hold message queues for each port
port_queues = {
    8085: queue.Queue(),
    8086: queue.Queue(),
    8087: queue.Queue()
}
orgs_port ={'organization_A':8085,'organization_B':8086,'organization_C':8087}
def handle_event_data(host, port):
    while True:
        try:
            # Create a socket object
            s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            # Connect to the server
            s.connect((host, port))
            print(f"Connected to {host}:{port}")
            # Receive and print data continuously
            event_splitted=""
            while True:
                data = s.recv(2048)
                if not data:
                    break
                xes_string=data.decode()
                event_string_list = xes_string.split('\n')
                if event_splitted != "":
                    event_string_list[0]=event_splitted+event_string_list[0]
                    event_splitted=""
                for event in event_string_list:
                    if event!="":
                        try:
                            log= xes_importer.deserialize(event)
                            if log[0][0]['organization'] in orgs_port.keys():
                                port_queues[orgs_port[log[0][0]['organization']]].put(event)
                            else:
                                port_queues[8085].put(event)
                        except Exception as e:
                            print(f"Error: {e}")
                            if type(e).__name__=='XMLSyntaxError':
                                event_splitted = event
                            continue
        except Exception as e:
            print(f"Error: {e}")


# Function to handle client connections and send messages from the queue
def handle_client(client_socket, message_queue):
    try:
        while True:
            message = message_queue.get()# Get a message from the queue
            print(message)
            client_socket.sendall((message + "\n").encode())
    except (BrokenPipeError, ConnectionResetError):
        print("Client disconnected")
    finally:
        client_socket.close()


# Function to set up a server on a given port
def server(port, message_queue):
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.bind(('0.0.0.0', port))
    server_socket.listen(5)
    print(f"Server listening on port {port}")
    while True:
        client_socket, addr = server_socket.accept()
        print(f"Accepted connection from {addr}")
        client_handler = threading.Thread(target=handle_client, args=(client_socket, message_queue))
        client_handler.start()

if __name__ == "__main__":
    # List of ports
    ports = [8085,8086,8087]
    # Start servers on specified ports
    for port in ports:
        server_thread = threading.Thread(target=server, args=(port, port_queues[port]))
        server_thread.start()
    host = '127.0.0.1'
    port = 1234
    handle_event_data(host, port)

"""
To run the event stream generator, run the following command in the terminal:
python3 main.py
"""

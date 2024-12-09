import socket
import sys
import threading
import time

student_directory = {}
sockets = []
running = True
server_mode = "primary"

def handle_message(message, client_socket):
    """Process client message for adding or looking up a student record."""
    time.sleep(2.7)
    components = message.split()
    perm_number = int(components[0])

    if len(components) > 1:
        phone_number = int(components[1])
        student_directory[perm_number] = phone_number
        print(f"Successfully inserted key {perm_number}")
        client_socket.send(f"Success{server_mode}".encode('utf-8'))
    else:
        if perm_number in student_directory:
            student_number = str(student_directory[perm_number])
            print(student_number)
            client_socket.send(f"{student_number}.{perm_number}.{server_mode}".encode('utf-8'))
        else:
            print("NOT FOUND")
            client_socket.send(f"NOT FOUND.{perm_number}.{server_mode}".encode('utf-8'))

def handle_client(client_socket):
    """Manage client connection, receiving and handling messages."""
    while running:
        try:
            message = client_socket.recv(1024).decode('utf-8')
            if not message:
                break
            threading.Thread(target=handle_message, args=(message, client_socket)).start()
        except ConnectionResetError:
            break
    
    client_socket.close()

def start_server(port, mode):
    """Initialize and start the server, setting up the socket and handling client connections."""
    global server_mode
    server_mode = "1" if mode == "primary" else "2"
    
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server_socket.bind(('localhost', port))
    server_socket.listen(5)
    sockets.append(server_socket)

    threading.Thread(target=input_handler).start()

    while running:
        try:
            client_socket, _ = server_socket.accept()
            threading.Thread(target=handle_client, args=(client_socket,)).start()
        except OSError:
            if not running:
                break

    sys.exit()

def handle_input():
    """Handle server commands from the command line."""
    return input().strip()

def input_handler():
    """Process server commands for dictionary display and shutdown."""
    global running
    while True:
        command = handle_input()

        if command == "dictionary":
            print_directory()
        elif command == "exit":
            shutdown_server()
            break

def print_directory():
    """Print the contents of the student directory."""
    items = [f"({perm_number}, {phone_number})" for perm_number, phone_number in student_directory.items()]
    print(f"{{{', '.join(items)}}}")

def shutdown_server():
    """Shut down the server gracefully."""
    global running
    running = False
    for sock in sockets:
        sock.close()
    sys.exit()

if __name__ == "__main__":
    server_mode = sys.argv[1]
    port = int(sys.argv[2])

    if server_mode == "primary":
        start_server(port, server_mode)
    elif server_mode == "secondary":
        start_server(port + 1, server_mode)

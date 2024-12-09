import socket
import threading
import sys
import time
import queue

class Client:
    def __init__(self, client_id, base_port):
        self.client_id = client_id
        self.my_port = base_port + 2 + client_id
        self.other_clients_ports = [base_port + 2 + i for i in range(1, 4) if port != self.my_port]
        self.client_ports = {i: base_port + 2 + i for i in range(1, 4)}
        self.server_ports = [base_port, base_port + 1]
        
        self.connected_clients = []
        self.connected_servers = []
        self.queue = queue.PriorityQueue()
        self.server_requests = queue.Queue()
        
        self.print_lock = threading.Lock()
        self.running = True
        self.init_client_state()

    def init_client_state(self):
        self.time_stamp = float(f"0.{self.client_id}")
        self.replies = 0
        self.critical_section = False
        self.success_replies = 0
        self.client_to_client_request = queue.PriorityQueue()

    def listen_for_clients(self, server_socket):
        while True:
            try:
                conn, addr = server_socket.accept()
                threading.Thread(target=self.handle_request, args=(conn,), daemon=True).start()
            except OSError:
                break

    def determine_logical_clock(self, received_time_stamp):
        max_clock = max(self.time_stamp, received_time_stamp) + 1.0
        return float(f"{int(max_clock)}.{self.client_id}")

    def queue_helper(self, time_stamp):
        logical_clk, process_id = time_stamp.strip("()").split(", ")
        return float(f"{logical_clk}.{process_id}") == self.queue.queue[0]

    def poll_success_replies(self, initial_request_time):
        while True:
            if self.success_replies > 1:
                self.notify_release(initial_request_time)
                self.success_replies = 0
                self.queue.get()
                self.critical_section = False
                break

    def notify_release(self, initial_request_time):
        c_x, c_y = self.get_client_ids()
        with self.print_lock:
            print(f"Sending release for request {initial_request_time} to Client {c_x}")
            print(f"Sending release for request {initial_request_time} to Client {c_y}")
        self.send_message("release")

    def critical_section_handler(self, initial_request_time):
        while True:
            if self.critical_section and self.queue_helper(initial_request_time):
                threading.Thread(target=self.poll_success_replies, args=(initial_request_time,), daemon=True).start()
                threading.Thread(target=self.send_message_to_servers, args=(initial_request_time,), daemon=True).start()
                break

    def handle_request(self, conn):
        while self.running:
            try:
                msg = conn.recv(1024).decode('utf-8')
                time.sleep(3)  # Simulate processing time
                self.process_message(msg)
            except ConnectionResetError:
                break
        conn.close()

    def process_message(self, msg):
        if msg.startswith("reply"):
            self.handle_reply(msg)
        elif msg == "release":
            self.handle_release()
        else:
            self.handle_request_message(msg)

    def handle_reply(self, msg):
        start = msg.find("(")
        end = msg.find(")") + 1
        time_stamp_of_init_request = msg[start:end]
        with self.print_lock:
            self.replies += 1
            print(f"Received reply for {time_stamp_of_init_request}, incrementing reply count to {self.replies}")
        
        if self.replies == 2:
            self.replies = 0
            self.critical_section = True
            threading.Thread(target=self.critical_section_handler, args=(time_stamp_of_init_request,), daemon=True).start()

    def handle_release(self):
        time_of_init_request = self.client_to_client_request.get()
        self.queue.get()
        with self.print_lock:
            print(f"Received release for request {time_of_init_request}")

    def handle_request_message(self, msg):
        try:
            # Clean up the message first
            msg = msg.split('reply')[0]  # Remove any reply text that got concatenated
            
            # Parse timestamp
            logical_time, process_id = self.parse_timestamp(msg)
            with self.print_lock:
                print(f"Received request ({logical_time}, {process_id})")
            self.client_to_client_request.put(f"({logical_time}, {process_id})")
            
            # Extract process ID more safely
            process_id = process_id.strip('() ')  # Remove parentheses and whitespace
            client_id_of_sender = int(process_id)
            received_time_stamp = float(f"{logical_time}.{process_id}")
            
            new_clock = self.determine_logical_clock(received_time_stamp)
            self.time_stamp = new_clock
            self.queue.put(received_time_stamp)
            
            response_msg = f"reply ({logical_time}, {process_id})"
            self.send_message(response_msg, client_id_of_sender)
        except (ValueError, IndexError) as e:
            print(f"Error processing message: {msg}, Error: {e}")

    def parse_timestamp(self, time_stamp):
        try:
            time_stamp = time_stamp.strip()  # Remove leading/trailing whitespace
            if '(' in time_stamp and ')' in time_stamp:
                # Handle formatted timestamp like "(1, 2)"
                content = time_stamp[time_stamp.find("(")+1:time_stamp.find(")")].split(', ')
                return float(content[0]), content[1]
            else:
                # Handle raw timestamp like "1.2"
                parts = time_stamp.split('.')
                return float(parts[0]), parts[1]
        except Exception as e:
            print(f"Error parsing timestamp {time_stamp}: {e}")
            raise

    def connect_to_other_clients(self):
        for port in self.other_clients_ports:
            self.connect_to_port(port, self.connected_clients)

    def connect_to_servers(self):
        for port in self.server_ports:
            self.connect_to_port(port, self.connected_servers)

    def connect_to_port(self, port, connection_list):
        while True:
            try:
                client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                client_socket.connect(('localhost', port))
                connection_list.append(client_socket)
                break
            except ConnectionRefusedError:
                time.sleep(1)

    def handle_server_messages(self, response, initial_request_time):
        time.sleep(2.8)
        if response.startswith("Success"):
            self.success_replies += 1
            with self.print_lock:
                print(f"Response from Server {response[-1]} for request {initial_request_time} : Success")
        else:
            self.handle_error_response(response)

    def handle_error_response(self, response):
        student_number, perm_number, server_id = response.split('.')
        with self.print_lock:
            print(f"Response from Server {server_id} for operation 'lookup {perm_number}' : {student_number}")

    def listen_for_server_messages(self, s, initial_request_time):
        while self.running:
            try:
                response = s.recv(1024).decode('utf-8')
                if not response:
                    break
                threading.Thread(target=self.handle_server_messages, args=(response, initial_request_time), daemon=True).start()
            except OSError:
                break

    def send_lookup_msg(self, server_request):
        for s in self.connected_servers:
            s.send(server_request.encode('utf-8'))
            threading.Thread(target=self.listen_for_server_messages, args=(s, server_request), daemon=True).start()

    def send_message_to_servers(self, initial_request_time):
        server_request = self.server_requests.get()
        components = server_request.split()
        for cnt, s in enumerate(self.connected_servers):
            server_type = "primary" if cnt < 1 else "secondary"
            with self.print_lock:
                print(f"Sending operation 'insert {components[0]} {components[1]}' for request {initial_request_time} to {server_type} server")
            s.send(server_request.encode('utf-8'))
            threading.Thread(target=self.listen_for_server_messages, args=(s, initial_request_time), daemon=True).start()

    def send_message(self, communication_type, target_client=None):
        if "(" in communication_type and ")" in communication_type:
            start = communication_type.find("(")
            end = communication_type.find(")") + 1
            time_stamp_reply = "reply" + communication_type[start:end]
        else:
            time_stamp_reply = communication_type

        for conn in self.connected_clients:
            if target_client is not None and conn.getpeername()[1] == self.client_ports[target_client]:
                conn.send(time_stamp_reply.encode('utf-8'))
            elif communication_type == "release":
                conn.send("release".encode('utf-8'))
            else:
                conn.send(f"{self.time_stamp}".encode('utf-8'))

    def handle_input(self):
        message = input().strip()
        components = message.split()
        request, perm_number, phone_number = components[0].lower(), None, None

        if request == "insert":
            perm_number, phone_number = components[1], components[2]
        elif request == "lookup":
            perm_number = components[1]

        return request, perm_number, phone_number

    def get_client_ids(self):
        return (2, 3) if self.client_id == 1 else (1, 3) if self.client_id == 2 else (1, 2)

    def input_handler(self):
        while True:
            request, perm_number, phone_number = self.handle_input()

            if request == "insert":
                self.handle_insert_request(perm_number, phone_number)
            elif request == "lookup":
                self.send_lookup_msg(perm_number)
            elif request == "exit":
                self.exit_client()

    def handle_insert_request(self, perm_number, phone_number):
        server_request = f"{perm_number} {phone_number}"
        self.server_requests.put(server_request)
        self.time_stamp += 1.0
        self.queue.put(self.time_stamp)

        logical_time, process_id = self.parse_timestamp(str(self.time_stamp))
        c_x, c_y = self.get_client_ids()
        with self.print_lock:
            print(f"Sending request ({logical_time}, {process_id}) to Client {c_x}")
            print(f"Sending request ({logical_time}, {process_id}) to Client {c_y}")

        threading.Thread(target=self.send_message, args=("",), daemon=True).start()

    def exit_client(self):
        self.running = False
        for sock in self.connected_servers:
            sock.close()
        sys.exit()

    def start_client(self):
        server_socket = self.create_server_socket()
        self.connect_to_servers()
        threading.Thread(target=self.listen_for_clients, args=(server_socket,), daemon=True).start()
        self.connect_to_other_clients()
        threading.Thread(target=self.input_handler).start()

    def create_server_socket(self):
        server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        server_socket.bind(('localhost', self.my_port))
        server_socket.listen(5)
        return server_socket

if __name__ == "__main__":
    client_id = int(sys.argv[1]) 
    port = int(sys.argv[2])
    client = Client(client_id, port)
    client.start_client()

import socket

# Define the server address and port
server_address = ('localhost', 8080)

# Create a socket object
sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

try:
    # Connect to the server
    sock.connect(server_address)
    print(f"Connected to {server_address}")

finally:
    # Close the socket
    sock.close()


import socket

sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
sock.connect("/tmp/wingo-ipc")

def recv(sock):
    data = ''
    while chr(0) not in data:
        data += sock.recv(4096)
    return data

def gribble(cmd):
    sock.send("%s%s" % (cmd, chr(0)))
    return recv(sock)

print gribble("GetClientName (GetActive)")

sock.close()


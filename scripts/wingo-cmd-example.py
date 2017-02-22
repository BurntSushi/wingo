import os
import os.path
import socket

sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
config_dir = os.getenv('XDG_RUNTIME_DIR')
f = os.path.join(config_dir, 'wingo', os.getenv('DISPLAY'))
try:
    sock.connect(f)
except socket.error:
    f += ".0"
    sock.connect(f)

def recv(sock):
    data = ''
    packet = ''
    while True:
        packet = sock.recv(4096)
        if chr(0) in packet:
            break
        data += packet
    if chr(0) in packet:
        data = packet[0:packet.index(chr(0))]
    return data


def gribble(cmd):
    sock.send("%s%s" % (cmd, chr(0)))
    return recv(sock)


print gribble("GetClientName (GetActive)")

sock.close()

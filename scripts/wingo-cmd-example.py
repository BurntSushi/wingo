import os
import os.path
import socket

sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
f = os.path.join(os.getenv('XDG_RUNTIME_DIR'), 'wingo', os.getenv('DISPLAY'))
sock.connect(f)


def recv(sock):
    data = ''
    while chr(0) not in data:
        data += sock.recv(4096)
    if chr(0) in data:
        data = data[0:data.index(chr(0))]
    return data


def gribble(cmd):
    sock.send("%s%s" % (cmd, chr(0)))
    return recv(sock)


print gribble("GetClientName (GetActive)")

sock.close()

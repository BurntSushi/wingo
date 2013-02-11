import os
import os.path
import socket
import sys

if len(sys.argv) != 2 or sys.argv[1] not in ('top', 'bot', 'left', 'right'):
    print >> sys.stderr, 'Usage: growto (top | bot | left | right)'
    sys.exit(1)

direction = sys.argv[1]
sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
disp = os.getenv('DISPLAY')
if len(disp) == 0:
    disp = ':0.0'
if ':' not in disp:
    disp = ':' + disp
if '.' not in disp:
    disp = disp + '.0'
f = os.path.join(os.getenv('XDG_RUNTIME_DIR'), 'wingo', disp)
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


win = int(gribble('GetActive'))
x = int(gribble('GetClientX %d' % win))
y = int(gribble('GetClientY %d' % win))
w = int(gribble('GetClientWidth %d' % win))
h = int(gribble('GetClientHeight %d' % win))

wrk = int(gribble('GetWorkspaceId (GetClientWorkspace %d)' % win))
headw = int(gribble('GetHeadWidth (WorkspaceHead %d)' % wrk))
headh = int(gribble('GetHeadHeight (WorkspaceHead %d)' % wrk))

if x == -9999 or y == -9999:
    print >> sys.stderr, 'Window %d is not visible.' % win
    sys.exit(1)

if direction == 'top':
    gribble('MoveRelative %d %d %d' % (win, x, 0))
    gribble('Resize %d %d %d' % (win, w, h + y))
elif direction == 'left':
    gribble('MoveRelative %d %d %d' % (win, 0, y))
    gribble('Resize %d %d %d' % (win, w + x, h))
elif direction == 'bot':
    gribble('Resize %d %d %d' % (win, w, headh - y))
elif direction == 'right':
    gribble('Resize %d %d %d' % (win, headw - x, h))

sock.close()

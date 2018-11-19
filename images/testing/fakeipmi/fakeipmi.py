#!/usr/bin/python

import sys
import SocketServer
from ipmisim.ipmisim import IpmiServer

if len(sys.argv) != 2:
    raise SystemExit("Port is not specified")

server = SocketServer.UDPServer(('0.0.0.0', int(sys.argv[1])), IpmiServer)
server.serve_forever()
